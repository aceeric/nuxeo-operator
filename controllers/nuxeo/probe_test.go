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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestProbes performs a basic test of default and explicit Probes. It defines a Nuxeo CR with an explicit probe
// and verifies that the explicit probe and a default probe were generated into the deployment.
func (suite *probeSuite) TestProbes() {
	nux := suite.probeSuiteNewNuxeo()
	dep := genTestDeploymentForProbeSuite()
	err := configureProbes(&dep, nux.Spec.NodeSets[0])
	require.Nil(suite.T(), err, "configureProbes failed")
	require.Equal(suite.T(), defaultProbe(false), dep.Spec.Template.Spec.Containers[0].LivenessProbe,
		"No explicit LivenessProbe was defined so a default should have been generated - but it was not. Or, it was generated incorrectly")
	// explicit probe - should match
	require.Equal(suite.T(), nux.Spec.NodeSets[0].ReadinessProbe, dep.Spec.Template.Spec.Containers[0].ReadinessProbe,
		"Explicit ReadinessProbe was defined. Actual ReadinessProbe should have been identical but was not")
}

// same as TestProbes except configures Nuxeo to terminate TLS which requires that the probes also use HTTPS
func (suite *probeSuite) TestProbesHttps() {
	nux := suite.probeSuiteNewNuxeo()
	dep := genTestDeploymentForProbeSuite()
	nux.Spec.NodeSets[0].NuxeoConfig.TlsSecret = "this-will-force-https-probes"
	err := configureProbes(&dep, nux.Spec.NodeSets[0])
	require.Nil(suite.T(), err, "configureProbes failed")
	require.Equal(suite.T(), int32(8443), dep.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.HTTPGet.Port.IntVal,
		"Probe not configured for HTTPS")
	require.Equal(suite.T(), corev1.URISchemeHTTPS, dep.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.HTTPGet.Scheme,
		"Probe not configured for HTTPS")
}

// probeSuite is the Probe test suite structure
type probeSuite struct {
	suite.Suite
	r         NuxeoReconciler
	nuxeoName string
	namespace string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *probeSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *probeSuite) AfterTest(_, _ string) {
	// nop for this suite
}

// This function runs the Probe unit test suite. It is called by 'go test' and will call every
// function in this file with a probeSuite receiver that begins with "Test..."
func TestProbeUnitTestSuite(t *testing.T) {
	suite.Run(t, new(probeSuite))
}

// probeSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *probeSuite) probeSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		// whatever else is needed for the suite
		Spec: v1alpha1.NuxeoSpec{
			NodeSets: []v1alpha1.NodeSet{{
				Name:     "test",
				Replicas: 1,
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"foo", "bar"},
						},
					},
					InitialDelaySeconds: 1,
					TimeoutSeconds:      2,
					PeriodSeconds:       3,
					SuccessThreshold:    4,
					FailureThreshold:    5,
				},
			}},
		},
	}
}

// genTestDeploymentForProbeSuite creates and returns a Deployment struct minimally
// configured to support this suite
func genTestDeploymentForProbeSuite() appsv1.Deployment {
	replicas := int32(1)
	dep := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: NuxeoServiceAccountName,
					Containers: []corev1.Container{{
						Name: "nuxeo",
					}},
				},
			},
		},
	}
	return dep
}
