package e2e

import (
	"testing"

	f "github.com/operator-framework/operator-sdk/pkg/test"
	_ "nuxeo-operator/pkg/apis"
)

// TestMain must be run using the operator-sdk. E.g.:
//  $ operator-sdk test local ./test/e2e --operator-namespace operator-test
// The e2e test has a few pre-requisites. These are documented in the README. If the pre-requisites are not
// provided, the e2e test will fail.
func TestMain(m *testing.M) {
	f.MainEntry(m)
}
