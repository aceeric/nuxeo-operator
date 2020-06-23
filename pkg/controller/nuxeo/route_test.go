package nuxeo

import (
	"context"
	"testing"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TestBasicRouteCreation tests the basic mechanics of creating a new OpenShift Route from the Nuxeo CR spec
// when a Route does not already exist
func (suite *routeSuite) TestBasicRouteCreation() {
	nux := suite.routeSuiteNewNuxeo()
	result, err := reconcileOpenShiftRoute(&suite.r, nux.Spec.Access, nux.Spec.NodeSets[0], nux, log)
	require.Nil(suite.T(), err, "reconcileOpenShiftRoute failed with err: %v\n", err)
	require.Equal(suite.T(), reconcile.Result{}, result, "reconcileOpenShiftRoute returned unexpected result: %v\n", result)
	found := &routev1.Route{}
	expectedRouteName := suite.nuxeoName + "-" + suite.deploymentName + "-" + "route"
	err = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedRouteName, Namespace: suite.namespace}, found)
	require.Nil(suite.T(), err, "Route creation failed with err: %v\n", err)
	require.Equal(suite.T(), suite.routeHostName, found.Spec.Host, "Route has incorrect host name: %v\n", found.Spec.Host)
}

// TestRouteHostChange creates a Route, then changes the hostname in the Nuxeo CR and does a reconciliation. Then
// it verifies the Route hostname was updated. Since all of the basic mechanics of Route reconciliation are verified
// in the TestBasicRouteCreation function, this function dispenses with the various require.Nil - etc. - checks.
// It seems redundant to me to repeat them here: if they would fail here, they would fail there.
func (suite *routeSuite) TestRouteHostChange() {
	nux := suite.routeSuiteNewNuxeo()
	// create the route
	_, _ = reconcileOpenShiftRoute(&suite.r, nux.Spec.Access, nux.Spec.NodeSets[0], nux, log)
	newHostName := "modified." + nux.Spec.Access.Hostname
	nux.Spec.Access.Hostname = newHostName
	// should update the route
	_, _ = reconcileOpenShiftRoute(&suite.r, nux.Spec.Access, nux.Spec.NodeSets[0], nux, log)
	expectedRouteName := suite.nuxeoName + "-" + suite.deploymentName + "-" + "route"
	found := &routev1.Route{}
	_ = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedRouteName, Namespace: suite.namespace}, found)
	require.Equal(suite.T(), newHostName, found.Spec.Host, "Route has incorrect host name: %v\n", found.Spec.Host)
}

func (suite *routeSuite) TestRouteToTLS() {
	// todo-me test when the Nuxeo CR is updated to indicate TLS passthrough
}

func (suite *routeSuite) TestRouteFromTLS() {
	// todo-me test when Nuxeo CR indicates TLS passthrough then is updated to standard HTTP
}

// routeSuite is the Route test suite structure
type routeSuite struct {
	suite.Suite
	r              ReconcileNuxeo
	nuxeoName      string
	routeHostName  string
	namespace      string
	deploymentName string
}

// SetupSuite initializes the Fake client, a ReconcileNuxeo struct, and various test suite constants
func (suite *routeSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.routeHostName = "test-host.corpdomain.io"
	suite.namespace = "testns"
	suite.deploymentName = "testclust"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *routeSuite) AfterTest(_, _ string) {
	obj := routev1.Route{}
	_ = suite.r.client.DeleteAllOf(context.TODO(), &obj)
}

// This function runs the Route unit test suite. It is called by 'go test' and will call every
// function in this file with a routeSuite receiver that begins with "Test..."
func TestRouteUnitTestSuite(t *testing.T) {
	suite.Run(t, new(routeSuite))
}

// routeSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite
func (suite *routeSuite) routeSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		Spec: v1alpha1.NuxeoSpec{
			Access: v1alpha1.NuxeoAccess{
				Hostname: suite.routeHostName,
			},
			NodeSets: []v1alpha1.NodeSet{{
				Name:     suite.deploymentName,
				Replicas: 1,
			}},
		},
	}
}
