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

package envtest

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/aceeric/nuxeo-operator/controllers/nuxeo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/yaml"

	nuxeov1alpha1 "github.com/aceeric/nuxeo-operator/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var k8sManager ctrl.Manager
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {

	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	// remove when cluster-scoped CRD issue resolved
	crds := readCRDs()
	var res []runtime.Object
	for _, obj := range crds {
		res = append(res, obj)
	}
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		// put back when cluster-scoped CRD issue resolved
		//CRDDirectoryPaths: []string{filepath.Join("../..", "config", "crd", "bases")},
		CRDs: res,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = nuxeov1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme:    scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	r := &nuxeo.NuxeoReconciler{
		Client: k8sManager.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("nuxeo-operator"),
		Scheme: k8sManager.GetScheme(),
	}
	err = r.SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

// todo this is a temporary hack. Can't get the envtest tools to support cluster scoped CRDs. The scaffolding
// provides an empty namespace to the reconciler. Problem goes away when the CRD is Namespace scoped. So as a
// work-around patch the CRD to be Namespaced. This borrows from: sigs.k8s.io/controller-runtime@v0.6.2/
// pkg/envtest/crd.go
func readCRDs() []*unstructured.Unstructured {
	path := filepath.Join("../..", "config", "crd", "bases")
	var crds []*unstructured.Unstructured
	crd := &unstructured.Unstructured{}
	files, _ :=  ioutil.ReadDir(path)
	for _, file := range files {
		docs, _ := readDocuments(filepath.Join(path, file.Name()))
		for _, doc := range docs {
			_ = yaml.Unmarshal(doc, crd)
			(crd.Object["spec"].(map[string]interface{}))["scope"] = "Namespaced"
			crds = append(crds, crd)
		}
	}
	return crds
}

// todo remove this when cluster scoped CRD and namespace issue resolved
func readDocuments(fp string) ([][]byte, error) {
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}
	docs := [][]byte{}
	reader := k8syaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(b)))
	for {
		// Read document
		doc, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if len(regexp.MustCompile(`\s`).ReplaceAll(doc, []byte(""))) > 0 {
			docs = append(docs, doc)
		}
	}
	return docs, nil
}