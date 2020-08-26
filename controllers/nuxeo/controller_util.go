package nuxeo

import (
	"context"
	"fmt"

	"github.com/aceeric/nuxeo-operator/controllers/util"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// detects Kubernetes/OpenShift and configures resources accordingly
func (r *NuxeoReconciler) controllerConfig() error {
	if r.clusterHasRoute() {
		util.SetIsOpenShift(true)
		if err := r.registerOpenShiftRoute(); err != nil {
			return err
		}
	} else if !r.clusterHasIngress() {
		return fmt.Errorf("unable to determine cluster type")
	} else if err := r.registerKubernetesIngress(); err != nil {
		return err
	}
	return nil
}

// returns true if the cluster contains an OpenShift Route type
// todo would like to do this and below without the default ns
func (r *NuxeoReconciler) clusterHasRoute() bool {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "route.openshift.io", Version: "v1", Kind: "Route"})
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: "foo", Namespace: "default"}, obj)
	if err != nil {
		if _, ok := err.(*meta.NoKindMatchError); ok {
			return false
		}
	}
	return true
}

// returns true if the cluster contains a Kubernetes Ingress type
func (r *NuxeoReconciler) clusterHasIngress() bool {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "networking.k8s.io", Version: "v1beta1", Kind: "Ingress"})
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: "foo", Namespace: "default"}, obj)
	if err != nil {
		if _, ok := err.(*meta.NoKindMatchError); ok {
			return false
		}
	}
	return true
}

// registerOpenShiftRoute registers OpenShift Route types with the Scheme Builder
func (r *NuxeoReconciler) registerOpenShiftRoute() error {
	const GroupName = "route.openshift.io"
	const GroupVersion = "v1"
	schemeGroupVersion := schema.GroupVersion{Group: GroupName, Version: GroupVersion}
	addKnownTypes := func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(schemeGroupVersion,
			&routev1.Route{},
			&routev1.RouteList{},
		)
		metav1.AddToGroupVersion(scheme, schemeGroupVersion)
		return nil
	}
	schemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	return schemeBuilder.AddToScheme(r.Scheme)
}

// registerKubernetesIngress registers Kubernetes Ingress types with the Scheme Builder. Note, according to:
//  https://kubernetes.io/blog/2019/07/18/api-deprecations-in-1-16/
// "Use the networking.k8s.io/v1beta1 API version, available since v1.14"
func (r *NuxeoReconciler) registerKubernetesIngress() error {
	const GroupName = "networking.k8s.io"
	const GroupVersion = "v1beta1"
	schemeGroupVersion := schema.GroupVersion{Group: GroupName, Version: GroupVersion}
	addKnownTypes := func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(schemeGroupVersion,
			&v1beta1.Ingress{},
			&v1beta1.IngressList{},
		)
		metav1.AddToGroupVersion(scheme, schemeGroupVersion)
		return nil
	}
	schemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	return schemeBuilder.AddToScheme(r.Scheme)
}
