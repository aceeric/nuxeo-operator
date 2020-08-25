package nuxeo

import (
	"context"
	"testing"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TestBasicServiceCreation tests the basic mechanics of creating a new Service from the Nuxeo CR spec
// when a Service does not already exist
func (suite *serviceSuite) TestBasicServiceCreation() {
	nux := suite.serviceSuiteNewNuxeo()
	err := suite.r.reconcileService(nux.Spec.Service, nux.Spec.NodeSets[0], nux)
	require.Nil(suite.T(), err, "reconcileService failed")
	found := &corev1.Service{}
	err = suite.r.Get(context.TODO(), types.NamespacedName{Name: serviceName(nux, nux.Spec.NodeSets[0]),
		Namespace: suite.namespace}, found)
	require.Nil(suite.T(), err, "Route creation failed")
	require.Equal(suite.T(), suite.servicePort, found.Spec.Ports[0].Port,
		"Service has incorrect port number")
}

// TestServiceTargetPortChanged creates a service, then changes target port in the Nuxeo CR and verifies the
// Service object was updated by the reconciler
func (suite *serviceSuite) TestServiceTargetPortChanged() {
	nux := suite.serviceSuiteNewNuxeo()
	_ = suite.r.reconcileService(nux.Spec.Service, nux.Spec.NodeSets[0], nux)
	newTargetPort := nux.Spec.Service.TargetPort + 1000
	nux.Spec.Service.TargetPort = newTargetPort
	_ = suite.r.reconcileService(nux.Spec.Service, nux.Spec.NodeSets[0], nux)
	found := &corev1.Service{}
	_ = suite.r.Get(context.TODO(), types.NamespacedName{Name: serviceName(nux, nux.Spec.NodeSets[0]),
		Namespace: suite.namespace}, found)
	require.Equal(suite.T(), newTargetPort, found.Spec.Ports[0].TargetPort.IntVal,
		"Route has incorrect target port number")
}

// serviceSuite is the Service test suite structure
type serviceSuite struct {
	suite.Suite
	r              NuxeoReconciler
	nuxeoName      string
	deploymentName string
	namespace      string
	servicePort    int32
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *serviceSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
	suite.deploymentName = "testclust"
	suite.servicePort = 1111
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *serviceSuite) AfterTest(_, _ string) {
	obj := corev1.Service{}
	_ = suite.r.DeleteAllOf(context.TODO(), &obj)
}

// This function runs the Service unit test suite. It is called by 'go test' and will call every
// function in this file with a serviceSuite receiver that begins with "Test..."
func TestServiceUnitTestSuite(t *testing.T) {
	suite.Run(t, new(serviceSuite))
}

// serviceSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *serviceSuite) serviceSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		Spec: v1alpha1.NuxeoSpec{
			Service: v1alpha1.ServiceSpec{
				Type:       corev1.ServiceTypeClusterIP,
				Port:       suite.servicePort,
				TargetPort: 2222,
			},
			NodeSets: []v1alpha1.NodeSet{{
				Name:     suite.deploymentName,
				Replicas: 1,
			}},
		},
	}
}
