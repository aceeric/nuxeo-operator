package util

import (
	"bytes"
	"crypto/md5"
	"fmt"

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
func SetIsOpenShift(isOpenShift bool) {
	if isOpenShift {
		cluster = openShift
	} else {
		cluster = kubernetes
	}
}

var NuxeoServiceAccountName = "nuxeo"

// Used for debugging
func ObjectsDiffer(expected interface{}, actual interface{}) (bool, error) {
	var expMd5, actMd5 [md5.Size]byte
	var err error
	var b []byte

	if b, err = yaml.Marshal(expected); err != nil {
		return false, err
	}
	debugExp := string(b)
	expMd5 = md5.Sum(b)
	if b, err = yaml.Marshal(actual); err != nil {
		return false, err
	}
	debugAct := string(b)
	_ = debugAct
	_ = debugExp
	actMd5 = md5.Sum(b)
	return expMd5 != actMd5, nil
}

// DebugDumpObj is used for debugging as needed. It dumps the YAML to the console for the passed object
func DebugDumpObj(obj runtime.Object) {
	if b, err := yaml.Marshal(obj); err != nil {
		return
	} else {
		manifest := string(b)
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
	return nil, fmt.Errorf("could not find a container named 'nuxeo' in the deployment")
}

// gets the named environment variable from the passed container or nil
func GetEnv(container *corev1.Container, envName string) *corev1.EnvVar {
	for i := 0; i < len(container.Env); i++ {
		if container.Env[i].Name == envName {
			return &container.Env[i]
		}
	}
	return nil
}

// gets the named volume mount from the passed container or nil
func getVolMnt(container *corev1.Container, mntName string) *corev1.VolumeMount {
	for i := 0; i < len(container.VolumeMounts); i++ {
		if container.VolumeMounts[i].Name == mntName {
			return &container.VolumeMounts[i]
		}
	}
	return nil
}

// gets the named volume mount from the passed deployment or nil
func getVol(dep *appsv1.Deployment, volName string) *corev1.Volume {
	for i := 0; i < len(dep.Spec.Template.Spec.Volumes); i++ {
		if dep.Spec.Template.Spec.Volumes[i].Name == volName {
			return &dep.Spec.Template.Spec.Volumes[i]
		}
	}
	return nil
}

// gets the named container from the passed deployment or nil
func getContainer(dep *appsv1.Deployment, containerName string) *corev1.Container {
	for i := 0; i < len(dep.Spec.Template.Spec.Containers); i++ {
		if dep.Spec.Template.Spec.Containers[i].Name == containerName {
			return &dep.Spec.Template.Spec.Containers[i]
		}
	}
	return nil
}

// MergeOrAddEnvVar searches the environment variable array in the passed container for an entry whose name matches
// the name of the passed environment variable struct. If found in the container array, the value of the passed
// variable is appended to the value of the existing variable, separated by the passed separator. Otherwise
// the passed environment variable struct is appended to the container env var array. E.g. given a container
// with an existing env var corev1.EnvVar{Name: "Z", Value "abc,123"}, then:
//
//    MergeOrAddEnvVar(someContainer, corev1.EnvVar{Name: "Z", Value "xyz,456"}, ",")
//
// updates the variable value to: "abc,123,xyz,456". If either the passed environment variable or the existing one
// (by name) is a 'ValueFrom' environment variable then an error is returned.
func MergeOrAddEnvVar(container *corev1.Container, env corev1.EnvVar, separator string) error {
	if env.ValueFrom != nil {
		return fmt.Errorf("MergeOrAddEnvVar cannot be used for 'ValueFrom' environment variables")
	}
	if existingEnv := GetEnv(container, env.Name); existingEnv == nil {
		container.Env = append(container.Env, env)
	} else {
		if existingEnv.ValueFrom != nil {
			return fmt.Errorf("MergeOrAddEnvVar cannot be used for 'ValueFrom' environment variables")
		}
		existingEnv.Value += separator + env.Value
	}
	return nil
}

// Adds the passed environment variable to the passed container if not present, otherwise errors
func OnlyAddEnvVar(container *corev1.Container, env corev1.EnvVar) error {
	if existingEnv := GetEnv(container, env.Name); existingEnv != nil {
		return fmt.Errorf("duplicate environment variable: %v", env.Name)
	}
	container.Env = append(container.Env, env)
	return nil
}

// Adds the passed volume mount to the passed container if not present in the container, otherwise errors
func OnlyAddVolMnt(container *corev1.Container, mnt corev1.VolumeMount) error {
	if existingMnt := getVolMnt(container, mnt.Name); existingMnt != nil {
		return fmt.Errorf("duplicate volume mount: %v", mnt.Name)
	}
	container.VolumeMounts = append(container.VolumeMounts, mnt)
	return nil
}

// Adds the passed volume to the passed deployment if not present in the container, otherwise errors
func OnlyAddVol(dep *appsv1.Deployment, vol corev1.Volume) error {
	if existingVol:= getVol(dep, vol.Name); existingVol != nil {
		return fmt.Errorf("duplicate volume: %v", vol.Name)
	}
	dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes, vol)
	return nil
}

// Adds the passed container to the passed deployment if not present in the deployment, otherwise errors
func OnlyAddContainer(dep *appsv1.Deployment, container corev1.Container) error {
	if existingContainer := getContainer(dep, container.Name); existingContainer != nil {
		return fmt.Errorf("duplicate container: %v", existingContainer.Name)
	}
	dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers, container)
	return nil
}

// GetJsonPathValue applies the passed JSONPath expression to the passed runtime object and returns the
// result of the parse. It's less friendly than the kubectl get -o=jsonpath= in that the passed JSON path
// has to be in the canonical format: "{.path...}" meaning 1) enclosed in curly braces and 2) starting with
// a period. A variety of errors are returned but an empty return value and nil error can also indicate
// that the provided JSON path didn't find anything in the passed resource.
func GetJsonPathValue(obj runtime.Object, jsonPath string) ([]byte, error) {
	if len(jsonPath) < 3 || jsonPath[0:2]+jsonPath[len(jsonPath)-1:] != "{.}" {
		return nil, fmt.Errorf("invalid JSONPath expression: '%v'", jsonPath)
	} else if unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj); err != nil {
		return nil, err
	} else {
		return GetJsonPathValueU(unstructured, jsonPath)
	}
}

// Same as GetJsonPathValue, except takes an unstructured instead of a runtime object
func GetJsonPathValueU(unstructured map[string]interface{}, jsonPath string) ([]byte, error) {
	jp := jsonpath.New("jp")
	// parse the JSON path expression
	err := jp.Parse(jsonPath)
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

// mapToStr takes map A:1,B:2 and returns string "A=1\nB=2\n"
func mapToStr(cfg map[string]string, delim string) string {
	b := new(bytes.Buffer)
	for key, value := range cfg {
		_, _ = fmt.Fprintf(b, "%s=%s%s", key, value, delim)
	}
	return b.String()
}
