package nuxeo

import (
	"context"
	"testing"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/aceeric/nuxeo-operator/controllers/util"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestBasicAccess calls 'reconcileAccess' with OpenShift=true and OpenShift=false. The ingress_test.go file
// and the route_test.go file do specific relevant unit tests so this is just testing the top-level function
func (suite *accessSuite) TestBasicAccess() {
	nux := suite.accessSuiteNewNuxeo()
	util.SetIsOpenShift(false)
	err := suite.r.reconcileAccess(nux.Spec.Access, nux.Spec.NodeSets[0], nux)
	require.Nil(suite.T(), err, "reconcileAccess (Kubernetes) failed")
	util.SetIsOpenShift(true)
	err = suite.r.reconcileAccess(nux.Spec.Access, nux.Spec.NodeSets[0], nux)
	require.Nil(suite.T(), err, "reconcileAccess (OpenShift) failed")
}

// accessSuite is the Access test suite structure
type accessSuite struct {
	suite.Suite
	r              NuxeoReconciler
	nuxeoName      string
	namespace      string
	hostName       string
	deploymentName string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *accessSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
	suite.hostName = "test-host.corp.io"
	suite.deploymentName = "testclust"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *accessSuite) AfterTest(_, _ string) {
	objI := v1beta1.Ingress{}
	_ = suite.r.DeleteAllOf(context.TODO(), &objI)
	objR := routev1.Route{}
	_ = suite.r.DeleteAllOf(context.TODO(), &objR)
}

// This function runs the Access unit test suite. It is called by 'go test' and will call every
// function in this file with a accessSuite receiver that begins with "Test..."
func TestAccessUnitTestSuite(t *testing.T) {
	//if err := registerOpenShiftRoute(); err != nil {
	//	t.Fatal("could not register openshift Route schema")
	//}
	suite.Run(t, new(accessSuite))
}

// accessSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *accessSuite) accessSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		Spec: v1alpha1.NuxeoSpec{
			Access: v1alpha1.NuxeoAccess{
				Hostname: suite.hostName,
			},
			NodeSets: []v1alpha1.NodeSet{{
				Name:     suite.deploymentName,
				Replicas: 1,
			}},
		},
	}
}
