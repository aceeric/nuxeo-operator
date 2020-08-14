package nuxeo

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
)

// configureProbes adds liveness and readiness probes to the Nuxeo container spec in the passed deployment. If probes
// are defined in the passed NodeSet spec then they are used. (If thresholds in the provided probes are not specified
// they are defaulted.) If no explicit probe is defined in the Nuxeo CR, then probes are defaulted:
//  httpGet:
//    path: /nuxeo/runningstatus
//    port: 8080 (or 8443)
//    scheme: HTTP (or HTTPS)
//  initialDelaySeconds: 15
//  timeoutSeconds: 3
//  periodSeconds: 10
//  successThreshold: 1
//	failureThreshold: 3
func configureProbes(dep *appsv1.Deployment, nodeSet v1alpha1.NodeSet) error {
	if nuxeoContainer, err := util.GetNuxeoContainer(dep); err != nil {
		return err
	} else {
		useHttps := false
		if nodeSet.NuxeoConfig.TlsSecret != "" {
			// if Nuxeo is going to terminate TLS, then it will be listening on HTTPS:8443. Otherwise Nuxeo
			// listens on HTTP:8080. This affects how the probes are configured immediately below.
			useHttps = true
		}
		nuxeoContainer.LivenessProbe = defaultProbe(useHttps)
		if nodeSet.LivenessProbe != nil {
			nodeSet.LivenessProbe.DeepCopyInto(nuxeoContainer.LivenessProbe)
			setProbeDefaults(nuxeoContainer.LivenessProbe)
		}
		nuxeoContainer.ReadinessProbe = defaultProbe(useHttps)
		if nodeSet.ReadinessProbe != nil {
			nodeSet.ReadinessProbe.DeepCopyInto(nuxeoContainer.ReadinessProbe)
			setProbeDefaults(nuxeoContainer.ReadinessProbe)
		}
		return nil
	}
}

// defaultProbe creates - and returns a pointer to - a default liveness/readiness probe struct. If useHttps is passed
// as true then the probe is configured to use HTTPS port 8443, else HTTP port 8080. Per Kubernetes spec, If the
// scheme field is set to HTTPS, the kubelet sends an HTTPS request skipping certificate verification. So this
// probe works even with Nuxeo terminating TLS using self-signed certs in a test-style environment.
func defaultProbe(useHttps bool) *corev1.Probe {
	scheme := corev1.URISchemeHTTP
	port := int32(8080)
	if useHttps {
		scheme = corev1.URISchemeHTTPS
		port = 8443
	}
	return &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/nuxeo/runningstatus",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: port,
				},
				Scheme: scheme,
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      3,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

// Simplifies reconciliation since the probes are embedded in the Deployment. If nothing is provided for these
// thresholds then Kubernetes will assign defaults (another case where Kubernetes takes partial ownership of the
// resource spec.) That adds complexity to the reconciliation so - explicit values are assigned here so Kubernetes
// will leave them alone.
func setProbeDefaults(probe *corev1.Probe) {
	util.SetInt32If(&probe.InitialDelaySeconds, 0, 15)
	util.SetInt32If(&probe.TimeoutSeconds, 0, 3)
	util.SetInt32If(&probe.PeriodSeconds, 0, 10)
	util.SetInt32If(&probe.SuccessThreshold, 0, 1)
	util.SetInt32If(&probe.FailureThreshold, 0, 3)
}
