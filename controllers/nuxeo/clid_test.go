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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Tests injection of CLID-related mounts into the Nuxeo container
func (suite *clidSuite) TestBasicClid() {
	nux := suite.clidSuiteNewNuxeo()
	dep := genTestDeploymentForClidSuite()
	err := configureClid(nux, &dep)
	require.Nil(suite.T(), err, "configureClid failed")
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Containers[0].VolumeMounts),
		"Volume Mounts not correctly defined")
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Volumes),
		"Volumes not correctly defined")
}

// Tests CLID formatting. The CLID configured into the Nuxeo CR has to be a single line, and has to contain
// the double-dash "--" separator
func (suite *clidSuite) TestClidFormat() {
	nux := suite.clidSuiteNewNuxeo()
	_, err := suite.r.defaultClidCM(nux, "NOTVALID")
	require.NotNil(suite.T(), err)
	_, err = suite.r.defaultClidCM(nux, "IS--VALID")
	require.Nil(suite.T(), err)
}

// TestClidReconcile tests the creation of a CLID ConfigMap and then removal of the ConfigMap based on the
// Nuxeo spec
func (suite *clidSuite) TestClidReconcile() {
	nux := suite.clidSuiteNewNuxeo()
	// a valid clid as the -- separator which to operator converts to newline
	nux.Spec.Clid = "test--clid"
	err := suite.r.reconcileClid(nux)
	require.Nil(suite.T(), err)
	cm := &corev1.ConfigMap{}
	err = suite.r.Client.Get(context.TODO(), types.NamespacedName{Name: nuxeoClidConfigMapName, Namespace: suite.namespace}, cm)
	require.Nil(suite.T(), err, "Should have created a CLID CM")
	nux.Spec.Clid = ""
	err = suite.r.reconcileClid(nux)
	require.Nil(suite.T(), err)
	err = suite.r.Client.Get(context.TODO(), types.NamespacedName{Name: nuxeoClidConfigMapName, Namespace: suite.namespace}, cm)
	require.True(suite.T(), apierrors.IsNotFound(err), "Should have removed the CLID CM")
}

// clidSuite is the Clid test suite structure
type clidSuite struct {
	suite.Suite
	r         NuxeoReconciler
	nuxeoName string
	namespace string
	clidVal   string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *clidSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
	suite.clidVal = "11111111111111111111111111111111111111111111111111111111111111111111" +
		"22222222222222222222222222222222222222222222222222222222222222222222"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *clidSuite) AfterTest(_, _ string) {
	// nop
}

// This function runs the Clid unit test suite. It is called by 'go test' and will call every
// function in this file with a clidSuite receiver that begins with "Test..."
func TestClidUnitTestSuite(t *testing.T) {
	suite.Run(t, new(clidSuite))
}

// clidSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *clidSuite) clidSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		// whatever else is needed for the suite
		Spec: v1alpha1.NuxeoSpec{
			Clid: suite.clidVal,
		},
	}
}

// genTestDeploymentForClidSuite creates and returns a Deployment struct minimally configured to support this suite
func genTestDeploymentForClidSuite() appsv1.Deployment {
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
