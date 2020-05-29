package nuxeo

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	nuxeov1alpha1 "nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
	"testing"
	"time"
)

// proves that DeepCopyInto completely overwrites the target struct from the source struct, including
// overwriting non-nil with nil
func TestScratch(t *testing.T) {
	n := nuxeov1alpha1.Nuxeo{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       nuxeov1alpha1.NuxeoSpec{},
		Status:     nuxeov1alpha1.NuxeoStatus{},
	}
	r := ReconcileNuxeo{
		client: nil,
		scheme: &runtime.Scheme{},
	}
	d1 := r.defaultDeployment(&n)
	d2 := r.defaultDeployment(&n)
	d2.ObjectMeta.Namespace = "d2ns"
	d1.DeepCopyInto(d2)
	fmt.Printf("d1ns=%v -- dsns=%v\n", d1.ObjectMeta.Namespace, d2.ObjectMeta.Namespace)
}

func TestCompare(t *testing.T) {
	n := nuxeov1alpha1.Nuxeo{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       nuxeov1alpha1.NuxeoSpec{},
		Status:     nuxeov1alpha1.NuxeoStatus{},
	}
	r := ReconcileNuxeo{
		client: nil,
		scheme: &runtime.Scheme{},
	}
	d1 := r.defaultDeployment(&n)
	d2 := r.defaultDeployment(&n)
	d2.ObjectMeta.UID = "FOO"
	fmt.Printf("d1Hash == d2Hash? -- %v\n", util.HashObject(d1.Spec) == util.HashObject(d2.Spec))
	replicas := int32(100)
	d2.Spec.Replicas = &replicas
	fmt.Printf("d1Hash == d2Hash? -- %v\n", util.HashObject(d1.Spec) == util.HashObject(d2.Spec))
}

func TestNoNodeSet(t *testing.T) {
	n := nuxeov1alpha1.NuxeoSpec{
		NodeSets: []nuxeov1alpha1.NodeSet{{
			Name:     "",
			Type:     "",
			Replicas: 0,
			PodTemplate: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "",
					GenerateName:    "",
					Namespace:       "",
					SelfLink:        "",
					UID:             "",
					ResourceVersion: "",
					Generation:      0,
					CreationTimestamp: metav1.Time{
						Time: time.Time{},
					},
					DeletionTimestamp: &metav1.Time{
						Time: time.Time{},
					},
					DeletionGracePeriodSeconds: nil,
					Labels:                     nil,
					Annotations:                nil,
					OwnerReferences:            nil,
					Finalizers:                 nil,
					ClusterName:                "",
					ManagedFields:              nil,
				},
				Spec: corev1.PodSpec{
					Volumes:                       nil,
					InitContainers:                nil,
					Containers:                    nil,
					EphemeralContainers:           nil,
					RestartPolicy:                 "",
					TerminationGracePeriodSeconds: nil,
					ActiveDeadlineSeconds:         nil,
					DNSPolicy:                     "",
					NodeSelector:                  nil,
					ServiceAccountName:            "",
					DeprecatedServiceAccount:      "",
					AutomountServiceAccountToken:  nil,
					NodeName:                      "",
					HostNetwork:                   false,
					HostPID:                       false,
					HostIPC:                       false,
					ShareProcessNamespace:         nil,
					SecurityContext: &corev1.PodSecurityContext{
						SELinuxOptions: &corev1.SELinuxOptions{
							User:  "",
							Role:  "",
							Type:  "",
							Level: "",
						},
						WindowsOptions: &corev1.WindowsSecurityContextOptions{
							GMSACredentialSpecName: nil,
							GMSACredentialSpec:     nil,
							RunAsUserName:          nil,
						},
						RunAsUser:          nil,
						RunAsGroup:         nil,
						RunAsNonRoot:       nil,
						SupplementalGroups: nil,
						FSGroup:            nil,
						Sysctls:            nil,
					},
					ImagePullSecrets: nil,
					Hostname:         "",
					Subdomain:        "",
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: nil,
							},
							PreferredDuringSchedulingIgnoredDuringExecution: nil,
						},
						PodAffinity: &corev1.PodAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution:  nil,
							PreferredDuringSchedulingIgnoredDuringExecution: nil,
						},
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution:  nil,
							PreferredDuringSchedulingIgnoredDuringExecution: nil,
						},
					},
					SchedulerName:     "",
					Tolerations:       nil,
					HostAliases:       nil,
					PriorityClassName: "",
					Priority:          nil,
					DNSConfig: &corev1.PodDNSConfig{
						Nameservers: nil,
						Searches:    nil,
						Options:     nil,
					},
					ReadinessGates:            nil,
					RuntimeClassName:          nil,
					EnableServiceLinks:        nil,
					PreemptionPolicy:          nil,
					Overhead:                  nil,
					TopologySpreadConstraints: nil,
				},
			},
		}},
	}
	hashNodeSetPodTemplateSpec := util.HashObject(&n.NodeSets[0].PodTemplate.Spec)
	hashPodSpec := util.HashObject(&corev1.PodSpec{})
	if hashNodeSetPodTemplateSpec != hashPodSpec {
		println("Empty NodeSet does not equal Empty NodeSet")
		s1 := util.SpewObject(&n.NodeSets[0].PodTemplate.Spec)
		s2 := util.SpewObject(&corev1.PodSpec{})
		println(s1 + "\n\n\n\n" + s2)
	} else {
		println("Hooray!")
	}
}

// https://github.com/kubernetes/client-go
func TestYamlToStruct(t *testing.T) {
}

// https://github.com/lightbend/akka-cluster-operator/blob/b62b3275a4e3b97908df79b0e869c93cce88e2d6/pkg/controller/akkacluster/subset.go#L22