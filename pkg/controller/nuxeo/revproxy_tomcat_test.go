package nuxeo

import (
	_ "context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
)

// TestBasicTomcatRevProxy defines a Nuxeo struct with a Tomcat rev proxy configured and a mock Deployment. The
// test performs very basic validation of the expected configuration: The resulting deployment should contain
// a volume, and the nuxeo container should contain three environment variables, and a volume mount. E.g.:
//  spec:
//   template:
//     spec:
//       containers:
//       - env:
//         - name: TOMCAT_KEYSTORE_PASS
//           valueFrom:
//             secretKeyRef:
//               key: keystorePass
//               name: testsecret
//         - name: NUXEO_CUSTOM_PARAM
//           value: |
//             nuxeo.server.https.port=8443
//             nuxeo.server.https.keystoreFile=/etc/secrets/tomcat_keystore/keystore.jks
//             nuxeo.server.https.keystorePass=${env:TOMCAT_KEYSTORE_PASS}
//         - name: NUXEO_TEMPLATES
//           value: https
//         name: nuxeo
//         resources: {}
//         volumeMounts:
//         - mountPath: /etc/secrets/tomcat_keystore
//           name: tomcat-keystore
//           readOnly: true
//       volumes:
//       - name: tomcat-keystore
//         secret:
//           items:
//           - key: keystore.jks
//             path: keystore.jks
//           secretName: testsecret
func (suite *tomcatRevProxySpecSuite) TestBasicTomcatRevProxy() {
	nux := suite.tomcatRevProxySpecSuiteNewNuxeo()
	dep := genTestDeploymentForTomcatSuite()
	err := configureTomcatForTLS(&dep, nux.Spec.RevProxy.Tomcat)
	require.Nil(suite.T(), err, "configureTomcatForTLS failed with err: %v\n", err)
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Volumes), "Incorrect volume configuration\n")
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Containers[0].VolumeMounts), "Incorrect volume mount configuration\n")
	require.Equal(suite.T(), 3, len(dep.Spec.Template.Spec.Containers[0].Env), "Incorrect environment configuration\n")
}

// tomcatRevProxySpecSuite is the TomcatRevProxySpec test suite structure
type tomcatRevProxySpecSuite struct {
	suite.Suite
	r              ReconcileNuxeo
	deploymentName string
	tomcatSecret   string
}

// SetupSuite initializes the Fake client, a ReconcileNuxeo struct, and various test suite constants
func (suite *tomcatRevProxySpecSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.deploymentName = "testclust"
	suite.tomcatSecret = "testsecret"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *tomcatRevProxySpecSuite) AfterTest(_, _ string) {
	// nop
}

// This function runs the TomcatRevProxySpec unit test suite. It is called by 'go test' and will call every
// function in this file with a tomcatRevProxySpecSuite receiver that begins with "Test..."
func TestTomcatRevProxySpecUnitTestSuite(t *testing.T) {
	suite.Run(t, new(tomcatRevProxySpecSuite))
}

// tomcatRevProxySpecSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *tomcatRevProxySpecSuite) tomcatRevProxySpecSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		Spec: v1alpha1.NuxeoSpec{
			RevProxy: v1alpha1.RevProxySpec{
				Tomcat: v1alpha1.TomcatRevProxySpec{
					Secret: suite.tomcatSecret,
				},
			},
			//NodeSets: []v1alpha1.NodeSet{{
			//	Name:     suite.deploymentName,
			//	Replicas: 1,
			//}},
		},
	}
}

// genTestDeploymentForTomcatSuite creates and returns a Deployment struct minimally configured to support this suite
func genTestDeploymentForTomcatSuite() appsv1.Deployment {
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
