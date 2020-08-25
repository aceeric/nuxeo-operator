package nuxeo

import (
	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/aceeric/nuxeo-operator/controllers/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	nuxeoClidConfigMapName = "nuxeo-clid"
	clidKey                = "instance.clid"
)

// configureClid configures the passed Deployment with a Volume and VolumeMount to project the CLID into the
// Nuxeo container at a hard-coded mount point: /var/lib/nuxeo/data/instance.clid. The volume references
// a hard-coded ConfigMap name managed by the operator: "nuxeo-clid". See the reconcileClid() function
// for the code that reconciles that actual ConfigMap.
func configureClid(instance *v1alpha1.Nuxeo, dep *appsv1.Deployment) error {
	if instance.Spec.Clid == "" {
		return nil
	}
	if nuxeoContainer, err := GetNuxeoContainer(dep); err != nil {
		return err
	} else {
		volMnt := corev1.VolumeMount{
			Name:      nuxeoClidConfigMapName,
			ReadOnly:  true,
			MountPath: "/var/lib/nuxeo/data/" + clidKey,
			SubPath:   clidKey,
		}
		if err := util.OnlyAddVolMnt(nuxeoContainer, volMnt); err != nil {
			return err
		}
		vol := corev1.Volume{
			Name: nuxeoClidConfigMapName,
		}
		vol.ConfigMap = &corev1.ConfigMapVolumeSource{
			DefaultMode:          util.Int32Ptr(420),
			LocalObjectReference: corev1.LocalObjectReference{Name: nuxeoClidConfigMapName},
			Items: []corev1.KeyToPath{{
				Key:  clidKey,
				Path: clidKey,
			}},
		}
		return util.OnlyAddVol(dep, vol)
	}
}

// reconcileClid creates, updates, or deletes the CLID ConfigMap. If the Clid is specified in the CR, then the
// corresponding CM is added/updated in the cluster. If Clid is not specified, then it is removed from the
// cluster if present
func (r *NuxeoReconciler) reconcileClid(instance *v1alpha1.Nuxeo) error {
	if instance.Spec.Clid != "" {
		expected := r.defaultClidCM(instance, instance.Spec.Clid)
		_, err := r.addOrUpdate(nuxeoClidConfigMapName, instance.Namespace, expected, &corev1.ConfigMap{},
			util.ConfigMapComparer)
		return err
	} else {
		return r.removeIfPresent(instance, nuxeoClidConfigMapName, instance.Namespace, &corev1.ConfigMap{})
	}
}

// defaultClidCM creates and returns a ConfigMap struct named "nuxeo-clid" to hold the passed CLID string
func (r *NuxeoReconciler) defaultClidCM(instance *v1alpha1.Nuxeo, clidValue string) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nuxeoClidConfigMapName,
			Namespace: instance.Namespace,
		},
		Data: map[string]string{clidKey: clidValue},
	}
	_ = controllerutil.SetControllerReference(instance, cm, r.Scheme)
	return cm
}
