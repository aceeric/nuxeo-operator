package nuxeo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TestBasicDeploymentCreation tests the basic mechanics of creating a new Deployment from the Nuxeo CR spec
// when a Deployment does not already exist
func (suite *nodeSetSuite) TestBasicDeploymentCreation() {
	nux := suite.nodeSetSuiteNewNuxeo()
	result, err := reconcileNodeSet(&suite.r, nux.Spec.NodeSets[0], nux, nux.Spec.RevProxy, log)
	require.Nil(suite.T(), err, "reconcileNodeSet failed with err: %v\n", err)
	require.Equal(suite.T(), reconcile.Result{Requeue: true}, result,
		"reconcileNodeSet returned unexpected result: %v\n", result)
	found := &appsv1.Deployment{}
	err = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: deploymentName(nux, nux.Spec.NodeSets[0]),
		Namespace: suite.namespace}, found)
	require.Nil(suite.T(), err, "Deployment creation failed with err: %v\n", err)
	require.Equal(suite.T(), suite.nuxeoContainerName, found.Spec.Template.Spec.Containers[0].Name,
		"Deployment has incorrect container name: %v\n", found.Spec.Template.Spec.Containers[0].Name)
}

// TestDeploymentUpdated creates a Deployment, updates the Nuxeo CR, and verifies the Deployment was updated
func (suite *nodeSetSuite) TestDeploymentUpdated() {
	nux := suite.nodeSetSuiteNewNuxeo()
	_, _ = reconcileNodeSet(&suite.r, nux.Spec.NodeSets[0], nux, nux.Spec.RevProxy, log)
	newReplicas := nux.Spec.NodeSets[0].Replicas + 2
	nux.Spec.NodeSets[0].Replicas = newReplicas
	_, _ = reconcileNodeSet(&suite.r, nux.Spec.NodeSets[0], nux, nux.Spec.RevProxy, log)
	found := &appsv1.Deployment{}
	_ = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: deploymentName(nux, nux.Spec.NodeSets[0]),
		Namespace: suite.namespace}, found)
	require.Equal(suite.T(), newReplicas, *found.Spec.Replicas,
		"Deployment has incorrect replica count: %v\n", *found.Spec.Replicas)
}

// TestDeploymentClustering tests the clustering configuration. If defines clustering as enabled, and also defines
// and inline nuxeo.conf. The operator code under test should create a nuxeo.conf ConfigMap from the inlined
// content and and append to that content specific values for clustering configuration.
func (suite *nodeSetSuite) TestDeploymentClustering() {
	var err error
	nux := suite.nodeSetSuiteNewNuxeoClustered()
	_, _ = reconcileNodeSet(&suite.r, nux.Spec.NodeSets[0], nux, nux.Spec.RevProxy, log)
	found := &appsv1.Deployment{}
	_ = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: deploymentName(nux, nux.Spec.NodeSets[0]),
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
	require.Equal(suite.T(), 2, envCount, "Environment incorrectly defined\n")
	_, err = reconcileNuxeoConf(&suite.r, nux, nux.Spec.NodeSets[0], log)
	require.Nil(suite.T(), err, "reconcileNuxeoConf failed with err: %v\n", err)
	foundCMap := &corev1.ConfigMap{}
	cmName := nux.Name + "-" + nux.Spec.NodeSets[0].Name + "-nuxeo-conf"
	err = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: cmName, Namespace: suite.namespace}, foundCMap)
	nuxeoConfExpected := suite.nuxeoConfContent +
		"\nrepository.binary.store=${env:NUXEO_BINARY_STORE}\n" +
		"nuxeo.cluster.enabled=true\n" +
		"nuxeo.cluster.nodeid=${env:POD_UID}\n"
	require.Equal(suite.T(), nuxeoConfExpected, foundCMap.Data["nuxeo.conf"],"nuxeo.conf ConfigMap has incorrect values: %v\n",
		foundCMap.Data)
}

func (suite *nodeSetSuite) TestDeploymentClusteringnoBinaries() {
	// todo-me test clustering when binaries not defined - should err
}

// TestInteractiveChanged tests when a nodeset is changed from interactive true to false and vice versa
func (suite *nodeSetSuite) TestInteractiveChanged() {
	// todo-me code this test
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
	_, _ = reconcileNodeSet(&suite.r, nux.Spec.NodeSets[0], nux, nux.Spec.RevProxy, log)
	found := &appsv1.Deployment{}
	_ = suite.r.client.Get(context.TODO(), types.NamespacedName{Name: deploymentName(nux, nux.Spec.NodeSets[0]),
		Namespace: suite.namespace}, found)
	require.Equal(suite.T(), suite.imagePullPolicy, found.Spec.Template.Spec.Containers[1].ImagePullPolicy,
		"Deployment sidecar has incorrect pull policy: %v\n", found.Spec.Template.Spec.Containers[1].ImagePullPolicy)
}

// nodeSetSuite is the NodeSet test suite structure
type nodeSetSuite struct {
	suite.Suite
	r                  ReconcileNuxeo
	nuxeoName          string
	deploymentName     string
	namespace          string
	nuxeoContainerName string
	imagePullPolicy    corev1.PullPolicy
	nuxeoConfContent   string
}

// SetupSuite initializes the Fake client, a ReconcileNuxeo struct, and various test suite constants
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
	_ = suite.r.client.DeleteAllOf(context.TODO(), &obj)
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
				Name:     suite.deploymentName,
				Replicas: 3,
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
						Value: suite.nuxeoConfContent,
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
