package util

import (
	"bytes"
	"crypto/md5"
	goerrors "errors"

	"github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/jsonpath"
)

type clusterType int

const (
	openShift  clusterType = 1
	kubernetes clusterType = 2
)

var cluster = kubernetes

// Returns true if the operator is running in an OpenShift cluster. Else false = Kubernetes. False
// by default, unless SetIsOpenShift() was called prior to this call
func IsOpenShift() bool {
	return cluster == openShift
}

// Sets operator state indicating that the operator believes it is running in an OpenShift cluster.
func SetIsOpenShift(_isOpenShift bool) {
	cluster = openShift
}

var NuxeoServiceAccountName = "nuxeo"

// ObjectsDiffer generates a YAML from each passed object then generates an MD5 sum of each YAML and returns
// true if the MD5 sums differ, and false if the MD5 sums are the same. If the YAML generation fails, then the
// resulting error is returned, otherwise a nil error is returned. This function is intended for comparing
// CR specs. The underlying assumption is that any difference in a spec is actionable for the operator. So this
// handles two cases: 1) the Nuxeo CR is modified, and a dependent CR should look different as a result. 2) A
// cluster CR owned by the Nuxeo CR is modified independently of the Operator and is therefore out of sync
// with how the Operator would expect it to look. Note that this works in most cases but not all. For example,
// a PVC can be generated with nil values in the Spec Volume field and the cluster will fill the volume field
// in. So this function is only useful for Specs that the cluster doesn't alter.
func ObjectsDiffer(expected interface{}, actual interface{}) (bool, error) {
	var expMd5, actMd5 [md5.Size]byte
	var err error
	var bytes []byte

	if bytes, err = yaml.Marshal(expected); err != nil {
		return false, err
	}
	debugExp := string(bytes)
	expMd5 = md5.Sum(bytes)
	if bytes, err = yaml.Marshal(actual); err != nil {
		return false, err
	}
	debugAct := string(bytes)
	_ = debugAct
	_ = debugExp
	actMd5 = md5.Sum(bytes)
	return expMd5 != actMd5, nil
}

// DebugDumpObj is used for debugging as needed. It dumps the YAML to the console for the passed object
func DebugDumpObj(obj runtime.Object) {
	if bytes, err := yaml.Marshal(obj); err != nil {
		return
	} else {
		manifest := string(bytes)
		println(manifest)
	}
}

// GetNuxeoContainer walks the container array in the passed deployment and returns a ref to the container
// named "nuxeo". If not found, returns a nil container ref and an error.
func GetNuxeoContainer(dep *appsv1.Deployment) (*corev1.Container, error) {
	for i := 0; i < len(dep.Spec.Template.Spec.Containers); i++ {
		if dep.Spec.Template.Spec.Containers[i].Name == "nuxeo" {
			return &dep.Spec.Template.Spec.Containers[i], nil
		}
	}
	return nil, goerrors.New("could not find a container named 'nuxeo' in the deployment")
}

// GetEnv searches the environment variable array in the passed container for an env var with the passed name.
// If found, returns a ref to the env var, else returns nil.
func GetEnv(container *corev1.Container, envName string) *corev1.EnvVar {
	for i := 0; i < len(container.Env); i++ {
		if container.Env[i].Name == envName {
			return &container.Env[i]
		}
	}
	return nil
}

// MergeOrAdd searches the environment variable array in the passed container for an entry whose name matches
// the name of the passed environment variable struct. If found in the container array, the value of the passed
// variable is appended to the value of the existing variable, separated by the passed separator. Otherwise
// the passed environment variable struct is appended to the container env var array. E.g. given a container
// with an existing env var corev1.EnvVar{Name: "Z", Value "abc,123"}, then:
//   MergeOrAdd(someContainer, corev1.EnvVar{Name: "Z", Value "xyz,456"}, ",")
// updates the container's variable value to: "abc,123,xyz,456"
func MergeOrAdd(container *corev1.Container, env corev1.EnvVar, separator string) error {
	if env.ValueFrom != nil {
		return goerrors.New("MergeOrAdd cannot be used for 'ValueFrom' environment variables")
	}
	if existingEnv := GetEnv(container, env.Name); existingEnv == nil {
		container.Env = append(container.Env, env)
	} else {
		if existingEnv.ValueFrom != nil {
			return goerrors.New("MergeOrAdd cannot be used for 'ValueFrom' environment variables")
		}
		existingEnv.Value += separator + env.Value
	}
	return nil
}

// Adds the passed environment variable to the passed container only if not already present. If already present
// in the container, returns an error
func OnlyAdd(container *corev1.Container, env corev1.EnvVar) error {
	if existingEnv := GetEnv(container, env.Name); existingEnv != nil {
		return goerrors.New("duplicate environment variable: "+env.Name)
	}
	container.Env = append(container.Env, env)
	return nil
}

// GetJsonPathValue applies the passed JSONPath expression to the passed runtime object and returns the
// result of the parse. It's less friendly than the kubectl get -o=jsonpath= in that the passed JSON path
// has to be included in curly braces. A variety of errors are returned but an empty return value and nil
// error can also indicate that the provided JSON path didn't find anything in the passed resource.
// todo-me clone RelaxedJSONPathExpression: https://github.com/kubernetes/kubectl/blob/master/pkg/cmd/get/customcolumn.go
func GetJsonPathValue(obj runtime.Object, jsonPath string) ([]byte, error) {
	if len(jsonPath) < 3 {
		return nil, goerrors.New("invalid JSONPath expression: " + jsonPath)
	}
	if jsonPath[0:1]+jsonPath[len(jsonPath)-1:] != "{}" {
		return nil, goerrors.New("JSONPath expression must be curly-brace enclosed: " + jsonPath)
	}
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	jp := jsonpath.New("jp")
	// parse the JSON path expression
	err = jp.Parse(jsonPath)
	if err != nil {
		return nil, err
	}
	result, err := jp.FindResults(&unstructured)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	for ix := range result {
		if err := jp.PrintResults(&buf, result[ix]); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// Returns a pointer to the passed value
func Int32Ptr(i int32) *int32 {
	return &i
}

// Returns a pointer to the passed value
func Int64Ptr(i int64) *int64 {
	return &i
}

// set v = thenVal if v == ifVal
func SetInt32If(v *int32, ifVal int32, thenVal int32) {
	if *v == ifVal {
		*v = thenVal
	}
}
