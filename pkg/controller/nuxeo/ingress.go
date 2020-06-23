package nuxeo

import (
	"context"
	goerrors "errors"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const nginxPassthroughAnnotation = "nginx.ingress.kubernetes.io/ssl-passthrough"

// reconcileIngress configures access to the Nuxeo cluster via a Kubernetes Ingress
func reconcileIngress(r *ReconcileNuxeo, access v1alpha1.NuxeoAccess, nodeSet v1alpha1.NodeSet,
	instance *v1alpha1.Nuxeo, reqLogger logr.Logger) (reconcile.Result, error) {
	found := &v1beta1.Ingress{}
	ingressName := ingressName(instance, nodeSet)
	expected, err := r.defaultIngress(instance, access, ingressName, nodeSet)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: ingressName, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Ingress", "Namespace", expected.Namespace, "Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Ingress", "Namespace", expected.Namespace, "Name", expected.Name)
			return reconcile.Result{}, err
		}
		// Ingress created successfully
		return reconcile.Result{}, nil
	} else if err != nil {
		reqLogger.Error(err, "Error attempting to get Ingress for Nuxeo cluster: "+ingressName)
		return reconcile.Result{}, err
	}
	shouldUpdate := false
	//if !equality.Semantic.DeepDerivative(expected.Spec, found.Spec) || tlsSpecChanged(expected.Spec.TLS, found.Spec.TLS) {
	//	expected.Spec.DeepCopyInto(&found.Spec)
	//	shouldUpdate = true
	//}
	// experiment
	if different, err := util.ObjectsDiffer(expected.Spec, found.Spec); err == nil && different {
		expected.Spec.DeepCopyInto(&found.Spec)
		shouldUpdate = true
	} else if err != nil {
		return reconcile.Result{}, err
	}
	// ingress tls termination behavior is affected by annotations rather than the ingress spec
	if a := updatedAnnotations(expected.ObjectMeta.Annotations, found.ObjectMeta.Annotations); a != nil {
		found.ObjectMeta.Annotations = a
		shouldUpdate = true
	}
	if shouldUpdate {
		reqLogger.Info("Updating Ingress", "Namespace", expected.Namespace, "Name", expected.Name)
		err = r.client.Update(context.TODO(), found)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

//// returns true if the passed expected and actual IngressTLS's are different, else false
//func tlsSpecChanged(expected []v1beta1.IngressTLS, found []v1beta1.IngressTLS) bool {
//	if expected == nil && found == nil {
//		return false
//	} else if expected == nil || found == nil {
//		return true
//	}
//	return reflect.DeepEqual(expected, found)
//}

// checks the annotations in the Ingress from the cluster. If the SSL passthrough annotations are different then
// add/remove the passthrough annotation in the cluster copy and return the modified copy. Caller must update
// the cluster from the return value. Returns nil if no change. The use case is: the watched Nuxeo CR was changed
// and TLS termination was added/removed and so the Ingress annotations that control this must be updated
// accordingly.
func updatedAnnotations(expected map[string]string, found map[string]string) map[string]string {
	if found != nil {
		if _, ok := found[nginxPassthroughAnnotation]; ok {
			if expected == nil {
				// Nuxeo CR was updated: change Ingress from passthrough TLS to normal HTTP
				delete(found, nginxPassthroughAnnotation)
				return found
			}
		}
	} else if expected != nil {
		// Nuxeo CR was updated: change Ingress from normal HTTP to passthrough TLS
		return map[string]string{nginxPassthroughAnnotation: "true"}
	}
	return nil
}

// defaultIngress generates and returns an Ingress struct from the passed params. The generated Ingress spec
// is illustrated below:
//  spec:
//    rules:
//    - host: <access.Hostname>
//      http:
//        paths:
//        - backend:
//            serviceName: <see serviceName function>
//            servicePort: web
//    tls: (if access.Termination specified)
//    - hosts:
//      - <access.Hostname>
// If the passed 'access' struct indicates TLS termination then an annotation is included in the
// returned object's metadata
func (r *ReconcileNuxeo) defaultIngress(nux *v1alpha1.Nuxeo, access v1alpha1.NuxeoAccess, ingressName string,
	nodeSet v1alpha1.NodeSet) (*v1beta1.Ingress, error) {
	targetPort := intstr.IntOrString{
		Type:   intstr.String,
		StrVal: "web",
	}
	if access.TargetPort != (intstr.IntOrString{}) {
		targetPort = access.TargetPort
	}
	ingress := v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: nux.Namespace,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{{
				Host: access.Hostname,
				IngressRuleValue: v1beta1.IngressRuleValue{
					HTTP: &v1beta1.HTTPIngressRuleValue{
						Paths: []v1beta1.HTTPIngressPath{{
							Backend: v1beta1.IngressBackend{
								ServiceName: serviceName(nux, nodeSet),
								ServicePort: targetPort,
							},
						}},
					},
				},
			}},
		},
	}
	if access.Termination != "" {
		if access.Termination != routev1.TLSTerminationPassthrough {
			return nil, goerrors.New("current operator version only supports TLS passthrough")
		}
		ingress.Spec.TLS = []v1beta1.IngressTLS{{
			Hosts: []string{access.Hostname},
		}}
		ingress.ObjectMeta.Annotations = map[string]string{nginxPassthroughAnnotation: "true"}
	}
	_ = controllerutil.SetControllerReference(nux, &ingress, r.scheme)
	return &ingress, nil
}

// ingressName generates an Ingress name from the passed Nuxeo CR, and the passed NodeSet. The generated
// name consists of the passed Nuxeo CR name + dash + the passed 'nodeSet' name + dash + 'ingress'. E.g. if
// 'nux.Name' is 'my-nuxeo' and 'nodeSet.Name' is 'cluster' then the function returns 'my-nuxeo-cluster-ingress'.
func ingressName(nux *v1alpha1.Nuxeo, nodeSet v1alpha1.NodeSet) string {
	return nux.Name + "-" + nodeSet.Name + "-ingress"
}
