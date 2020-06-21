package e2e

import (
	"flag"
	"testing"

	f "github.com/operator-framework/operator-sdk/pkg/test"
	_ "nuxeo-operator/pkg/apis"
)

// TestMain must be run using the operator-sdk. E.g.:
//  $ operator-sdk test local ./test/e2e --operator-namespace operator-test
// The e2e test has a few pre-requisites. These are documented in the README. If the pre-requisites are not
// provided, the e2e test will fail. The tests expect to receive a nuxeo image arg, e.g.:
// '--nuxeo-image=localhost:32000/images/nuxeo:10.10'
type testArgs struct {
	nuxeoImageName *string
}

var args = &testArgs{}

func TestMain(m *testing.M) {
	args.nuxeoImageName = flag.String("nuxeo-image", "", "Nuxeo Image name")
	f.MainEntry(m)
}
