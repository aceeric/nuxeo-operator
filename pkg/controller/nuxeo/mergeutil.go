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
//
// Deprecated: in favor of addVolumeProjectionAndItems
func addVolumeAndItems(dep *appsv1.Deployment, toAdd corev1.Volume) error {
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

// addVolumeProjectionAndItems builds up a volume with a projection source by adding new sources and merging
// keys to existing sources. The passed volume must define exactly one projection source or an error is returned.
// When merging items, if an incoming volume source has an item like {key: x, path: y} and an matching source in
// an existing volume has key like {key: x, path: z} then an error is returned.
func addVolumeProjectionAndItems(dep *appsv1.Deployment, toAdd corev1.Volume) error {
	if toAdd.Projected == nil || len(toAdd.Projected.Sources) != 1 {
		return goerrors.New("exactly one projection is supported for " + toAdd.Name)
	}
	for _, vol := range dep.Spec.Template.Spec.Volumes {
		if vol.Name == toAdd.Name {
			if vol.Projected == nil {
				return goerrors.New("attempt to merge projected volume " + toAdd.Name + " into non-projected volume " + vol.Name)
			}
			for _, src := range vol.Projected.Sources {
				if curItems, toAddItems, same := sameSrc(src, toAdd.Projected.Sources[0]); same {
					// add keys from incoming vol projection
					for _, itemToAdd := range *toAddItems {
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
				} else {
					vol.Projected.Sources = append(vol.Projected.Sources, toAdd.Projected.Sources[0])
					return nil
				}
			}
		}
	}
	dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes, toAdd)
	return nil
}

// returns true if the two volume projections are the same type (e.g. Secret) and same name, and returns pointers to
// each projection's Items array if same. Otherwise returns nil Items pointers and false.
func sameSrc(src1 corev1.VolumeProjection, src2 corev1.VolumeProjection) (*[]corev1.KeyToPath, *[]corev1.KeyToPath, bool) {
	if src1.Secret != nil && src2.Secret != nil && src1.Secret.Name == src2.Secret.Name {
		return &src1.Secret.Items, &src2.Secret.Items, true
	} else if src1.ConfigMap != nil && src2.ConfigMap != nil && src1.ConfigMap.Name == src2.ConfigMap.Name {
		return &src1.ConfigMap.Items, &src2.ConfigMap.Items, true
	} else {
		return nil, nil, false
	}
}

// if the passed container does not already have the passed mount, then the passed mount is added. If the container
// does have the passed mount, and the mounts are identical, no action is taken. Otherwise an error is returned.
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
