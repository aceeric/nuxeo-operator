/*
Copyright 2020 Eric Ace.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nuxeo

import (
	"context"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// updateNuxeoStatus updates the status field in the Nuxeo CR being watched by the operator
func (r *NuxeoReconciler) updateNuxeoStatus(instance *v1alpha1.Nuxeo) error {
	deployments := appsv1.DeploymentList{}
	opts := []client.ListOption{
		client.InNamespace(instance.Namespace),
	}
	desiredNodes := int32(0)
	for _, nodeSet := range instance.Spec.NodeSets {
		desiredNodes += nodeSet.Replicas
	}
	availableNodes := int32(0)
	if err := r.List(context.TODO(), &deployments, opts...); err != nil {
		return err
	} else {
		for _, dep := range deployments.Items {
			if instance.IsOwner(dep.ObjectMeta) {
				availableNodes += dep.Status.AvailableReplicas
			}
		}
		instance.Status.DesiredNodes = desiredNodes
		instance.Status.AvailableNodes = availableNodes
		switch {
		case availableNodes == 0:
			instance.Status.Status = v1alpha1.StatusUnavailable
		case availableNodes == desiredNodes:
			instance.Status.Status = v1alpha1.StatusHealthy
		default:
			instance.Status.Status = v1alpha1.StatusDegraded
		}
	}
	if err := r.Status().Update(context.TODO(), instance); err != nil {
		return err
	}
	return nil
}
