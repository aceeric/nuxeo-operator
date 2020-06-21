package util

import (
	"crypto/md5"
	"fmt"
	"hash/fnv"

	"github.com/davecgh/go-spew/spew"
	"github.com/ghodss/yaml"
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
// by default, unless SetIsOpenShift() was called prior to this call
func IsOpenShift() bool {
	return isOpenShift
}

// Sets operator state indicating that the operator believes it is running in an OpenShift cluster.
func SetIsOpenShift() {
	isOpenShift = true
}

var NuxeoServiceAccountName = "nuxeo"

// ObjectsDiffer generates a YAML from each passed object then generates an MD5 sum of each YAML and returns
// true if the MD5 sums differ, and false if the MD5 sums are the same. If the YAML generation fails, then the
// resulting error is returned, otherwise a nil error is returned. This function is intended for comparing
// CR specs. The underlying assumption is that any difference is a spec is actionable for the operator. So this
// handles two cases: 1) the Nuxeo CR is modified, and a dependent CR should look different as a result. 2) A
// cluster CR owned by the Nuxeo CR is modified independently of the Operator and is therefore out of sync
// with how the Operator would expect it to look.
func ObjectsDiffer(expected interface{}, actual interface{}) (bool, error) {
	var expMd5, actMd5 [md5.Size]byte
	var err error
	var bytes []byte

	if bytes, err = yaml.Marshal(expected); err != nil{
		return false, err
	}
	expMd5 = md5.Sum(bytes)
	if bytes, err = yaml.Marshal(actual); err != nil{
		return false, err
	}
	actMd5 = md5.Sum(bytes)
	return expMd5 != actMd5, nil
}