package nuxeo

import (
	"context"
	"fmt"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Main reconciler. Isolated to a separate file to minimize mods to the OSDK-generated controller.
func (r *NuxeoReconciler) doReconcile(request reconcile.Request) (reconcile.Result, error) {
	emptyResult := reconcile.Result{}
	kv := []interface{}{"nuxeo", request.NamespacedName}
	r.Log.Info("reconciling Nuxeo", kv...)
	// Get the Nuxeo CR from the request namespace
	instance := &v1alpha1.Nuxeo{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("nuxeo resource not found. Ignoring since object must be deleted", kv...)
			return emptyResult, nil
		}
		return reconcile.Result{Requeue: true}, err
	}
	// only configure service/ingress/route for the interactive NodeSet
	var interactiveNodeSet v1alpha1.NodeSet
	if interactiveNodeSet, err = getInteractiveNodeSet(instance.Spec.NodeSets); err != nil {
		return emptyResult, err
	}
	if err = r.reconcileService(instance.Spec.Service, interactiveNodeSet, instance); err != nil {
		return emptyResult, err
	}
	if err = r.reconcileAccess(instance.Spec.Access, interactiveNodeSet, instance); err != nil {
		return emptyResult, err
	}
	if err = r.reconcileServiceAccount(instance); err != nil {
		return emptyResult, err
	}
	if err = r.reconcilePvc(instance); err != nil {
		return emptyResult, err
	}
	if err = r.reconcileClid(instance); err != nil {
		return emptyResult, err
	}
	if requeue, err := r.reconcileNodeSets(instance); err != nil {
		return emptyResult, err
	} else if requeue {
		return reconcile.Result{Requeue: true}, nil
	}
	if err := r.updateNuxeoStatus(instance); err != nil {
		return emptyResult, err
	}
	r.Log.Info("finished", kv...)
	return emptyResult, nil
}

// Reconciles each NodeSet to a Deployment. Return true to requeue, else false=don't requeue
func (r *NuxeoReconciler) reconcileNodeSets(instance *v1alpha1.Nuxeo) (bool, error) {
	for _, nodeSet := range instance.Spec.NodeSets {
		if requeue, err := r.reconcileNodeSet(nodeSet, instance); err != nil {
			return requeue, err
		} else if requeue {
			return requeue, nil
		}
	}
	return false, nil
}

// Returns the interactive NodeSet from the passed array, or non-nil error if a) there is no interactive NodeSet
// defined, or b) there is more than one interactive NodeSet defined
func getInteractiveNodeSet(nodeSets []v1alpha1.NodeSet) (v1alpha1.NodeSet, error) {
	toReturn := v1alpha1.NodeSet{}
	for _, nodeSet := range nodeSets {
		if nodeSet.Interactive {
			if toReturn.Name != "" {
				return toReturn, fmt.Errorf("exactly one interactive NodeSet is required in the Nuxeo CR")
			}
			toReturn = nodeSet
		}
	}
	return toReturn, nil
}
