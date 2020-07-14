package util

import (
	"reflect"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Compares expected Data to found Data in two secrets. If same, returns true, otherwise updates found.Data from
// expected.Data and returns false. False means found must be written back to the cluster. This is intended to be
// used on Secrets that have been obtained from the cluster, meaning the entire contents of the secret are in
// the Data stanza. (The function ignores the StringData field)
func SecretCompare(expected runtime.Object, found runtime.Object) bool {
	if !reflect.DeepEqual(expected.(*v1.Secret).Data, found.(*v1.Secret).Data) {
		found.(*v1.Secret).Data = expected.(*v1.Secret).Data
		return false
	}
	return true
}

//func ConfigMapCompareKey - compare one key in a configmap
