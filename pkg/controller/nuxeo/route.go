package nuxeo

import (
	"context"
	errors2 "errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/openshift/api/route/v1"
	v13 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcileOpenShiftRoute configures access to the Nuxeo cluster via an OpenShift Route
func reconcileOpenShiftRoute(r *ReconcileNuxeo, access v1alpha1.NuxeoAccess, forcePassthrough bool, nodeSet v1alpha1.NodeSet,
	instance *v1alpha1.Nuxeo, reqLogger logr.Logger) (reconcile.Result, error) {
	found := &v1.Route{}
	routeName := routeName(instance, nodeSet)
	expected, err := r.defaultRoute(instance, access, forcePassthrough, routeName, nodeSet)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: routeName, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Route", "Namespace", expected.Namespace, "Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Route", "Namespace", expected.Namespace, "Name", expected.Name)
			return reconcile.Result{}, err
		}
		// Route created successfully
		return reconcile.Result{}, nil
	} else if err != nil {
		reqLogger.Error(err, "Error attempting to get Route for Nuxeo cluster: "+routeName)
		return reconcile.Result{}, err
	}
	//if !equality.Semantic.DeepDerivative(expected.Spec, found.Spec) {
	//	reqLogger.Info("Updating Route", "Namespace", expected.Namespace, "Name", expected.Name)
	//	expected.Spec.DeepCopyInto(&found.Spec)
	//	if err = r.client.Update(context.TODO(), found); err != nil {
	//		return reconcile.Result{}, err
	//	}
	//}
	// experiment
	if different, err := util.ObjectsDiffer(expected.Spec, found.Spec); err == nil && different {
		reqLogger.Info("Updating Route", "Namespace", expected.Namespace, "Name", expected.Name)
		expected.Spec.DeepCopyInto(&found.Spec)
		if err = r.client.Update(context.TODO(), found); err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

// defaultRoute generates and returns a Route struct from the passed params. The generated Route is illustrated
// below. The 'tls' section of the route is only populated if the passed 'access' arg specifies a TLSSecret and/or
// Termination.
//
//  apiVersion: v1
//  kind: Route
//  metadata:
//    name: <see routeName function>
//  spec:
//    host: <access.Hostname>
//    port:
//      targetPort: web
//    to:
//      kind: Service
//      name: <see serviceName function>
//    tls:
//      termination: <access.Termination>
//      key: <from secret named 'access.TLSSecret' if 'key' specified in that secret>
//      certificate: "
//      caCertificate: "
//      destinationCACertificate: "
//      insecureEdgeTerminationPolicy: "
func (r *ReconcileNuxeo) defaultRoute(nux *v1alpha1.Nuxeo, access v1alpha1.NuxeoAccess, forcePassthrough bool,
	routeName string, nodeSet v1alpha1.NodeSet) (*v1.Route, error) {
	targetPort := intstr.IntOrString{
		Type:   intstr.String,
		StrVal: "web",
	}
	if access.TargetPort != (intstr.IntOrString{}) {
		targetPort = access.TargetPort
	}
	route := v1.Route{
		ObjectMeta: v12.ObjectMeta{
			Name:      routeName,
			Namespace: nux.Namespace,
		},
		Spec: v1.RouteSpec{
			Host: access.Hostname,
			To: v1.RouteTargetReference{
				Kind:   "Service",
				Name:   serviceName(nux, nodeSet),
				Weight: nil,
			},
			Port: &v1.RoutePort{TargetPort: targetPort},
			TLS:  nil,
		},
	}
	if access.Termination != "" || forcePassthrough {
		route.Spec.TLS = &v1.TLSConfig{
			Termination: access.Termination,
		}
	}
	if access.TLSSecret != "" && access.Termination != "" {
		s := &v13.Secret{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: access.TLSSecret, Namespace: nux.Namespace}, s)
		if err != nil {
			return nil, errors2.New(fmt.Sprintf("TLS Secret not found: %v", access.TLSSecret))
		}
		var cert []byte
		cert, _ = s.Data["certificate"]
		var key []byte
		key, _ = s.Data["key"]
		var caCert []byte
		caCert, _ = s.Data["caCertificate"]
		var destCaCert []byte
		destCaCert, _ = s.Data["destinationCACertificate"]
		var insecterm []byte
		insecterm, _ = s.Data["insecureEdgeTerminationPolicy"]
		route.Spec.TLS.Certificate = string(cert)
		route.Spec.TLS.Key = string(key)
		route.Spec.TLS.CACertificate = string(caCert)
		route.Spec.TLS.DestinationCACertificate = string(destCaCert)
		route.Spec.TLS.InsecureEdgeTerminationPolicy = v1.InsecureEdgeTerminationPolicyType(insecterm)
	}
	_ = controllerutil.SetControllerReference(nux, &route, r.scheme)
	return &route, nil
}

// routeName generates a Route name from the passed Nuxeo CR, and the passed NodeSet. The generated
// name consists of the passed Nuxeo CR name + dash + the passed 'nodeSet' name + dash + 'route'. E.g. if
// 'nux.Name' is 'my-nuxeo' and 'nodeSet.Name' is 'cluster' then the function returns 'my-nuxeo-cluster-route'.
func routeName(nux *v1alpha1.Nuxeo, nodeSet v1alpha1.NodeSet) string {
	return nux.Name + "-" + nodeSet.Name + "-route"
}
