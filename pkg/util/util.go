package util

import (
	"fmt"
	"hash/fnv"

	"github.com/davecgh/go-spew/spew"
)

// see IsOpenShift() / SetIsOpenShift()
var isOpenShift = false

func HashObject(object interface{}) string {
	hf := fnv.New32()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	_, _ = printer.Fprintf(hf, "%#v", object)
	return fmt.Sprint(hf.Sum32())
}

func SpewObject(object interface{}) string {
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	return printer.Sdump(object)
}

// Returns true if the operator is running in an OpenShift cluster. Else false = Kubernetes. False
// by default, unless SetIsOpenShift() was called
func IsOpenShift() bool {
	return isOpenShift
}

// Sets operator state indicating that the operator believes it is running in an OpenShift cluster.
func SetIsOpenShift() {
	isOpenShift = true
}

var NuxeoServiceAccountName = "nuxeo"