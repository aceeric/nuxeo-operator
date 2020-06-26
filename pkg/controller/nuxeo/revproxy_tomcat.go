package nuxeo

import (
	"bytes"
	goerrors "errors"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
)

// configureTomcatForTLS configures Tomcat as the reverse proxy as follows:
//  1) Creates a volume and volume mount referencing the keystore.jks key from the secret in the passed
//     tomcat rev proxy spec
//  2) Creates env var TOMCAT_KEYSTORE_PASS referencing keystorePass in the same secret
//  3) Configures the NUXEO_CUSTOM_PARAM env var to reference the keystore and password. Nuxeo will incorporate
//     this into nuxeo.conf when it starts the server
//  4) Adds an https entry to the NUXEO_TEMPLATES env var
func configureTomcatForTLS(dep *appsv1.Deployment, tomcat v1alpha1.TomcatRevProxySpec) error {
	var nuxeoContainer *corev1.Container
	var err error
	if nuxeoContainer, err = util.GetNuxeoContainer(dep); err != nil {
		return err
	}
	keystoreVol := corev1.Volume{
		Name: "tomcat-keystore",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: tomcat.Secret,
				Items: []corev1.KeyToPath{{
					Key:  "keystore.jks",
					Path: "keystore.jks",
				}},
			}},
	}
	dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes, keystoreVol)
	keystoreVolMnt := corev1.VolumeMount{
		Name:      "tomcat-keystore",
		ReadOnly:  true,
		MountPath: "/etc/secrets/tomcat_keystore",
	}
	nuxeoContainer.VolumeMounts = append(nuxeoContainer.VolumeMounts, keystoreVolMnt)

	// TOMCAT_KEYSTORE_PASS env var
	keystorePassEnv := getEnv(nuxeoContainer, "TOMCAT_KEYSTORE_PASS")
	if keystorePassEnv != nil {
		return goerrors.New("TOMCAT_KEYSTORE_PASS already defined - operator cannot override")
	} else {
		keystorePassEnv = &corev1.EnvVar{
			Name: "TOMCAT_KEYSTORE_PASS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: tomcat.Secret},
					Key:                  "keystorePass",
				},
			},
		}
		nuxeoContainer.Env = append(nuxeoContainer.Env, *keystorePassEnv)
	}

	// NUXEO_CUSTOM_PARAM env var (refs TOMCAT_KEYSTORE_PASS set above). We use this environment var
	// because the configurer can also specify nuxeo.conf and if they did, it could clash with this code path
	// if this code path also tried to use nuxeo.conf
	tomcatConfig := map[string]string{
		"nuxeo.server.https.port":         "8443",
		"nuxeo.server.https.keystoreFile": "/etc/secrets/tomcat_keystore/keystore.jks",
		"nuxeo.server.https.keystorePass": "${env:TOMCAT_KEYSTORE_PASS}",
	}
	customParamEnv := getEnv(nuxeoContainer, "NUXEO_CUSTOM_PARAM")
	if customParamEnv == nil {
		customParamEnv = &corev1.EnvVar{
			Name:  "NUXEO_CUSTOM_PARAM",
			Value: mapToStr(tomcatConfig, "\n"),
		}
	} else {
		if customParamEnv.ValueFrom != nil {
			// if the configurer defines an external source for NUXEO_CUSTOM_PARAM then the operator cannot
			// touch that. Since the operator uses that environment variable to configure Tomcat for Nuxeo 10.x
			// the operator cannot proceed with the configuration
			return goerrors.New("operator Tomcat TLS config conflicts with externally defined NUXEO_CUSTOM_PARAM env var")
		}
		for key, _ := range tomcatConfig {
			if strings.Contains(customParamEnv.Value, key) {
				return goerrors.New("NUXEO_CUSTOM_PARAM already defines configuration: " + key)
			}
		}
		customParamEnv.Value += "\n" + mapToStr(tomcatConfig, "\n")
	}
	nuxeoContainer.Env = append(nuxeoContainer.Env, *customParamEnv)

	// NUXEO_TEMPLATES env var
	templatesEnv := getEnv(nuxeoContainer, "NUXEO_TEMPLATES")
	if templatesEnv == nil {
		templatesEnv = &corev1.EnvVar{
			Name:  "NUXEO_TEMPLATES",
			Value: "https",
		}
		nuxeoContainer.Env = append(nuxeoContainer.Env, *templatesEnv)
	} else {
		if templatesEnv.ValueFrom != nil {
			// same as above
			return goerrors.New("operator Tomcat TLS config conflicts with externally defined NUXEO_TEMPLATES env var")
		}
		templatesEnv.Value += ",https"
	}
	return nil
}

// getEnv searches the environment variable array in the passed container for an env var with the passed name.
// If found, returns a ref to the env var, else returns nil.
func getEnv(container *corev1.Container, envName string) *corev1.EnvVar {
	for i := 0; i < len(container.Env); i++ {
		if container.Env[i].Name == envName {
			return &container.Env[i]
		}
	}
	return nil
}

// mapToStr takes map A:1,B:2 and returns string "A=1\nB=2\n"
func mapToStr(cfg map[string]string, delim string) string {
	b := new(bytes.Buffer)
	for key, value := range cfg {
		_, _ = fmt.Fprintf(b, "%s=%s%s", key, value, delim)
	}
	return b.String()
}
