package nuxeo

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcileServiceAccount creates a service account for the Nuxeo deployments to run under. At present, there isn't
// anything in the service account spec - so this is just a placeholder in case any special service-related
// capabilities are needed in the future
func reconcileServiceAccount(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, reqLogger logr.Logger) (reconcile.Result, error) {
	found := &corev1.ServiceAccount{}
	svcAcctName := util.NuxeoServiceAccountName
	expected, err := r.defaultServiceAccount(instance, svcAcctName)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: svcAcctName, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ServiceAccount", "Namespace", expected.Namespace, "Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil {
			reqLogger.Error(err, "Failed to create new ServiceAccount", "Namespace", expected.Namespace, "Name", expected.Name)
			return reconcile.Result{}, err
		}
		// ServiceAccount created successfully
		return reconcile.Result{}, nil
	} else if err != nil {
		reqLogger.Error(err, "Error attempting to get ServiceAccount for Nuxeo cluster: "+svcAcctName)
		return reconcile.Result{}, err
	}
	// at present, there isn't anything in the service account so - nothing to do
	return reconcile.Result{}, nil
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
