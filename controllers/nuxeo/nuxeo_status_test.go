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
	"github.com/aceeric/nuxeo-operator/controllers/util"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TestNuxeoStatus performs a basic status test. Creates a Nuxeo CR. Updates the status to reflect the various
// supported status values. The fake client doesn't update certain values so this test code works around that.
func (suite *nuxeoStatusSuite) TestNuxeoStatus() {
	nux := suite.nuxeoStatusSuiteNewNuxeo()
	// fake doesn't assign a UID
	nux.UID = "12345678-1234-1234-1234-123456789012"
	err := suite.r.Create(context.TODO(), nux)
	require.Nil(suite.T(), err)
	err = suite.r.updateNuxeoStatus(nux)
	require.Nil(suite.T(), err)
	err = suite.r.Client.Get(context.TODO(), types.NamespacedName{Name: suite.nuxeoName, Namespace: suite.namespace}, nux)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), v1alpha1.StatusUnavailable, nux.Status.Status)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "x",
			Namespace: suite.namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "v1alpha1",
				Kind:       "Nuxeo",
				Name:       suite.nuxeoName,
				UID:        nux.UID,
			}},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: util.Int32Ptr(4),
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
	// fake doesn't update available replicas
	dep.Status.AvailableReplicas = *dep.Spec.Replicas
	err = suite.r.Create(context.TODO(), dep)
	require.Nil(suite.T(), err)
	err = suite.r.updateNuxeoStatus(nux)
	require.Nil(suite.T(), err)
	err = suite.r.Client.Get(context.TODO(), types.NamespacedName{Name: suite.nuxeoName, Namespace: suite.namespace}, nux)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), v1alpha1.StatusDegraded, nux.Status.Status)
	dep.Spec.Replicas = util.Int32Ptr(5)
	dep.Status.AvailableReplicas = *dep.Spec.Replicas
	err = suite.r.Client.Update(context.TODO(), dep)
	require.Nil(suite.T(), err)
	err = suite.r.updateNuxeoStatus(nux)
	require.Nil(suite.T(), err)
	err = suite.r.Client.Get(context.TODO(), types.NamespacedName{Name: suite.nuxeoName, Namespace: suite.namespace}, nux)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), v1alpha1.StatusHealthy, nux.Status.Status)
}

// nuxeoStatusSuite is the NuxeoStatus test suite structure
type nuxeoStatusSuite struct {
	suite.Suite
	r         NuxeoReconciler
	nuxeoName string
	namespace string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *nuxeoStatusSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *nuxeoStatusSuite) AfterTest(_, _ string) {
	obj := v1alpha1.Nuxeo{}
	_ = suite.r.Client.DeleteAllOf(context.TODO(), &obj)
}

// This function runs the NuxeoStatus unit test suite. It is called by 'go test' and will call every
// function in this file with a nuxeoStatusSuite receiver that begins with "Test..."
func TestNuxeoStatusUnitTestSuite(t *testing.T) {
	suite.Run(t, new(nuxeoStatusSuite))
}

// nuxeoStatusSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *nuxeoStatusSuite) nuxeoStatusSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		Spec: v1alpha1.NuxeoSpec{
			NodeSets: []v1alpha1.NodeSet{{
				Name:     "foo",
				Replicas: 3,
			}, {
				Name:     "bar",
				Replicas: 2,
			}},
		},
	}
}
