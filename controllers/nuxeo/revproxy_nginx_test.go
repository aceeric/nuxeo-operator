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

// TestBasicNginxRevProxy defines a Nuxeo struct with an Nginx rev proxy configured and a mock Deployment. The
// test performs some basic validation of the expected Nginx sidecar configuration.
func (suite *nginxRevProxySpecSuite) TestBasicNginxRevProxy() {
	nux := suite.nginxRevProxySpecSuiteNewNuxeo()
	dep := genTestDeploymentForNginxSuite()
	err := configureNginx(&dep, nux.Spec.RevProxy.Nginx)
	require.Nil(suite.T(), err, "configureNginx failed")
	require.Equal(suite.T(), 2, len(dep.Spec.Template.Spec.Containers), "Container not created")
	nginx := dep.Spec.Template.Spec.Containers[1]
	require.Equal(suite.T(), "nginx", nginx.Name, "Container name incorrect")
	volCnt := 0
	for _, vol := range dep.Spec.Template.Spec.Volumes {
		if vol.Name == "nginx-conf" && vol.VolumeSource.ConfigMap.LocalObjectReference.Name == suite.nginxConfigMap {
			volCnt += 1
		} else if vol.Name == "nginx-cert" && vol.VolumeSource.Secret.SecretName == suite.nginxSecret {
			volCnt += 1
		}
	}
	require.Equal(suite.T(), 2, volCnt, "Deployment volumes incorrect")
}

// TestNginxRevProxyNoCM is the same as TestBasicNginxRevProxy except it does not specify a config map, causing
// the operator to auto-generate one
func (suite *nginxRevProxySpecSuite) TestNginxRevProxyNoCM() {
	nux := suite.nginxRevProxySpecSuiteNewNuxeo()
	nux.Spec.RevProxy.Nginx.ConfigMap = "" // cause the operator to auto-gen
	dep := genTestDeploymentForNginxSuite()
	nginxCmName, err := suite.r.reconcileNginxCM(nux, nux.Spec.RevProxy.Nginx.ConfigMap)
	require.Nil(suite.T(), err, "reconcileNginxCM failed")
	nux.Spec.RevProxy.Nginx.ConfigMap = nginxCmName
	err = configureNginx(&dep, nux.Spec.RevProxy.Nginx)
	require.Nil(suite.T(), err, "configureNginx failed")
	require.Equal(suite.T(), 2, len(dep.Spec.Template.Spec.Containers), "Container not created")
	nginx := dep.Spec.Template.Spec.Containers[1]
	require.Equal(suite.T(), "nginx", nginx.Name, "Container name incorrect")
	autoGen := false
	expCMName := defaultNginxCMName(nux.Name)
	for _, vol := range dep.Spec.Template.Spec.Volumes {
		if vol.Name == "nginx-conf" && vol.VolumeSource.ConfigMap.LocalObjectReference.Name == expCMName {
			autoGen = true
			break
		}
	}
	require.True(suite.T(), autoGen, "Auto-gen ConfigMap failed")
}

// nginxRevProxySpecSuite is the NginxRevProxySpec test suite structure
type nginxRevProxySpecSuite struct {
	suite.Suite
	r              NuxeoReconciler
	deploymentName string
	nginxConfigMap string
	nginxSecret    string
	nginxImage     string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *nginxRevProxySpecSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.deploymentName = "testclust"
	suite.nginxConfigMap = "testcm"
	suite.nginxSecret = "testsecret"
	suite.nginxImage = "testimage"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *nginxRevProxySpecSuite) AfterTest(_, _ string) {
	// nop
}

// This function runs the NginxRevProxySpec unit test suite. It is called by 'go test' and will call every
// function in this file with a nginxRevProxySpecSuite receiver that begins with "Test..."
func TestNginxRevProxySpecUnitTestSuite(t *testing.T) {
	suite.Run(t, new(nginxRevProxySpecSuite))
}

// nginxRevProxySpecSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *nginxRevProxySpecSuite) nginxRevProxySpecSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "z",
			Namespace: "z",
		},
		Spec: v1alpha1.NuxeoSpec{
			RevProxy: v1alpha1.RevProxySpec{
				Nginx: v1alpha1.NginxRevProxySpec{
					ConfigMap: suite.nginxConfigMap,
					Secret:    suite.nginxSecret,
					Image:     suite.nginxImage,
				},
			},
			NodeSets: []v1alpha1.NodeSet{{
				Name:     suite.deploymentName,
				Replicas: 1,
			}},
		},
	}
}

// genTestDeploymentForNginxSuite creates and returns a Deployment struct minimally configured to support this suite
func genTestDeploymentForNginxSuite() appsv1.Deployment {
	replicas := int32(1)
	dep := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name: "nuxeo",
					}},
				},
			},
		},
	}
	return dep
}
