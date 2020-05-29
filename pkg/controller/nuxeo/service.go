package nuxeo

import (
	"context"
	goerrors "errors"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcileService reconciles the passed ServiceSpec from the Nuxeo CR this operator is watching to the ServiceSpec's
// corresponding in-cluster Service. If no Service exists, a Service is created from the ServiceSpec. If a
// Service exists and its state differs from the ServiceSpec, the Service is conformed to the ServiceSpec.
// Otherwise, the fall-through case is that a Service exists that matches the ServiceSpec and so in this
// case - cluster state is not modified.
func reconcileService(r *ReconcileNuxeo, svc v1alpha1.ServiceSpec, nodeSet v1alpha1.NodeSet,
	instance *v1alpha1.Nuxeo, reqLogger logr.Logger) (reconcile.Result, error) {
	found := &corev1.Service{}
	svcName := serviceName(instance, nodeSet)
	expected, err := r.defaultService(instance, svc, svcName)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: svcName, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Namespace", expected.Namespace, "Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "Namespace", expected.Namespace, "Name", expected.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully
		return reconcile.Result{}, nil
	} else if err != nil {
		reqLogger.Error(err, "Error attempting to get Service for Nuxeo cluster: "+svcName)
		return reconcile.Result{}, err
	}
	if !equality.Semantic.DeepDerivative(expected.Spec, found.Spec) {
		expected.Spec.DeepCopyInto(&found.Spec)
		r.client.Update(context.TODO(), found) // TODO HANDLE ERROR
	}
	return reconcile.Result{}, nil
}

// defaultService generates and returns a Service struct from the passed params. The default Service
// generated - if no overrides are provided - is as follows:
//  apiVersion: v1
//    kind: Service
//  metadata:
//    name: <from the svcName arg>
//    namespace: <nux.ObjectMeta.Namespace>
//  spec:
//    type: ClusterIP
//    selector:
//      app: "nuxeo",
//		nuxeoCr: <nux.ObjectMeta.Name>,
//      interactive: "true"
//    ports:
//      - name: web
//        port: 80
//        targetPort: 8080
func (r *ReconcileNuxeo) defaultService(nux *v1alpha1.Nuxeo, svc v1alpha1.ServiceSpec, svcName string) (*corev1.Service, error) {
	var svcType = corev1.ServiceTypeClusterIP
	var port int32 = 80
	var targetPort int32 = 8080
	if svc != (v1alpha1.ServiceSpec{}) {
		port = int32(svc.Port)
		targetPort = int32(svc.TargetPort)
		svcType = svc.Type
	}
	switch svcType {
	case "ClusterIP":
		s := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      svcName,
				Namespace: nux.Namespace,
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
				Selector: labelsForNuxeo(nux, true),
				Type:     corev1.ServiceTypeClusterIP,
			},
		}
		_ = controllerutil.SetControllerReference(nux, &s, r.scheme)
		return &s, nil
	case "NodePort":
		fallthrough
	case "LoadBalancer":
		fallthrough
	default:
		return nil, goerrors.New(fmt.Sprintf("Unsupported Service Type: %v", svcType))
	}
}

// serviceName generates a service name from the passed Nuxeo CR, and the passed NodeSet. The generated
// name consists of the passed Nuxeo CR name + dash + the passed 'nodeSet' name + dash + 'service'. E.g. if
// 'nux.Name' is 'my-nuxeo' and 'nodeSet.Name' is 'cluster' then the function returns 'my-nuxeo-cluster-service'.
func serviceName(nux *v1alpha1.Nuxeo, nodeSet v1alpha1.NodeSet) string {
	return nux.Name + "-" + nodeSet.Name + "-service"
}
