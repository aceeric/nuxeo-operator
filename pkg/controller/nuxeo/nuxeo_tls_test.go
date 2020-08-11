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

// TestBasicNuxeoTLS defines a Nuxeo struct with Nuxeo TLS configured and a mock Deployment. The
// test performs very basic validation of the expected configuration: The resulting deployment should contain
// a volume, and the nuxeo container should contain two environment variables, and a volume mount. In addition,
// nuxeo.conf settings should be returned. E.g.:
//  spec:
//   template:
//     spec:
//       containers:
//       - env:
//         - name: TLS_KEYSTORE_PASS
//           valueFrom:
//             secretKeyRef:
//               key: keystorePass
//               name: testsecret
//         - name: NUXEO_TEMPLATES
//           value: https
//         name: nuxeo
//         resources: {}
//         volumeMounts:
//         - mountPath: /etc/secrets/tls-keystore
//           name: tls-keystore
//           readOnly: true
//       volumes:
//       - name: tls-keystore
//         secret:
//           items:
//           - key: keystore.jks
//             path: keystore.jks
//           secretName: testsecret
func (suite *nuxeoTLSSuite) TestBasicNuxeoTLS() {
	nux := suite.nuxeoTLSSuiteNewNuxeo()
	dep := genTestDeploymentForNuxeoTLSSuite()
	nuxeoconf, err := configureNuxeoForTLS(&dep, nux.Spec.NodeSets[0].NuxeoConfig.TlsSecret)
	require.Nil(suite.T(), err, "configureNuxeoForTLS failed")
	require.Equal(suite.T(), suite.nuxeoConf, nuxeoconf, "Incorrect nuxeo.conf string")
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Volumes), "Incorrect volume configuration")
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Containers[0].VolumeMounts), "Incorrect volume mount configuration")
	require.Equal(suite.T(), 2, len(dep.Spec.Template.Spec.Containers[0].Env), "Incorrect environment configuration")
}

// nuxeoTLSSuite is the NuxeoTLS test suite structure
type nuxeoTLSSuite struct {
	suite.Suite
	r              ReconcileNuxeo
	deploymentName string
	tlsSecret      string
	nuxeoConf      string
}

// SetupSuite initializes the Fake client, a ReconcileNuxeo struct, and various test suite constants
func (suite *nuxeoTLSSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.deploymentName = "testclust"
	suite.tlsSecret = "testsecret"
	suite.nuxeoConf = "nuxeo.server.https.port=8443\n" +
		"nuxeo.server.https.keystoreFile=/etc/secrets/tls-keystore/keystore.jks\n"  +
		"nuxeo.server.https.keystorePass=${env:TLS_KEYSTORE_PASS}\n"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *nuxeoTLSSuite) AfterTest(_, _ string) {
	// nop
}

// This function runs the NuxeoTLS unit test suite. It is called by 'go test' and will call every
// function in this file with a nuxeoTLSSuite receiver that begins with "Test..."
func TestNuxeoTLSUnitTestSuite(t *testing.T) {
	suite.Run(t, new(nuxeoTLSSuite))
}

// nuxeoTLSSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *nuxeoTLSSuite) nuxeoTLSSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		Spec: v1alpha1.NuxeoSpec{
			NodeSets: []v1alpha1.NodeSet{{
				Name:     suite.deploymentName,
				Replicas: 1,
				NuxeoConfig: v1alpha1.NuxeoConfig{
					TlsSecret:      suite.tlsSecret,
				},
			}},
		},
	}
}

// genTestDeploymentForNuxeoTLSSuite creates and returns a Deployment struct minimally configured to support this suite
func genTestDeploymentForNuxeoTLSSuite() appsv1.Deployment {
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
