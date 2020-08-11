package nuxeo

import (
	goerrors "errors"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// Adds a volume with projections to the passed deployment, or adds source from passed volume to existing volume
// projection in deployment, or adds keys from passed projection source to sources in existing volume projection.
// When adding keys, if an incoming volume source has an item like {key: x, path: y} and a matching source in
// an existing volume has key like {key: x, path: z} then an error is returned.
func addVolumeProjectionAndItems(dep *appsv1.Deployment, toAdd corev1.Volume) error {
	if toAdd.Projected == nil || len(toAdd.Projected.Sources) != 1 {
		return goerrors.New("exactly one projection is supported for " + toAdd.Name)
	}
	for _, vol := range dep.Spec.Template.Spec.Volumes {
		if vol.Name == toAdd.Name {
			// volume exists
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
					// source does not exist, so add
					vol.Projected.Sources = append(vol.Projected.Sources, toAdd.Projected.Sources[0])
					return nil
				}
			}
		}
	}
	// no matching volume so add
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
