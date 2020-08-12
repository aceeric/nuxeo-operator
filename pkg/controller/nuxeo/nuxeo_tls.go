package nuxeo

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"nuxeo-operator/pkg/util"
)

// configureNuxeoForTLS configures Nuxeo to terminate TLS as follows:
//  1) Creates a volume and volume mount referencing the keystore.jks key from the passed secret name
//  2) Creates env var TLS_KEYSTORE_PASS referencing keystorePass key in the same secret
//  3) Adds an https entry to the NUXEO_TEMPLATES env var
//  4) Returns entries to be merged into nuxeo.conf
func configureNuxeoForTLS(dep *appsv1.Deployment, tlsSecret string) (string, error) {
	var nuxeoContainer *corev1.Container
	var err error
	if nuxeoContainer, err = util.GetNuxeoContainer(dep); err != nil {
		return "", err
	}
	keystoreVol := corev1.Volume{
		Name: "tls-keystore",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  tlsSecret,
				DefaultMode: util.Int32Ptr(420),
				Items: []corev1.KeyToPath{{
					Key:  "keystore.jks",
					Path: "keystore.jks",
				}},
			}},
	}
	if err := util.OnlyAddVol(dep, keystoreVol); err != nil {
		return "", err
	}
	keystoreVolMnt := corev1.VolumeMount{
		Name:      "tls-keystore",
		ReadOnly:  true,
		MountPath: "/etc/secrets/tls-keystore",
	}
	if err := util.OnlyAddVolMnt(nuxeoContainer, keystoreVolMnt); err != nil {
		return "", err
	}
	// TLS_KEYSTORE_PASS env var
	keystorePassEnv := corev1.EnvVar{
		Name: "TLS_KEYSTORE_PASS",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: tlsSecret},
				Key:                  "keystorePass",
			},
		},
	}
	if err := util.OnlyAddEnvVar(nuxeoContainer, keystorePassEnv); err != nil {
		return "", err
	}

	// NUXEO_TEMPLATES env var
	templatesEnv := corev1.EnvVar{
		Name:  "NUXEO_TEMPLATES",
		Value: "https",
	}
	tlsConfig := "nuxeo.server.https.port=8443\n" +
		"nuxeo.server.https.keystoreFile=/etc/secrets/tls-keystore/keystore.jks\n" +
		"nuxeo.server.https.keystorePass=${env:TLS_KEYSTORE_PASS}\n"
	return tlsConfig, util.MergeOrAddEnvVar(nuxeoContainer, templatesEnv, ",")
}
