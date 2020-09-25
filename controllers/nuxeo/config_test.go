/*
Copyright 2020 Eric Ace.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nuxeo

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Tests basic configuration functionality. Defines a Nuxeo CR with all available 'NuxeoConfig' options
// configured. Those that result in environment variables in the container are verified. An inline nuxeo.conf
// is specified which should result in the creation of a volume mount and a volume with a ConfigMap volume
// reference to a hard-coded CM name. This is also verified. (The actual config map reconciliation is tested
// in the nuxeo_conf_test.go file.)
func (suite *nuxeoConfigSuite) TestBasicConfig() {
	nux := suite.nuxeoConfigSuiteNewNuxeo()
	dep := genTestDeploymentForConfigSuite()
	sec := genTestJvmPkiSecret()
	err := configureConfig(&dep, nux.Spec.NodeSets[0], sec)
	require.Nil(suite.T(), err, "configureConfig failed")
	validActualEnvCnt := 0
	for _, env := range dep.Spec.Template.Spec.Containers[0].Env {
		switch {
		case env.Name == "JAVA_OPTS" && strings.Contains(env.Value, suite.javaOpts) &&
			strings.Contains(env.Value, "-Djavax.net.ssl."):
			validActualEnvCnt += 1
		case env.Name == "NUXEO_TEMPLATES" && env.Value == strings.Join(suite.nuxeoTemplates, ","):
			validActualEnvCnt += 1
		case env.Name == "NUXEO_PACKAGES" && env.Value == strings.Join(suite.nuxeoPackages, " "):
			validActualEnvCnt += 1
		case env.Name == "NUXEO_URL" && env.Value == suite.nuxeoUrl:
			validActualEnvCnt += 1
		case env.Name == "NUXEO_ENV_NAME" && env.Value == suite.nuxeoEnvName:
			validActualEnvCnt += 1
		}
	}
	require.Equal(suite.T(), 5, validActualEnvCnt,
		"Configuration environment variables were not created correctly")
	require.Equal(suite.T(), 3, len(dep.Spec.Template.Spec.Containers[0].VolumeMounts),
		"Volume Mounts not correctly defined")
	require.Equal(suite.T(), 3, len(dep.Spec.Template.Spec.Volumes),
		"Volumes not correctly defined")
}

// nuxeoConfigSuite is the NuxeoConfig test suite structure
type nuxeoConfigSuite struct {
	suite.Suite
	r                NuxeoReconciler
	nuxeoName        string
	namespace        string
	deploymentName   string
	javaOpts         string
	nuxeoTemplates   []string
	nuxeoPackages    []string
	nuxeoEnvName     string
	nuxeoUrl         string
	nuxeoConfContent string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *nuxeoConfigSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
	suite.deploymentName = "testclust"
	suite.javaOpts = "-Xms8m"
	suite.nuxeoTemplates = []string{"template1", "template2"}
	suite.nuxeoPackages = []string{"package1", "package2"}
	suite.nuxeoEnvName = "thisnuxeo"
	suite.nuxeoUrl = "http://nuxeo.mycorp.io"
	suite.nuxeoConfContent = "this.is.a=test"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *nuxeoConfigSuite) AfterTest(_, _ string) {
	// nop
}

// This function runs the NuxeoConfig unit test suite. It is called by 'go test' and will call every
// function in this file with a nuxeoConfigSuite receiver that begins with "Test..."
func TestNuxeoConfigUnitTestSuite(t *testing.T) {
	suite.Run(t, new(nuxeoConfigSuite))
}

// nuxeoConfigSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *nuxeoConfigSuite) nuxeoConfigSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		// whatever else is needed for the suite
		Spec: v1alpha1.NuxeoSpec{
			NodeSets: []v1alpha1.NodeSet{{
				Name:     suite.deploymentName,
				Replicas: 1,
				NuxeoConfig: v1alpha1.NuxeoConfig{
					JavaOpts:       suite.javaOpts,
					NuxeoTemplates: suite.nuxeoTemplates,
					NuxeoPackages:  suite.nuxeoPackages,
					NuxeoUrl:       suite.nuxeoUrl,
					NuxeoName:      suite.nuxeoEnvName,
					NuxeoConf: v1alpha1.NuxeoConfigSetting{
						Inline: suite.nuxeoConfContent,
					},
					// this is ignored by the unit test because the unit test tests at a lower layer than this is
					// looked at by the operator but it seems better to init the struct the way it would actually
					// be used
					JvmPKISecret: "jvm-pki-secret",

					OfflinePackages: []v1alpha1.OfflinePackage{{
						PackageName: "test-package1-4.5.6.zip",
						ValueFrom: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "test-package1-configmap"},
						}},
					}, {
						PackageName: "test-package2-1.2.3.zip",
						ValueFrom: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
							SecretName: "test-package2-secret",
						}},
					}},
				},
			}},
		},
	}
}

// genTestDeploymentForConfigSuite creates and returns a Deployment struct minimally
// configured to support this suite
func genTestDeploymentForConfigSuite() appsv1.Deployment {
	replicas := int32(1)
	dep := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: NuxeoServiceAccountName,
					Containers: []corev1.Container{{
						Image:           "test",
						ImagePullPolicy: corev1.PullAlways,
						Name:            "nuxeo",
					}},
				},
			},
		},
	}
	return dep
}

// getTestJvmPkiSecret creates and returns a Secret with all six fields supported by the Operator for
// configuring the JVM KeyStore/TrustStore. It simulates an existing secret referenced by the NodeSet.
// NuxeoConfig.JvmPKISecret property
func genTestJvmPkiSecret() corev1.Secret {
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "jvm-pki-secret",
		},
		Data: map[string][]byte{},
	}
	storeType, storePass := "jks", "frobozz"
	storeTypeEncoded := base64.StdEncoding.EncodeToString([]byte(storeType))
	storePassEncoded := base64.StdEncoding.EncodeToString([]byte(storePass))
	secret.Data["keyStore"] = []byte{}
	secret.Data["keyStoreType"] = []byte(storeTypeEncoded)
	secret.Data["keyStorePassword"] = []byte(storePassEncoded)
	secret.Data["trustStore"] = []byte{}
	secret.Data["trustStoreType"] = []byte(storeTypeEncoded)
	secret.Data["trustStorePassword"] = []byte(storePassEncoded)
	return secret
}
