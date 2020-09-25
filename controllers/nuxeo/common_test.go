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
