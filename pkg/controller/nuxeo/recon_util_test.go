package nuxeo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
)

// Tests the addOrUpdate function with a secret. Note that the fake client will not encode the Secret Data so the
// test creates a secret with unencoded data in the Data member, and does an unencoded comparison on the created
// secret.
func (suite *reconUtilSuite) TestReconUtilSecret() {
	var err error
	exp := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.secretName,
			Namespace: suite.namespace,
		},
		Data: map[string][]byte{suite.secretKey: suite.secretData},
		Type: v1.SecretTypeOpaque,
	}
	_, err = suite.r.addOrUpdate(suite.secretName, suite.namespace, &exp, &v1.Secret{}, util.SecretComparer)
	require.Nil(suite.T(), err, "addOrUpdate failed with error")
	created := v1.Secret{}
	err = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: suite.secretName, Namespace: suite.namespace}, &created)
	require.Nil(suite.T(), err, "addOrUpdate did not create secret")
	require.Equal(suite.T(), created.Data[suite.secretKey], suite.secretData, "Secret data is incorrect")
}

// Tests the removeIfPresent function. Creates a nuxeo struct with a UID. Creates a secret with an owner reference
// indicating the Nuxeo CR is an owner. Confirms the secret was deleted.
func (suite *reconUtilSuite) TestRemoveSecret() {
	nux := suite.reconUtilSuiteNewNuxeo()
	exp := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.secretName,
			Namespace: suite.namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "v1",
				Kind:       "certificatesToPEM",
				Name:       "bar",
				UID:        "12312312-1231-1231-1231-123123123123",
			}, {
				APIVersion: "v2",
				Kind:       "ting",
				Name:       "tang",
				UID:        suite.nuxeoUID,
			}},
		},
		Data: map[string][]byte{suite.secretKey: suite.secretData},
		Type: v1.SecretTypeOpaque,
	}
	err := suite.r.client.Create(context.TODO(), &exp)
	require.Nil(suite.T(), err, "failed to create secret")
	err = suite.r.removeIfPresent(nux, suite.secretName, suite.namespace, &exp)
	require.Nil(suite.T(), err, "removeIfPresent failed")
}

// Tests removeIfPresent with an object that doesn't exist in the cluster and then again with an object that exists
// but isn't owned by Nuxeo to ensure that both cases are handled. In the first case, it's basically a NOP, in the
// second case the resource should not be removed because it is not owned by the Nuxeo CR
func (suite *reconUtilSuite) TestShouldNotRemove() {
	nux := suite.reconUtilSuiteNewNuxeo()
	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.cmName,
			Namespace: suite.namespace,
		},
		Data: map[string]string{"x": "y"},
	}
	err := suite.r.removeIfPresent(nux, suite.cmName, suite.namespace, &cm)
	require.Nil(suite.T(), err, "removeIfPresent failed")
	err = suite.r.client.Create(context.TODO(), &cm)
	require.Nil(suite.T(), err, "failed to create object")
	err = suite.r.removeIfPresent(nux, suite.cmName, suite.namespace, &cm)
	require.Nil(suite.T(), err, "removeIfPresent failed")
	exists := v1.ConfigMap{}
	err = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: suite.cmName, Namespace: suite.namespace}, &exists)
	require.Nil(suite.T(), err, "failed to get object")
}

// reconUtilSuite is the ReconUtil test suite structure
type reconUtilSuite struct {
	suite.Suite
	r          ReconcileNuxeo
	nuxeoName  string
	namespace  string
	secretKey  string
	secretData []byte
	secretName string
	cmName     string
	nuxeoUID   types.UID
}

// SetupSuite initializes the Fake client, a ReconcileNuxeo struct, and various test suite constants
func (suite *reconUtilSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
	suite.secretKey = "certificatesToPEM"
	suite.secretData = []byte("bar")
	suite.secretName = "testsecret"
	suite.cmName = "mycm"
	suite.nuxeoUID = "00000000-0000-0000-0000-000000000000"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *reconUtilSuite) AfterTest(_, _ string) {
	s := v1.Secret{}
	_ = suite.r.client.DeleteAllOf(context.TODO(), &s)
	cm := v1.ConfigMap{}
	_ = suite.r.client.DeleteAllOf(context.TODO(), &cm)
}

// This function runs the ReconUtil unit test suite. It is called by 'go test' and will call every
// function in this file with a reconUtilSuite receiver that begins with "Test..."
func TestReconUtilUnitTestSuite(t *testing.T) {
	suite.Run(t, new(reconUtilSuite))
}

// reconUtilSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *reconUtilSuite) reconUtilSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
			UID:       suite.nuxeoUID,
		},
		// whatever else is needed for the suite
		Spec: v1alpha1.NuxeoSpec{
		},
	}
}
