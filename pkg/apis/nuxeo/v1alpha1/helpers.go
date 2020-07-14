package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// IsOwner inspects the passed object metadata for an owner reference to the Nuxeo CR in the receiver.
// If the Nuxeo receiver UID is found in any of the passed metadata owner references, then the function returns
// true - meaning Nuxeo is an owner of the object. Else false is returned
func (nux *Nuxeo) IsOwner(objMeta metav1.ObjectMeta) bool {
	for _, ref := range objMeta.OwnerReferences {
		if ref.UID == nux.UID {
			return true
		}
	}
	return false
}

// same as IsOwner except takes an array of UIDs
func (nux *Nuxeo) IsOwnerUids(uids []string) bool {
	for _, uid := range uids {
		if uid == string(nux.UID) {
			return true
		}
	}
	return false
}