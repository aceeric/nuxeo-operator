package nuxeo

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcileOpenShiftRoute configures access to the Nuxeo cluster via a Kubernetes Ingress
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
		// Route created successfully
		return reconcile.Result{}, nil
	} else if err != nil {
		reqLogger.Error(err, "Error attempting to get Ingress for Nuxeo cluster: "+ingressName)
		return reconcile.Result{}, err
	}
	if !equality.Semantic.DeepDerivative(expected.Spec, found.Spec) {
		reqLogger.Info("Updating Ingress", "Namespace", expected.Namespace, "Name", expected.Name)
		expected.Spec.DeepCopyInto(&found.Spec)
		if err = r.client.Update(context.TODO(), found); err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

// defaultIngress generates and returns an Ingress struct from the passed params. The generated Ingress is illustrated
// below.
// todo-me log when non-kubernetes options are provided in the CR
// todo-me TLS
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
			Backend: &v1beta1.IngressBackend{
				ServiceName: serviceName(nux, nodeSet),
				ServicePort: targetPort,
			},
			//todo-me
			//TLS: []v1beta1.IngressTLS{{
			//	Hosts:      nil,
			//	SecretName: "",
			//}},
		},
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
