package nuxeo

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// updateNuxeoStatus updates the status field in the Nuxeo CR being watched by the operator. This is a
// very crude implementation and will be expanded in a later version
func (r *ReconcileNuxeo) updateNuxeoStatus(instance *v1alpha1.Nuxeo) error {
	deployments := appsv1.DeploymentList{}
	opts := []client.ListOption{
		client.InNamespace(instance.Namespace),
	}
	availableNodes := int32(0)
	if err := r.client.List(context.TODO(), &deployments, opts...); err != nil {
		return err
	} else {
		for _, dep := range deployments.Items {
			if instance.IsOwner(dep.ObjectMeta) {
				availableNodes += dep.Status.AvailableReplicas
			}
		}
		instance.Status.AvailableNodes = availableNodes
	}
	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		return err
	}
	return nil
}
