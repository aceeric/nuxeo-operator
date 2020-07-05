package nuxeo

import (
	"context"
	goerrors "errors"
	"fmt"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
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
	found := &routev1.Route{}
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

// defaultRoute generates and returns a Route struct from the passed params. The 'tls' section of the route is
// only populated if the passed 'access' arg specifies a TLSSecret and/or Termination - or - the forcePassthrough
// arg is true
func (r *ReconcileNuxeo) defaultRoute(nux *v1alpha1.Nuxeo, access v1alpha1.NuxeoAccess, forcePassthrough bool,
	routeName string, nodeSet v1alpha1.NodeSet) (*routev1.Route, error) {
	if access.Termination != "" && forcePassthrough {
		return nil, goerrors.New("invalid to explicitly specify route/ingress termination if Nuxeo is terminating TLS")
	}
	targetPort := intstr.IntOrString{
		Type:   intstr.String,
		StrVal: "web",
	}
	if access.TargetPort != (intstr.IntOrString{}) {
		targetPort = access.TargetPort
	}
	route := routev1.Route{
		ObjectMeta: v12.ObjectMeta{
			Name:      routeName,
			Namespace: nux.Namespace,
		},
		Spec: routev1.RouteSpec{
			Host: access.Hostname,
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   serviceName(nux, nodeSet),
				Weight: nil,
			},
			Port: &routev1.RoutePort{TargetPort: targetPort},
			TLS:  nil,
		},
	}
	if access.Termination != "" || forcePassthrough {
		term := access.Termination
		if forcePassthrough {
			term = routev1.TLSTerminationPassthrough
		}
		route.Spec.TLS = &routev1.TLSConfig{
			Termination: term,
		}
	}
	if access.TLSSecret != "" && access.Termination != "" {
		s := &v13.Secret{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: access.TLSSecret, Namespace: nux.Namespace}, s)
		if err != nil {
			return nil, goerrors.New(fmt.Sprintf("TLS Secret not found: %v", access.TLSSecret))
		}
		// accept "certificate" and "tls.crt" keys in the secret
		var cert, key []byte
		var ok bool
		if cert, ok = s.Data["certificate"]; !ok {
			cert, _ = s.Data["tls.crt"]
		}
		// accept "key" and "tls.key" keys in the secret
		if key, ok = s.Data["key"]; !ok {
			key, _ = s.Data["tls.key"]
		}
		caCert, _ := s.Data["caCertificate"]
		destCaCert, _ := s.Data["destinationCACertificate"]
		insTermPol, _ := s.Data["insecureEdgeTerminationPolicy"]
		route.Spec.TLS.Certificate = string(cert)
		route.Spec.TLS.Key = string(key)
		route.Spec.TLS.CACertificate = string(caCert)
		route.Spec.TLS.DestinationCACertificate = string(destCaCert)
		route.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyType(insTermPol)
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
