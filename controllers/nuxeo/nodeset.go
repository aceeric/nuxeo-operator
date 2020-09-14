package nuxeo

import (
	"context"
	"fmt"
	"strings"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/aceeric/nuxeo-operator/controllers/common"
	"github.com/aceeric/nuxeo-operator/controllers/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileNodeSet reconciles the passed NodeSet from the Nuxeo CR this operator is watching to the NodeSet's
// corresponding in-cluster Deployment. If no Deployment exists, a Deployment is created from the NodeSet. If a
// Deployment exists and its state differs from the NodeSet, the Deployment is conformed to the NodeSet.
// Otherwise, the fall-through case is that a Deployment exists that matches the NodeSet and so in this
// case - cluster state is not modified.
//
// Returns:
//   requeue true to requeue, else false (true means success but requeue to update status)
//   error or nil
func (r *NuxeoReconciler) reconcileNodeSet(nodeSet v1alpha1.NodeSet, instance *v1alpha1.Nuxeo) (bool, error) {
	var expected *appsv1.Deployment
	var err error
	depName := deploymentName(instance, nodeSet)
	if expected, err = r.defaultDeployment(instance, depName, nodeSet); err != nil {
		return false, err
	}
	if err := r.configureDeploymentFromNuxeo(nodeSet, expected, instance); err != nil {
		return false, err
	}
	if op, err := r.addOrUpdate(depName, instance.Namespace, expected, &appsv1.Deployment{},
		util.DeploymentComparer); err != nil {
		return false, err
	} else if op == Created {
		return true, nil
	}
	return false, nil
}

// configureDeploymentFromNuxeo applies the various Nuxeo CR configurations to the passed 'expected' deployment
// and reconciles any dependent resources that the operator generates to support those deployment configurations.
// For example, if the operator generates a ConfigMap for nuxeo.conf, then that ConfigMap will be in the cluster
// by the time the function exits so the caller can create the deployment and it should not error due to missing
// dependencies.
func (r *NuxeoReconciler) configureDeploymentFromNuxeo(nodeSet v1alpha1.NodeSet, expected *appsv1.Deployment,
	instance *v1alpha1.Nuxeo) error {
	var backingNuxeoConf, tlsNuxeoConf string

	if err := configureProbes(expected, nodeSet); err != nil {
		return err
	}
	if err := configureStorage(expected, nodeSet); err != nil {
		return err
	}
	jvmPkiSecret := corev1.Secret{}
	if nodeSet.NuxeoConfig.JvmPKISecret != "" {
		if err := r.Get(context.TODO(), types.NamespacedName{Name: nodeSet.NuxeoConfig.JvmPKISecret,
			Namespace: instance.ObjectMeta.Namespace}, &jvmPkiSecret); err != nil {
			return fmt.Errorf("configuration specifies JVM PKI secret that does not exist: %v",
				nodeSet.NuxeoConfig.JvmPKISecret)
		}
	}
	if err := configureContainers(instance, expected); err != nil {
		return err
	}
	if err := configureConfig(expected, nodeSet, jvmPkiSecret); err != nil {
		return err
	}
	if err := configureClid(instance, expected); err != nil {
		return err
	}
	if err := configureClustering(expected, nodeSet); err != nil {
		return err
	}
	if err := r.configureContributions(instance, expected, nodeSet); err != nil {
		return err
	}
	if tmp, err := r.configureBackingServices(instance, expected); err != nil {
		return err
	} else {
		backingNuxeoConf = tmp
	}
	if nodeSet.Interactive {
		revProxy := instance.Spec.RevProxy
		if revProxy.Nginx != (v1alpha1.NginxRevProxySpec{}) {
			// nginx will terminate TLS
			nginxCmName, err := r.reconcileNginxCM(instance, revProxy.Nginx.ConfigMap)
			if err != nil {
				return err
			}
			revProxy.Nginx.ConfigMap = nginxCmName
			if err := configureNginx(expected, revProxy.Nginx); err != nil {
				return err
			}
		} else if nodeSet.NuxeoConfig.TlsSecret != "" {
			// Nuxeo will terminate TLS
			if tmp, err := configureNuxeoForTLS(expected, nodeSet.NuxeoConfig.TlsSecret); err != nil {
				return err
			} else {
				tlsNuxeoConf = tmp
			}
		}
	}
	for _, vol := range instance.Spec.Volumes {
		// explicit volume config - configurer must ensure they exist in the cluster
		if err := util.OnlyAddVol(expected, vol); err != nil {
			return err
		}
	}
	if err := configureNuxeoConf(instance, expected, nodeSet, backingNuxeoConf, tlsNuxeoConf); err != nil {
		return err
	}
	if nxconfHash, err := r.reconcileNuxeoConf(instance, nodeSet, backingNuxeoConf, tlsNuxeoConf); err != nil {
		return err
	} else if nxconfHash != "" {
		util.AnnotateTemplate(expected, common.NuxeoConfHashAnnotation, nxconfHash)
	}
	return nil
}

// defaultDeployment returns a nuxeo Deployment object with hard-coded default values, and the passed arg
// values. Note - many cluster defaults are explicitly specified here because it simplifies reconciliation. The
// deployment generated is just a basic shell, and the caller will apply all the configurations to it from the
// Nuxeo CR. The deployment always contains one container named "nuxeo".
func (r *NuxeoReconciler) defaultDeployment(instance *v1alpha1.Nuxeo, depName string,
	nodeSet v1alpha1.NodeSet) (*appsv1.Deployment, error) {
	nuxeoImage := "nuxeo:latest"
	if instance.Spec.NuxeoImage != "" {
		nuxeoImage = instance.Spec.NuxeoImage
	}
	var pullPolicy = corev1.PullIfNotPresent
	if instance.Spec.ImagePullPolicy == "" {
		if strings.HasSuffix(nuxeoImage, ":latest") {
			pullPolicy = corev1.PullAlways
		}
	} else {
		pullPolicy = instance.Spec.ImagePullPolicy
	}
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      depName,
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsForNuxeo(instance, nodeSet.Interactive),
			},
			Replicas:                util.Int32Ptr(nodeSet.Replicas),
			ProgressDeadlineSeconds: util.Int32Ptr(600),
			RevisionHistoryLimit:    util.Int32Ptr(10),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%",
					},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%",
					},
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelsForNuxeo(instance, nodeSet.Interactive),
				},
				Spec: corev1.PodSpec{
					// comes back from the cluster anyway:
					DeprecatedServiceAccount:      NuxeoServiceAccountName,
					ServiceAccountName:            NuxeoServiceAccountName,
					TerminationGracePeriodSeconds: util.Int64Ptr(30),
					DNSPolicy:                     corev1.DNSClusterFirst,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					SchedulerName:                 corev1.DefaultSchedulerName,
					SecurityContext:               &corev1.PodSecurityContext{},
					Containers: []corev1.Container{{
						Image:           nuxeoImage,
						ImagePullPolicy: pullPolicy,
						Name:            "nuxeo",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Protocol:      corev1.ProtocolTCP,
						}},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: corev1.TerminationMessageReadFile,
						VolumeMounts:             nil,
						Env:                      nodeSet.Env,
						Resources:                nodeSet.Resources,
					}},
					Volumes: nil,
				},
			},
		},
	}
	_ = controllerutil.SetControllerReference(instance, dep, r.Scheme)
	return dep, nil
}

// deploymentName generates a deployment name from the passed Nuxeo CR, and the passed NodeSet. The generated
// name consists of the passed Nuxeo CR name + dash + the passed 'nodeSet' name. E.g. if 'instance.Name' is 'my-nuxeo'
// and 'nodeSet.Name' is 'cluster' then the function returns 'my-nuxeo-cluster'.
func deploymentName(instance *v1alpha1.Nuxeo, nodeSet v1alpha1.NodeSet) string {
	return instance.Name + "-" + nodeSet.Name
}

// labelsForNuxeo returns a map of labels that are intended for the following specific purposes 1) a
// Deployment's match labels / pod template labels, and 2) a Service's selectors that enable the service to
// select a Nuxeo pod for TCP/IP traffic routing
func labelsForNuxeo(instance *v1alpha1.Nuxeo, interactive bool) map[string]string {
	m := map[string]string{
		"app":     "nuxeo",
		"nuxeoCr": instance.ObjectMeta.Name,
	}
	if interactive {
		m["interactive"] = "true"
	}
	return m
}

// configureClustering adds an environment variable POD_UID to the Nuxeo container in the passed deployment. The
// env var is defined using the downward API to get the UID of the Pod. This environment variable is referenced by
// the defaultNuxeoConfCM() function to build a ConfigMap of nuxeo.conf properties to project into the Nuxeo Pod.
// The CM built by that function has a variable 'nuxeo.cluster.nodeid=${env:POD_UID}' referencing the env var.
// If the POD_UID environment variable is already present in the Nuxeo container then the function returns an error.
//
// The function also verifies that the passed NodeSet defines a binary storage type. The is necessary because in
// clustered mode, the binary store must be shared by all nodes in the cluster. The binary storage type must be
// supplied by the configurer since it references cluster storage and will therefore be site-specific. If a binary
// storage is not defined, then an error is returned.
func configureClustering(dep *appsv1.Deployment, nodeSet v1alpha1.NodeSet) error {
	if !nodeSet.ClusterEnabled {
		return nil
	}
	if !binaryStorageIsDefined(nodeSet) {
		return fmt.Errorf("configuration must define a Binaries storage in storageType")
	}
	if nuxeoContainer, err := GetNuxeoContainer(dep); err != nil {
		return err
	} else {
		return util.OnlyAddEnvVar(nuxeoContainer, corev1.EnvVar{
			Name: "POD_UID",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.uid",
				},
			},
		})
	}
}

// binaryStorageIsDefined returns true if the passed nodeset defines a 'Binaries' storage type, otherwise
// returns false
func binaryStorageIsDefined(nodeSet v1alpha1.NodeSet) bool {
	for _, storage := range nodeSet.Storage {
		if storage.StorageType == v1alpha1.NuxeoStorageBinaries {
			return true
		}
	}
	return false
}

// configureContainers copies the InitContainers and Containers arrays from the passed Nuxeo struct into
// the passed deployment struct. The Containers array is first checked to ensure that it does not define a
// container named "nuxeo", since this container name is reserved by the Operator.
func configureContainers(instance *v1alpha1.Nuxeo, expected *appsv1.Deployment) error {
	for idx, container := range instance.Spec.Containers {
		if container.Name == "nuxeo" {
			return fmt.Errorf("container at ordinal position %v uses reserved 'nuxeo' name", idx)
		}
	}
	expected.Spec.Template.Spec.InitContainers = instance.Spec.InitContainers
	expected.Spec.Template.Spec.Containers = append(expected.Spec.Template.Spec.Containers, instance.Spec.Containers...)
	return nil
}
