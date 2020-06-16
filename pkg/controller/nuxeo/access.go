package nuxeo

import (
	"github.com/go-logr/logr"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcileAccess configures external access to the Nuxeo cluster either through an OpenShift Route object
// or a Kubernetes Ingress object
func reconcileAccess(r *ReconcileNuxeo, access v1alpha1.NuxeoAccess, nodeSet v1alpha1.NodeSet,
	instance *v1alpha1.Nuxeo, reqLogger logr.Logger) (reconcile.Result, error) {
	if util.IsOpenShift() {
		return reconcileOpenShiftRoute(r, access, nodeSet, instance, reqLogger)
	} else {
		return reconcileIngress(r, access, nodeSet, instance, reqLogger)
	}
}
