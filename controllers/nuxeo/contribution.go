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
	"fmt"
	"strings"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/aceeric/nuxeo-operator/controllers/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// configureContributions injects custom contributions into the Nuxeo container from Kubernetes storage resources
// like Secrets, ConfigMaps, and other Volume Sources. These are added to the NUXEO_TEMPLATES environment variable
// in the Deployment descriptor. As a result. when Nuxeo starts, these contributions will be merged into the
// nuxeo properties and the contributions will go into /opt/nuxeo/server/nxserver/config when Nuxeo starts.
func (r *NuxeoReconciler) configureContributions(instance *v1alpha1.Nuxeo, dep *appsv1.Deployment, nodeSet v1alpha1.NodeSet) error {
	var err error
	var nuxeoContainer *corev1.Container
	if len(nodeSet.Contributions) == 0 {
		return nil
	}
	// verify either Secret/ConfigMap Volume Sources OR other types of Volume Sources
	cfgMapSecretCnt, nonCfgMapSecretCnt := 0, 0
	for _, contrib := range nodeSet.Contributions {
		if contrib.VolumeSource.ConfigMap != nil || contrib.VolumeSource.Secret != nil {
			if len(contrib.Templates) != 1 {
				return fmt.Errorf("ConfigMap/Secret contributions can only supply one template name")
			}
			cfgMapSecretCnt += 1
		} else {
			nonCfgMapSecretCnt += 1
		}
	}
	if cfgMapSecretCnt != 0 && nonCfgMapSecretCnt != 0 {
		return fmt.Errorf("cannot define both ConfigMap/Secret contributions and non-ConfigMap/Secret contributions")
	} else if nonCfgMapSecretCnt > 1 {
		return fmt.Errorf("only one non-ConfigMap/Secret contribution is supported")
	}
	if nuxeoContainer, err = GetNuxeoContainer(dep); err != nil {
		return err
	}
	var templates []string
	for _, contrib := range nodeSet.Contributions {
		templates = append(templates, contrib.Templates...)
		if contrib.VolumeSource.ConfigMap != nil {
			err = r.configureCmContrib(dep, instance.Namespace, nuxeoContainer, contrib.Templates[0], contrib.VolumeSource.ConfigMap.Name)
		} else if contrib.VolumeSource.Secret != nil {
			err = r.configureSecretContrib(dep, instance.Namespace, nuxeoContainer, contrib.Templates[0], contrib.VolumeSource.Secret.SecretName)
		} else {
			// only one of these is supported for now
			err = configureVolDeployment(dep, contrib.VolumeSource, nuxeoContainer)
		}
		if err != nil {
			return err
		}
	}
	// configure NUXEO_TEMPLATES with absolute path refs to where the contributions are mounted:
	// /etc/nuxeo/nuxeo-operator-config
	for i := 0; i < len(templates); i++ {
		templates[i] = "/etc/nuxeo/nuxeo-operator-config/" + templates[i]
	}
	templatesEnv := corev1.EnvVar{
		Name:  "NUXEO_TEMPLATES",
		Value: strings.Join(templates, ","),
	}
	return util.MergeOrAddEnvVar(nuxeoContainer, templatesEnv, ",")
}

// If the Volume Source containing the contribution is a ConfigMap then gets all the keys from the resource
// because each will have to be explicitly configured in the Volume specifier. Then calls 'configureDeployment' to
// actually configure the passed Deployment.
func (r *NuxeoReconciler) configureCmContrib(dep *appsv1.Deployment, namespace string, nuxeoContainer *corev1.Container,
	contribName string, cmName string) error {
	cm := &corev1.ConfigMap{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: cmName, Namespace: namespace}, cm); err != nil {
		return err
	}
	var keys []string
	for k, _ := range cm.Data {
		keys = append(keys, k)
	}
	return configureDeployment(dep, cm, nuxeoContainer, contribName, cmName, keys)
}

// If the Volume Source containing the contribution is a Secret then gets all the keys from the resource
// because each will have to be explicitly configured in the Volume specifier. Then calls 'configureDeployment' to
// actually configure the passed Deployment.
func (r *NuxeoReconciler) configureSecretContrib(dep *appsv1.Deployment, namespace string, nuxeoContainer *corev1.Container,
	contribName string, secretName string) error {
	secret := &corev1.Secret{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: namespace}, secret); err != nil {
		return err
	}
	var keys []string
	for k, _ := range secret.Data {
		keys = append(keys, k)
	}
	return configureDeployment(dep, secret, nuxeoContainer, contribName, secretName, keys)
}

// Configures the passed Deployment when the contribution Volume Source is other than Secret/ConfigMap because
// the entire volume is mounted as a unit and only the NUXEO_TEMPLATES is used to add the contribution to Nuxeo.
// Presently, only one of these is supported with the expectation that all contributions will be present in the
// volume and so multiple volumes aren't needed
func configureVolDeployment(dep *appsv1.Deployment, volSrc corev1.VolumeSource, nuxeoContainer *corev1.Container) error {
	volMnt := corev1.VolumeMount{
		Name:      "nuxeo-operator-config",
		ReadOnly:  true,
		MountPath: "/etc/nuxeo/nuxeo-operator-config",
	}
	if err := util.OnlyAddVolMnt(nuxeoContainer, volMnt); err != nil {
		return err
	}
	vol := corev1.Volume{
		Name:         "nuxeo-operator-config",
		VolumeSource: volSrc,
	}
	return util.OnlyAddVol(dep, vol)
}

// Configures the deployment by adding Volumes and Volume Mounts
func configureDeployment(dep *appsv1.Deployment, typ interface{}, nuxeoContainer *corev1.Container,
	contribName string, volSrc string, keys []string) error {
	volMnt := corev1.VolumeMount{
		Name:      "nuxeo-operator-config-" + contribName,
		ReadOnly:  true,
		MountPath: "/etc/nuxeo/nuxeo-operator-config/" + contribName,
	}
	if err := util.OnlyAddVolMnt(nuxeoContainer, volMnt); err != nil {
		return err
	}
	vol := corev1.Volume{
		Name: "nuxeo-operator-config-" + contribName,
	}
	if _, ok := typ.(*corev1.ConfigMap); ok {
		vol.ConfigMap = &corev1.ConfigMapVolumeSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: volSrc},
			DefaultMode:          util.Int32Ptr(420),
		}
		mapKeysToItems(keys, &vol.ConfigMap.Items)
	} else if _, ok := typ.(*corev1.Secret); ok {
		vol.Secret = &corev1.SecretVolumeSource{
			SecretName:  volSrc,
			DefaultMode: util.Int32Ptr(420),
		}
		mapKeysToItems(keys, &vol.Secret.Items)
	} else {
		return fmt.Errorf("configureDeployment only valid for ConfigMap and Secret volume sources")
	}
	return util.OnlyAddVol(dep, vol)
}

// mapKeysToItems takes the passed keys (from a ConfigMap or Secret volume source) that each represent
// individual files for a contribution, like "nuxeo.defaults" or "my-config.xml", and it sets the key/path
// pairs in the passed 'items' ref. Basically what it does is the nuxeo.defaults key gets a path nuxeo.defaults
// (i.e. in the contribution root subdirectory) and every other key gets a path nxserver/config/ + key which is
// where Nuxeo expects to find the files that it will copy into /opt/nuxeo/server/nxserver/config.
//
// The passed 'items' array is modified by the function
func mapKeysToItems(keys []string, items *[]corev1.KeyToPath) {
	for _, key := range keys {
		path := key
		if key != "nuxeo.defaults" {
			path = "nxserver/config/" + key
		}
		*items = append(*items, corev1.KeyToPath{
			Key:  key,
			Path: path,
		})
	}
}
