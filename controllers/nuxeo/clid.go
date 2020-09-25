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
	"fmt"
	"strings"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/aceeric/nuxeo-operator/controllers/common"
	"github.com/aceeric/nuxeo-operator/controllers/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	nuxeoClidConfigMapName = "nuxeo-clid"
	clidKey                = "instance.clid"
	clidSeparator          = "--"
)

// configureClid configures the passed Deployment with a Volume and VolumeMount to project the CLID into the
// Nuxeo container at a hard-coded mount point: /var/lib/nuxeo/data/instance.clid. The volume references
// a hard-coded ConfigMap name managed by the operator: "nuxeo-clid". See the reconcileClid() function
// for the code that reconciles that actual ConfigMap.
func configureClid(instance *v1alpha1.Nuxeo, dep *appsv1.Deployment) error {
	if instance.Spec.Clid == "" {
		return nil
	}
	if nuxeoContainer, err := GetNuxeoContainer(dep); err != nil {
		return err
	} else {
		volMnt := corev1.VolumeMount{
			Name:      nuxeoClidConfigMapName,
			ReadOnly:  true,
			MountPath: "/var/lib/nuxeo/data/" + clidKey,
			SubPath:   clidKey,
		}
		if err := util.OnlyAddVolMnt(nuxeoContainer, volMnt); err != nil {
			return err
		}
		vol := corev1.Volume{
			Name: nuxeoClidConfigMapName,
		}
		vol.ConfigMap = &corev1.ConfigMapVolumeSource{
			DefaultMode:          util.Int32Ptr(420),
			LocalObjectReference: corev1.LocalObjectReference{Name: nuxeoClidConfigMapName},
			Items: []corev1.KeyToPath{{
				Key:  clidKey,
				Path: clidKey,
			}},
		}
		util.AnnotateTemplate(dep, common.ClidHashAnnotation, util.CRC(instance.Spec.Clid))
		return util.OnlyAddVol(dep, vol)
	}
}

// reconcileClid creates, updates, or deletes the CLID ConfigMap. If the Clid is specified in the CR, then the
// corresponding CM is added/updated in the cluster. If Clid is not specified, then it is removed from the
// cluster if present
func (r *NuxeoReconciler) reconcileClid(instance *v1alpha1.Nuxeo) error {
	if instance.Spec.Clid != "" {
		if expected, err := r.defaultClidCM(instance, instance.Spec.Clid); err != nil {
			return err
		} else {
			_, err := r.addOrUpdate(nuxeoClidConfigMapName, instance.Namespace, expected, &corev1.ConfigMap{},
				util.ConfigMapComparer)
			return err
		}
	} else {
		return r.removeIfPresent(instance, nuxeoClidConfigMapName, instance.Namespace, &corev1.ConfigMap{})
	}
}

// defaultClidCM creates and returns a ConfigMap struct named "nuxeo-clid" to hold the passed CLID string. The CLID
// string has to conform to the format that you would get from the Nuxeo registration site. Specifically it has to
// contain the double dash separator that Nuxeo uses to split the single CLID into two lines. This function will
// split the clid and newline format it so it has the correct two-line format in the CLID file.
func (r *NuxeoReconciler) defaultClidCM(instance *v1alpha1.Nuxeo, clidValue string) (*corev1.ConfigMap, error) {
	if len(strings.Split(clidValue, clidSeparator)) != 2 {
		return nil, fmt.Errorf("configured CLID value '%v' does not contain required separator '%v'",
			clidValue, clidSeparator)
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nuxeoClidConfigMapName,
			Namespace: instance.Namespace,
		},
		Data: map[string]string{clidKey: strings.Replace(clidValue, clidSeparator, "\n", 1)},
	}
	_ = controllerutil.SetControllerReference(instance, cm, r.Scheme)
	return cm, nil
}
