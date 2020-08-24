package nuxeo

import (
	"fmt"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
)

const (
	nuxeoConfVolumeName = "nuxeo-conf"
	nuxeoConfName       = "nuxeo.conf"
)

// configureConfig examines the NuxeoConfig field of the passed NodeSet and configures the passed Deployment accordingly
// by updating the Nuxeo container and Deployment. This injects configuration settings to support things like
// Java Opts, nuxeo.conf, etc. See 'NuxeoConfig' in the NodeSet for more info.
func configureConfig(dep *appsv1.Deployment, nodeSet v1alpha1.NodeSet, jvmPkiSecret corev1.Secret) error {
	var nuxeoContainer *corev1.Container
	var err error

	if nuxeoContainer, err = util.GetNuxeoContainer(dep); err != nil {
		return err
	}
	if err := configureJavaOpts(nuxeoContainer, nodeSet); err != nil {
		return err
	}
	if err := configureNuxeoTemplates(nuxeoContainer, nodeSet); err != nil {
		return err
	}
	if err := configureNuxeoPackages(nuxeoContainer, nodeSet); err != nil {
		return err
	}
	if err := configureNuxeoURL(nuxeoContainer, nodeSet); err != nil {
		return err
	}
	if err := configureNuxeoEnvName(nuxeoContainer, nodeSet); err != nil {
		return err
	}
	if err := configureJvmPki(dep, nuxeoContainer, jvmPkiSecret); err != nil {
		return err
	}
	if err := configureOfflinePackages(dep, nuxeoContainer, nodeSet); err != nil {
		return err
	}
	return nil
}

// configureJavaOpts defines a JAVA_OPTS environment variable in the passed container with a default value, or,
// with the value specified in nodeSet.NuxeoConfig.JavaOpts
func configureJavaOpts(nuxeoContainer *corev1.Container, nodeSet v1alpha1.NodeSet) error {
	env := corev1.EnvVar{
		Name:  "JAVA_OPTS",
		Value: "-XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap -XX:MaxRAMFraction=1",
	}
	if nodeSet.NuxeoConfig.JavaOpts != "" {
		env.Value = nodeSet.NuxeoConfig.JavaOpts
	}
	return util.MergeOrAddEnvVar(nuxeoContainer, env, " ")
}

// configureJavaOpts defines a NUXEO_TEMPLATES environment variable in the passed container iff
// nodeSet.NuxeoConfig.NuxeoTemplates was specified in the CR
func configureNuxeoTemplates(nuxeoContainer *corev1.Container, nodeSet v1alpha1.NodeSet) error {
	if nodeSet.NuxeoConfig.NuxeoTemplates != nil || len(nodeSet.NuxeoConfig.NuxeoTemplates) != 0 {
		env := corev1.EnvVar{
			Name:  "NUXEO_TEMPLATES",
			Value: strings.Join(nodeSet.NuxeoConfig.NuxeoTemplates, ","),
		}
		return util.OnlyAddEnvVar(nuxeoContainer, env)
	}
	return nil
}

// configureNuxeoPackages defines a NUXEO_PACKAGES environment variable in the passed container iff
// nodeSet.NuxeoConfig.NuxeoPackages was specified in the CR
func configureNuxeoPackages(nuxeoContainer *corev1.Container, nodeSet v1alpha1.NodeSet) error {
	if nodeSet.NuxeoConfig.NuxeoPackages != nil || len(nodeSet.NuxeoConfig.NuxeoPackages) != 0 {
		env := corev1.EnvVar{
			Name:  "NUXEO_PACKAGES",
			Value: strings.Join(nodeSet.NuxeoConfig.NuxeoPackages, ","),
		}
		return util.OnlyAddEnvVar(nuxeoContainer, env)
	}
	return nil
}

// configureNuxeoURL defines a NUXEO_URL environment variable in the passed container iff
// nodeSet.NuxeoConfig.NuxeoUrl was specified in the CR
func configureNuxeoURL(nuxeoContainer *corev1.Container, nodeSet v1alpha1.NodeSet) error {
	if nodeSet.NuxeoConfig.NuxeoUrl != "" {
		env := corev1.EnvVar{
			Name:  "NUXEO_URL",
			Value: nodeSet.NuxeoConfig.NuxeoUrl,
		}
		return util.OnlyAddEnvVar(nuxeoContainer, env)
	}
	return nil
}

// configureNuxeoEnvName defines a NUXEO_ENV_NAME environment variable in the passed container iff
// nodeSet.NuxeoConfig.NuxeoName was specified in the CR
func configureNuxeoEnvName(nuxeoContainer *corev1.Container, nodeSet v1alpha1.NodeSet) error {
	if nodeSet.NuxeoConfig.NuxeoName != "" {
		env := corev1.EnvVar{
			Name:  "NUXEO_ENV_NAME",
			Value: nodeSet.NuxeoConfig.NuxeoName,
		}
		return util.OnlyAddEnvVar(nuxeoContainer, env)
	}
	return nil
}

// configureNuxeoConf handles the nuxeo.conf configuration from the Nuxeo CR. The function initializes a
// volume mount, and a config map volume to reference a cluster resource holding nuxeo.conf content. This function
// only configures the volume and volume mount in the deployment. See the reconcileNuxeoConf function for the
// code that reconciles the actual ConfigMap resource. If the nodeSet.NuxeoConfig.NuxeoConf.ValueFrom
// field is initialized then the volume and mount are still initialized here, but the volume source is
// expected to have been provided by the configurer, external to the operator.
func configureNuxeoConf(instance *v1alpha1.Nuxeo, dep *appsv1.Deployment, nodeSet v1alpha1.NodeSet,
	backingNuxeoConf string, tlsNuxeoConf string) error {
	if !shouldReconNuxeoConf(nodeSet, backingNuxeoConf, tlsNuxeoConf) &&
		nodeSet.NuxeoConfig.NuxeoConf.ValueFrom == (corev1.VolumeSource{}) {
		// there is no nuxeo.conf configuration anywhere in the CR
		return nil
	}
	if shouldReconNuxeoConf(nodeSet, backingNuxeoConf, tlsNuxeoConf) &&
		nodeSet.NuxeoConfig.NuxeoConf.ValueFrom != (corev1.VolumeSource{}) {
		return fmt.Errorf("external nuxeo.conf volume source clashes with operator-managed nuxeo.conf")
	}
	if nodeSet.NuxeoConfig.NuxeoConf.ValueFrom != (corev1.VolumeSource{}) &&
		nodeSet.NuxeoConfig.NuxeoConf.ValueFrom.ConfigMap == nil &&
		nodeSet.NuxeoConfig.NuxeoConf.ValueFrom.Secret == nil {
		return fmt.Errorf("only ConfigMap and Secret volume sources are currently supported")
	}
	volMnt := corev1.VolumeMount{
		Name:      nuxeoConfVolumeName,
		ReadOnly:  false,
		MountPath: "/docker-entrypoint-initnuxeo.d/nuxeo.conf",
		SubPath:   nuxeoConfName,
	}
	if nuxeoContainer, err := util.GetNuxeoContainer(dep); err != nil {
		return err
	} else if err := addVolMnt(nuxeoContainer, volMnt); err != nil {
		return err
	}
	vol := corev1.Volume{
		Name: nuxeoConfVolumeName,
	}
	if shouldReconNuxeoConf(nodeSet, backingNuxeoConf, tlsNuxeoConf) {
		cmName := nuxeoConfCMName(instance, nodeSet.Name)
		vol.ConfigMap = &corev1.ConfigMapVolumeSource{
			DefaultMode:          util.Int32Ptr(420),
			LocalObjectReference: corev1.LocalObjectReference{Name: cmName},
			Items: []corev1.KeyToPath{{
				Key:  nuxeoConfName,
				Path: nuxeoConfName,
			}},
		}
	} else {
		// configurer is responsible to ensure that nuxeo.conf key is present in the config map volume
		// source or secret volume source.
		vol.VolumeSource = nodeSet.NuxeoConfig.NuxeoConf.ValueFrom
	}
	return util.OnlyAddVol(dep, vol)
}

// configureJvmPki adds a new - or appends to an existing - JAVA_OPTS env var in the passed container's env var
// array based on the contents of the passed secret. The secret is expected to have been provided by the configurer.
// The function looks at the following keys in the secret: keyStore, keyStoreType, keyStorePassword, trustStore,
// trustStoreType and trustStorePassword and sets the corresponding -Djavax.net.ssl... variables accordingly. For
// the keystore and truststore components of the secret, volumes and volume mounts are created like
// /etc/pki/jvm/keystore.??? and /etc/pki/jvm/truststore.??? with extensions based on store type. If no store
// type is populated in the secret then the store file will have no extension.
func configureJvmPki(dep *appsv1.Deployment, nuxeoContainer *corev1.Container, jvmPkiSecret corev1.Secret) error {
	if jvmPkiSecret.Name == "" {
		return nil
	}
	optVal, keystoreType, truststoreType, trustStoreName, keyStoreName := "", "", "", "", ""

	// key store
	if val, ok := jvmPkiSecret.Data["keyStoreType"]; ok {
		keystoreType = string(val)
		optVal += " -Djavax.net.ssl.keyStoreType=" + keystoreType
	}
	if _, ok := jvmPkiSecret.Data["keyStore"]; ok {
		keyStoreName = "keystore" + storeTypeToFileExtension(keystoreType)
		optVal += " -Djavax.net.ssl.keyStore=/etc/pki/jvm/" + keyStoreName
	}
	if val, ok := jvmPkiSecret.Data["keyStorePassword"]; ok {
		optVal += " -Djavax.net.ssl.keyStorePassword=" + string(val)
	}

	// trust store
	if val, ok := jvmPkiSecret.Data["trustStoreType"]; ok {
		truststoreType = string(val)
		optVal += " -Djavax.net.ssl.trustStoreType=" + truststoreType
	}
	if _, ok := jvmPkiSecret.Data["trustStore"]; ok {
		trustStoreName = "truststore" + storeTypeToFileExtension(truststoreType)
		optVal += " -Djavax.net.ssl.trustStore=/etc/pki/jvm/" + trustStoreName
	}
	if val, ok := jvmPkiSecret.Data["trustStorePassword"]; ok {
		optVal += " -Djavax.net.ssl.trustStorePassword=" + string(val)
	}
	env := corev1.EnvVar{
		Name:  "JAVA_OPTS",
		Value: optVal,
	}
	if err := util.MergeOrAddEnvVar(nuxeoContainer, env, " "); err != nil {
		return err
	}
	// create a volume and volume mount for the keystore/truststore if defined
	if keyStoreName != "" || trustStoreName != "" {
		jvmPkiVolMnt := corev1.VolumeMount{
			Name:      "jvm-pki",
			ReadOnly:  true,
			MountPath: "/etc/pki/jvm",
		}
		if err := util.OnlyAddVolMnt(nuxeoContainer, jvmPkiVolMnt); err != nil {
			return err
		}
		jvmPkiVol := corev1.Volume{
			Name: "jvm-pki",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: util.Int32Ptr(420),
					SecretName:  jvmPkiSecret.Name,
				}},
		}
		if trustStoreName != "" {
			jvmPkiVol.VolumeSource.Secret.Items = append(jvmPkiVol.VolumeSource.Secret.Items,
				corev1.KeyToPath{
					Key:  "trustStore",
					Path: trustStoreName,
				})
		}
		if keyStoreName != "" {
			jvmPkiVol.VolumeSource.Secret.Items = append(jvmPkiVol.VolumeSource.Secret.Items,
				corev1.KeyToPath{
					Key:  "keyStore",
					Path: keyStoreName,
				})
		}
		return util.OnlyAddVol(dep, jvmPkiVol)
	}
	return nil
}

// Given "PKCS12" (or "pkcs12"), returns ".p12", else returns storeType in lower case prefixed with a period.
// E.g. given "FOO", returns ".foo". Given "", returns "". Note that the file name of the store
// is irrelevant to Java, but by convention, most folks would expect to see .p12 or .jks in the container.
func storeTypeToFileExtension(storeType string) string {
	lower := strings.ToLower(storeType)
	if lower == "pkcs12" {
		return ".p12"
	} else if lower != "" {
		return "." + lower
	}
	return lower
}

// configureOfflinePackages creates a volume and volume mount for each marketplace package in the list of
// offline packages. The results is that each ZIP file is projected into /docker-entrypoint-initnuxeo.d in the Nuxeo
// container, causing Nuxeo to
func configureOfflinePackages(dep *appsv1.Deployment, nuxeoContainer *corev1.Container,
	nodeSet v1alpha1.NodeSet) error {
	for i, pkg := range nodeSet.NuxeoConfig.OfflinePackages {
		if pkg.ValueFrom.ConfigMap == nil && pkg.ValueFrom.Secret == nil {
			return fmt.Errorf("only ConfigMaps and Secrets are currently supported for offline packages")
		}
		mntName := "offline-package-" + strconv.Itoa(i)
		volMnt := corev1.VolumeMount{
			Name:      mntName,
			ReadOnly:  true,
			MountPath: "/docker-entrypoint-initnuxeo.d/" + pkg.PackageName,
			SubPath:   pkg.PackageName,
		}
		if err := util.OnlyAddVolMnt(nuxeoContainer, volMnt); err != nil {
			return err
		}
		vol := corev1.Volume{
			Name: mntName,
		}
		if pkg.ValueFrom.ConfigMap != nil {
			vol.ConfigMap = &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: pkg.ValueFrom.ConfigMap.Name},
				DefaultMode:          util.Int32Ptr(420),
				Items: []corev1.KeyToPath{{
					Key:  pkg.PackageName,
					Path: pkg.PackageName,
				}},
			}
		} else {
			vol.Secret = &corev1.SecretVolumeSource{
				SecretName:  pkg.ValueFrom.Secret.SecretName,
				DefaultMode: util.Int32Ptr(420),
				Items: []corev1.KeyToPath{{
					Key:  pkg.PackageName,
					Path: pkg.PackageName,
				}},
			}
		}
		if err := util.OnlyAddVol(dep, vol); err != nil {
			return err
		}
	}
	return nil
}
