package util

import (
	"fmt"
	"hash/fnv"

	"github.com/davecgh/go-spew/spew"
)

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

// future: detect Kubernetes vs OpenShift
func IsOpenShift() bool {
	return true
}

var NuxeoServiceAccountName = "nuxeo"