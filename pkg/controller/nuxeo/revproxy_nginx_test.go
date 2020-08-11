package nuxeo

import (
	_ "context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
)

// TestBasicNginxRevProxy defines a Nuxeo struct with an Nginx rev proxy configured and a mock Deployment. The
// test performs some basic validation of the expected Nginx sidecar configuration.
func (suite *nginxRevProxySpecSuite) TestBasicNginxRevProxy() {
	nux := suite.nginxRevProxySpecSuiteNewNuxeo()
	dep := genTestDeploymentForNginxSuite()
	configureNginx(&dep, nux.Spec.RevProxy.Nginx)
	require.Equal(suite.T(), 2, len(dep.Spec.Template.Spec.Containers), "Container not created")
	nginx := dep.Spec.Template.Spec.Containers[1]
	require.Equal(suite.T(), "nginx", nginx.Name, "Container name incorrect")
	volCnt := 0
	for _, vol := range dep.Spec.Template.Spec.Volumes {
		if vol.Name == "nginx-conf" && vol.VolumeSource.ConfigMap.LocalObjectReference.Name == suite.nginxConfigMap {
			volCnt += 1
		} else if vol.Name == "nginx-cert" && vol.VolumeSource.Secret.SecretName == suite.nginxSecret {
			volCnt += 1
		}
	}
	require.Equal(suite.T(), 2, volCnt, "Deployment volumes incorrect")
}

// nginxRevProxySpecSuite is the NginxRevProxySpec test suite structure
type nginxRevProxySpecSuite struct {
	suite.Suite
	r              ReconcileNuxeo
	deploymentName string
	nginxConfigMap string
	nginxSecret    string
	nginxImage     string
}

// SetupSuite initializes the Fake client, a ReconcileNuxeo struct, and various test suite constants
func (suite *nginxRevProxySpecSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.deploymentName = "testclust"
	suite.nginxConfigMap = "testcm"
	suite.nginxSecret = "testsecret"
	suite.nginxImage = "testimage"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *nginxRevProxySpecSuite) AfterTest(_, _ string) {
	// nop
}

// This function runs the NginxRevProxySpec unit test suite. It is called by 'go test' and will call every
// function in this file with a nginxRevProxySpecSuite receiver that begins with "Test..."
func TestNginxRevProxySpecUnitTestSuite(t *testing.T) {
	suite.Run(t, new(nginxRevProxySpecSuite))
}

// nginxRevProxySpecSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *nginxRevProxySpecSuite) nginxRevProxySpecSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "z",
			Namespace: "z",
		},
		Spec: v1alpha1.NuxeoSpec{
			RevProxy: v1alpha1.RevProxySpec{
				Nginx: v1alpha1.NginxRevProxySpec{
					ConfigMap: suite.nginxConfigMap,
					Secret:    suite.nginxSecret,
					Image:     suite.nginxImage,
				},
			},
			NodeSets: []v1alpha1.NodeSet{{
				Name:     suite.deploymentName,
				Replicas: 1,
			}},
		},
	}
}

// genTestDeploymentForNginxSuite creates and returns a Deployment struct minimally configured to support this suite
func genTestDeploymentForNginxSuite() appsv1.Deployment {
	replicas := int32(1)
	dep := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name: "nuxeo",
					}},
				},
			},
		},
	}
	return dep
}
