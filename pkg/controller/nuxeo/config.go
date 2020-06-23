package nuxeo

import (
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
)

// handleConfig examines the NuxeoConfig field of the passed NodeSet and configures the passed Deployment accordingly
// by updating the Nuxeo container and Deployment. This injects configuration settings to support things like
// Java Opts, nuxeo.conf, etc. See 'NuxeoConfig' in the NodeSet for more info.
func handleConfig(nux *v1alpha1.Nuxeo, dep *appsv1.Deployment, nodeSet v1alpha1.NodeSet) error {
	var nuxeoContainer *corev1.Container
	var err error
	var env corev1.EnvVar

	if nuxeoContainer, err = util.GetNuxeoContainer(dep); err != nil {
		return err
	}

	// Java Opts
	env = corev1.EnvVar{
		Name:  "JAVA_OPTS",
		Value: "-XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap -XX:MaxRAMFraction=1",
	}
	if nodeSet.NuxeoConfig.JavaOpts != "" {
		env.Value = nodeSet.NuxeoConfig.JavaOpts
	}
	nuxeoContainer.Env = append(nuxeoContainer.Env, env)

	// Nuxeo Templates
	env = corev1.EnvVar{
		Name:  "NUXEO_TEMPLATES",
		Value: "default",
	}
	// todo-me it seems like if the CR configurator specified templates they should be taken exactly as
	//  provided rather than appending to 'default'
	if nodeSet.NuxeoConfig.NuxeoTemplates != nil || len(nodeSet.NuxeoConfig.NuxeoTemplates) != 0 {
		env.Value += "," + strings.Join(nodeSet.NuxeoConfig.NuxeoTemplates, ",")
	}
	nuxeoContainer.Env = append(nuxeoContainer.Env, env)

	// Nuxeo Packages
	if nodeSet.NuxeoConfig.NuxeoPackages != nil || len(nodeSet.NuxeoConfig.NuxeoPackages) != 0 {
		env = corev1.EnvVar{
			Name:  "NUXEO_PACKAGES",
			Value: strings.Join(nodeSet.NuxeoConfig.NuxeoPackages, ","),
		}
		nuxeoContainer.Env = append(nuxeoContainer.Env, env)
	}

	// Nuxeo URL
	if nodeSet.NuxeoConfig.NuxeoUrl != "" {
		env = corev1.EnvVar{
			Name:  "NUXEO_URL",
			Value: nodeSet.NuxeoConfig.NuxeoUrl,
		}
		nuxeoContainer.Env = append(nuxeoContainer.Env, env)
	}

	// Nuxeo Env Name
	if nodeSet.NuxeoConfig.NuxeoName != "" {
		env = corev1.EnvVar{
			Name:  "NUXEO_ENV_NAME",
			Value: nodeSet.NuxeoConfig.NuxeoName,
		}
		nuxeoContainer.Env = append(nuxeoContainer.Env, env)
	}

	// nuxeo.conf. If the nodeSet.NuxeoConfig.NuxeoConf.Value is defined then this represents an
	// inlined nuxeo.conf. In this case, this code initializes a volume and config map volume mount
	// to reference a ConfigMap holding the inlined nuxeo.conf content. This section of code simply
	// initializes the volume and volume mount. See the reconcileNuxeoConf function for the logic
	// that reconciles the actual ConfigMap. If the nodeSet.NuxeoConfig.NuxeoConf.ValueFrom field is
	// initialized rather than nodeSet.NuxeoConfig.NuxeoConf.Value then the volume and mount are
	// still initialized but the ConfigMap is expected to have been provided externally to the
	// operator.
	// todo-me need to completely replace nuxeo.conf? How to reconcile user-provided nuxeo.conf with
	//  auto-generated '### BEGIN - DO NOT EDIT BETWEEN BEGIN AND END ###' in the generated nuxeo.conf?
	if nodeSet.NuxeoConfig.NuxeoConf != (v1alpha1.NuxeoConfigSetting{}) {
		volMnt := corev1.VolumeMount{
			Name:      "nuxeoconf",
			ReadOnly:  false,
			MountPath: "/docker-entrypoint-initnuxeo.d",
		}
		nuxeoContainer.VolumeMounts = append(nuxeoContainer.VolumeMounts, volMnt)
		vol := corev1.Volume{
			Name: "nuxeoconf",
		}
		if nodeSet.NuxeoConfig.NuxeoConf.Value != "" {
			cmName := nux.Name + "-" + nodeSet.Name + "-nuxeo-conf"
			vol.ConfigMap = &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: cmName},
			}
		} else {
			vol.VolumeSource = nodeSet.NuxeoConfig.NuxeoConf.ValueFrom
		}
		dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes, vol)
	}
	return nil
}
