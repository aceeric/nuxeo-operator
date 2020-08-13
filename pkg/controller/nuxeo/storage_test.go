package nuxeo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
)

// Tests basic storage configuration. Creates a Nuxeo spec with NuxeoStorageData, NuxeoStorageBinaries
// and TransientStores defined, and verifies that these correctly result in volumes, volume mounts, and
// environment variables in the Nuxeo deployment.
func (suite *nuxeoStorageSpecSuite) TestBasicNuxeoStorage() {
	nux := suite.nuxeoStorageSpecSuiteNewNuxeo()
	dep := genTestDeploymentForStorageSuite()
	err := configureStorage(&dep, nux.Spec.NodeSets[0])
	require.Nil(suite.T(), err, "configureStorage failed")
	require.Equal(suite.T(), 3, len(dep.Spec.Template.Spec.Volumes), "Volumes were not created")
	require.Equal(suite.T(), 3, len(dep.Spec.Template.Spec.Containers[0].VolumeMounts),
		"Volume mounts were not created")
	require.Equal(suite.T(), 2, len(dep.Spec.Template.Spec.Containers[0].Env),
		"Environment variables were not created")
}

// nuxeoStorageSpecSuite is the NuxeoStorageSpec test suite structure
type nuxeoStorageSpecSuite struct {
	suite.Suite
	r              ReconcileNuxeo
	nuxeoName      string
	deploymentName string
	namespace      string
}

// SetupSuite initializes the Fake client, a ReconcileNuxeo struct, and various test suite constants
func (suite *nuxeoStorageSpecSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
	suite.deploymentName = "testclust"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *nuxeoStorageSpecSuite) AfterTest(_, _ string) {
	dep := appsv1.Deployment{}
	_ = suite.r.client.DeleteAllOf(context.TODO(), &dep)
}

// This function runs the NuxeoStorageSpec unit test suite. It is called by 'go test' and will call every
// function in this file with a nuxeoStorageSpecSuite receiver that begins with "Test..."
func TestNuxeoStorageSpecUnitTestSuite(t *testing.T) {
	suite.Run(t, new(nuxeoStorageSpecSuite))
}

// nuxeoStorageSpecSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *nuxeoStorageSpecSuite) nuxeoStorageSpecSuiteNewNuxeo() *v1alpha1.Nuxeo {
	testStorageClass := "certificatesToPEM-storage-class"
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
				Storage: []v1alpha1.NuxeoStorageSpec{{
					StorageType: v1alpha1.NuxeoStorageBinaries,
					Size:        "10M",
				}, {
					StorageType: v1alpha1.NuxeoStorageTransientStore,
					Size:        "1Gi",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				}, {
					StorageType: v1alpha1.NuxeoStorageData,
					VolumeClaimTemplate: corev1.PersistentVolumeClaim{
						ObjectMeta: metav1.ObjectMeta{
							Name: "explicit-pvc",
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
							Resources:        corev1.ResourceRequirements{},
							StorageClassName: &testStorageClass,
						},
					},
				}},
			}},
		},
	}
}

// genTestDeploymentForStorageSuite creates and returns a Deployment struct minimally configured to support this suite
func genTestDeploymentForStorageSuite() appsv1.Deployment {
	replicas := int32(1)
	dep := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: util.NuxeoServiceAccountName,
					Containers: []corev1.Container{{
						Image:           "test",
						ImagePullPolicy: corev1.PullAlways,
						Name:            "nuxeo",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
						}},
					}},
				},
			},
		},
	}
	return dep
}
