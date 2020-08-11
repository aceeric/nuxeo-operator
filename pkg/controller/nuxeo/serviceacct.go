package nuxeo

import (
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileServiceAccount creates a service account for the Nuxeo deployments to run under. At present, there isn't
// anything in the service account spec - so this is just a placeholder in case any special service-related
// capabilities are needed in the future
func reconcileServiceAccount(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, reqLogger logr.Logger) error {
	svcAcctName := util.NuxeoServiceAccountName
	expected, err := r.defaultServiceAccount(instance, svcAcctName)
	if err != nil {
		return err
	}
	_, err = addOrUpdate(r, svcAcctName, instance.Namespace, expected, &corev1.ServiceAccount{}, util.NopComparer, reqLogger)
	return err
}

// defaultServiceAccount creates and returns a service account struct
func (r *ReconcileNuxeo) defaultServiceAccount(nux *v1alpha1.Nuxeo, svcAcctName string) (*corev1.ServiceAccount, error) {
	sa := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcAcctName,
			Namespace: nux.Namespace,
		},
	}
	_ = controllerutil.SetControllerReference(nux, &sa, r.scheme)
	return &sa, nil
}
