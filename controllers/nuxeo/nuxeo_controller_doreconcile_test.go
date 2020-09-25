package nuxeo

import (
	"context"
	"testing"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TestReconcileNuxeoCR calls the Reconcile function which is the top-level reconciler
// by the controller manager
func (suite *nuxeoCRSuite) TestReconcileNuxeoCR() {
	rq := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: suite.namespace,
			Name:      suite.nuxeoName,
		},
	}
	// handles the case where the requested object was deleted
	_, err := suite.r.Reconcile(rq)
	require.Nil(suite.T(), err)

	// now create the CR and reconcile it
	nux := suite.nuxeoCRSuiteNewNuxeo()
	err = suite.r.Client.Create(context.TODO(), nux)
	require.Nil(suite.T(), err)
	result, err := suite.r.Reconcile(rq)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), reconcile.Result{Requeue: true}, result)
}

// nuxeoCRSuite is the NuxeoCR test suite structure
type nuxeoCRSuite struct {
	suite.Suite
	r              NuxeoReconciler
	nuxeoName      string
	namespace      string
	deploymentName string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *nuxeoCRSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
	suite.deploymentName = "nuxeo"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *nuxeoCRSuite) AfterTest(_, _ string) {
	//obj := the type of object you are testing. E.g.: obj := corev1.Service{}
	//_ = suite.r.client.DeleteAllOf(context.TODO(), &obj)
}

// This function runs the NuxeoCR unit test suite. It is called by 'go test' and will call every
// function in this file with a nuxeoCRSuite receiver that begins with "Test..."
func TestNuxeoCRUnitTestSuite(t *testing.T) {
	suite.Run(t, new(nuxeoCRSuite))
}

// nuxeoCRSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *nuxeoCRSuite) nuxeoCRSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		// whatever else is needed for the suite
		Spec: v1alpha1.NuxeoSpec{
			NodeSets: []v1alpha1.NodeSet{{
				Name:        suite.deploymentName,
				Replicas:    1,
				Interactive: true,
			}},
		},
	}
}
