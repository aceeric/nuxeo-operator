package nuxeo

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ConfigMap utils

// addOrUpdateConfigMap reconciles an expected ConfigMap into the cluster
func addOrUpdateConfigMap(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, expected *corev1.ConfigMap,
	reqLogger logr.Logger) (reconcile.Result, error) {
	found := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: expected.Name, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ConfigMap", "Namespace", instance.Namespace, "Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil {
			reqLogger.Error(err, "Failed to create ConfigMap", "Namespace", expected.Namespace, "Name", expected.Name)
			return reconcile.Result{}, err
		}
		// CM created successfully
		return reconcile.Result{}, nil
	} else if err != nil {
		reqLogger.Error(err, "Error attempting to get ConfigMap: "+expected.Name)
		return reconcile.Result{}, err
	}
	if !reflect.DeepEqual(expected.Data, found.Data) {
		reqLogger.Info("Updating ConfigMap", "Namespace", expected.Namespace, "Name", expected.Name)
		found.Data = expected.Data
		if err = r.client.Update(context.TODO(), found); err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

// removeConfigMapIfPresent looks for a ConfigMap in the cluster matching the passed name. If found, and owned
// by this Nuxeo, then it is removed. Otherwise cluster state is not modified. The use case is: the configurer
// creates a CR that causes the Operator to create a ConfigMap. The operator creates a ConfigMap. The configurer
// then edits the CR such that the ConfigMap is no longer needed. Then the operator removes the ConfigMap.
func removeConfigMapIfPresent(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, cmName string,
	reqLogger logr.Logger) (reconcile.Result, error) {
	found := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: cmName, Namespace: instance.Namespace}, found)
	if err == nil {
		if instance.IsOwner(found.ObjectMeta) {
			if err := r.client.Delete(context.TODO(), found); err != nil {
				reqLogger.Error(err, "Error attempting to delete nuxeo conf ConfigMap: "+cmName)
				return reconcile.Result{}, err
			}
		}
	} else if !errors.IsNotFound(err) {
		reqLogger.Error(err, "Error attempting to get nuxeo conf ConfigMap: "+cmName)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
