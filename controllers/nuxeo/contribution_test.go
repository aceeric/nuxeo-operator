package nuxeo

import (
	"context"
	"testing"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestSecretConfigMapContribution performs a very basic contribution test. It configures one ConfigMap contribution
// and one Secret contribution and confirms that this resulted in two volumes and two volume mounts. It also verifies
// the the nuxeo container's 'NUXEO_TEMPLATES' environment variable was updated to reference the two contributions
// via absolute paths - with the paths matching how the operator configures the volume mounts.
func (suite *contributionSuite) TestSecretConfigMapContribution() {
	var err error
	nux := suite.contributionSuiteNewNuxeo()
	err = suite.genConfigMapForContrib()
	require.Nil(suite.T(), err, "genConfigMapForContrib failed")
	err = suite.genSecretForContrib()
	require.Nil(suite.T(), err, "genSecretForContrib failed")
	dep := genTestDeploymentForContributionSuite()
	err = suite.r.configureContributions(nux, &dep, nux.Spec.NodeSets[0])
	require.Nil(suite.T(), err, "configureConfig failed")
	require.Equal(suite.T(), 2, len(dep.Spec.Template.Spec.Volumes),
		"incorrect volume configuration")
	require.Equal(suite.T(), 2, len(dep.Spec.Template.Spec.Containers[0].VolumeMounts),
		"incorrect volume mount configuration")
	expectedTemplates := "test,/etc/nuxeo/nuxeo-operator-config/" + suite.cmContribName +
		",/etc/nuxeo/nuxeo-operator-config/" + suite.secretContribName
	require.Equal(suite.T(), expectedTemplates, dep.Spec.Template.Spec.Containers[0].Env[0].Value,
		"Templates incorrectly added to NUXEO_TEMPLATES env var")
}

// Same as TestSecretConfigMapContribution except the contribution comes from a PVC
func (suite *contributionSuite) TestPVCContribution() {
	var err error
	nux := suite.contributionSuitePvcNewNuxeo()
	dep := genTestDeploymentForContributionSuite()
	err = suite.r.configureContributions(nux, &dep, nux.Spec.NodeSets[0])
	require.Nil(suite.T(), err, "configureConfig failed")
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Volumes),
		"incorrect volume configuration")
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Containers[0].VolumeMounts),
		"incorrect volume mount configuration")
	expectedTemplates := "test,/etc/nuxeo/nuxeo-operator-config/" + suite.pvcContribName
	require.Equal(suite.T(), expectedTemplates, dep.Spec.Template.Spec.Containers[0].Env[0].Value,
		"Templates incorrectly added to NUXEO_TEMPLATES env var")
}

// contributionSuite is the Contribution test suite structure
type contributionSuite struct {
	suite.Suite
	r                 NuxeoReconciler
	nuxeoName         string
	namespace         string
	deploymentName    string
	cmContribName     string
	secretContribName string
	pvcContribName    string
	configMapName     string
	secretName        string
	pvName            string
	pvcName           string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *contributionSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
	suite.deploymentName = "testclust"
	suite.cmContribName = "test-contrib-from-cm"
	suite.secretContribName = "test-contrib-from-secret"
	suite.pvcContribName = "test-contrib-from-pvc"
	suite.configMapName = "my-cm"
	suite.secretName = "my-secret"
	suite.pvName = "test-pv"
	suite.pvcName = "test-pvc"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *contributionSuite) AfterTest(_, _ string) {
	obj := corev1.ConfigMap{}
	_ = suite.r.DeleteAllOf(context.TODO(), &obj)
	objSecret := corev1.Secret{}
	_ = suite.r.DeleteAllOf(context.TODO(), &objSecret)
	objPv := corev1.PersistentVolume{}
	_ = suite.r.DeleteAllOf(context.TODO(), &objPv)
	objPvc := corev1.PersistentVolumeClaim{}
	_ = suite.r.DeleteAllOf(context.TODO(), &objPvc)
}

// This function runs the Contribution unit test suite. It is called by 'go test' and will call every
// function in this file with a contributionSuite receiver that begins with "Test..."
func TestContributionUnitTestSuite(t *testing.T) {
	suite.Run(t, new(contributionSuite))
}

// contributionSuiteNewNuxeo creates a test Nuxeo struct with a Secret and a ConfigMap as two
// contribution sources
func (suite *contributionSuite) contributionSuiteNewNuxeo() *v1alpha1.Nuxeo {
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
				Contributions: []v1alpha1.Contribution{{
					Templates: []string{suite.cmContribName},
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: suite.configMapName},
						},
					},
				}, {
					Templates: []string{suite.secretContribName},
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: suite.secretName,
						},
					},
				}},
			}},
		},
	}
}

// contributionSuiteNewNuxeo creates a test Nuxeo struct with a PVC as the contribution source
func (suite *contributionSuite) contributionSuitePvcNewNuxeo() *v1alpha1.Nuxeo {
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
				Contributions: []v1alpha1.Contribution{{
					Templates: []string{suite.pvcContribName},
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: suite.pvcName,
							ReadOnly:  true,
						},
					},
				}},
			}},
		},
	}
}

// genConfigMapForContrib generates a ConfigMap containing a contribution
func (suite *contributionSuite) genConfigMapForContrib() error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.configMapName,
			Namespace: suite.namespace,
		},
		Data: map[string]string{
			"nuxeo.defaults": "test.test.test=100",
			"contrib-1.xml":  "TEST-TEST-TEST",
			"contrib-2.xml":  "TEST2-TEST2-TEST2",
		},
	}
	return suite.r.Create(context.TODO(), cm)
}

// genConfigMapForContrib generates a Secret containing a contribution
func (suite *contributionSuite) genSecretForContrib() error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.secretName,
			Namespace: suite.namespace,
		},
		Data: map[string][]byte{
			"nuxeo.defaults":       []byte("secret.secret.secret=100"),
			"secret-contrib-1.xml": []byte("SECRET-SECRET-SECRET"),
			"secret-contrib-2.xml": []byte("SECRET2-SECRET2-SECRET2"),
		},
	}
	return suite.r.Create(context.TODO(), secret)
}

// genTestDeploymentForConfigSuite creates and returns a Deployment struct minimally
// configured to support this suite
func genTestDeploymentForContributionSuite() appsv1.Deployment {
	replicas := int32(1)
	dep := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: NuxeoServiceAccountName,
					Containers: []corev1.Container{{
						Name: "nuxeo",
						Env: []corev1.EnvVar{{
							Name:  "NUXEO_TEMPLATES",
							Value: "test",
						}},
					}},
				},
			},
		},
	}
	return dep
}
