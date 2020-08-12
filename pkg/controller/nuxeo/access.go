package nuxeo

import (
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
)

// reconcileAccess configures external access to the Nuxeo cluster either through an OpenShift Route object
// or a Kubernetes Ingress object. This function simply delegates to 'reconcileOpenShiftRoute' or
// 'reconcileIngress'
func reconcileAccess(r *ReconcileNuxeo, access v1alpha1.NuxeoAccess, nodeSet v1alpha1.NodeSet,
	instance *v1alpha1.Nuxeo) error {
	forcePassthrough := false
	if nodeSet.NuxeoConfig.TlsSecret != "" {
		// if Nuxeo is terminating TLS then force tls passthrough termination in the route/ingress
		forcePassthrough = true
	}
	if util.IsOpenShift() {
		return reconcileOpenShiftRoute(r, access, forcePassthrough, nodeSet, instance)
	} else {
		return reconcileIngress(r, access, forcePassthrough, nodeSet, instance)
	}
}
