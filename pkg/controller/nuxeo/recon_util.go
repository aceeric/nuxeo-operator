package nuxeo

import (
	"context"
	goerrors "errors"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
)

// Compares expected to found. If identical, returns true, meaning no action required.
// Otherwise updates found from expected and returns false, meaning caller must write
// found back to the cluster.
type comparer func (expected runtime.Object, found runtime.Object) bool

// indicates whether the 'addOrUpdate' function updated or created a resource, or did nothing
type reconOp int
const (
	Updated reconOp = 1
	Created = 2
	NA = 3
)

// Performs the standard reconciliation logic with expected and found. Expected is what the caller expects
// to find in the cluster. Found is a ref of the same type to receive what exists in the cluster. If no instance
// of expected exists in the cluster, then expected is created in the cluster. If an instance of expected (i.e. found)
// exists and differs, then the cluster is updated from expected. Otherwise cluster is not altered. The comparer
// function is called to do two things: 1) determine logical equality of expected and found, and 2) if unequal
// to set the state of found from expected so this function can write found back into the cluster.
//
// Caller is expected to have set the Nuxeo CR as the owner of 'expected' if that is the intent (this function
// performs no modifications to 'expected')
func addOrUpdate(r *ReconcileNuxeo, name string, namespace string, expected runtime.Object, found runtime.Object,
	comparer comparer, reqLogger logr.Logger) (reconOp, error)  {
	var kind string
	var err error
	if kind, err = getKind(r.scheme, expected); err != nil {
		return NA, err
	}
	knv := []interface{}{"Namespace", namespace, "Name", name}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new "+kind, knv...)
		err = r.client.Create(context.TODO(), expected)
		if err != nil {
			reqLogger.Error(err, "Failed to create "+kind, knv...)
			return NA, err
		}
		return Created, nil
	} else if err != nil {
		reqLogger.Error(err, "Error attempting to get "+kind, knv...)
		return NA, err
	}
	if !comparer(expected, found) {
		reqLogger.Info("Updating "+kind, knv...)
		if err = r.client.Update(context.TODO(), found); err != nil {
			return Updated, err
		}
	}
	return NA, nil
}

// removeIfPresent looks for an object in the cluster matching the passed name and type (as expressed in the 'found'
// arg.) If such an object exists, and it is owned by the passed Nuxeo instance, then the object is removed.
// Otherwise cluster state is not modified.
func removeIfPresent(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, name string, namespace string,
	found runtime.Object, reqLogger logr.Logger) error {
	var kind string
	var err error
	var uids []string
	knv := []interface{}{"Namespace", namespace, "Name", name}
	if kind, err = getKind(r.scheme, found); err != nil {
		return err
	}
	if err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, found); err == nil {
		if uids, err = getOwnerRefs(found); err != nil {
			reqLogger.Error(err, "Error attempting to get owner refs for "+kind, knv...)
			return err
		} else if instance.IsOwnerUids(uids) {
			if err := r.client.Delete(context.TODO(), found); err != nil {
				reqLogger.Error(err, "Error attempting to delete "+kind, knv...)
				return err
			}
		}
	} else if !errors.IsNotFound(err) {
		reqLogger.Error(err, "Error attempting to get "+kind, knv...)
		return err
	}
	return nil
}

// getOwnerRefs returns owner reference UIDs from the passed object. If there are no owner references then an empty
// array is returned. If there is any error manipulating the passed object, a non-nil error is returned.
func getOwnerRefs(obj runtime.Object) ([]string, error) {
	if unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj); err != nil {
		return nil, err
	} else {
		if m, ok := unstructured["metadata"]; !ok {
			return nil, goerrors.New("no metadata in passed object")
		} else {
			metadata, ok := m.(map[string]interface{})
			if !ok {
				return nil, goerrors.New("unexpected metadata structure")
			}
			if ownerRefs, ok := metadata["ownerReferences"]; !ok {
				// no owner refs
				return []string{}, nil
			} else {
				var uids []string
				for _, ref := range ownerRefs.([]interface{}) {
					uid := ref.(map[string]interface{})["uid"]
					uids = append(uids, uid.(string))
				}
				return uids, nil
			}
		}
	}
}

// Gets the Kind for the passed object. Returns non-nil error if any error was encountered attempting to do that.
func getKind(scheme *runtime.Scheme, obj runtime.Object) (string, error){
	// use the scheme to get the GVK of the object then get the Kind from the GVK
	if gvk, _, err := scheme.ObjectKinds(obj); err != nil {
		return "", err
	} else if len(gvk) > 1 {
		return "", goerrors.New("unexpected result from scheme.ObjectKinds")
	} else {
		return gvk[0].Kind, nil
	}
}