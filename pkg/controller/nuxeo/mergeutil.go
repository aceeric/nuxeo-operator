package nuxeo

import (
	goerrors "errors"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// Scans the Volume array in the passed deployment for one with a name matching the passed volume name. If not
// found in the array, then adds the passed volume to the array. If found, then compares the volume source of the
// existing volume in the array with the volume source of the passed volume. If a different volume source types
// this indicates that there are collisions on volume definition. So returns a non-nil error. Otherwise
// attempts to merge the items collection from the passed volume into the existing volume. Collisions on items are
// also returned as non-nil errors. Otherwise on completion the deployment will have one volume with all items
func addVolumeAndItems(dep *appsv1.Deployment, toAdd corev1.Volume) error{
	for _, vol := range dep.Spec.Template.Spec.Volumes {
		if vol.Name == toAdd.Name {
			var curItems *[]corev1.KeyToPath
			var toAddItems []corev1.KeyToPath
			if vol.VolumeSource.ConfigMap != nil && toAdd.VolumeSource.ConfigMap != nil {
				curItems = &vol.VolumeSource.ConfigMap.Items
				toAddItems = toAdd.VolumeSource.ConfigMap.Items
			} else if vol.VolumeSource.Secret != nil && toAdd.VolumeSource.Secret != nil {
				curItems = &vol.VolumeSource.Secret.Items
				toAddItems = toAdd.VolumeSource.Secret.Items
			} else {
				return goerrors.New("attempt to add volume " + toAdd.Name + " but a volume already exists with a different volume source")
			}
			// add keys from incoming vol to this vol
			for _, itemToAdd := range toAddItems {
				exists := false
				for i := 0; i < len(*curItems); i++ {
					if (*curItems)[i].Key == itemToAdd.Key {
						if (*curItems)[i].Path != itemToAdd.Path {
							return goerrors.New("collision on item " + itemToAdd.Key + " in volume " + toAdd.Name)
						}
						exists = true
					}
				}
				if !exists {
					*curItems = append(*curItems, itemToAdd)
				}
			}
			return nil
		}
	}
	dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes, toAdd)
	return nil
}

func addVolMnt(container *corev1.Container, mntToAdd corev1.VolumeMount) error {
	for _, mnt := range container.VolumeMounts {
		if mnt.Name == mntToAdd.Name {
			if !reflect.DeepEqual(mnt, mntToAdd) {
				return goerrors.New("collision trying to add volume mount " + mntToAdd.Name)
			}
			return nil // already present
		}
	}
	container.VolumeMounts = append(container.VolumeMounts, mntToAdd)
	return nil
}