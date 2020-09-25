package nuxeo

import (
	"os"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// initUnitTestReconcile registers schema types, creates a Fake client, and returns the Fake client
// wrapped in a NuxeoReconciler struct.
func initUnitTestReconcile() NuxeoReconciler {
	objs := []runtime.Object{&v1alpha1.Nuxeo{}}
	s := scheme.Scheme
	s.AddKnownTypes(schema.GroupVersion{Group: "nuxeo.com", Version: "v1alpha1"}, &v1alpha1.Nuxeo{}, &v1alpha1.NuxeoList{})
	cl := fake.NewFakeClientWithScheme(s, objs...)
	r := NuxeoReconciler{
		Client: cl,
		Scheme: s,
		Log:    log.Log.WithName("controller_nuxeo"),
	}
	if err := r.registerOpenShiftRoute(); err != nil {
		log.Log.Error(err, "registerOpenShiftRoute failed")
		os.Exit(1)
	} else if err := r.registerKubernetesIngress(); err != nil {
		log.Log.Error(err, "registerKubernetesIngress failed")
		os.Exit(1)
	}
	return r
}
