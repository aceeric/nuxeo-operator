package nuxeo

/*
this is a template for a unit test file for the FooType. Paste into the target
directory. E.g.: if testing a ConfigMap then replace all 'fooType' with 'configMap' and
all 'FooType' with 'ConfigMap'

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

func (suite *fooTypeSuite) TestFooTypeAlwaysPasses() {}

// fooTypeSuite is the FooType test suite structure
type fooTypeSuite struct {
	suite.Suite
	r         NuxeoReconciler
	nuxeoName string
	namespace string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *fooTypeSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *fooTypeSuite) AfterTest(_, _ string) {
	//obj := the type of object you are testing. E.g.: obj := corev1.Service{}
	//_ = suite.r.client.DeleteAllOf(context.TODO(), &obj)
}


// This function runs the FooType unit test suite. It is called by 'go test' and will call every
// function in this file with a fooTypeSuite receiver that begins with "Test..."
func TestFooTypeUnitTestSuite(t *testing.T) {
	suite.Run(t, new(fooTypeSuite))
}

// fooTypeSuiteNewNuxeo creates a test Nuxeo struct suitable for the test cases in this suite.
func (suite *fooTypeSuite) fooTypeSuiteNewNuxeo() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
        // whatever else is needed for the suite
		Spec: v1alpha1.NuxeoSpec{
		},
	}
}
*/