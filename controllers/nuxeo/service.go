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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileService reconciles the passed ServiceSpec from the Nuxeo CR this operator is watching to the ServiceSpec's
// corresponding in-cluster Service. If no Service exists, a Service is created from the ServiceSpec. If a
// Service exists and its state differs from the ServiceSpec, the Service is conformed to the ServiceSpec.
// Otherwise, the fall-through case is that a Service exists that matches the ServiceSpec and so in this
// case - cluster state is not modified.
func (r *NuxeoReconciler) reconcileService(svc v1alpha1.ServiceSpec, nodeSet v1alpha1.NodeSet,
	instance *v1alpha1.Nuxeo) error {
	svcName := serviceName(instance, nodeSet)
	isTLS := false
	if nodeSet.NuxeoConfig.TlsSecret != "" {
		// if Nuxeo is terminating TLS then configure the service for TLS
		isTLS = true
	}
	expected, err := r.defaultService(instance, svc, svcName, isTLS)
	if err != nil {
		return err
	}
	_, err = r.addOrUpdate(svcName, instance.Namespace, expected, &corev1.Service{}, util.ServiceComparer)
	return err
}

// defaultService generates and returns a Service struct from the passed params. The default Service
// generated - if no overrides are provided - is as follows:
//  apiVersion: v1
//    kind: Service
//  metadata:
//    name: <from the svcName arg>
//    namespace: <instance.ObjectMeta.Namespace>
//  spec:
//    type: ClusterIP
//    selector:
//      app: "nuxeo",
//		nuxeoCr: <instance.ObjectMeta.Name>,
//      interactive: "true"
//    ports:
//      - name: web
//        port: 80 (or 443)
//        targetPort: 8080 (or 8443)
func (r *NuxeoReconciler) defaultService(instance *v1alpha1.Nuxeo, svc v1alpha1.ServiceSpec,
	svcName string, isTLS bool) (*corev1.Service, error) {
	var svcType = corev1.ServiceTypeClusterIP
	var port int32 = 80
	var targetPort int32 = 8080
	if isTLS || instance.Spec.RevProxy != (v1alpha1.RevProxySpec{}) {
		port = 443
		targetPort = 8443
	}
	if svc != (v1alpha1.ServiceSpec{}) {
		port = svc.Port
		targetPort = svc.TargetPort
		svcType = svc.Type
	}
	switch svcType {
	case "ClusterIP":
		s := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      svcName,
				Namespace: instance.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Name:     "web",
					Protocol: corev1.ProtocolTCP,
					Port:     port,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: targetPort,
					},
				}},
				Selector: labelsForNuxeo(instance, true),
				Type:     corev1.ServiceTypeClusterIP,
			},
		}
		_ = controllerutil.SetControllerReference(instance, &s, r.Scheme)
		return &s, nil
	case "NodePort":
		fallthrough
	case "LoadBalancer":
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported Service Type: %v", svcType)
	}
}

// serviceName generates a service name from the passed Nuxeo CR, and the passed NodeSet. The generated
// name consists of the passed Nuxeo CR name + dash + the passed 'nodeSet' name + dash + 'service'. E.g. if
// 'instance.Name' is 'my-nuxeo' and 'nodeSet.Name' is 'cluster' then the function returns 'my-nuxeo-cluster-service'.
func serviceName(instance *v1alpha1.Nuxeo, nodeSet v1alpha1.NodeSet) string {
	return instance.Name + "-" + nodeSet.Name + "-service"
}
