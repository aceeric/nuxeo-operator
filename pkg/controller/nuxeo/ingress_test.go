package nuxeo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TestBasicIngressCreation tests the basic mechanics of creating a new Kubernetes Ingress from the Nuxeo CR spec
// when an Ingress does not already exist
func (suite *ingressSuite) TestBasicIngressCreation() {
	nux := suite.ingressSuiteNewNuxeo()
	result, err := reconcileIngress(&suite.r, nux.Spec.Access, nux.Spec.NodeSets[0], nux, log)
	require.Nil(suite.T(), err, "reconcileIngress failed with err: %v\n", err)
	require.Equal(suite.T(), reconcile.Result{}, result, "reconcileIngress returned unexpected result: %v\n", result)
	found := &v1beta1.Ingress{}
	expectedIngressName := suite.nuxeoName + "-" + suite.deploymentName + "-" + "ingress"
	err = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedIngressName, Namespace: suite.namespace}, found)
	require.Nil(suite.T(), err, "Ingress creation failed with err: %v\n", err)
	require.Equal(suite.T(), suite.ingressHostName, found.Spec.Rules[0].Host,
		"Ingress has incorrect host name: %v\n", found.Spec.Rules[0].Host)
}

// TestIngressHostChange creates an Ingress, then changes the hostname in the Nuxeo CR and does a reconciliation. Then
// it verifies the Ingress hostname was updated.
func (suite *ingressSuite) TestIngressHostChange() {
	nux := suite.ingressSuiteNewNuxeo()
	// create the ingress
	_, _ = reconcileIngress(&suite.r, nux.Spec.Access, nux.Spec.NodeSets[0], nux, log)
	newHostName := "modified." + nux.Spec.Access.Hostname
	nux.Spec.Access.Hostname = newHostName
	// should update the ingress
	_, _ = reconcileIngress(&suite.r, nux.Spec.Access, nux.Spec.NodeSets[0], nux, log)
	expectedIngressName := suite.nuxeoName + "-" + suite.deploymentName + "-" + "ingress"
	found := &v1beta1.Ingress{}
	_ = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedIngressName, Namespace: suite.namespace}, found)
	require.Equal(suite.T(), newHostName, found.Spec.Rules[0].Host,
		"Ingress has incorrect host name: %v\n", found.Spec.Rules[0].Host)
}

func (suite *ingressSuite) TestIngressToTLS() {
	// todo-me test when the Nuxeo CR is updated to indicate TLS passthrough
}

func (suite *ingressSuite) TestIngressFromTLS() {
	// todo-me test when Nuxeo CR indicates TLS passthrough then is updated to standard HTTP
}

// ingressSuite is the Ingress test suite structure
type ingressSuite struct {
	suite.Suite
	r               ReconcileNuxeo
	nuxeoName       string
	namespace       string
	ingressHostName string
	deploymentName  string
}

// SetupSuite initializes the Fake client, a ReconcileNuxeo struct, and various test suite constants
func (suite *ingressSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
	suite.ingressHostName = "test-host.corpdomain.io"
	suite.deploymentName = "testclust"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *ingressSuite) AfterTest(_, _ string) {
	obj := v1beta1.Ingress{}
	_ = suite.r.client.DeleteAllOf(context.TODO(), &obj)
}

// This function runs the Ingress unit test suite. It is called by 'go test' and will call every
// function in this file with a ingressSuite receiver that begins with "Test..."
func TestIngressUnitTestSuite(t *testing.T) {
	suite.Run(t, new(ingressSuite))
}

// ingressSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *ingressSuite) ingressSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		Spec: v1alpha1.NuxeoSpec{
			Access: v1alpha1.NuxeoAccess{
				Hostname: suite.ingressHostName,
			},
			NodeSets: []v1alpha1.NodeSet{{
				Name:     suite.deploymentName,
				Replicas: 1,
			}},
		},
	}
}
