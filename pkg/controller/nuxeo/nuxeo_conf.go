package nuxeo

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcileNuxeoConf inspects the Nuxeo CR to see if it contains an inline nuxeo.conf. If it does, this the function
// creates a ConfigMap struct to hold the contents, identified by the key 'nuxeo.conf' and reconciles this with the
// cluster. The caller must have defined a Volume and VolumeMount elsewhere to reference the ConfigMap. If the Nuxeo
// CR indicates that an inline nuxeo.conf should not exist, then the function makes sure a ConfigMap does not exist
// in the cluster. The ConfigMap is given a hard-coded name: nuxeo cluster name + "-" + node set name + "-nuxeo-conf".
// E.g 'my-nuxeo-cluster-nuxeo-conf'.
func reconcileNuxeoConf(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, nodeSet v1alpha1.NodeSet,
	reqLogger logr.Logger) (reconcile.Result, error) {
	if nodeSet.NuxeoConfig.NuxeoConf.Value != "" {
		return addOrUpdate(r, instance, nodeSet, reqLogger)
	} else {
		return removeIfPresent(r, instance, nodeSet, reqLogger)
	}
}

// removeIfPresent should be called if the Nuxeo CR indicates that no nuxeo.conf ConfigMap should be present
// in the cluster. The function looks for a nuxeo.conf ConfigMap having the internally defined name: nuxeo cluster
// name + "-" + node set name + "-nuxeo-conf". If found, and owned by this Nuxeo, then it is removed. Otherwise
// cluster state is not modified. Note that if the Nuxeo CR contains a nuxeo.conf configuration entry with
// ValueFrom defined, and that name matches the internally defined name and it is owned by the Nuxeo CR then
// it will be removed. So - configurers should not name their nuxeo.conf ConfigMap the way this code does.
func removeIfPresent(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, nodeSet v1alpha1.NodeSet,
	reqLogger logr.Logger) (reconcile.Result, error) {
	found := &corev1.ConfigMap{}
	cmName := instance.Name + "-" + nodeSet.Name + "-nuxeo-conf"
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

// addOrUpdate reconciles an expected nuxeo.conf ConfigMap into the cluster. Expectation is the caller only calls
// this if the passed NodeSet contains a NuxeoConf inline config.
func addOrUpdate(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, nodeSet v1alpha1.NodeSet,
	reqLogger logr.Logger) (reconcile.Result, error) {
	expected := r.defaultNuxeoConfCM(instance, nodeSet.Name, nodeSet.NuxeoConfig.NuxeoConf.Value)
	found := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: expected.Name, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Nuxeo Conf ConfigMap", "Namespace", instance.Namespace, "Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil {
			reqLogger.Error(err, "Failed to create Nuxeo Conf ConfigMap", "Namespace", expected.Namespace, "Name", expected.Name)
			return reconcile.Result{}, err
		}
		// CM created successfully
		return reconcile.Result{}, nil
	} else if err != nil {
		reqLogger.Error(err, "Error attempting to get nuxeo conf ConfigMap: "+expected.Name)
		return reconcile.Result{}, err
	}
	if !reflect.DeepEqual(expected.Data, found.Data) {
		reqLogger.Info("Updating Nuxeo Conf ConfigMap", "Namespace", expected.Namespace, "Name", expected.Name)
		expected.Data = found.Data
		if err = r.client.Update(context.TODO(), found); err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

// defaultNuxeoConfCM generates a ConfigMap struct in a standard internally defined form to hold the passed
// inline nuxeo conf string data. A ref to the generated struct is returned.
func (r *ReconcileNuxeo) defaultNuxeoConfCM(nux *v1alpha1.Nuxeo, nodeSetName string, nuxeoConf string) *corev1.ConfigMap {
	cmName := nux.Name + "-" + nodeSetName + "-nuxeo-conf"
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: nux.Namespace,
		},
		Data: map[string]string{"nuxeo.conf": nuxeoConf},
	}
	_ = controllerutil.SetControllerReference(nux, cm, r.scheme)
	return cm
}
