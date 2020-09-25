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
)

// TestControllerUtilDetectIngress tests the ability to detect a Kubernetes Ingress. Since the test setup
// configures this, the result should always be true
func (suite *controllerUtilSuite) TestControllerUtilDetectIngress() {
	ok := suite.r.clusterHasIngress()
	require.True(suite.T(), ok)
}

// controllerUtilSuite is the ControllerUtil test suite structure
type controllerUtilSuite struct {
	suite.Suite
	r         NuxeoReconciler
	nuxeoName string
	namespace string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *controllerUtilSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *controllerUtilSuite) AfterTest(_, _ string) {
	// NOP
}

// This function runs the ControllerUtil unit test suite. It is called by 'go test' and will call every
// function in this file with a controllerUtilSuite receiver that begins with "Test..."
func TestControllerUtilUnitTestSuite(t *testing.T) {
	suite.Run(t, new(controllerUtilSuite))
}
