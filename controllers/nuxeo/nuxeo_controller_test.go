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
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Just a gimmick to be able to unit test nuxeo_controller.go
type dummyRESTMapper struct{}

func (_ dummyRESTMapper) KindFor(_ schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	return schema.GroupVersionKind{}, nil
}
func (_ dummyRESTMapper) KindsFor(resource schema.GroupVersionResource) ([]schema.GroupVersionKind, error) {
	return []schema.GroupVersionKind{}, nil
}
func (_ dummyRESTMapper) ResourceFor(input schema.GroupVersionResource) (schema.GroupVersionResource, error) {
	return schema.GroupVersionResource{}, nil
}
func (_ dummyRESTMapper) ResourcesFor(input schema.GroupVersionResource) ([]schema.GroupVersionResource, error) {
	return []schema.GroupVersionResource{}, nil
}
func (_ dummyRESTMapper) RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	return nil, nil
}
func (_ dummyRESTMapper) RESTMappings(gk schema.GroupKind, versions ...string) ([]*meta.RESTMapping, error) {
	return nil, nil
}
func (_ dummyRESTMapper) ResourceSingularizer(resource string) (singular string, err error) {
	return "", nil
}

// Invokes TestSetupWithManager in nuxeo_controller.go
func (suite *nuxeoControllerSuite) TestSetupWithManager() {
	cfg := &rest.Config{}
	drm := dummyRESTMapper{}
	mgr, err := manager.New(cfg, manager.Options{
		MapperProvider: func(c *rest.Config) (meta.RESTMapper, error) { return drm, nil },
	})
	require.Nil(suite.T(), err)
	err = suite.r.SetupWithManager(mgr)
	require.Nil(suite.T(), err)
}

// nuxeoControllerSuite is the NuxeoController test suite structure
type nuxeoControllerSuite struct {
	suite.Suite
	r         NuxeoReconciler
	nuxeoName string
	namespace string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *nuxeoControllerSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *nuxeoControllerSuite) AfterTest(_, _ string) {
	// NOP
}

// This function runs the NuxeoController unit test suite. It is called by 'go test' and will call every
// function in this file with a nuxeoControllerSuite receiver that begins with "Test..."
func TestNuxeoControllerUnitTestSuite(t *testing.T) {
	suite.Run(t, new(nuxeoControllerSuite))
}
