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
	"k8s.io/apimachinery/pkg/types"
)

// TestBasicDeploymentCreation tests the basic mechanics of creating a new Deployment from the Nuxeo CR spec
// when a Deployment does not already exist
func (suite *nodeSetSuite) TestBasicDeploymentCreation() {
	nux := suite.nodeSetSuiteNewNuxeo()
	requeue, err := suite.r.reconcileNodeSet(nux.Spec.NodeSets[0], nux)
	require.Nil(suite.T(), err, "reconcileNodeSet failed")
	require.Equal(suite.T(), true, requeue, "reconcileNodeSet returned unexpected result")
	found := &appsv1.Deployment{}
	err = suite.r.Get(context.TODO(), types.NamespacedName{Name: deploymentName(nux, nux.Spec.NodeSets[0]),
		Namespace: suite.namespace}, found)
	require.Nil(suite.T(), err, "Deployment creation failed")
	require.Equal(suite.T(), suite.nuxeoContainerName, found.Spec.Template.Spec.Containers[0].Name,
		"Deployment has incorrect container name")
}

// TestDeploymentUpdated creates a Deployment, updates the Nuxeo CR, and verifies the Deployment was updated
func (suite *nodeSetSuite) TestDeploymentUpdated() {
	nux := suite.nodeSetSuiteNewNuxeo()
	_, _ = suite.r.reconcileNodeSet(nux.Spec.NodeSets[0], nux)
	newReplicas := nux.Spec.NodeSets[0].Replicas + 2
	nux.Spec.NodeSets[0].Replicas = newReplicas
	_, _ = suite.r.reconcileNodeSet(nux.Spec.NodeSets[0], nux)
	found := &appsv1.Deployment{}
	_ = suite.r.Get(context.TODO(), types.NamespacedName{Name: deploymentName(nux, nux.Spec.NodeSets[0]),
		Namespace: suite.namespace}, found)
	require.Equal(suite.T(), newReplicas, *found.Spec.Replicas,
		"Deployment has incorrect replica count")
}

// TestDeploymentClustering tests the clustering configuration. If defines clustering as enabled, and also defines
// an inline nuxeo.conf. The operator code under test should create a nuxeo.conf ConfigMap from the inlined
// content and and append to that content specific values for clustering configuration.
func (suite *nodeSetSuite) TestDeploymentClustering() {
	var err error
	nux := suite.nodeSetSuiteNewNuxeoClustered()
	_, _ = suite.r.reconcileNodeSet(nux.Spec.NodeSets[0], nux)
	found := &appsv1.Deployment{}
	_ = suite.r.Get(context.TODO(), types.NamespacedName{Name: deploymentName(nux, nux.Spec.NodeSets[0]),
		Namespace: suite.namespace}, found)
	envCount := 0
	for _, envVar := range found.Spec.Template.Spec.Containers[0].Env {
		switch {
		case envVar.Name == "POD_UID" && envVar.ValueFrom.FieldRef.FieldPath == "metadata.uid":
			envCount += 1
		case envVar.Name == "NUXEO_BINARY_STORE" && envVar.Value == "/var/lib/nuxeo/binaries/binaries":
			envCount += 1
		}
	}
	require.Equal(suite.T(), 2, envCount, "Environment incorrectly defined")
	_, err = suite.r.reconcileNuxeoConf(nux, nux.Spec.NodeSets[0], "", "")
	require.Nil(suite.T(), err, "reconcileNuxeoConf failed")
	foundCMap := &corev1.ConfigMap{}
	cmName := nuxeoConfCMName(nux, nux.Spec.NodeSets[0].Name)
	err = suite.r.Get(context.TODO(), types.NamespacedName{Name: cmName, Namespace: suite.namespace}, foundCMap)
	nuxeoConfExpected := suite.nuxeoConfContent +
		"\nrepository.binary.store=${env:NUXEO_BINARY_STORE}\n" +
		"nuxeo.cluster.enabled=true\n" +
		"nuxeo.cluster.nodeid=${env:POD_UID}\n"
	require.Equal(suite.T(), nuxeoConfExpected, foundCMap.Data["nuxeo.conf"], "nuxeo.conf ConfigMap has incorrect values")
}

// TestDeploymentClusteringNoBinaries validates that when clustering is defined, the configurer also defined
// storage for Nuxeo binaries
func (suite *nodeSetSuite) TestDeploymentClusteringNoBinaries() {
	nux := suite.nodeSetSuiteNewNuxeoClustered()
	nux.Spec.NodeSets[0].Storage = []v1alpha1.NuxeoStorageSpec{}
	_, err := suite.r.reconcileNodeSet(nux.Spec.NodeSets[0], nux)
	require.NotNil(suite.T(), err, "TODO")
}

// TestRevProxyDeploymentCreation is the same as TestBasicDeploymentCreation except it includes an Nginx rev proxy
func (suite *nodeSetSuite) TestRevProxyDeploymentCreation() {
	nux := suite.nodeSetSuiteNewNuxeo()
	nux.Spec.RevProxy = v1alpha1.RevProxySpec{
		Nginx: v1alpha1.NginxRevProxySpec{
			Image:           "foo",
			ImagePullPolicy: corev1.PullAlways,
		},
	}
	_, _ = suite.r.reconcileNodeSet(nux.Spec.NodeSets[0], nux)
	found := &appsv1.Deployment{}
	_ = suite.r.Get(context.TODO(), types.NamespacedName{Name: deploymentName(nux, nux.Spec.NodeSets[0]),
		Namespace: suite.namespace}, found)
	require.Equal(suite.T(), suite.imagePullPolicy, found.Spec.Template.Spec.Containers[1].ImagePullPolicy,
		"Deployment sidecar has incorrect pull policy")
}

// TestSideCar tests explicit definition of a container and a volume. The use case being tested is where someone
// configures their own reverse-proxy container and a volume that contains the configuration of the proxy
func (suite *nodeSetSuite) TestSideCar() {
	nux := suite.nodeSetSuiteNewNuxeo()
	nux.Spec.Containers = []corev1.Container{{
		Name: "test-container",
	}}
	nux.Spec.Volumes = []corev1.Volume{{
		Name: "test-volume",
	}}
	_, _ = suite.r.reconcileNodeSet(nux.Spec.NodeSets[0], nux)
	found := &appsv1.Deployment{}
	_ = suite.r.Get(context.TODO(), types.NamespacedName{Name: deploymentName(nux, nux.Spec.NodeSets[0]),
		Namespace: suite.namespace}, found)
	require.Equal(suite.T(), 2, len(found.Spec.Template.Spec.Containers),
		"Configured sidecar not added")
	require.Equal(suite.T(), 1, len(found.Spec.Template.Spec.Volumes),
		"Configured volume not added")
}

// TestDupVolume tests explicit configuration of a volume whose name clashes with a volume that the operator creates
// to hold the auto-generated nginx configuration
func (suite *nodeSetSuite) TestDupVolume() {
	nux := suite.nodeSetSuiteNewNuxeo()
	nux.Spec.RevProxy = v1alpha1.RevProxySpec{
		Nginx: v1alpha1.NginxRevProxySpec{
			Image: "nginx",
		},
	}
	// clashes with vol generated by the operator:
	nux.Spec.Volumes = []corev1.Volume{{
		Name: "nginx-conf",
	}}
	_, err := suite.r.reconcileNodeSet(nux.Spec.NodeSets[0], nux)
	require.NotNil(suite.T(), err, "reconcileNodeSet should have detected dup volume")
}

// nodeSetSuite is the NodeSet test suite structure
type nodeSetSuite struct {
	suite.Suite
	r                  NuxeoReconciler
	nuxeoName          string
	deploymentName     string
	namespace          string
	nuxeoContainerName string
	imagePullPolicy    corev1.PullPolicy
	nuxeoConfContent   string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *nodeSetSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
	suite.deploymentName = "testclust"
	suite.nuxeoContainerName = "nuxeo"
	suite.imagePullPolicy = corev1.PullAlways
	suite.nuxeoConfContent = "test.property.one=TESTONE\ntest.property.two=TESTTWO"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *nodeSetSuite) AfterTest(_, _ string) {
	obj := appsv1.Deployment{}
	_ = suite.r.DeleteAllOf(context.TODO(), &obj)
}

// This function runs the NodeSet unit test suite. It is called by 'go test' and will call every
// function in this file with a nodeSetSuite receiver that begins with "Test..."
func TestNodeSetUnitTestSuite(t *testing.T) {
	suite.Run(t, new(nodeSetSuite))
}

// nodeSetSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *nodeSetSuite) nodeSetSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		Spec: v1alpha1.NuxeoSpec{
			NodeSets: []v1alpha1.NodeSet{{
				Name:        suite.deploymentName,
				Interactive: true,
				Replicas:    3,
			}},
		},
	}
}

// nodeSetSuiteNewNuxeo creates a clustered test Nuxeo struct suitable for the test cases in this suite.
func (suite *nodeSetSuite) nodeSetSuiteNewNuxeoClustered() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		Spec: v1alpha1.NuxeoSpec{
			NodeSets: []v1alpha1.NodeSet{{
				Name:           suite.deploymentName,
				Replicas:       3,
				ClusterEnabled: true,
				NuxeoConfig: v1alpha1.NuxeoConfig{
					NuxeoConf: v1alpha1.NuxeoConfigSetting{
						Inline: suite.nuxeoConfContent,
					},
				},
				Storage: []v1alpha1.NuxeoStorageSpec{{
					StorageType:  v1alpha1.NuxeoStorageBinaries,
					Size:         "10M",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				}},
			}},
		},
	}
}
