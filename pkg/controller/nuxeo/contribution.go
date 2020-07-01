package nuxeo

import (
	"context"
	goerrors "errors"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
)

// TODO-ME HAVE TO ADD TO NUXEO_TEMPLATES

func configureContributions(r *ReconcileNuxeo, nux *v1alpha1.Nuxeo, dep *appsv1.Deployment, nodeSet v1alpha1.NodeSet) error {
	var err error
	var nuxeoContainer *corev1.Container
	if len(nodeSet.Contributions) == 0 {
		return nil
	}
	cfgMapSecretCnt, nonCfgMapSecretCnt := 0, 0
	for _, contrib := range nodeSet.Contributions {
		if contrib.VolumeSource.ConfigMap != nil || contrib.VolumeSource.Secret != nil {
			if len(contrib.Templates) != 1 {
				return goerrors.New("ConfigMap/Secret contributions can only supply one template name")
			}
			cfgMapSecretCnt += 1
		} else {
			nonCfgMapSecretCnt += 1
		}
	}
	if cfgMapSecretCnt != 0 && nonCfgMapSecretCnt != 0 {
		return goerrors.New("cannot define both ConfigMap/Secret contributions and non-ConfigMap/Secret contributions")
	} else if nonCfgMapSecretCnt > 1 {
		return goerrors.New("only one non-ConfigMap/Secret contribution is supported")
	}
	if nuxeoContainer, err = util.GetNuxeoContainer(dep); err != nil {
		return err
	}

	var templates []string
	for _, contrib := range nodeSet.Contributions {
		templates = append(templates, contrib.Templates...)
		if contrib.VolumeSource.ConfigMap != nil {
			err = configureCmContrib(r, dep, nux.Namespace, nuxeoContainer, contrib.Templates[0], contrib.VolumeSource.ConfigMap.Name)
		} else if contrib.VolumeSource.Secret.SecretName != "" {
			err = configureSecretContrib(r, dep, nux.Namespace, nuxeoContainer, contrib.Templates[0], contrib.VolumeSource.Secret.SecretName)
		} else {
			return goerrors.New("NOT IMPLEMENTED YET!")
		}
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(templates); i++ {
		templates[i] = "/etc/nuxeo/nuxeo-operator-config/" + templates[i]
	}
	templatesEnv := corev1.EnvVar{
		Name:  "NUXEO_TEMPLATES",
		Value: strings.Join(templates, ","),
	}
	return util.MergeOrAdd(nuxeoContainer, templatesEnv, ",")
}

func configureCmContrib(r *ReconcileNuxeo, dep *appsv1.Deployment, namespace string, nuxeoContainer *corev1.Container,
	contribName string, cmName string) error {
	cm := &corev1.ConfigMap{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: cmName, Namespace: namespace}, cm); err != nil {
		return err
	}
	var keys []string
	for k, _ := range cm.Data {
		keys = append(keys, k)
	}
	return configureDeployment(dep, cm, nuxeoContainer, contribName, cmName, keys)
}

func configureSecretContrib(r *ReconcileNuxeo, dep *appsv1.Deployment, namespace string, nuxeoContainer *corev1.Container,
	contribName string, secretName string) error {
	secret := &corev1.Secret{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: namespace}, secret); err != nil {
		return err
	}
	var keys []string
	for k, _ := range secret.Data {
		keys = append(keys, k)
	}
	return configureDeployment(dep, secret, nuxeoContainer, contribName, secretName, keys)
}

func configureDeployment(dep *appsv1.Deployment, typ interface{}, nuxeoContainer *corev1.Container,
	contribName string, volSrc string, keys []string) error {
	volMnt := corev1.VolumeMount{
		Name:      "nuxeo-operator-config-" + contribName,
		ReadOnly:  true,
		MountPath: "/etc/nuxeo/nuxeo-operator-config/" + contribName,
		//SubPath:   contribName,
	}
	nuxeoContainer.VolumeMounts = append(nuxeoContainer.VolumeMounts, volMnt)
	vol := corev1.Volume{
		Name: "nuxeo-operator-config-" + contribName,
	}
	if _, ok := typ.(*corev1.ConfigMap); ok {
		vol.ConfigMap = &corev1.ConfigMapVolumeSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: volSrc},
			Items:                []corev1.KeyToPath{},
		}
		for _, key := range keys {
			if key == "nuxeo.defaults" {
				vol.ConfigMap.Items = append(vol.ConfigMap.Items, corev1.KeyToPath{
					Key:  key,
					Path: key,
				})
			} else {
				vol.ConfigMap.Items = append(vol.ConfigMap.Items, corev1.KeyToPath{
					Key:  key,
					Path: "nxserver/config/" + key,
				})
			}
		}
	} else if _, ok := typ.(*corev1.Secret); ok {
		vol.Secret = &corev1.SecretVolumeSource{
			SecretName: volSrc,
			Items:      []corev1.KeyToPath{},
		}
		for _, key := range keys {
			if key == "nuxeo.defaults" {
				vol.Secret.Items = append(vol.Secret.Items, corev1.KeyToPath{
					Key:  key,
					Path: key,
				})
			} else {
				vol.Secret.Items = append(vol.Secret.Items, corev1.KeyToPath{
					Key:  key,
					Path: "nxserver/config/" + key,
				})
			}
		}
	} else {
		// todo-me
	}
	dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes, vol)
	return nil
}
