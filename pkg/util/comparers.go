package util

import (
	"reflect"

	"k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// These comparers compare expected to found with logic unique to the various resource types. If expected ==
// found then the functions return true. Otherwise the functions update found from expected again within specific
// logic by type and return false. False means found must be written back to the cluster.

// Secret comparer (operator may annotate secrets)
func SecretCompare(expected runtime.Object, found runtime.Object) bool {
	if !reflect.DeepEqual(expected.(*v1.Secret).Data, found.(*v1.Secret).Data) {
		found.(*v1.Secret).Data = expected.(*v1.Secret).Data
		found.(*v1.Secret).Annotations = expected.(*v1.Secret).Annotations
		return false
	} else if !reflect.DeepEqual(expected.(*v1.Secret).Annotations, found.(*v1.Secret).Annotations) {
		found.(*v1.Secret).Annotations = expected.(*v1.Secret).Annotations
		return false
	}
	return true
}

// Service comparer (Kubeternes will update components of the service spec with values it chooses so the
// comparer has to ignore those)
func ServiceComparer(expected runtime.Object, found runtime.Object) bool {
	if !reflect.DeepEqual(expected.(*v1.Service).Spec.Ports, found.(*v1.Service).Spec.Ports) ||
		!reflect.DeepEqual(expected.(*v1.Service).Spec.Selector, found.(*v1.Service).Spec.Selector) ||
		expected.(*v1.Service).Spec.Type != found.(*v1.Service).Spec.Type {
		found.(*v1.Service).Spec.Ports = expected.(*v1.Service).Spec.Ports
		found.(*v1.Service).Spec.Selector = expected.(*v1.Service).Spec.Selector
		found.(*v1.Service).Spec.Type = expected.(*v1.Service).Spec.Type
		return false
	}
	return true
}

// Ingress comparer (ingress annotations control passthrough)
func IngressComparer(expected runtime.Object, found runtime.Object) bool {
	same := true
	const nginxPassthroughAnnotation = "nginx.ingress.kubernetes.io/ssl-passthrough"
	if !reflect.DeepEqual(expected.(*v1beta1.Ingress).Spec, found.(*v1beta1.Ingress).Spec) {
		expected.(*v1beta1.Ingress).Spec.DeepCopyInto(&found.(*v1beta1.Ingress).Spec)
		same = false
	}
	expAnnotations := expected.(*v1beta1.Ingress).Annotations
	foundAnnotations := found.(*v1beta1.Ingress).Annotations
	if foundAnnotations != nil {
		if _, ok := foundAnnotations[nginxPassthroughAnnotation]; ok {
			// found ingress is annotated with passthrough
			if expAnnotations == nil {
				// Nuxeo CR was updated: change Ingress from passthrough TLS to normal HTTP
				delete(foundAnnotations, nginxPassthroughAnnotation)
				same = false
			}
		} else if expAnnotations != nil{
			// found ingress is not annotated with passthrough and expected is thusly annotated
			foundAnnotations[nginxPassthroughAnnotation] = "true"
		}
	} else if expAnnotations != nil {
		// Nuxeo CR was updated: change Ingress from normal HTTP to passthrough TLS
		foundAnnotations = map[string]string{nginxPassthroughAnnotation: "true"}
		same = false
	}
	return same
}