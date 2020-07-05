package nuxeo

import (
	"context"
	goerrors "errors"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcilePvc examines the Storage definitions in each NodeSet of the passed Nuxeo CR, gathers a list of PVCs,
// then conforms actual PVCs in the cluster to those expected PVCs. If the Nuxeo CR changes the definition of a PVC,
// and there is an existing PVC with the same name, then the existing PVC is deleted and re-created.
func reconcilePvc(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo) (reconcile.Result, error) {
	var expectedPvcs []corev1.PersistentVolumeClaim
	for _, nodeSet := range instance.Spec.NodeSets {
		for _, storage := range nodeSet.Storage {
			claimDefined, err := util.ObjectsDiffer(storage.VolumeClaimTemplate, corev1.PersistentVolumeClaim{})
			if err != nil {
				return reconcile.Result{}, err
			} else if claimDefined {
				// CR defines an explicit PVC for the storage
				storage.VolumeClaimTemplate.Namespace = instance.Namespace
				_ = controllerutil.SetControllerReference(instance, &storage.VolumeClaimTemplate, r.scheme)
				expectedPvcs = append(expectedPvcs, storage.VolumeClaimTemplate)
			} else if storage.VolumeSource == (corev1.VolumeSource{}) {
				// default PVC - let the operator define the PVC struct
				volName := volumeNameForStorage(storage.StorageType)
				pvc := corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      volName + "-pvc",
						Namespace: instance.Namespace,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(storage.Size),
							},
						},
					},
				}
				_ = controllerutil.SetControllerReference(instance, &pvc, r.scheme)
				expectedPvcs = append(expectedPvcs, pvc)
			} else {
				// volume source explicitly defined in CR so do not reconcile a PVC for this storage spec
			}
		}
	}
	// we now have a (possibly empty) list of all PVCs that are expected in the cluster - conform the cluster
	var actualPvcs corev1.PersistentVolumeClaimList
	opts := []client.ListOption{
		client.InNamespace(instance.Namespace),
	}
	if err := r.client.List(context.TODO(), &actualPvcs, opts...); err != nil {
		return reconcile.Result{}, err
	} else {
		if err := addPvcs(r, instance, expectedPvcs, actualPvcs.Items); err != nil {
			return reconcile.Result{}, err
		}
		if err := deletePvcs(r, instance, expectedPvcs, actualPvcs.Items); err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

// getPvc searches the passed array of PVCs for one with a Name matching the passed pvc Name. If found, returns
// a ref to the item in the array. Else returns nil.
func getPvc(pvcs []corev1.PersistentVolumeClaim, pvcName string) *corev1.PersistentVolumeClaim {
	for _, pvc := range pvcs {
		if pvc.Name == pvcName {
			return &pvc
		}
	}
	return nil
}

// addPvcs creates expected PVCs in the cluster if 1) not already existent, or 2) the specs differ. If there
// is an existing PVC for an expected PVC (same name) and that existing PVC is not owned by the passed Nuxeo
// CR, then that is an error condition, and a non-nil error is returned.
func addPvcs(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, expected []corev1.PersistentVolumeClaim,
	actual []corev1.PersistentVolumeClaim) error {
	for _, expectedPvc := range expected {
		if actualPvc := getPvc(actual, expectedPvc.Name); actualPvc != nil {
			if !instance.IsOwner(actualPvc.ObjectMeta) {
				return goerrors.New(fmt.Sprintf("Existing PVC '%v' is not owned by this Nuxeo '%v' and cannot be reconciled",
					actualPvc.Name, instance.UID))
			}
			different := !reflect.DeepEqual(expectedPvc.Spec.AccessModes, actualPvc.Spec.AccessModes) ||
				!reflect.DeepEqual(expectedPvc.Spec.Resources, actualPvc.Spec.Resources) ||
				(expectedPvc.Spec.VolumeMode != nil && expectedPvc.Spec.VolumeMode != actualPvc.Spec.VolumeMode)
			if different {
				if err := r.client.Delete(context.TODO(), actualPvc); err != nil {
					return err
				}
				if err := r.client.Create(context.TODO(), &expectedPvc); err != nil {
					return err
				}
			} else {
				// same - nop
			}
		} else if err := r.client.Create(context.TODO(), &expectedPvc); err != nil {
			return err
		}
	}
	return nil
}

// deletePvcs removes orphaned PVCs. The use case is: a Nuxeo CR is deployed with a PVC defined for, say, Data.
// Someone edits the Nuxeo CR and changes the name of the PVC. This function removes the previous PVC. Only PVCs
// owned by the passed Nuxeo CR are removed.
func deletePvcs(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, expected []corev1.PersistentVolumeClaim,
	actual []corev1.PersistentVolumeClaim) error {
	for _, actualPvc := range actual {
		if expectedPvc := getPvc(expected, actualPvc.Name); expectedPvc == nil && instance.IsOwner(actualPvc.ObjectMeta) {
			if err := r.client.Delete(context.TODO(), &actualPvc); err != nil {
				return err
			}
		}
	}
	return nil
}
