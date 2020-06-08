package nuxeo

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcileNodeSet reconciles the passed NodeSet from the Nuxeo CR this operator is watching to the NodeSet's
// corresponding in-cluster Deployment. If no Deployment exists, a Deployment is created from the NodeSet. If a
// Deployment exists and its state differs from the NodeSet, the Deployment is conformed to the NodeSet.
// Otherwise, the fall-through case is that a Deployment exists that matches the NodeSet and so in this
// case - cluster state is not modified.
func reconcileNodeSet(r *ReconcileNuxeo, nodeSet v1alpha1.NodeSet, instance *v1alpha1.Nuxeo, revProxy v1alpha1.RevProxySpec, reqLogger logr.Logger) (reconcile.Result, error) {
	actual := &v1.Deployment{}
	depName := deploymentName(instance, nodeSet)
	expected := r.defaultDeployment(instance, depName, nodeSet, revProxy)
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: depName, Namespace: instance.Namespace}, actual)
	if err != nil && errors.IsNotFound(err) {
		// Add any custom labels from NodeSet > PodTemplate > Metadata > Labels into Deployment > Spec > template >
		// metadata > labels. Don't overwrite default labels assigned by this operator - only add new labels. This
		// is in case any functionality of the operator relies on operator-generated labels being present in the
		// pod template.
		for label, value := range nodeSet.PodTemplate.Labels {
			if _, ok := expected.Spec.Template.Labels[label]; !ok {
				expected.Spec.Template.Labels[label] = value
			} else {
				reqLogger.Info("NodeSet PodTemplate label clashes with built-in label and will be ignored",
					"NodeSet", nodeSet.Name, "label", label)
			}
		}
		reqLogger.Info("Creating a new Deployment", "Namespace", expected.Namespace,
			"Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Namespace",
				expected.Namespace, "Name", expected.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Error attempting to get Deployment for NodeSet: "+nodeSet.Name)
		return reconcile.Result{}, err
	}
	if !equality.Semantic.DeepDerivative(expected.Spec, actual.Spec) {
		reqLogger.Info("Updating Deployment", "Namespace", expected.Namespace, "Name", expected.Name)
		expected.Spec.DeepCopyInto(&actual.Spec)
		if err = r.client.Update(context.TODO(), actual); err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

// defaultDeployment returns a nuxeo Deployment object with hard-coded default values, and the passed arg
// values. Returns a Deployment struct like so:
//
//   apiVersion: apps/v1
//   kind: Deployment
//   metadata:
//     name: <passed 'depName' arg>
//   spec:
//     replicas: <passed 'replicas' arg>
//     selector:
//       matchLabels:
//         app: nuxeo
//         nuxeoCr: <from your Nuxeo CR metadata.name field>
//     template:
//       metadata:
//         labels:
//           app: nuxeo
//           nuxeoCr: <from your Nuxeo CR metadata.name field>
//       spec:
//         serviceAccountName: nuxeo
//         containers:
//         - name: nuxeo
//           imagePullPolicy: Always
//           image: 'nuxeo:latest'
//           ports:
//           - containerPort: 8080
//
// If the revProxy arg indicates that a reverse proxy is to be included in the deployment, then that results in
// another (TLS sidecar) container being added to the deployment
func (r *ReconcileNuxeo) defaultDeployment(nux *v1alpha1.Nuxeo, depName string, nodeSet v1alpha1.NodeSet, revProxy v1alpha1.RevProxySpec) *v1.Deployment {
	replicas32 := int32(nodeSet.Replicas)
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
	dep := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      depName,
			Namespace: nux.Namespace,
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsForNuxeo(nux, nodeSet.Interactive),
			},
			Replicas: &replicas32,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelsForNuxeo(nux, nodeSet.Interactive),
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: util.NuxeoServiceAccountName,
					Containers: []corev1.Container{{
						Image:           nuxeoImage,
						ImagePullPolicy: pullPolicy,
						Name:            "nuxeo",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
						}},
						VolumeMounts: []corev1.VolumeMount{},
						Env: nodeSet.Env,
					}},
					Volumes: []corev1.Volume{},
				},
			},
		},
	}
	if util.HashObject(nodeSet.PodTemplate.Spec) != util.HashObject(corev1.PodSpec{}) {
		// if the passed NodeSet specifies a pod template Spec, then use that in the Deployment, rather than
		// the hard-coded default pod template spec generated by the code above
		nodeSet.PodTemplate.Spec.DeepCopyInto(&dep.Spec.Template.Spec)
	}
	if revProxy.Nginx != (v1alpha1.NginxRevProxySpec{}) {
		// if a Nginx is specified as the reverse proxy, then configure an additional Container and supporting
		// Volumes in the Deployment
		dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers, nginxContainer(revProxy.Nginx))
		dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes, nginxVolumes(revProxy.Nginx)...)
	}
	// Set Nuxeo as the owner and controller
	_ = controllerutil.SetControllerReference(nux, dep, r.scheme)
	return dep
}

// nginxContainer creates and returns a Container struct defining the Nginx reverse proxy. It defines various
// volume mounts which - therefore - must also be defined in the deployment that ultimately holds this container
// struct.
func nginxContainer(nginx v1alpha1.NginxRevProxySpec) corev1.Container {
	nginxImage := "nginx:latest"
	if nginx.Image != "" {
		nginxImage = nginx.Image
	}
	var pullPolicy = corev1.PullIfNotPresent
	if nginx.ImagePullPolicy == "" {
		if strings.HasSuffix(nginxImage, ":latest") {
			pullPolicy = corev1.PullAlways
		}
	} else {
		pullPolicy = nginx.ImagePullPolicy
	}
	c := corev1.Container{
		Name:  "nginx",
		Image: nginxImage,
		ImagePullPolicy: pullPolicy,
		Ports: []corev1.ContainerPort{{
			Name:          "nginx-port",
			ContainerPort: 8443,
			Protocol:      "TCP",
		}},
		VolumeMounts: []corev1.VolumeMount{{
			Name:      "nginx-cert",
			ReadOnly:  true,
			MountPath: "/etc/secrets/",
		}, {
			Name:      "nginx-conf",
			ReadOnly:  true,
			MountPath: "/etc/nginx/",
		}, {
			Name:      "nginx-cache",
			ReadOnly:  false,
			MountPath: "/var/cache/nginx",
		}, {
			Name:      "nginx-tmp",
			ReadOnly:  false,
			MountPath: "/var/tmp",
		}},
	}
	return c
}

// nginxVolumes creates and returns a slice of Volume specs that support the VolumeMounts generated by the
// 'nginxContainer' function. Expectation is that these items will be added by the caller into a Deployment
func nginxVolumes(nginx v1alpha1.NginxRevProxySpec) []corev1.Volume {
	vols := []corev1.Volume{{
		Name: "nginx-conf",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: nginx.ConfigMap},
			},
		},
	}, {
		Name: "nginx-cert",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: nginx.Secret,
			},
		},
	}, {
		Name: "nginx-cache",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}, {
		Name: "nginx-tmp",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}}
	return vols
}

// deploymentName generates a deployment name from the passed Nuxeo CR, and the passed NodeSet. The generated
// name consists of the passed Nuxeo CR name + dash + the passed 'nodeSet' name. E.g. if 'nux.Name' is 'my-nuxeo'
// and 'nodeSet.Name' is 'cluster' then the function returns 'my-nuxeo-cluster'.
func deploymentName(nux *v1alpha1.Nuxeo, nodeSet v1alpha1.NodeSet) string {
	return nux.Name + "-" + nodeSet.Name
}

// labelsForNuxeo returns a map of labels that are intended for the following specific purposes 1) a
// Deployment's match labels / pod template labels, and 2) a Service's selectors
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
