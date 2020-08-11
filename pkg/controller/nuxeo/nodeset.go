package nuxeo

import (
	"context"
	goerrors "errors"
	"strings"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileNodeSet reconciles the passed NodeSet from the Nuxeo CR this operator is watching to the NodeSet's
// corresponding in-cluster Deployment. If no Deployment exists, a Deployment is created from the NodeSet. If a
// Deployment exists and its state differs from the NodeSet, the Deployment is conformed to the NodeSet.
// Otherwise, the fall-through case is that a Deployment exists that matches the NodeSet and so in this
// case - cluster state is not modified.
func reconcileNodeSet(r *ReconcileNuxeo, nodeSet v1alpha1.NodeSet, instance *v1alpha1.Nuxeo,
	revProxy v1alpha1.RevProxySpec, reqLogger logr.Logger) (bool, error) {
	var expected *appsv1.Deployment
	var err error
	var backingNuxeoConf, tlsNuxeoConf string
	depName := deploymentName(instance, nodeSet)
	if expected, err = r.defaultDeployment(instance, depName, nodeSet); err != nil {
		reqLogger.Error(err, "Error attempting to create default Deployment for NodeSet: "+nodeSet.Name)
		return false, err
	}
	if err = configureContributions(r, instance, expected, nodeSet); err != nil {
		reqLogger.Error(err, "Error attempting to configure contributions for NodeSet: "+nodeSet.Name)
		return false, err
	}
	if backingNuxeoConf, err = configureBackingServices(r, instance, expected, reqLogger); err != nil {
		return false, err
	}
	if nodeSet.Interactive {
		if revProxy.Nginx != (v1alpha1.NginxRevProxySpec{}) {
			// nginx will terminate TLS
			if err = configureNginx(expected, revProxy.Nginx); err != nil {
				return false, err
			}
		} else if nodeSet.NuxeoConfig.TlsSecret != "" {
			// nuxeo will terminate TLS
			if tlsNuxeoConf, err = configureNuxeoForTLS(expected, nodeSet.NuxeoConfig.TlsSecret); err != nil {
				return false, err
			}
		}
	}
	if err = configureNuxeoConf(instance, expected, nodeSet, backingNuxeoConf, tlsNuxeoConf); err != nil {
		return false, err
	}
	if err = reconcileNuxeoConf(r, instance, nodeSet, backingNuxeoConf, tlsNuxeoConf, reqLogger); err != nil {
		return false, err
	}
	if op, err := addOrUpdate(r, depName, instance.Namespace, expected, &appsv1.Deployment{},
		util.DeploymentComparer, reqLogger); err != nil {
		return false, err
	} else if op == Created {
		return true, nil
	}
	return false, nil
}

// defaultDeployment returns a nuxeo Deployment object with hard-coded default values, and the passed arg
// values. If the revProxy arg indicates that a reverse proxy is to be included in the deployment, then that results
// in another (TLS sidecar) container being added to the deployment. Note - many cluster defaults are explicitly
// specified here because it simplifies reconciliation
func (r *ReconcileNuxeo) defaultDeployment(nux *v1alpha1.Nuxeo, depName string, nodeSet v1alpha1.NodeSet) (*appsv1.Deployment, error) {
	nuxeoImage := "nuxeo:latest"
	if nux.Spec.NuxeoImage != "" {
		nuxeoImage = nux.Spec.NuxeoImage
	}
	var pullPolicy = corev1.PullIfNotPresent
	if nux.Spec.ImagePullPolicy == "" {
		if strings.HasSuffix(nuxeoImage, ":latest") {
			pullPolicy = corev1.PullAlways
		}
	} else {
		pullPolicy = nux.Spec.ImagePullPolicy
	}
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      depName,
			Namespace: nux.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsForNuxeo(nux, nodeSet.Interactive),
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
					Labels: labelsForNuxeo(nux, nodeSet.Interactive),
				},
				Spec: corev1.PodSpec{
					DeprecatedServiceAccount:      util.NuxeoServiceAccountName, // comes back from the cluster anyway...
					ServiceAccountName:            util.NuxeoServiceAccountName,
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
						VolumeMounts:             []corev1.VolumeMount{},
						Env:                      nodeSet.Env,
						Resources:                nodeSet.Resources,
					}},
					Volumes: []corev1.Volume{},
				},
			},
		},
	}
	// liveness/readiness
	useHttpsForProbes := false
	if nodeSet.NuxeoConfig.TlsSecret != "" {
		// if Nuxeo is going to terminate TLS, then it will be listening on HTTPS:8443. Otherwise Nuxeo
		// listens on HTTP:8080. This affects how the probes are configured immediately below.
		useHttpsForProbes = true
	}
	if err := addProbes(dep, nodeSet, useHttpsForProbes); err != nil {
		return nil, err
	}
	if err := handleStorage(dep, nodeSet); err != nil {
		return nil, err
	}
	jvmPkiSecret := corev1.Secret{}
	if nodeSet.NuxeoConfig.JvmPKISecret != "" {
		// if the config specifies a JVM PKI secret, get it here so lower layers in the call stack aren't doing
		// a lot of cluster I/O and can instead be focused on basic struct initialization
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: nodeSet.NuxeoConfig.JvmPKISecret,
			Namespace: nux.ObjectMeta.Namespace}, &jvmPkiSecret); err != nil {
			return nil, goerrors.New("Nuxeo configuration specifies JVM PKI secret that does not exist: " + nodeSet.NuxeoConfig.JvmPKISecret)
		}
	}
	if err := handleConfig(nux, dep, nodeSet, jvmPkiSecret); err != nil {
		return nil, err
	}
	if err := handleClid(nux, dep); err != nil {
		return nil, err
	}
	if err := configureClustering(dep, nodeSet); err != nil {
		return nil, err
	}
	_ = controllerutil.SetControllerReference(nux, dep, r.scheme)
	return dep, nil
}

// deploymentName generates a deployment name from the passed Nuxeo CR, and the passed NodeSet. The generated
// name consists of the passed Nuxeo CR name + dash + the passed 'nodeSet' name. E.g. if 'nux.Name' is 'my-nuxeo'
// and 'nodeSet.Name' is 'cluster' then the function returns 'my-nuxeo-cluster'.
func deploymentName(nux *v1alpha1.Nuxeo, nodeSet v1alpha1.NodeSet) string {
	return nux.Name + "-" + nodeSet.Name
}

// labelsForNuxeo returns a map of labels that are intended for the following specific purposes 1) a
// Deployment's match labels / pod template labels, and 2) a Service's selectors that enable the service to
// select a Nuxeo pod for TCP/IP traffic routing
func labelsForNuxeo(nux *v1alpha1.Nuxeo, interactive bool) map[string]string {
	m := map[string]string{
		"app":     "nuxeo",
		"nuxeoCr": nux.ObjectMeta.Name,
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
		return goerrors.New("configuration must define a Binaries storage in storageType")
	}
	if nuxeoContainer, err := util.GetNuxeoContainer(dep); err != nil {
		return err
	} else {
		if env := util.GetEnv(nuxeoContainer, "POD_UID"); env != nil {
			return goerrors.New("'POD_UID' environment variable already defined")
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
