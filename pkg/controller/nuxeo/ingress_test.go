package nuxeo

import (
	"context"
	"testing"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
)

// TestBasicIngressCreation tests the basic mechanics of creating a new Kubernetes Ingress from the Nuxeo CR spec
// when an Ingress does not already exist
func (suite *ingressSuite) TestBasicIngressCreation() {
	nux := suite.ingressSuiteNewNuxeo()
	err := suite.r.reconcileIngress(nux.Spec.Access, false, nux.Spec.NodeSets[0], nux)
	require.Nil(suite.T(), err, "reconcileIngress failed")
	found := &v1beta1.Ingress{}
	expectedIngressName := suite.nuxeoName + "-" + suite.deploymentName + "-" + "ingress"
	err = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedIngressName, Namespace: suite.namespace}, found)
	require.Nil(suite.T(), err, "Ingress creation failed")
	require.Equal(suite.T(), suite.ingressHostName, found.Spec.Rules[0].Host,
		"Ingress has incorrect host name")
}

// TestIngressHostChange creates an Ingress, then changes the hostname in the Nuxeo CR and does a reconciliation. Then
// it verifies the Ingress hostname was updated.
func (suite *ingressSuite) TestIngressHostChange() {
	nux := suite.ingressSuiteNewNuxeo()
	// create the ingress
	_ = suite.r.reconcileIngress(nux.Spec.Access, false, nux.Spec.NodeSets[0], nux)
	newHostName := "modified." + nux.Spec.Access.Hostname
	nux.Spec.Access.Hostname = newHostName
	// should update the ingress
	_ = suite.r.reconcileIngress(nux.Spec.Access, false, nux.Spec.NodeSets[0], nux)
	expectedIngressName := suite.nuxeoName + "-" + suite.deploymentName + "-" + "ingress"
	found := &v1beta1.Ingress{}
	_ = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedIngressName, Namespace: suite.namespace}, found)
	require.Equal(suite.T(), newHostName, found.Spec.Rules[0].Host,
		"Ingress has incorrect host name")
}

// TestIngressToTLS creates a basic HTTP ingress from a Nuxeo CR, then updates the CR to indicate TLS. Reconciles the
// Nuxeo CR and confirms the ingress was changed to support TLS.
func (suite *ingressSuite) TestIngressToTLS() {
	nux := suite.ingressSuiteNewNuxeo()
	_ = suite.r.reconcileIngress(nux.Spec.Access, false, nux.Spec.NodeSets[0], nux)
	_ = createTlsIngressSecret(suite)
	nux.Spec.Access.TLSSecret = suite.tlsSecretName
	nux.Spec.Access.Termination = routev1.TLSTerminationPassthrough
	_ = suite.r.reconcileIngress(nux.Spec.Access, false, nux.Spec.NodeSets[0], nux)
	expectedIngressName := suite.nuxeoName + "-" + suite.deploymentName + "-" + "ingress"
	found := &v1beta1.Ingress{}
	_ = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedIngressName, Namespace: suite.namespace}, found)
	require.Equal(suite.T(), suite.ingressHostName, found.Spec.TLS[0].Hosts[0], "Ingress not configured")
}

// TestIngressFromTLS is the opposite of TestIngressToTLS
func (suite *ingressSuite) TestIngressFromTLS() {
	nux := suite.ingressSuiteNewNuxeo()
	nux.Spec.Access.TLSSecret = suite.tlsSecretName
	nux.Spec.Access.Termination = routev1.TLSTerminationPassthrough
	_ = suite.r.reconcileIngress(nux.Spec.Access, false, nux.Spec.NodeSets[0], nux)
	expectedIngressName := suite.nuxeoName + "-" + suite.deploymentName + "-" + "ingress"
	found := &v1beta1.Ingress{}
	_ = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedIngressName, Namespace: suite.namespace}, found)
	require.Equal(suite.T(), suite.ingressHostName, found.Spec.TLS[0].Hosts[0], "Ingress not configured")
	// un-configure TLS. Should cause the ingress to become plain HTTP
	nux.Spec.Access.TLSSecret = ""
	nux.Spec.Access.Termination = ""
	_ = suite.r.reconcileIngress(nux.Spec.Access, false, nux.Spec.NodeSets[0], nux)
	foundUpdated := &v1beta1.Ingress{}
	_ = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedIngressName, Namespace: suite.namespace}, foundUpdated)
	require.Nil(suite.T(), foundUpdated.Spec.TLS, "Ingress not updated")
}

// TestIngressForcePassthrough tests the logic where configuring Nuxeo to terminate TLS causes the Ingress to be
// configured for TLS Passthrough
func (suite *ingressSuite) TestIngressForcePassthrough() {
	nux := suite.ingressSuiteNewNuxeo()
	nux.Spec.NodeSets[0].NuxeoConfig.TlsSecret = "dummy"
	_ = suite.r.reconcileAccess(nux.Spec.Access, nux.Spec.NodeSets[0], nux)
	expectedIngressName := suite.nuxeoName + "-" + suite.deploymentName + "-" + "ingress"
	found := &v1beta1.Ingress{}
	_ = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: expectedIngressName, Namespace: suite.namespace}, found)
	require.Equal(suite.T(), suite.ingressHostName, found.Spec.TLS[0].Hosts[0], "Ingress not configured")
}

// ingressSuite is the Ingress test suite structure
type ingressSuite struct {
	suite.Suite
	r               ReconcileNuxeo
	nuxeoName       string
	namespace       string
	ingressHostName string
	deploymentName  string
	tlsSecretName   string
	tlsCert         string
	tlsKey          string
}

// SetupSuite initializes the Fake client, a ReconcileNuxeo struct, and various test suite constants
func (suite *ingressSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
	suite.ingressHostName = "test-host.corpdomain.io"
	suite.deploymentName = "testclust"
	suite.tlsCert = "THECERT"
	suite.tlsKey = "THEKEY"
	util.SetIsOpenShift(false)
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

// Generate connection secrets the way ECK generates them
func createTlsIngressSecret(suite *ingressSuite) error {
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
