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
	"fmt"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/aceeric/nuxeo-operator/controllers/util"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileIngress configures access to the Nuxeo cluster via a Kubernetes Ingress
func (r *NuxeoReconciler) reconcileIngress(access v1alpha1.NuxeoAccess, forcePassthrough bool, nodeSet v1alpha1.NodeSet,
	instance *v1alpha1.Nuxeo) error {
	ingressName := ingressName(instance, nodeSet)
	if access != (v1alpha1.NuxeoAccess{}) {
		if expected, err := r.defaultIngress(instance, access, forcePassthrough, ingressName, nodeSet); err != nil {
			return err
		} else {
			_, err = r.addOrUpdate(ingressName, instance.Namespace, expected, &v1beta1.Ingress{}, util.IngressComparer)
			return err
		}
	} else {
		return r.removeIfPresent(instance, ingressName, instance.Namespace, &v1beta1.Ingress{})
	}
}

// defaultIngress generates and returns an Ingress struct from the passed params. If the passed 'access' struct
// indicates TLS termination, or forcePassthrough==true, then an annotation is included in the returned object's
// metadata
func (r *NuxeoReconciler) defaultIngress(instance *v1alpha1.Nuxeo, access v1alpha1.NuxeoAccess, forcePassthrough bool,
	ingressName string, nodeSet v1alpha1.NodeSet) (*v1beta1.Ingress, error) {
	const nginxPassthroughAnnotation = "nginx.ingress.kubernetes.io/ssl-passthrough"
	targetPort := intstr.IntOrString{
		Type:   intstr.String,
		StrVal: "web",
	}
	if access.TargetPort != (intstr.IntOrString{}) {
		targetPort = access.TargetPort
	}
	ingress := v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: instance.Namespace,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{{
				Host: access.Hostname,
				IngressRuleValue: v1beta1.IngressRuleValue{
					HTTP: &v1beta1.HTTPIngressRuleValue{
						Paths: []v1beta1.HTTPIngressPath{{
							Backend: v1beta1.IngressBackend{
								ServiceName: serviceName(instance, nodeSet),
								ServicePort: targetPort,
							},
						}},
					},
				},
			}},
		},
	}
	if access.Termination != "" || forcePassthrough {
		if access.Termination != "" && access.Termination != routev1.TLSTerminationPassthrough &&
			access.Termination != routev1.TLSTerminationEdge {
			return nil, fmt.Errorf("only passthrough and edge termination are supported")
		}
		ingress.Spec.TLS = []v1beta1.IngressTLS{{
			Hosts: []string{access.Hostname},
		}}
		if access.Termination == routev1.TLSTerminationPassthrough || forcePassthrough {
			ingress.ObjectMeta.Annotations = map[string]string{nginxPassthroughAnnotation: "true"}
		} else {
			// the Ingress will terminate TLS
			if access.TLSSecret == "" {
				return nil, fmt.Errorf("the Ingress was configured for TLS termination but no secret was provided")
			}
			// secret needs keys 'tls.crt' and 'tls.key' and cert must have CN=<access.Hostname>
			ingress.Spec.TLS[0].SecretName = access.TLSSecret
		}
	}
	_ = controllerutil.SetControllerReference(instance, &ingress, r.Scheme)
	return &ingress, nil
}

// ingressName generates an Ingress name from the passed Nuxeo CR, and the passed NodeSet. The generated
// name consists of the passed Nuxeo CR name + dash + the passed 'nodeSet' name + dash + 'ingress'. E.g. if
// 'instance.Name' is 'my-nuxeo' and 'nodeSet.Name' is 'cluster' then the function returns 'my-nuxeo-cluster-ingress'.
func ingressName(instance *v1alpha1.Nuxeo, nodeSet v1alpha1.NodeSet) string {
	return instance.Name + "-" + nodeSet.Name + "-ingress"
}
