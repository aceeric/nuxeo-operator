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
	"context"
	"testing"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TestBasicServiceAccountCreation tests the basic mechanics of creating a new ServiceAccount from the Nuxeo CR spec
// when a ServiceAccount does not already exist
func (suite *serviceAccountSuite) TestBasicServiceAccountCreation() {
	nux := suite.serviceAccountSuiteNewNuxeo()
	err := suite.r.reconcileServiceAccount(nux)
	require.Nil(suite.T(), err, "reconcileServiceAccount failed")
	found := &corev1.ServiceAccount{}
	err = suite.r.Get(context.TODO(), types.NamespacedName{Name: NuxeoServiceAccountName, Namespace: suite.namespace}, found)
	require.Nil(suite.T(), err, "ServiceAccount creation failed")
}

// serviceAccountSuite is the ServiceAccount test suite structure
type serviceAccountSuite struct {
	suite.Suite
	r         NuxeoReconciler
	namespace string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *serviceAccountSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.namespace = "testns"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *serviceAccountSuite) AfterTest(_, _ string) {
	obj := corev1.ServiceAccount{}
	_ = suite.r.DeleteAllOf(context.TODO(), &obj)
}

// This function runs the ServiceAccount unit test suite. It is called by 'go test' and will call every
// function in this file with a serviceAccountSuite receiver that begins with "Test..."
func TestServiceAccountUnitTestSuite(t *testing.T) {
	suite.Run(t, new(serviceAccountSuite))
}

// serviceAccountSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite. There isn't
// muc functionality in the ServiceAccount at this time so this is mostly a shell
func (suite *serviceAccountSuite) serviceAccountSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: suite.namespace,
		},
	}
}
