package nuxeo

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// updateNuxeoStatus updates the status field in the Nuxeo CR being watched by the operator. This is a
// very crude implementation and will be expanded in a later version
func updateNuxeoStatus(r *ReconcileNuxeo, nux *v1alpha1.Nuxeo, reqLogger logr.Logger) error {
	deployments := appsv1.DeploymentList{}
	opts := []client.ListOption{
		client.InNamespace(nux.Namespace),
	}
	availableNodes := int32(0)
	if err := r.client.List(context.TODO(), &deployments, opts...); err != nil {
		return errors.Wrap(err, "Failed to list deployments for Nuxeo")
	} else {
		for _, dep := range deployments.Items {
			if nux.IsOwner(dep.ObjectMeta) {
				availableNodes += dep.Status.AvailableReplicas
			}
		}
		nux.Status.AvailableNodes = availableNodes
	}
	if err := r.client.Status().Update(context.TODO(), nux); err != nil {
		return errors.Wrap(err, "Failed to update Nuxeo status")
	}
	return nil
}
