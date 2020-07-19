package nuxeo

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
)

// TestProbes performs a basic test of default and explicit Probes. It defines a Nuxeo CR with an explicit probe
// and verifies that the explicit probe and a default probe were generated into the deployment.
func (suite *probeSuite) TestProbes() {
	nux := suite.probeSuiteNewNuxeo()
	dep := genTestDeploymentForProbeSuite()
	err := addProbes(&dep, nux.Spec.NodeSets[0], false)
	require.Nil(suite.T(), err, "addProbes failed with err: %v\n", err)
	require.Equal(suite.T(), dep.Spec.Template.Spec.Containers[0].LivenessProbe, defaultProbe(false),
		"No explicit LivenessProbe was defined so a default should have been generated - but it was not. Or, it was generated incorrectly\n")
	// explicit probe - should match
	require.Equal(suite.T(), dep.Spec.Template.Spec.Containers[0].ReadinessProbe, nux.Spec.NodeSets[0].ReadinessProbe,
		"Explicit ReadinessProbe was defined. Actual ReadinessProbe should have been identical but was not\n")
}

func (suite *probeSuite) TestProbesHttps() {
	// todo-me test with useHttp=true
}

// probeSuite is the Probe test suite structure
type probeSuite struct {
	suite.Suite
	r         ReconcileNuxeo
	nuxeoName string
	namespace string
}

// SetupSuite initializes the Fake client, a ReconcileNuxeo struct, and various test suite constants
func (suite *probeSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *probeSuite) AfterTest(_, _ string) {
	// nop for this suite
}

// This function runs the Probe unit test suite. It is called by 'go test' and will call every
// function in this file with a probeSuite receiver that begins with "Test..."
func TestProbeUnitTestSuite(t *testing.T) {
	suite.Run(t, new(probeSuite))
}

// probeSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *probeSuite) probeSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		// whatever else is needed for the suite
		Spec: v1alpha1.NuxeoSpec{
			NodeSets: []v1alpha1.NodeSet{{
				Name:     "test",
				Replicas: 1,
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"certificatesToPEM", "bar"},
						},
					},
					InitialDelaySeconds: 1,
					TimeoutSeconds:      2,
					PeriodSeconds:       3,
					SuccessThreshold:    4,
					FailureThreshold:    5,
				},
			}},
		},
	}
}

// genTestDeploymentForProbeSuite creates and returns a Deployment struct minimally
// configured to support this suite
func genTestDeploymentForProbeSuite() appsv1.Deployment {
	replicas := int32(1)
	dep := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: util.NuxeoServiceAccountName,
					Containers: []corev1.Container{{
						Name: "nuxeo",
					}},
				},
			},
		},
	}
	return dep
}
