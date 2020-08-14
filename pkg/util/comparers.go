package util

import (
	"reflect"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// These comparers compare expected to found with logic unique to the various resource types. If expected ==
// found then the functions return true. Otherwise the functions update found from expected again with specific
// logic by type and return false. False means found must be written back to the cluster.

// Secret comparer (operator may annotate secrets)
func SecretComparer(expected runtime.Object, found runtime.Object) bool {
	if !reflect.DeepEqual(expected.(*corev1.Secret).Data, found.(*corev1.Secret).Data) {
		found.(*corev1.Secret).Data = expected.(*corev1.Secret).Data
		found.(*corev1.Secret).Annotations = expected.(*corev1.Secret).Annotations
		return false
	} else if !reflect.DeepEqual(expected.(*corev1.Secret).Annotations, found.(*corev1.Secret).Annotations) {
		found.(*corev1.Secret).Annotations = expected.(*corev1.Secret).Annotations
		return false
	}
	return true
}

// Service comparer (Kubernetes will update components of the service spec with values it chooses so the
// comparer has to ignore those)
func ServiceComparer(expected runtime.Object, found runtime.Object) bool {
	if !reflect.DeepEqual(expected.(*corev1.Service).Spec.Ports, found.(*corev1.Service).Spec.Ports) ||
		!reflect.DeepEqual(expected.(*corev1.Service).Spec.Selector, found.(*corev1.Service).Spec.Selector) ||
		expected.(*corev1.Service).Spec.Type != found.(*corev1.Service).Spec.Type {
		found.(*corev1.Service).Spec.Ports = expected.(*corev1.Service).Spec.Ports
		found.(*corev1.Service).Spec.Selector = expected.(*corev1.Service).Spec.Selector
		found.(*corev1.Service).Spec.Type = expected.(*corev1.Service).Spec.Type
		return false
	}
	return true
}

// Ingress comparer (ingress annotations control passthrough) so - while these annotations are not
// part of the spec - if they change then its a reconcilement event.
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
		} else if expAnnotations != nil {
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

// OpenShift Route comparer
func RouteComparer(expected runtime.Object, found runtime.Object) bool {
	if !reflect.DeepEqual(expected.(*routev1.Route).Spec, found.(*routev1.Route).Spec) {
		expected.(*routev1.Route).Spec.DeepCopyInto(&found.(*routev1.Route).Spec)
		return false
	}
	return true
}

// Deployment comparer. The Nuxeo Operator doesn't annotate Spec > Template > Metadata > Annotations but
// Kubernetes uses that to trigger a rolling update if someone executes 'kubectl rollout restart deployment
// nuxeo-cluster' so this code nils the found annotations so the comparer doesn't consider them. If anything
// else in the deployment changes - triggering creation of a new deployment - then this will result in the
// annotations being reset to empty which is consistent with the state of a new deployment created by the
// operator.
func DeploymentComparer(expected runtime.Object, found runtime.Object) bool {
	found.(*appsv1.Deployment).Spec.Template.Annotations = nil
	if !reflect.DeepEqual(expected.(*appsv1.Deployment).Spec, found.(*appsv1.Deployment).Spec) {
		expected.(*appsv1.Deployment).Spec.DeepCopyInto(&found.(*appsv1.Deployment).Spec)
		return false
	}
	return true
}

// ConfigMap comparer
func ConfigMapComparer(expected runtime.Object, found runtime.Object) bool {
	if !reflect.DeepEqual(expected.(*corev1.ConfigMap).Data, found.(*corev1.ConfigMap).Data) {
		found.(*corev1.ConfigMap).Data = expected.(*corev1.ConfigMap).Data
		return false
	}
	return true
}

// PersistentVolumeClaim comparer
func PvcComparer(expected runtime.Object, found runtime.Object) bool {
	exp := expected.(*corev1.PersistentVolumeClaim)
	fnd := found.(*corev1.PersistentVolumeClaim)
	if !reflect.DeepEqual(exp.Spec.AccessModes, fnd.Spec.AccessModes) ||
		!reflect.DeepEqual(exp.Spec.Resources, fnd.Spec.Resources) ||
		(exp.Spec.VolumeMode != nil && exp.Spec.VolumeMode != fnd.Spec.VolumeMode) {
		fnd.Spec.AccessModes = exp.Spec.AccessModes
		fnd.Spec.Resources = exp.Spec.Resources
		fnd.Spec.AccessModes = exp.Spec.AccessModes
		return false
	}
	return true
}

// objects are always the same
func NopComparer(runtime.Object, runtime.Object) bool {
	return true
}
