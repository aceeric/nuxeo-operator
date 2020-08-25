package nuxeo

import (
	"context"
	"testing"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Tests the basic functionality to create a nuxeo.conf ConfigMap from inlined data in the Nuxeo CR. Defines
// a Nuxeo CR with inline nuxeo.conf and calls the config map reconciliation. Verifies that a ConfigMap was
// created that contains the matching nuxeo.conf content.
func (suite *nuxeoConfSuite) TestBasicInlineNuxeoConf() {
	nux := suite.nuxeoConfSuiteNewNuxeo()
	err := suite.r.reconcileNuxeoConf(nux, nux.Spec.NodeSets[0], "", "")
	require.Nil(suite.T(), err, "reconcileNuxeoConf failed")
	found := &corev1.ConfigMap{}
	cmName := nuxeoConfCMName(nux, nux.Spec.NodeSets[0].Name)
	err = suite.r.Get(context.TODO(), types.NamespacedName{Name: cmName, Namespace: suite.namespace}, found)
	require.Nil(suite.T(), err, "Nuxeo conf ConfigMap creation failed")
	require.Equal(suite.T(), suite.nuxeoConfContent, found.Data[suite.nuxeoConfKey],
		"ConfigMap has incorrect nuxeo.conf content")
}

// joinCompact is used to build the combined nuxeo.conf ConfigMap
func (suite *nuxeoConfSuite) TestJoinCompact() {
	s0 := ""
	s1 := "Good morning,\n"
	s2 := "\n\nand in case I don't see ya,\n\n\n\n"
	s3 := "Good afternoon,\n\n\n\n"
	s4 := "\n\n\n\ngood evening,"
	s5 := "and"
	s6 := "goodnight!\n\n\n\n"
	act := joinCompact("\n", s0, s1, s2, s3, s4, s5, s6)
	exp := "Good morning,\nand in case I don't see ya,\nGood afternoon,\ngood evening,\nand\ngoodnight!\n"
	require.Equal(suite.T(), exp, act, "joinCompact Failed")
}

// TestExternalNuxeoConf tests when an external source for nuxeo conf is defined by the configurer and another
// Nuxeo CR setting is configured that requires the operator to own nuxeo.conf - these are incompatible together
func (suite *nuxeoConfSuite) TestExternalNuxeoConf() {
	nux := suite.nuxeoConfSuiteNewNuxeo()
	// this indicates that the operator should own a nuxeo.conf ConfigMap
	nux.Spec.NodeSets[0].ClusterEnabled = true
	// this indicates that the configurer is mounting an externally provisioned nuxeo.conf source (in this
	// case a dummy empty dir) which is in conflict with operator-managed nuxeo.conf above
	nux.Spec.NodeSets[0].NuxeoConfig.NuxeoConf = v1alpha1.NuxeoConfigSetting{
		ValueFrom: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium: corev1.StorageMediumDefault,
			},
		},
	}
	err := configureNuxeoConf(nux, nil, nux.Spec.NodeSets[0], "", "")
	require.NotNil(suite.T(), err, "nuxeo.conf conflict not detected")
}

// nuxeoConfSuite is the NuxeoConf test suite structure
type nuxeoConfSuite struct {
	suite.Suite
	r                NuxeoReconciler
	nuxeoName        string
	deploymentName   string
	namespace        string
	nuxeoConfContent string
	nuxeoConfKey     string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *nuxeoConfSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.deploymentName = "testclust"
	suite.namespace = "testns"
	suite.nuxeoConfContent = "test.test.test=100\n"
	suite.nuxeoConfKey = "nuxeo.conf"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *nuxeoConfSuite) AfterTest(_, _ string) {
	obj := corev1.ConfigMap{}
	_ = suite.r.DeleteAllOf(context.TODO(), &obj)
}

// This function runs the NuxeoConf unit test suite. It is called by 'go test' and will call every
// function in this file with a nuxeoConfSuite receiver that begins with "Test..."
func TestNuxeoConfUnitTestSuite(t *testing.T) {
	suite.Run(t, new(nuxeoConfSuite))
}

// nuxeoConfSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *nuxeoConfSuite) nuxeoConfSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		Spec: v1alpha1.NuxeoSpec{
			NodeSets: []v1alpha1.NodeSet{{
				Name:     suite.deploymentName,
				Replicas: 1,
				NuxeoConfig: v1alpha1.NuxeoConfig{
					NuxeoConf: v1alpha1.NuxeoConfigSetting{
						// inline
						Inline: suite.nuxeoConfContent,
					},
				},
			}},
		},
	}
}
