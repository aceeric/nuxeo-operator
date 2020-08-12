package nuxeo

import (
	"context"
	"testing"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
)

// TestBasicRouteCreation tests the basic mechanics of creating a new OpenShift Route from the Nuxeo CR spec
// when a Route does not already exist
func (suite *routeSuite) TestBasicRouteCreation() {
	nux := suite.routeSuiteNewNuxeo()
	err := reconcileOpenShiftRoute(&suite.r, nux.Spec.Access, false, nux.Spec.NodeSets[0], nux)
	require.Nil(suite.T(), err, "reconcileOpenShiftRoute failed")
	found := &routev1.Route{}
	expectedRouteName := suite.nuxeoName + "-" + suite.deploymentName + "-" + "route"
	err = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedRouteName, Namespace: suite.namespace}, found)
	require.Nil(suite.T(), err, "Route creation failed")
	require.Equal(suite.T(), suite.routeHostName, found.Spec.Host, "Route has incorrect host name")
}

// TestRouteHostChange creates a Route, then changes the hostname in the Nuxeo CR and does a reconciliation. Then
// it verifies the Route hostname was updated. Since all of the basic mechanics of Route reconciliation are verified
// in the TestBasicRouteCreation function, this function dispenses with the various require.Nil - etc. - checks.
// It seems redundant to me to repeat them here: if they would fail here, they would fail there.
func (suite *routeSuite) TestRouteHostChange() {
	nux := suite.routeSuiteNewNuxeo()
	// create the route
	_ = reconcileOpenShiftRoute(&suite.r, nux.Spec.Access, false, nux.Spec.NodeSets[0], nux)
	newHostName := "modified." + nux.Spec.Access.Hostname
	nux.Spec.Access.Hostname = newHostName
	// should update the route
	_ = reconcileOpenShiftRoute(&suite.r, nux.Spec.Access, false, nux.Spec.NodeSets[0], nux)
	expectedRouteName := suite.nuxeoName + "-" + suite.deploymentName + "-" + "route"
	found := &routev1.Route{}
	_ = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedRouteName, Namespace: suite.namespace}, found)
	require.Equal(suite.T(), newHostName, found.Spec.Host, "Route has incorrect host name")
}

// TestRouteToTLS creates a basic HTTP route from a Nuxeo CR, then updates the CR to indicate TLS. Reconciles the
// Nuxeo CR and confirms the route was changed to support TLS.
func (suite *routeSuite) TestRouteToTLS() {
	nux := suite.routeSuiteNewNuxeo()
	_ = reconcileOpenShiftRoute(&suite.r, nux.Spec.Access, false, nux.Spec.NodeSets[0], nux)
	_ = createTlsSecret(suite)
	nux.Spec.Access.TLSSecret = suite.tlsSecretName
	nux.Spec.Access.Termination = routev1.TLSTerminationPassthrough
	_ = reconcileOpenShiftRoute(&suite.r, nux.Spec.Access, false, nux.Spec.NodeSets[0], nux)
	expectedRouteName := suite.nuxeoName + "-" + suite.deploymentName + "-" + "route"
	found := &routev1.Route{}
	_ = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedRouteName, Namespace: suite.namespace}, found)
	require.Equal(suite.T(), suite.tlsCert, found.Spec.TLS.Certificate, "Route not updated")
}

// TestRouteFromTLS is the opposite of TestRouteToTLS
func (suite *routeSuite) TestRouteFromTLS() {
	nux := suite.routeSuiteNewNuxeo()
	_ = createTlsSecret(suite)
	nux.Spec.Access.TLSSecret = suite.tlsSecretName
	nux.Spec.Access.Termination = routev1.TLSTerminationPassthrough
	_ = reconcileOpenShiftRoute(&suite.r, nux.Spec.Access, false, nux.Spec.NodeSets[0], nux)
	expectedRouteName := suite.nuxeoName + "-" + suite.deploymentName + "-" + "route"
	found := &routev1.Route{}
	_ = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedRouteName, Namespace: suite.namespace}, found)
	require.Equal(suite.T(), suite.tlsCert, found.Spec.TLS.Certificate, "Route incorrectly configured")
	// un-configure TLS. Should cause the route to become plain HTTP
	nux.Spec.Access.TLSSecret = ""
	nux.Spec.Access.Termination = ""
	_ = reconcileOpenShiftRoute(&suite.r, nux.Spec.Access, false, nux.Spec.NodeSets[0], nux)
	foundUpdated := &routev1.Route{}
	_ = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedRouteName, Namespace: suite.namespace}, foundUpdated)
	require.Nil(suite.T(), foundUpdated.Spec.TLS, "Route not updated")
}

// TestRouteForcePassthrough tests the logic where configuring Nuxeo to terminate TLS causes the Route to be configured
// for TLS Passthrough
func (suite *routeSuite) TestRouteForcePassthrough() {
	nux := suite.routeSuiteNewNuxeo()
	nux.Spec.NodeSets[0].NuxeoConfig.TlsSecret = "dummy"
	_ = reconcileAccess(&suite.r, nux.Spec.Access, nux.Spec.NodeSets[0], nux)
	expectedRouteName := suite.nuxeoName + "-" + suite.deploymentName + "-" + "route"
	found := &routev1.Route{}
	_ = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedRouteName, Namespace: suite.namespace}, found)
	require.Equal(suite.T(), routev1.TLSTerminationPassthrough, found.Spec.TLS.Termination, "Route not configured")
}

// routeSuite is the Route test suite structure
type routeSuite struct {
	suite.Suite
	r              ReconcileNuxeo
	nuxeoName      string
	routeHostName  string
	namespace      string
	deploymentName string
	tlsSecretName  string
	tlsCert        string
	tlsKey         string
}

// SetupSuite initializes the Fake client, a ReconcileNuxeo struct, and various test suite constants
func (suite *routeSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.routeHostName = "test-host.corpdomain.io"
	suite.namespace = "testns"
	suite.deploymentName = "testclust"
	suite.tlsSecretName = "testsecret"
	suite.tlsCert = "THECERT"
	suite.tlsKey = "THEKEY"
	util.SetIsOpenShift(true)
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
				Name:        suite.deploymentName,
				Interactive: true,
				Replicas:    1,
			}},
		},
	}
}

// Generate connection secrets the way ECK generates them
func createTlsSecret(suite *routeSuite) error {
	userSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.tlsSecretName,
			Namespace: suite.namespace,
		},
		Data: map[string][]byte{
			"certificate": []byte(suite.tlsCert),
			"key":         []byte(suite.tlsKey),
		},
		Type: corev1.SecretTypeOpaque,
	}
	return suite.r.client.Create(context.TODO(), &userSecret)
}
