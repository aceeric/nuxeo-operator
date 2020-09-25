/*
Copyright 2020 Eric Ace.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nuxeo

import (
	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/aceeric/nuxeo-operator/controllers/util"
)

// reconcileAccess configures external access to the Nuxeo cluster either through an OpenShift Route object
// or a Kubernetes Ingress object. This function simply delegates to 'reconcileOpenShiftRoute' or
// 'reconcileIngress'
func (r *NuxeoReconciler) reconcileAccess(access v1alpha1.NuxeoAccess, nodeSet v1alpha1.NodeSet,
	instance *v1alpha1.Nuxeo) error {
	forcePassthrough := false
	if nodeSet.NuxeoConfig.TlsSecret != "" {
		// if Nuxeo is terminating TLS then force tls passthrough termination in the route/ingress
		forcePassthrough = true
	}
	if util.IsOpenShift() {
		return r.reconcileOpenShiftRoute(access, forcePassthrough, nodeSet, instance)
	} else {
		return r.reconcileIngress(access, forcePassthrough, nodeSet, instance)
	}
}
