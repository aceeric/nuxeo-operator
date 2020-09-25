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

package v1alpha1

import (
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NuxeoStorage defines a type of persistent storage
type NuxeoStorage string

const (
	// NuxeoStorageBinaries holds the blobs that are attached to documents
	NuxeoStorageBinaries NuxeoStorage = "Binaries"
	// NuxeoStorageTransientStore holds transient data with configurable expiration
	NuxeoStorageTransientStore = "TransientStore"
	// NuxeoStorageConnect is for Nuxeo NuxeoStorageConnect
	NuxeoStorageConnect = "Connect"
	// NuxeoStorageData holds various Nuxeo system data
	NuxeoStorageData = "Data"
	// NuxeoStorageNuxeoTmp is like /tmp for Nuxeo
	NuxeoStorageNuxeoTmp = "NuxeoTmp"
)

// By default, all filesystem access inside a Pod is ephemeral and data is lost when the Pod terminates. The
// NuxeoStorageSpec enables definition of persistent storage. By default, the Nuxeo Operator will create a PVC
// for each specified storage with volumeMode=Filesystem, accessMode=ReadWriteOnce, and no storage class.
// This Operator will define a volume and a volume mount for the PVC with a hard-coded path that is reasonable
// for the storage. If a default PVC as described is not desired, the Volume Source can be overridden by
// specifying the 'volumeSource'.
type NuxeoStorageSpec struct {
	// Defines the type of Nuxeo data for of the storage
	// +kubebuilder:validation:Enum=Binaries;TransientStore;Connect;Data;NuxeoTmp
	StorageType NuxeoStorage `json:"storageType"`

	// Defines the amount of storage to request. E.g.: 2Gi, 100M, etc.
	Size string `json:"size"`

	// Enables explicit definition of a PVC supporting this storage. If specified, then overrides size and
	// volumeSource.
	// +optional
	VolumeClaimTemplate corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`

	// Enables the Volume Source to be explicitly specified. Ignored if volumeClaimTemplate is specified. This could
	// be used, for example, to define an EmptyDir volume source for testing/troubleshooting.
	// +optional
	VolumeSource corev1.VolumeSource `json:"volumeSource,omitempty"`
}

// Contributions allow a configurer to add ad-hoc or persistent contributions to the Nuxeo server. Two scenarios are
// envisioned. For an ad-hoc contribution, you define a ConfigMap or Secret with the contribution contents, and define
// the name of the contribution in the templates list. The operator configures that single contribution into Nuxeo by
// mounting the files, and adding one entry into the nuxeo templates.
//
// For persistent contributions, you configure a persistent storage resource in the cluster that can contain multiple
// contributions, each in its own sub-directory. You then configure the templates list with the contributions from
// the store that you want configured into Nuxeo. The operator mounts the entire store, but only adds the specified
// contributions into the nuxeo templates.
type Contribution struct {
	// For a ConfigMap or Secret contribution, only one entry is supported: the name that you want assigned to this
	// contribution. E.g. if you specify '["my-contrib"]', then the operator mounts files into /etc/nuxeo/nuxeo-operator-config/my-contrib
	// and sets NUXEO_TEMPLATES=...,/etc/nuxeo/nuxeo-operator-config/my-contrib. For other volume sources, this
	// is a list of directories in the storage resource, and each one is added to NUXEO_TEMPLATES, but the entire
	// volume is mounted into /etc/nuxeo/nuxeo-operator-config
	Templates []string `json:"templates"`

	// For a ConfigMap or Secret, a key 'nuxeo.defaults' causes they value to be mounted as
	// /etc/nuxeo/nuxeo-operator-config/<your contrib>/nuxeo.defaults. For all other keys, they are mounted as
	// files in /etc/nuxeo/nuxeo-operator-config/<your contrib>/nxserver/config. For other volume sources, the
	// entire volume is mounted under /etc/nuxeo/nuxeo-operator-config with the assumption that the tree structure
	// is valid for a nuxeo contribution. See the documentation for additional details.
	VolumeSource corev1.VolumeSource `json:"volumeSource"`
}

// NodeSet defines the structure of the Nuxeo cluster. Each NodeSet results in a Deployment. This supports the
// capability to define different configurations for a Deployment of interactive Nuxeo nodes vs a Deployment
// of worker Nuxeo nodes.
type NodeSet struct {
	// The name of this node set. In cases where only one node set is needed, a recommended naming strategy is
	// to name this node set 'cluster'. For example, if you generate a Nuxeo CR named 'my-nuxeo' into the namespace
	// being watched by the Nuxeo Operator, and you name this node set 'cluster'. Then the operator will create
	// a deployment from the node set named 'my-nuxeo-cluster'
	Name string `json:"name"`

	// Populates the 'spec.replicas' property of the Deployment generated by this node set.
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas"`

	// Indicates whether this NodeSet will be accessible outside the cluster. Default is 'false'. If 'true', then
	// the Service created by the operator will be have its selectors defined such that it selects the Pods
	// created by this NodeSet. Exactly one NodeSet must be configured for external access.
	// +optional
	Interactive bool `json:"interactive,omitempty"`

	// Turns on repository clustering per https://doc.nuxeo.com/nxdoc/next/nuxeo-clustering-configuration/.
	// Sets nuxeo.conf properties: repository.binary.store=/var/lib/nuxeo/binaries/binaries. Sets
	// nuxeo.cluster.enabled=true and nuxeo.cluster.nodeid={env:POD_UID}. Sets POD_UID env var using the
	// downward API. Requires the configurer to specify storage.storageType.Binaries and errors if this is not
	// the configured.
	// +optional
	ClusterEnabled bool `json:"clusterEnabled,omitempty"`

	// Supports manually defining environment variables for the Nuxeo container created by the Operator for
	// this NodeSet.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Compute Resources required by containers. Cannot be updated.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Supports a custom readiness probe. If not explicitly specified in the CR then a default httpGet readiness
	// probe on /nuxeo/runningstatus:8080 will be defined by the operator. To disable a probe, define an exec
	// probe that invokes the command 'true'
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`

	// Supports a custom liveness probe. If not explicitly specified in the CR then a default httpGet liveness
	// probe on /nuxeo/runningstatus:8080 will be defined by the operator. To disable a probe, define an exec
	// probe that invokes the command 'true'
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`

	// Storage provides the ability to configure persistent filesystem storage for the Nuxeo Pods
	// +optional
	Storage []NuxeoStorageSpec `json:"storage,omitempty"`

	// NuxeoConfig defines some common configuration settings to customize Nuxeo
	// +optional
	NuxeoConfig NuxeoConfig `json:"nuxeoConfig,omitempty"`

	// Provides the ability to add custom or ad-hoc contributions directly into the Nuxeo server
	// +optional
	Contributions []Contribution `json:"contribs,omitempty"`
}

// ServiceSpec provides the ability to minimally customize the the type of Service generated by the Operator.
type ServiceSpec struct {
	// Specifies the Service type to create
	// +optional
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	Type corev1.ServiceType `json:"type,omitempty"`

	// Specifies the port exposed by the service
	// +optional
	Port int32 `json:"port,omitempty"`

	// Specifies the port that the service will use internally to communicate with the Nuxeo cluster
	// +optional
	TargetPort int32 `json:"targetPort,omitempty"`
}

// NuxeoAccess supports creation of an OpenShift Route or Kubernetes Ingress supporting access to the Nuxeo Service
// from outside of the cluster.
type NuxeoAccess struct {
	// Specifies the host name. This is incorporated by the Operator into the operator-generated
	// OpenShift Route and should be accessible from outside the cluster via DNS or some other suitable
	// name resolution mechanism
	Hostname string `json:"hostname"`

	// Selects a target port in the Service backed by this NuxeoAccess spec. By default, 'web' is
	// populated by the Operator - which finds the default 'web' port in the Service generated by the Operator
	// +optional
	TargetPort intstr.IntOrString `json:"targetPort,omitempty"`

	// Specifies the name of a secret with fields required to configure ingress for TLS, as determined by
	// the termination field. For OpenShift (Route) supported fields are 'key'/'tls.key', 'certificate'/'tls.crt',
	// 'caCertificate', 'destinationCACertificate', and 'insecureEdgeTerminationPolicy'. These map to Route properties.
	// For Kubernetes (Ingress) supported fields are: 'tls.crt' and 'tls.key'. These are required by Kubernetes.
	// This setting is ignored unless 'termination' is specified
	// +optional
	TLSSecret string `json:"tlsSecret,omitempty"`

	// Specifies the TLS termination type. E.g. 'edge', 'passthrough', etc.
	// +optional
	Termination routev1.TLSTerminationType `json:"termination,omitempty"`
}

// NginxRevProxySpec defines the configuration elements needed to configure the Nginx reverse proxy.
type NginxRevProxySpec struct {
	// Defines a ConfigMap that contains an 'nginx.conf' key, and a 'proxy.conf' key, each of which provide
	// configuration to the Nginx container. If not provided, then the operator will auto-generate a ConfigMap with
	// defaults and mount it into the container/deployment.
	// +optional
	ConfigMap string `json:"configMap"`

	// References a secret containing keys 'tls.key', 'tls.cert', and 'dhparam' which are used to terminate
	// the Nginx TLS connection.
	// +optional
	Secret string `json:"secret"`

	// Specifies the Nginx image. If not provided, defaults to "nginx:latest"
	// +optional
	Image string `json:"image,omitempty"`

	// Image pull policy. If not specified, then if 'image' is specified with the :latest tag,
	// then this is 'Always', otherwise it is 'IfNotPresent'. Note that this flows through to a Pod ultimately,
	// and pull policy is immutable in a Pod spec. Therefore if any changes are made to this value in a Nuxeo
	// CR once the Operator has generated a Deployment from the CR, subsequent Deployment reconciliations will fail.
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
}

// RevProxySpec defines the reverse proxies supported by the Nuxeo Operator. Details are provided in the individual
// specs.
type RevProxySpec struct {
	// nginx supports configuration of Nginx as the reverse proxy
	// +optional
	Nginx NginxRevProxySpec `json:"nginx,omitempty"`
}

// NuxeoConfigSetting Supports configuration settings that can be specified with inline values, or from
// Secrets or ConfigMaps
type NuxeoConfigSetting struct {
	// Specifies an inline value for the setting. Either this, or the valueFrom must be specified, but not
	// both.
	// +optional
	Inline string `json:"inline,omitempty"`

	// Source for the configuration settings's value. Either this, or the value must be specified, but not
	// both. Only Secrets and Config maps are supported at the present time. Any other type of volume source
	// will cause the operator to error. A later release may support the other volume sources.
	// +optional
	ValueFrom corev1.VolumeSource `json:"valueFrom,omitempty"`
}

// OfflinePackage supports installing Marketplace packages in a Kubernetes cluster without connectivity
// to the Nuxeo Marketplace. A configurer creates or downloads a marketplace package ZIP, and
// configures a storage resource containing the ZIP for that package. A Nuxeo CR is configured that references
// that resource. The Operator configures Nuxeo so that - on startup - Nuxeo installs the package into
// the running Nuxeo instance. In the current version, only ConfigMaps and Secrets can be used to hold the
// package binaries. And only one package ZIP per ConfigMap/Secret is supported. The reason for this is
// the the Nuxeo container init shell script only supports installing packages from a single directory level.
// In order to support a persistent volume and claim, the Nuxeo script needs to be modified to support sub-directories.
// If that change is made by Nuxeo, then this Operator will be updated to support mounting a volume into a
// subdirectory of the Nuxeo container init directory. This will make it possible to do offline installation from
// a single volume containing multiple packages.
type OfflinePackage struct {
	// For Secret and ConfigMap volume sources (currently the only two supported) this is the key in the
	// object that contains the package ZIP. This becomes the file name of the ZIP in the Nuxeo container.
	// E.g.: nuxeo-sample-2.5.3.zip
	PackageName string `json:"packageName,omitempty"`

	// Source for ZIP binary data. Only Secrets and Config maps are supported at the present time. Any other type
	// of volume source will cause the operator to error. A later release may support the other volume sources.
	ValueFrom corev1.VolumeSource `json:"valueFrom,omitempty"`
}

// NuxeoConfig provides the ability to configure the Nuxeo server. These settings are added to each Deployment
// generated from the NodeSet.
type NuxeoConfig struct {
	// JavaOpts define environment variables that are passed on to the JVM in the container
	// +optional
	JavaOpts string `json:"javaOpts,omitempty"`

	// NuxeoTemplates defines a list of templates to load when starting Nuxeo
	// +optional
	NuxeoTemplates []string `json:"nuxeoTemplates,omitempty"`

	// NuxeoPackages defines a list of packages to install when starting Nuxeo. Except for packages that are
	// pre-included in the Nuxeo container (like nuxeo-web-ui) these packages can only be installed if the Nuxeo
	// cluster has internet access to Nuxeo Connect.
	//+optional
	NuxeoPackages []string `json:"nuxeoPackages,omitempty"`

	// NuxeoUrl is the redirect url used by Nuxeo
	// +optional
	NuxeoUrl string `json:"nuxeoUrl,omitempty"`

	// NuxeoName defines a human-friendly name for this cluster
	// +optional
	NuxeoName string `json:"nuxeoName,omitempty"`

	// NuxeoConf specifies values to append to nuxeo.conf. Values can be provided inline, or from a Secret
	// or ConfigMap
	// +optional
	NuxeoConf NuxeoConfigSetting `json:"nuxeoConf,omitempty"`

	// tlsSecret enables TLS termination by the Nuxeo Pod. The field specifies the name of a secret containing
	// keys keystore.jks and keystorePass. As of Nuxeo 10.10, only JKS is supported. (This is a Nuxeo constraint.)
	// +optional
	TlsSecret string `json:"tlsSecret,omitempty"`

	// JvmPKISecret names a secret containing six keys that are used to configure the JVM-wide keystore/truststore
	// for the Nuxeo container. The operator mounts the keystore and truststore files into the Nuxeo container, and
	// sets environment variables which the Nuxeo loader passes through into the JVM. All of the following keys will
	// be configured from the secret into JVM keystore/truststore properties: keyStore, keyStorePassword, keyStoreType,
	// trustStore, trustStorePassword, and trustStoreType.
	// +optional
	JvmPKISecret string `json:"jvmPKISecret,omitempty"`

	// offlinePackages configures a list of Nuxeo marketplace packages (ZIP files) that have been made available to
	// the Operator as externally configured storage resources. In the current version, only ConfigMaps and Secrets
	// can be used to hold offline packages. And only one ZIP per ConfigMap/Secret is supported.
	// +optional
	OfflinePackages []OfflinePackage `json:"offlinePackages,omitempty"`
}

// CertTransformType defines a type of certificate transformation that the Nuxeo Operator can perform
type CertTransformType string

const (
	// Convert a certificate to a trust store containing certs only - no keys. Supports one-way TLS
	TrustStore CertTransformType = "TrustStore"

	// Convert a certificate and private key to a key store supporting mutual TLS
	KeyStore CertTransformType = "KeyStore"
)

// CertTransform supports the Operator's requirement to take incoming PKI assets in the form of CRTs/PEMs etc.
// and to transform those into Java key stores and trust stores for the Nuxeo server. The Operator creates a
// secondary secret and puts the resulting key store / trust store into that secondary secret. The Operator
// also generates a store password, places that password into the secondary secret, and projects that
// password into the Nuxeo Pod. The secondary secret name is derived from the nuxeo CR name, and the backing service
// name. E.g.: If nuxeo.name == my-nuxeo and backingService.name == elastic then secondary secret name is
// my-nuxeo-binding-elastic. If multiple transformations are specified within a single binding then all the
// keys are created within the same secondary binding secret.
type CertTransform struct {
	// certTransformType establishes the type of certificate transform to apply.
	// +kubebuilder:validation:Enum=TrustStore;KeyStore
	Type CertTransformType `json:"type"`

	// Same as the "from" field of the resource projection: if the resource being transformed is a Secret or
	// ConfigMap then this is a key in the .Data stanza. Otherwise it is a JSON Path expression. Either way, this
	// finds the certificate bits in the upstream resource.
	// (e.g. client or CA) is specified
	Cert string `json:"cert"`

	// Either a key, or a JSON Path - see the cert documentation. For keystore transforms, this represents the
	// private key bits in the upstream resource. Ignored for truststore transforms.
	// +optional
	PrivateKey string `json:"privateKey"`

	// store defines the name of the key/trust store to create from a source PKI elements. This becomes a key in
	// the secondary secret and also a mounted file on the file system in the Pod. Ensure this is unique across all
	// transformations defined in the backing service.
	Store string `json:"store"`

	// When the Operator creates the store, it generates a random store password and places that password into the
	// secondary secret's .Data identified by this key.  Ensure this key is unique across all transformations defined
	// in the backing service.
	Password string `json:"password"`

	// passEnv defines the name of an environment variable for the Operator to project into the Nuxeo Pod, with the
	// source of that env var being the secondary secret password. This password can then be referenced in nuxeo.conf.
	// E.g.: elasticsearch.restClient.truststore.password=${env:ENV_YOU_SPECIFY_HERE}.
	PassEnv string `json:"passEnv"`
}

// ResourceProjection determines how a value from a backing service resource is projected into the Nuxeo Pod. Values
// can be projected as environment variables, mounts, or transformed - which always results in a mount. There can
// be multiple projections from a single resource. E.g. one resource might contain one value that is projected
// as an env var, and another value projected as a filesystem object.
type ResourceProjection struct {
	// If the backing service resource is a Secret or ConfigMap, this is the key of the resource value
	// of interest from the .Data stanza of the Secret or ConfigMap. If the backing service resource is not a
	// Secret or ConfigMap, this is a JSON Path expression that finds the value of interest in the resource.
	// Ignored for transforms.
	// +optional
	From string `json:"from"`

	// If the backing service resource is a Secret or ConfigMap, and the desire is to project the key value as an
	// environment variable, then this is the name of the environment variable to project. The Operator will define
	// a valueFrom environment variable. Environment projection is only supported for Secrets and ConfigMaps
	// at present. Specify only one of env, mount, or transform. Ignored for transforms.
	// +optional
	Env string `json:"env"`

	// Supports the ability to define environment variables with values directly from upstream resources. For
	// example, an upstream resource (not a secret or config map) might contain a port number that is needed
	// for backing service connectivity. By setting this to true, the operator will define an environment variable
	// named by the Env field, and set the value directly in the env var, rather then as a value from.
	// +optional
	Value bool `json:"value"`

	// If the backing service resource can be used without transformation, and the desire is to mount it as a file,
	// then provide the name of a file for the Operator to mount the resource value as. The operator will copy the
	// value into a secondary secret and mount it from there. Specify only one of env, mount, or transform.
	// Ignored for transforms.
	// +optional
	Mount string `json:"mount"`

	// If the backing service resource requires transformation - e.g. from PEM to JKS - specify the transformation
	// using the transform field. If a transform is specified, then from, env, and mount are ignored.
	// +optional
	Transform CertTransform `json:"transform"`
}

// PreconfigType defines a pre-configured backing service that the Operator knows about and can bind Nuxeo to
// using a terse Nuxeo CR. This relieves the configurer of worrying about the details of the backing service.
type PreconfigType string

const (
	// Elastic Cloud on Kubernetes
	ECK PreconfigType = "ECK"
	// Strimzi Kafka
	Strimzi PreconfigType = "Strimzi"
	// Crunchy Postgres
	Crunchy PreconfigType = "Crunchy"
	// mongo.com Enterprise
	MongoEnterprise PreconfigType = "MongoEnterprise"
)

// A PreconfiguredBackingService is a short-hand way to bind Nuxeo to a backing service. It's a preconfigured
// type that has corresponding go code in the operator to generate the generic backing structures to bing to
// a backing service.
type PreconfiguredBackingService struct {
	// type identifies the preconfigured backing service
	// +kubebuilder:validation:Enum=ECK;Strimzi;Crunchy
	Type PreconfigType `json:"type"`

	// resource identifies the name of the top-level backing service resource. For example, for Elastic Cloud on
	// Kubernetes, this is the cluster resource that would be returned if one executed: 'kubectl get
	// elasticsearch'. The Nuxeo Operator already knows the GVK, so all the configurer needs to provide
	// is the resource name in the same namespace as the Nuxeo CR.
	Resource string `json:"resource"`

	// Optional configuration settings that tune the backing service binding, depending on how the
	// backing service was configured. For example, with Strimzi, it is possible to allow both plain text and
	// tls connections. If you want Nuxeo to connect one way or the other, then specify that here. See
	// the documentation for the settings that are valid for the various pre-configured backing services.
	// +optional
	Settings map[string]string `json:"settings"`
}

// A BackingServiceResource provides the ability to extract values from a Kubernetes cluster resource, and to
// project those values into the Nuxeo Pod. For example, a password can be obtained from a backing service secret
// and projected into the Nuxeo Pod as an environment variable or mount.
type BackingServiceResource struct {
	// GVK of the cluster resource from which to obtain a value or values.
	metav1.GroupVersionKind `json:",inline"`

	// name is the name of the cluster resource from which to obtain a value or values.
	Name string `json:"name"`

	// Each projection defines one value to get from the resource specified by GVK+Name, and how to project
	// that one value into the Nuxeo Pod.
	Projections []ResourceProjection `json:"projections"`
}

// A backing service specifies three things: 1) a list of cluster resources from which to obtain connection
// configuration values like passwords and certificates, and the corresponding projections of those values into
// the Nuxeo Pod. 2) A nuxeo.conf string that can reference the projected resource values. 3) A name. The name
// is important because it is used by the operator as a base directory into which to mount files. Files are
// mounted under /etc/nuxeo-operator/binding/<the name you assign>.
//
// Once the operator is finished configuring all of the backing service bindings, all of the nuxeo.conf entries are
// concatenated and appended to the operator-managed nuxeo.conf ConfigMap. The Nuxeo CR offers the ability
// to specify inline nuxeo.conf values in the Nuxeo CR. If these are specified, the backing service settings are
// appended. The Nuxeo CR also offers the ability to define nuxeo.conf content in an externally provisioned ConfigMap
// or Secret and to reference that in the Nuxeo CR. An externally provisioned nuxeo.conf ConfigMap or Secret is not
// compatible with backing services and will result in a reconciliation error. Only inlined nuxeo.conf content is
// supported with backing services - because the Operator has to have ownership of the cluster resource holding
// the nuxeo.conf content and it can't do that if the resource is provisioned by the configurer. (Because as of
// Nuxeo 10.10 only one nuxeo.conf can exist in /docker-entrypoint-initnuxeo.d to be processed by the Nuxeo
// startup script.)
type BackingService struct {
	// The name of the backing service, as well as the directory under which to mount any files. Required
	// if preConfigured is empty. If name is specified, then resources and nuxeoConf must also be specified,
	// and preConfigured is ignored.
	// +optional
	Name string `json:"name"`

	// Resources and projections control how backing service cluster resources are referenced within the Nuxeo
	// Pod. Required if name is specified. Ignored if preConfigured is specified.
	// +optional
	Resources []BackingServiceResource `json:"resources"`

	// nuxeo.conf entries - some of which will be static, and some of which will reference resource projections
	// via environment variables or filesystem mounts. Required if name is specified. Ignored if preConfigured
	// is specified.
	// +optional
	NuxeoConf string `json:"nuxeoConf"`

	// Some backing services (e.g. PostgreSQL) require that a template be added to the list of templates. If that
	// is the case, then specify the template here. E.g.: 'postgresql', 'mongodb', etc.
	// +optional
	Template string `json:"template"`

	// For a set of known backing service (e.g. Strimzi Kafka) the configuration can be tersely specified
	// using a preConfigured entry. The Nuxeo Operator will perform all the configuration using just a couple
	// of additional settings. If this is specified, then name, resources, and nuxeoConf are all ignored.
	// +optional
	Preconfigured PreconfiguredBackingService `json:"preConfigured"`
}

// Defines the desired state of a Nuxeo cluster
type NuxeoSpec struct {
	// Overrides the default Nuxeo container image selected by the Operator. By default, the Operator
	// uses 'nuxeo:latest' as the container image. To override that, include the image spec here. Any allowable
	// form is supported.
	// +optional
	NuxeoImage string `json:"nuxeoImage,omitempty"`

	// The Nuxeo version. This isn't presently used but is forward-looking for when the Operator needs to implement
	// different behaviors for different Nuxeo versions.
	// +optional
	Version string `json:"version,omitempty"`

	// Image pull policy. If not specified, then if 'nuxeoImage' is specified with the :latest tag, then this is
	// 'Always', otherwise it is 'IfNotPresent'. Note that this flows through to a Pod ultimately, and pull policy
	// is immutable in a Pod spec. Therefore if any changes are made to this value in a Nuxeo CR once the
	// Operator has generated a Deployment from the CR, subsequent Deployment reconciliations will fail.
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Causes a reverse proxy to be included in the Nuxeo interactive deployment. The reverse proxy will
	// receive traffic from the Route/Ingress object created by the Operator, and forward that traffic to the Nuxeo
	// Service created by the operator, which in turn will forward traffic to the Nuxeo interactive Pods. Presently,
	// Nginx is the only supported option but the structure is intended to allow other implementations in the future.
	// If omitted, then no reverse proxy is created and traffic goes directly to the Nuxeo Pods.
	// +optional
	RevProxy RevProxySpec `json:"revProxy,omitempty"`

	// Provides the ability to minimally customize the type of Service generated by the Operator.
	// +optional
	Service ServiceSpec `json:"serviceSpec,omitempty"`

	// Defines how Nuxeo will be accessed externally to the cluster. It results in the creation of an
	// OpenShift Route object. In the future, it will also support generation of a Kubernetes Ingress object
	// +optional
	Access NuxeoAccess `json:"access,omitempty"`

	// Each nodeSet causes a Deployment to be created with the specified number of replicas, and other
	// characteristics specified within the nodeSet spec. At least one nodeSet is required
	// +kubebuilder:validation:MinItems=1
	NodeSets []NodeSet `json:"nodeSets"`

	// Nuxeo CLID. Must be formatted as it would be obtained from the Nuxeo registration site, with the double
	// dash separator
	// +optional
	Clid string `json:"clid,omitempty"`

	// Backing Services are used to bind Nuxeo to cluster backing services like Kafka, MongoDB, ElasticSearch,
	// and Postgres
	// +optional
	BackingServices []BackingService `json:"backingServices,omitempty"`

	// initContainers provides the ability to add custom init containers
	// +optional
	InitContainers []corev1.Container `json:"initContainers,omitempty"`

	// containers provides the ability to add "sidecar" containers. Note - the Operator will create a container named
	// "nuxeo". Therefore, the "nuxeo" container name is reserved and will cause a reconcile error if specified here.
	// +optional
	Containers []corev1.Container `json:"containers"`

	// provides explicit volume configuration, mainly to support the container fields since containers can
	// specify volume mounts, which therefore requires a volume.
	// +optional
	Volumes []corev1.Volume `json:"volumes"`
}

type StatusValue string

const (
	StatusUnavailable StatusValue = "unavailable"
	StatusHealthy     StatusValue = "healthy"
	StatusDegraded    StatusValue = "degraded"
)

// NuxeoStatus defines the observed state of a Nuxeo cluster
type NuxeoStatus struct {
	DesiredNodes   int32       `json:"desiredNodes,omitempty"`
	AvailableNodes int32       `json:"availableNodes,omitempty"`
	Status         StatusValue `json:"status,omitempty"`
}

// Nuxeo is the Schema for the nuxeos API
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=nuxeos,scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="health",type="string",JSONPath=".status.status"
// +kubebuilder:printcolumn:name="available",type="integer",JSONPath=".status.availableNodes"
// +kubebuilder:printcolumn:name="desired",type="integer",JSONPath=".status.desiredNodes"
type Nuxeo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NuxeoSpec   `json:"spec,omitempty"`
	Status NuxeoStatus `json:"status,omitempty"`
}

// NuxeoList contains a list of Nuxeo
// +kubebuilder:object:root=true
type NuxeoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Nuxeo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Nuxeo{}, &NuxeoList{})
}
