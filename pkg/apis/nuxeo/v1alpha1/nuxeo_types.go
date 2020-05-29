package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NodeSet defines the structure of the Nuxeo cluster. Each NodeSet results in a Deployment. This supports the
// capability to defined different configurations for a deployment of interactive Nuxeo nodes vs a deployment
// of worker Nuxeo nodes.
type NodeSet struct {
	// The name of this node set. In cases where only one node set is needed, a recommended naming strategy is
	// to name this node set 'cluster'. For example, say you generate a Nuxeo CR named 'my-nuxeo' into the namespace
	// being watched by the Nuxeo Operator, and you name this node set 'cluster'. Then the operator will create
	// a deployment from the node set named 'my-nuxeo-cluster'
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Optional
	// Indicates whether this NodeSet will be accessible outside the cluster. Default is 'false'. If 'true', then
	// the Service created by the operator will be have its selectors defined such that it finds the Pods
	// created by this NodeSet. Exactly one NodeSet must be configured for external access.
	// TODO does this make sense? Might someone want to stand up a non-interactive cluster for testing?
	Interactive bool `json:"interactive,omitempty"`

	// +kubebuilder:validation:Minimum=1
	// Populates the 'spec.replicas' property of the 'Deployment' generated by this node set
	Replicas int `json:"replicas,omitempty"`

	// PodTemplate provides the ability to override hard-coded pod defaults
	// +kubebuilder:validation:Optional
	PodTemplate corev1.PodTemplateSpec `json:"podTemplate,omitempty"`
}

// Provides the ability to minimally customize the the type of Service generated by the Operator
type ServiceSpec struct {
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	// Type specifies the Service type to create
	Type corev1.ServiceType `json:"type,omitempty"`
	// Port specifies the port exposed by the service
	Port int32 `json:"port,omitempty"`
	// TargetPort specifies the port that the servuce will use internally to communicate with the Nuxeo cluster
	TargetPort int32 `json:"targetPort,omitempty"`
}

// NuxeoAccess supports creation of an OpenShift Route supporting access to the Nuxeo Service from outside of the
// cluster. TODO Kubernetes Ingress
type NuxeoAccess struct {
	// Hostname specifies the host name
	Hostname string `json:"hostname,omitempty"`

	// +kubebuilder:validation:Optional
	// TargetPort selects a target port in the Service backed by this NuxeoAccess spec. By default, 'web' is
	// populated by the Operator - which finds the default 'web' port in the Service generated by the Operator
	TargetPort intstr.IntOrString `json:"targetPort,omitempty"`

	// +kubebuilder:validation:Optional
	// TLSSecret specifies the name of a secret with fields required to configure ingress for TLS, as determined by
	// the 'Termination' field. Example fields expected in such a secret are - 'key', 'certificate', and
	// 'caCertificate'. If this element is specified, then TLS is enabled. Otherwise, TLS is not enabled and the
	// 'Termination' property is ignored.
	TLSSecret string `json:"tlsSecret,omitempty"`

	// +kubebuilder:validation:Optional
	// Termination specifies the TLS termination type. E.g. 'edge', 'passthrough', etc. Ignored unless 'TLSSecret'
	// is specified.
	Termination routev1.TLSTerminationType `json:"termination,omitempty"`
}

// NginxRevProxySpec defines the configuration elements needed for the Nginx reverse proxy. The config map specifies
// a CM resources that contains am 'nginx.conf' key, and a 'proxy.conf' key, each of which provide required
// configuration to the Nginx container. The Secret field references a secret containing keys 'tls.key', 'tls.cert',
// and 'dhparam'.
type NginxRevProxySpec struct {
	ConfigMap string `json:"configMap,omitempty"`
	Secret    string `json:"secret,omitempty"`
}

// DummyRevProxySpec supports testing
type DummyRevProxySpec struct{}

// RevProxySpec defines the reverse proxies supported by the Nuxeo Operator. Details are provided in the individual
// structs.
type RevProxySpec struct {
	// +kubebuilder:validation:Optional
	// nginx supports configuration of Nginx as the reverse proxy
	Nginx NginxRevProxySpec `json:"nginx,omitempty"`

	// +kubebuilder:validation:Optional
	// dummy supports testing
	Dummy DummyRevProxySpec `json:"dummy,omitempty"`
}

// NuxeoSpec Defines the desired state of a Nuxeo cluster
type NuxeoSpec struct {

	// +kubebuilder:validation:Optional
	// NuxeoImage overrides the default Nuxeo container image selection. By default, the Operator uses 'nuxeo:latest'
	// as the container image. To override that, include the image spec here. Any allowable form is supported. For
	// example, to reference an imagestream in a 'custom-images' OpenShift namespace, the following would be valid -
	// 'image-registry.openshift-image-registry.svc.cluster.local:5000/custom-images/nuxeo:custom'
	NuxeoImage string `json:"nuxeoImage,omitempty"`

	// +kubebuilder:validation:Optional
	// RevProxy causes a reverse proxy to be included in the Nuxeo interactive deployment. The reverse proxy will
	// receive traffic from the Route/Ingress object created by the Operator, and forward that traffic to the Nuxeo
	// Service created by the operator, which in turn will forward traffic to the Nuxeo interactive Pods. Presently,
	// Nginx is the only supported option but the structure is intended to allow other implementations in the future.
	// If omitted, then no reverse proxy is created and traffic goes directly to the Nuxeo Pods.
	RevProxy RevProxySpec `json:"revProxy,omitempty"`

	// ServiceSpec provides the ability to minimally customize the type of Service generated by the
	// Operator. To fully override the service, populate the TODO attribute in the Nuxeo spec
	Service ServiceSpec `json:"serviceSpec,omitempty"`

	// Access defines how Nuxeo will be accessed externally to the cluster. It results in the creation of an
	// OpenShift Route object or TODO a Kubernetes Ingress object
	Access NuxeoAccess `json:"access,omitempty"`

	// Each Node Set causes a Deployment to be created with the specified number of replicas, and other
	// characteristics
	NodeSets []NodeSet `json:"nodeSets,omitempty"`
}

// Defines the observed state of a Nuxeo cluster
// TODO IMPLEMENT THIS
type NuxeoStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Nodes are the names of the nuxeo pods
	Nodes []string `json:"nodes,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Nuxeo is the Schema for the nuxeos API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=nuxeos,scope=Namespaced
type Nuxeo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NuxeoSpec   `json:"spec,omitempty"`
	Status NuxeoStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NuxeoList contains a list of Nuxeo
type NuxeoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Nuxeo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Nuxeo{}, &NuxeoList{})
}
