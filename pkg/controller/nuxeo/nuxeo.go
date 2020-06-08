package nuxeo

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// updateNuxeoStatus updates the status field in the Nuxeo CR being watched by the operator
func updateNuxeoStatus(r *ReconcileNuxeo, nux *v1alpha1.Nuxeo, reqLogger logr.Logger) {
	deployments := appsv1.DeploymentList{}
	opts := []client.ListOption{
		client.InNamespace(nux.Namespace),
	}
	availableNodes := int32(0)
	if err := r.client.List(context.TODO(), &deployments, opts...); err == nil {
		for _, dep := range deployments.Items {
			if ownedBy(nux, dep) {
				availableNodes += dep.Status.AvailableReplicas
			}
		}
		nux.Status.AvailableNodes = availableNodes
	} else {
		reqLogger.Error(err, "Failed to list deployments for Nuxeo", "Namespace",
			nux.Namespace, "Name", nux.Name)
	}
}

// ownedBy returns true if the passed deployment is owned by the passed Nuxeo CR, otherwise returns false
func ownedBy(nux *v1alpha1.Nuxeo, dep appsv1.Deployment) bool {
	for _, ref := range dep.OwnerReferences {
		if ref.UID == nux.UID {
			return true
		}
	}
	return false
}