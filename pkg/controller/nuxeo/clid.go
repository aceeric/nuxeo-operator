package nuxeo

import (
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	nuxeoClidConfigMapName = "nuxeo-clid"
	clidKey = "instance.clid"
)

// handleClid configures the passed Deployment with a Volume and VolumeMount to project the CLID into the
// Nuxeo container at a hard-coded mount point: /var/lib/nuxeo/data/instance.clid. The volume references
// a hard-coded ConfigMap name managed by the operator: "nuxeo-clid". See the reconcileClid() function
// for the code that reconciles that actual ConfigMap.
func handleClid(nux *v1alpha1.Nuxeo, dep *appsv1.Deployment) error {
	if nux.Spec.Clid == "" {
		return nil
	}
	if nuxeoContainer, err := util.GetNuxeoContainer(dep); err != nil {
		return err
	} else {
		volMnt := corev1.VolumeMount{
			Name:      nuxeoClidConfigMapName,
			ReadOnly:  true,
			MountPath: "/var/lib/nuxeo/data/"+clidKey,
			SubPath:   clidKey,
		}
		nuxeoContainer.VolumeMounts = append(nuxeoContainer.VolumeMounts, volMnt)
		vol := corev1.Volume{
			Name: nuxeoClidConfigMapName,
		}
		vol.ConfigMap = &corev1.ConfigMapVolumeSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: nuxeoClidConfigMapName},
			Items: []corev1.KeyToPath{{
				Key:  clidKey,
				Path: clidKey,
			}},
		}
		dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes, vol)
	}
	return nil
}

// reconcileClid creates, updates, or deletes the CLID ConfigMap. If the Clid is specified in the CR, then the
// corresponding CM is added/updated in the cluster. If Clid is not specified, then it is removed from the
// cluster if present
func reconcileClid(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, reqLogger logr.Logger) (reconcile.Result, error) {
	if instance.Spec.Clid != "" {
		expected := r.defaultClidCM(instance, instance.Spec.Clid)
		return addOrUpdateConfigMap(r, instance, expected, reqLogger)
	} else {
		return removeConfigMapIfPresent(r, instance, nuxeoClidConfigMapName, reqLogger)
	}
}

// defaultClidCM creates and returns a ConfigMap struct named "nuxeo-clid" to hold the passed CLID string
func (r *ReconcileNuxeo) defaultClidCM(nux *v1alpha1.Nuxeo, clidValue string) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nuxeoClidConfigMapName,
			Namespace: nux.Namespace,
		},
		Data: map[string]string{clidKey: clidValue},
	}
	_ = controllerutil.SetControllerReference(nux, cm, r.scheme)
	return cm
}
