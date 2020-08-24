package nuxeo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Performs a basic PVC test. Defines a Nuxeo CR with storage configuration that should result in two PVCs.
// Creates an orphaned PVC. Ensures that the two PVCs were created and the orphan was removed.
func (suite *persistentVolumeClaimSuite) TestBasicPVC() {
	var err error
	nux := suite.persistentVolumeClaimSuiteNewNuxeo()
	err = createOrphanPVC(nux, suite.orphanPVCName, suite.r)
	require.Nil(suite.T(), err, "Error creating orphaned")
	err = suite.r.reconcilePvc(nux)
	require.Nil(suite.T(), err, "reconcilePvc failed")
	var pvcsInCluster corev1.PersistentVolumeClaimList
	opts := []client.ListOption{
		client.InNamespace(nux.Namespace),
	}
	err = suite.r.client.List(context.TODO(), &pvcsInCluster, opts...)
	require.Nil(suite.T(), err, "Error getting PVCs")
	require.Equal(suite.T(), 2, len(pvcsInCluster.Items), "Failed to obtain the expected number of PVCs")
	for _, pvc := range pvcsInCluster.Items {
		require.NotEqual(suite.T(), suite.orphanPVCName, pvc.Name, "Orphaned PVC was not removed")
	}
}

// persistentVolumeClaimSuite is the PersistentVolumeClaim test suite structure
type persistentVolumeClaimSuite struct {
	suite.Suite
	r             ReconcileNuxeo
	nuxeoName     string
	namespace     string
	orphanPVCName string
}

// SetupSuite initializes the Fake client, a ReconcileNuxeo struct, and various test suite constants
func (suite *persistentVolumeClaimSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
	suite.orphanPVCName = "zzz"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *persistentVolumeClaimSuite) AfterTest(_, _ string) {
	obj := corev1.PersistentVolumeClaim{}
	_ = suite.r.client.DeleteAllOf(context.TODO(), &obj)
}

// This function runs the PersistentVolumeClaim unit test suite. It is called by 'go test' and will call every
// function in this file with a persistentVolumeClaimSuite receiver that begins with "Test..."
func TestPersistentVolumeClaimUnitTestSuite(t *testing.T) {
	suite.Run(t, new(persistentVolumeClaimSuite))
}

// persistentVolumeClaimSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *persistentVolumeClaimSuite) persistentVolumeClaimSuiteNewNuxeo() *v1alpha1.Nuxeo {
	testStorageClass := "foo-storage-class"
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
			UID:       "bdd9861e-4ad8-4681-a93c-1f48cd49a01e",
		},
		// whatever else is needed for the suite
		Spec: v1alpha1.NuxeoSpec{
			NodeSets: []v1alpha1.NodeSet{{
				Name:     "test",
				Replicas: 1,
				Storage: []v1alpha1.NuxeoStorageSpec{{
					// should result in the creation of a default PVC
					StorageType: v1alpha1.NuxeoStorageBinaries,
					Size:        "10M",
				}, {
					// should not result in the creation of a PVC because volume source explicitly provided
					StorageType: v1alpha1.NuxeoStorageTransientStore,
					Size:        "1Gi",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				}, {
					// should result in the creation of an explicit PVC matching this spec
					StorageType: v1alpha1.NuxeoStorageData,
					VolumeClaimTemplate: corev1.PersistentVolumeClaim{
						ObjectMeta: metav1.ObjectMeta{
							Name: "explicit-pvc",
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
							Resources: corev1.ResourceRequirements{
								Limits: nil,
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: resource.MustParse("10M"),
								},
							},
							StorageClassName: &testStorageClass,
						},
					},
				}},
			}},
		},
	}
}

// Creates an "orphaned" PVC that doesn't match any of the PVCs in the test deployment. This mimics a case where
// a storage spec was defined in the Nuxeo CR, the reconciliation loop fired, a PVC was created, then the
// storage spec was removed. In that case we want the reconciler to remove any associated PVC that was created
// in a prior reconciliation loop (when the storage configuration was present.)
func createOrphanPVC(instance *v1alpha1.Nuxeo, pvcName string, r ReconcileNuxeo) error {
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: instance.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1M"),
				},
			},
		},
	}
	_ = controllerutil.SetControllerReference(instance, &pvc, r.scheme)
	if err := r.client.Create(context.TODO(), &pvc); err != nil {
		return err
	}
	return nil
}
