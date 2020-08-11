package nuxeo

import (
	"context"
	goerrors "errors"
	"fmt"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	v13 "k8s.io/api/core/v1"
	// todo-me clean up all these weird aliases created by GoLand
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func reconcileOpenShiftRoute(r *ReconcileNuxeo, access v1alpha1.NuxeoAccess, forcePassthrough bool,
	nodeSet v1alpha1.NodeSet, instance *v1alpha1.Nuxeo, reqLogger logr.Logger) error {
	routeName := routeName(instance, nodeSet)
	if access != (v1alpha1.NuxeoAccess{}) {
		if expected, err := r.defaultRoute(instance, access, forcePassthrough, routeName, nodeSet); err != nil {
			return err
		} else {
			_, err = addOrUpdate(r, routeName, instance.Namespace, expected, &routev1.Route{}, util.RouteComparer, reqLogger)
			return err
		}
	} else {
		return removeIfPresent(r, instance, routeName, instance.Namespace, &routev1.Route{}, reqLogger)
	}
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
				Weight: util.Int32Ptr(100),
			},
			Port: &routev1.RoutePort{TargetPort: targetPort},
			WildcardPolicy: routev1.WildcardPolicyNone,
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
