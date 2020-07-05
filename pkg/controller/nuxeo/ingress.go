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
func reconcileIngress(r *ReconcileNuxeo, access v1alpha1.NuxeoAccess, forcePassthrough bool, nodeSet v1alpha1.NodeSet,
	instance *v1alpha1.Nuxeo, reqLogger logr.Logger) (reconcile.Result, error) {
	found := &v1beta1.Ingress{}
	ingressName := ingressName(instance, nodeSet)
	expected, err := r.defaultIngress(instance, access, forcePassthrough, ingressName, nodeSet)
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

// defaultIngress generates and returns an Ingress struct from the passed params. If the passed 'access' struct
// indicates TLS termination, or forcePassthrough==true, then an annotation is included in the returned object's
// metadata
func (r *ReconcileNuxeo) defaultIngress(nux *v1alpha1.Nuxeo, access v1alpha1.NuxeoAccess, forcePassthrough bool, ingressName string,
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
	if access.Termination != "" || forcePassthrough {
		if access.Termination != "" && access.Termination != routev1.TLSTerminationPassthrough &&
		    access.Termination != routev1.TLSTerminationEdge {
			return nil, goerrors.New("only passthrough and edge termination are supported")
		}
		ingress.Spec.TLS = []v1beta1.IngressTLS{{
			Hosts: []string{access.Hostname},
		}}
		if access.Termination == routev1.TLSTerminationPassthrough || forcePassthrough {
			ingress.ObjectMeta.Annotations = map[string]string{nginxPassthroughAnnotation: "true"}
		} else {
			// the Ingress will terminate TLS
			if access.TLSSecret == "" {
				return nil, goerrors.New("the Ingress was configured for TLS termination but no secret was provided")
			}
			// secret needs keys 'tls.crt' and 'tls.key' and cert must have CN=<access.Hostname>
			ingress.Spec.TLS[0].SecretName  = access.TLSSecret
		}
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
