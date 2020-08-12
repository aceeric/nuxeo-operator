package nuxeo

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// initUnitTestReconcile registers schema types, creates a Fake client, and returns the Fake client
// wrapped in a ReconcileNuxeo struct.
func initUnitTestReconcile() ReconcileNuxeo {
	objs := []runtime.Object{&v1alpha1.Nuxeo{}}
	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.Nuxeo{}, &v1alpha1.NuxeoList{})
	_ = registerOpenShiftRoute()
	_ = registerKubernetesIngress()
	cl := fake.NewFakeClientWithScheme(s, objs...)
	return ReconcileNuxeo{
		client: cl,
		scheme: s,
		logger: log,
	}
}
