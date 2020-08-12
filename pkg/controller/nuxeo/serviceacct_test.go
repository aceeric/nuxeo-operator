package nuxeo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
)

// TestBasicServiceAccountCreation tests the basic mechanics of creating a new ServiceAccount from the Nuxeo CR spec
// when a ServiceAccount does not already exist
func (suite *serviceAccountSuite) TestBasicServiceAccountCreation() {
	nux := suite.serviceAccountSuiteNewNuxeo()
	err := reconcileServiceAccount(&suite.r, nux)
	require.Nil(suite.T(), err, "reconcileServiceAccount failed")
	found := &corev1.ServiceAccount{}
	err = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: util.NuxeoServiceAccountName, Namespace: suite.namespace}, found)
	require.Nil(suite.T(), err, "ServiceAccount creation failed")
}

// serviceAccountSuite is the ServiceAccount test suite structure
type serviceAccountSuite struct {
	suite.Suite
	r         ReconcileNuxeo
	namespace string
}

// SetupSuite initializes the Fake client, a ReconcileNuxeo struct, and various test suite constants
func (suite *serviceAccountSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.namespace = "testns"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *serviceAccountSuite) AfterTest(_, _ string) {
	obj := corev1.ServiceAccount{}
	_ = suite.r.client.DeleteAllOf(context.TODO(), &obj)
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
