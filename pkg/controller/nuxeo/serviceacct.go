package nuxeo

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileServiceAccount creates a service account for the Nuxeo deployments to run under. At present, there isn't
// anything in the service account spec - so this is just a placeholder in case any special service-related
// capabilities are needed in the future
func (r *ReconcileNuxeo) reconcileServiceAccount(instance *v1alpha1.Nuxeo) error {
	svcAcctName := util.NuxeoServiceAccountName
	expected, err := r.defaultServiceAccount(instance, svcAcctName)
	if err != nil {
		return err
	}
	_, err = r.addOrUpdate(svcAcctName, instance.Namespace, expected, &corev1.ServiceAccount{}, util.NopComparer)
	return err
}

// defaultServiceAccount creates and returns a service account struct
func (r *ReconcileNuxeo) defaultServiceAccount(instance *v1alpha1.Nuxeo,
	svcAcctName string) (*corev1.ServiceAccount, error) {
	sa := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcAcctName,
			Namespace: instance.Namespace,
		},
	}
	_ = controllerutil.SetControllerReference(instance, &sa, r.scheme)
	return &sa, nil
}
