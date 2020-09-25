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

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// Tests the addVolumeAndItems function
func (suite *mergeUtilSuite) TestTodo() {
}

// mergeUtilSuite is the MergeUtil test suite structure
type mergeUtilSuite struct {
	suite.Suite
	r          NuxeoReconciler
	nuxeoName  string
	namespace  string
	volName    string
	secretName string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *mergeUtilSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
	suite.volName = "testvol"
	suite.secretName = "test-secret"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *mergeUtilSuite) AfterTest(_, _ string) {
	// nop
}

// This function runs the MergeUtil unit test suite. It is called by 'go test' and will call every
// function in this file with a mergeUtilSuite receiver that begins with "Test..."
func TestMergeUtilUnitTestSuite(t *testing.T) {
	suite.Run(t, new(mergeUtilSuite))
}

// mergeUtilSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *mergeUtilSuite) mergeUtilSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		// un-needed for now
		//ObjectMeta: metav1.ObjectMeta{
		//	Name:      suite.nuxeoName,
		//	Namespace: suite.namespace,
		//},
		//// whatever else is needed for the suite
		//Spec: v1alpha1.NuxeoSpec{
		//},
	}
}

// genTestDeploymentForMergeUtilSuite creates and returns a Deployment struct minimally configured to support this suite
func genTestDeploymentForMergeUtilSuite(volName string, secretName string) appsv1.Deployment {
	replicas := int32(1)
	dep := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: volName,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: secretName,
								Items: []corev1.KeyToPath{{
									Key:  "Z",
									Path: "W",
								}},
							},
						},
					}},
				},
			},
		},
	}
	return dep
}
