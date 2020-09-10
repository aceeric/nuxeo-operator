package nuxeo

import (
	"testing"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Tests injection of CLID-related mounts into the Nuxeo container
func (suite *clidSuite) TestBasicClid() {
	nux := suite.clidSuiteNewNuxeo()
	dep := genTestDeploymentForClidSuite()
	err := configureClid(nux, &dep)
	require.Nil(suite.T(), err, "configureClid failed")
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Containers[0].VolumeMounts),
		"Volume Mounts not correctly defined")
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Volumes),
		"Volumes not correctly defined")
}

// Tests CLID formatting. The CLID configured into the Nuxeo CR has to be a single line, and has to contain
// the double-dash "--" separator
func (suite *clidSuite) TestClidFormat() {
	nux := suite.clidSuiteNewNuxeo()
	_, err := suite.r.defaultClidCM(nux, "NOTVALID")
	require.NotNil(suite.T(), err)
	_, err = suite.r.defaultClidCM(nux, "IS--VALID")
	require.Nil(suite.T(), err)
}

// clidSuite is the Clid test suite structure
type clidSuite struct {
	suite.Suite
	r         NuxeoReconciler
	nuxeoName string
	namespace string
	clidVal   string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *clidSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
	suite.clidVal = "11111111111111111111111111111111111111111111111111111111111111111111" +
		"22222222222222222222222222222222222222222222222222222222222222222222"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *clidSuite) AfterTest(_, _ string) {
	// nop
}

// This function runs the Clid unit test suite. It is called by 'go test' and will call every
// function in this file with a clidSuite receiver that begins with "Test..."
func TestClidUnitTestSuite(t *testing.T) {
	suite.Run(t, new(clidSuite))
}

// clidSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *clidSuite) clidSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		// whatever else is needed for the suite
		Spec: v1alpha1.NuxeoSpec{
			Clid: suite.clidVal,
		},
	}
}

// genTestDeploymentForClidSuite creates and returns a Deployment struct minimally configured to support this suite
func genTestDeploymentForClidSuite() appsv1.Deployment {
	replicas := int32(1)
	dep := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: NuxeoServiceAccountName,
					Containers: []corev1.Container{{
						Name: "nuxeo",
					}},
				},
			},
		},
	}
	return dep
}
