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
// for the storage. The Mount path can be overridden. If a default PVC as described is not desired, the Volume
// Source can be overridden by specifying the 'volumeSource'.
type NuxeoStorageSpec struct {
	// +kubebuilder:validation:Enum=Binaries;TransientStore;Connect;Data;NuxeoTmp
	// Defines the type of Nuxeo data for of the storage
	// todo-me need a better designator than "storage type"? "storage role"?
	StorageType NuxeoStorage `json:"storageType"`

	// Defines the amount of storage to request. E.g.: 2Gi, 100M, etc.
	Size string `json:"size"`

	// +kubebuilder:validation:Optional
	// Enables explicit definition of a PVC supporting this storage. If specified, then overrides size and
	// volumeSource.
	// +optional
	VolumeClaimTemplate corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`

	// todo-me since these paths are just mappings maybe it should not be allowed to change them
	//  because only binaries and transient stores appear to support ENV vars to set - others require
	//  nuxeo.conf...
	// +kubebuilder:validation:Optional
	// Path within the container at which the volume should be mounted. Defaults are: NuxeoStorageBinaries=/var/lib/nuxeo/binaries.
	// NuxeoStorageTransientStore=/var/lib/nuxeo/transientstore. NuxeoStorageConnect=/opt/nuxeo/connect.
	// NuxeoStorageData=/var/lib/nuxeo/data. NuxeoStorageNuxeoTmp=/opt/nuxeo/server/tmp.
	// +optional
	MountPath string `json:"mountPath,omitempty"`

	// +kubebuilder:validation:Optional
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

	// +kubebuilder:validation:Minimum=1
	// Populates the 'spec.replicas' property of the Deployment generated by this node set.
	Replicas int32 `json:"replicas"`

	// +kubebuilder:validation:Optional
	// Indicates whether this NodeSet will be accessible outside the cluster. Default is 'false'. If 'true', then
	// the Service created by the operator will be have its selectors defined such that it selects the Pods
	// created by this NodeSet. Exactly one NodeSet must be configured for external access.
	// +optional
	Interactive bool `json:"interactive,omitempty"`

	// +kubebuilder:validation:Optional
	// Turns on repository clustering per https://doc.nuxeo.com/nxdoc/next/nuxeo-clustering-configuration/.
	// Sets nuxeo.conf properties: repository.binary.store=/var/lib/nuxeo/binaries/binaries. Sets
	// nuxeo.cluster.enabled=true and nuxeo.cluster.nodeid={env:POD_UID}. Sets POD_UID env var using the
	// downward API. Requires the configurer to specify storage.storageType.Binaries and errors if this is not
	// the configured.
	// +optional
	ClusterEnabled bool `json:"clusterEnabled,omitempty"`

	// +kubebuilder:validation:Optional
	// Supports adding environment variables into the Nuxeo container created by the Operator for this NodeSet. If
	// the PodTemplate is specified, these environment variables are ignored and the environment variables from the
	// PodTemplate - whether they are explicitly defined or not - are used.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// +kubebuilder:validation:Optional
	// Compute Resources required by containers. Cannot be updated.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:validation:Optional
	// Supports a custom readiness probe. If not explicitly specified in the CR then a default httpGet readiness
	// probe on /nuxeo/runningstatus:8080 will be defined by the operator. To disable a probe, define an exec
	// probe that invokes the command 'true'
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`

	// +kubebuilder:validation:Optional
	// Supports a custom liveness probe. If not explicitly specified in the CR then a default httpGet liveness
	// probe on /nuxeo/runningstatus:8080 will be defined by the operator. To disable a probe, define an exec
	// probe that invokes the command 'true'
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`

	// +kubebuilder:validation:Optional
	// Storage provides the ability to configure persistent filesystem storage for the Nuxeo Pods
	// +optional
	Storage []NuxeoStorageSpec `json:"storage,omitempty"`

	// +kubebuilder:validation:Optional
	// NuxeoConfig defines some common configuration settings to customize Nuxeo
	// +optional
	NuxeoConfig NuxeoConfig `json:"nuxeoConfig,omitempty"`

	// +kubebuilder:validation:Optional
	// Provides the ability to add custom or ad-hoc contributions directly into the Nuxeo server
	// +optional
	Contributions []Contribution `json:"contribs,omitempty"`

	// +kubebuilder:validation:Optional
	// Provides the ability to override hard-coded pod defaults, enabling fine-grained control over the
	// configuration of the Pods in the Deployment.
	// todo-me given the complexity of configuring a pod maybe this just be removed
	// +optional
	PodTemplate corev1.PodTemplateSpec `json:"podTemplate,omitempty"`
}

// ServiceSpec provides the ability to minimally customize the the type of Service generated by the Operator.
type ServiceSpec struct {
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	// Specifies the Service type to create
	// +optional
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

	// +kubebuilder:validation:Optional
	// Specifies the name of a secret with fields required to configure ingress for TLS, as determined by
	// the termination field. For OpenShift (Route) supported fields are 'key'/'tls.key', 'certificate'/'tls.crt',
	// 'caCertificate', 'destinationCACertificate', and 'insecureEdgeTerminationPolicy'. These map to Route properties.
	// For Kubernetes (Ingress) supported fields are: 'tls.crt' and 'tls.key'. These are required by Kubernetes.
	// This setting is ignored unless 'termination' is specified
	// +optional
	TLSSecret string `json:"tlsSecret,omitempty"`

	// +kubebuilder:validation:Optional
	// Specifies the TLS termination type. E.g. 'edge', 'passthrough', etc.
	// +optional
	// todo-me consider operator-defined (platform-agnostic) Type and associated constants rather than OpenShift
	Termination routev1.TLSTerminationType `json:"termination,omitempty"`
}

// NginxRevProxySpec defines the configuration elements needed to configure the Nginx reverse proxy.
type NginxRevProxySpec struct {
	// Defines a ConfigMap that contains an 'nginx.conf' key, and a 'proxy.conf' key, each of which provide required
	// configuration to the Nginx container
	ConfigMap string `json:"configMap"`

	// References a secret containing keys 'tls.key', 'tls.cert', and 'dhparam' which are used to terminate
	// the Nginx TLS connection.
	Secret string `json:"secret"`

	// +kubebuilder:validation:Optional
	// Specifies the Nginx image. If not provided, defaults to "nginx:latest"
	// +optional
	Image string `json:"image,omitempty"`

	// +kubebuilder:validation:Optional
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
	// +kubebuilder:validation:Optional
	// nginx supports configuration of Nginx as the reverse proxy
	// +optional
	Nginx NginxRevProxySpec `json:"nginx,omitempty"`
}

// NuxeoConfigSetting Supports configuration settings that can be specified with inline values, or from
// Secrets or ConfigMaps
// todo-me inline should either be a secret, or, configurer should have the option of specifying secret or cfgmap
type NuxeoConfigSetting struct {
	// +kubebuilder:validation:Optional
	// Specifies an inline value for the setting. Either this, or the valueFrom must be specified, but not
	// both.
	// +optional
	Value string `json:"value,omitempty"`

	// +kubebuilder:validation:Optional
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
	// +kubebuilder:validation:Optional
	// JavaOpts define environment variables that are passed on to the JVM in the container
	// +optional
	JavaOpts string `json:"javaOpts,omitempty"`

	// +kubebuilder:validation:Optional
	// NuxeoTemplates defines a list of templates to load when starting Nuxeo
	// +optional
	NuxeoTemplates []string `json:"nuxeoTemplates,omitempty"`

	// +kubebuilder:validation:Optional
	// NuxeoPackages defines a list of packages to install when starting Nuxeo. These packages can only
	// be installed if the Nuxeo cluster has internet access to Nuxeo Connect.
	// todo-me consider ConnectPackages
	//+optional
	NuxeoPackages []string `json:"nuxeoPackages,omitempty"`

	// +kubebuilder:validation:Optional
	// NuxeoUrl is the redirect url used by Nuxeo
	// +optional
	NuxeoUrl string `json:"nuxeoUrl,omitempty"`

	// +kubebuilder:validation:Optional
	// NuxeoName defines a human-friendly name for this cluster
	// +optional
	NuxeoName string `json:"nuxeoName,omitempty"`

	// todo-me raise up a level
	// +kubebuilder:validation:Optional
	// NuxeoConf specifies values to append to nuxeo.conf. Values can be provided inline, or from a Secret
	// or ConfigMap
	// +optional
	NuxeoConf NuxeoConfigSetting `json:"nuxeoConf,omitempty"`

	// +kubebuilder:validation:Optional
	// tlsSecret enables TLS termination by the Nuxeo Pod. The field specifies the name of a secret containing
	// keys keystore.jks and keystorePass. As of Nuxeo 10.10, only JKS is supported. (This is a Nuxeo constraint.)
	// +optional
	TlsSecret string `json:"tlsSecret,omitempty"`

	// +kubebuilder:validation:Optional
	// JvmPKISecret names a secret containing six keys that are used to configure the JVM-wide keystore/truststore
	// for the Nuxeo container. The operator mounts the keystore and truststore files into the Nuxeo container, and
	// sets environment variables which the Nuxeo loader passes through into the JVM. All of the following keys will
	// be configured from the secret into JVM keystore/truststore properties: keyStore, keyStorePassword, keyStoreType,
	// trustStore, trustStorePassword, and trustStoreType.
	// +optional
	JvmPKISecret string `json:"jvmPKISecret,omitempty"`

	// +kubebuilder:validation:Optional
	// offlinePackages configures a list of Nuxeo marketplace packages (ZIP files) that have been made available to
	// the Operator as externally configured storage resources. In the current version, only ConfigMaps and Secrets
	// can be used to hold offline packages. And only one ZIP per ConfigMap/Secret is supported.
	// +optional
	OfflinePackages []OfflinePackage `json:"offlinePackages,omitempty"`
}

// CertTransformType defines a type of certificate transformation that the Nuxeo Operator can perform
type CertTransformType string

const (
	// Convert a CRT to a P12 trust store containing certs only - no keys. Assuming the CRT contains a root CA, then
	// the resulting trust store can be used by Nuxeo and the JVM to validate a backing service certificate during
	// a one-way SSL handshake
	CrtToTrustStore CertTransformType = "CrtToTrustStore"

	// todo-me ???ToKeyStore CERT AND KEY? CERT OR KEY?
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
	// +kubebuilder:validation:Enum=CrtToTrustStore
	// certTransformType establishes the type of certificate transform to apply.
	Type CertTransformType `json:"type"`

	// store defines the name of the key/trust store to create from a source PEM. This becomes a key in the secondary
	// secret and also a mounted file on the file system. Ensure this is unique across all transformations defined
	// in the backing service.
	Store string `json:"store"`

	// When the Operator creates the store, it generates a random store password and places that password into the
	// secondary secret's data identified by this key.  Ensure this key is unique across all transformations defined
	// in the backing service.
	Password string `json:"password"`

	// passEnv defines the name of an environment variable for the Operator to project into the Nuxeo Pod, with the
	// source of that env var being the secondary secret password. This password can then be referenced in nuxeo.conf.
	// E.g.: elasticsearch.restClient.truststore.password=${env:ENV_YOU_SPECIFY_HERE}.
	PassEnv string `json:"passEnv"`
}

// resourceProjection determines how a value from a backing service resource is projected into the Nuxeo Pod. If
// the backing service resource is a Secret or ConfigMap and the resource contains a value can be used as is, then
// the key field specifies the key in the Secret or ConfigMap. If the backing service resource is *not* a Secret or
// ConfigMap but the resource value can still be used without transformation, then the 'path' field is used to locate
// the resource value. E.g.: assume the following expression returns a value that is needed for the backing service
// connection configuration:
//   kubectl get elasticsearch my-elastic-cluster -ojsonpath='{.spec.nodeSets[0].config.node\.data}'
// Then in this case, the path field in the resource projection would specify "{.spec.nodeSets[0].config.node\.data}"
//
// Having used the path or key fields to obtain the value from the resource, there are three options to project
// the value into the Nuxeo Pod. If the backing service resource is a Secret or ConfigMap then the env field specifies
// an environment variable for the operator to project into the Nuxeo Pod. The environment variable will be defined
// as a 'valueFrom' environment variable. E.g. the Operator will project something like:
//   env:
//   - name: USER_PASSWORD
//     valueFrom:
//       secretKeyRef:
//         name: backing-service-secret-name
//         key: user.password (for example)
//
// If the backing service resource is not a Secret or ConfigMap but the resource value can still be used without
// transformation, then the mount field is used to mount the resource value into the Pod. The mount field
// specifies the name of the file to mount. All files are mounted by the Operator under
// /etc/nuxeo-operator/binding/<backing service name>. Mount is also a valid option for Secrets and ConfigMaps if,
// for example, a backing service secret contains a P12 key store and the goal is to simply mount that key store.
//
// Finally, the certTransform field supports transforming a backing service PKI resource such as a PEM into a key
// Store or trust store. This results in the creation of a secondary secret whose keys can be projected via
// mounts or environment variables.
type ResourceProjection struct {
	// If the backing service resource is a Secret or ConfigMap, this is the key of the resource value of interest.
	// Specify this, or path, but not both.
	Key string `json:"key"`

	// If the backing service resource is not a Secret or ConfigMap, this is a JSON Path expression that finds
	// the value of interest in the resource. Specify this, or key, but not both.
	Path string `json:"path"`

	// todo-me consider this to make explicit the key in the secondary secret that will hold the result of a jsonPath
	//  projection otherwise the operator has to take JSONPath like .spec.backingServices[0].resources[0].name and
	//  turn it into secondary secret key .spec.backingServices0.resources0.name
	//  NewKey string `json:"newKey"`

	// If the backing service resource is a Secret or ConfigMap, and the desire is to project the key value as an
	// environment variable, then this is the name of the environment variable to project. The Operator will define
	// a valueFrom environment variable. Specify only one of env, certTransform, or mount.
	Env string `json:"env"`

	// If the backing service resource can be used without transformation, and the desire is to mount it as a file,
	// then provide the name of a file for the Operator to mount the resource value as. Specify only one of env,
	// certTransform, or mount.
	Mount string `json:"mount"`

	// If the backing service resource requires transformation - e.g. from PEM to P12 - specify the transformation
	// using the transform field
	Transform CertTransform `json:"transform"`
}

// A BackingServiceResource provides the ability to extract values from a Kubernetes cluster resource, and to
// project those values into the Nuxeo Pod. For example, a password can be obtained from a backing service secret
// and projected into the Nuxeo Pod as an environment variable or mount.
type BackingServiceResource struct {
	// name is GVK of the cluster resource from which to obtain a value or values.
	metav1.GroupVersionKind `json:"kind"`

	// name is the name of the cluster resource from which to obtain a value or values.
	Name string `json:"name"`

	// Each projection defines one value to get from the resource specified by GVK+Name, and how to project
	// that one value into the Nuxeo Pod.
	Projections []ResourceProjection `json:"projections"`
}

// todo-me /etc/nuxeo-operator/binding/ or /etc/nuxeo-operator/backing/ NEED A CONST!
// A backing service specifies three things: 1) a list of cluster resources from which to obtain connection
// configuration values like passwords and certificates, and the corresponding projections of those values into
// the Nuxeo Pod. 2) A nuxeo.conf string that can reference the projected resource values. 3) A name. The name
// is important because it is used by the operator as a base directory into which to mount files. Files are
// mounted under /etc/nuxeo-operator/binding/<the name you assign>. E.g. if a backing service is defined thus:
//   spec:
//     backingServices:
//     - name: elastic
//       ...
//
// then all files mounted by the operator will mount under: /etc/nuxeo-operator/binding/elastic. If a nuxeo.conf
// entry references a mounted file, then it will reference it relative to that path.
//
// Once the operator is finished configuring all of the backing service bindings, all of the nuxeo.conf entries are
// concatenated and appended to the operator-managed nuxeo.conf ConfigMap. Note the Nuxeo CR offers the ability
// to specify inline nuxeo.conf values in the Nuxeo CR:
//   spec:
//     nodeSets:
//     - name: my-nuxeo
//       nuxeoConfig:
//         nuxeoConf:
//           value: |
//             # inline
//             a.b.c=test1
//             p.d.q=test2
// The nuxeo.conf entries specified in the backing services are appended to any inline content. The Nuxeo CR also
// offers the ability to define nuxeo.conf content in an externally provisioned ConfigMap or Secret and to reference
// that in the Nuxeo CR:
//   spec:
//     nodeSets:
//     - name: my-nuxeo
//       nuxeoConfig:
//         nuxeoConf:
//           valueFrom:
//             configMap:
//               name: my-externally-provisioned-nuxeo-conf-config-map
// An externally provisioned nuxeo.conf ConfigMap or Secret is not compatible with backing services and will result
// in a reconciliation error. Only inlined nuxeo.conf content is supported with backing services - because the
// Operator has to have ownership of the cluster resource holding the nuxeo.conf content and it can't do that if
// the resource is provisioned by the configurer.
type BackingService struct {
	// The name of the backing service, as well as the directory under which to mount any files
	Name string `json:"name"`

	// Resources and projections control how backing service cluster resources are referenced within the Nuxeo Pod
	Resources []BackingServiceResource `json:"resources"`

	// nuxeo.conf entries - some of which will be static, and some of which will reference resource projections
	// via environment variables or filesystem mounts
	NuxeoConf string `json:"nuxeoConf"`
}

// Defines the desired state of a Nuxeo cluster
type NuxeoSpec struct {

	// +kubebuilder:validation:Optional
	// Overrides the default Nuxeo container image selected by the Operator. By default, the Operator
	// uses 'nuxeo:latest' as the container image. To override that, include the image spec here. Any allowable
	// form is supported. E.g. 'image-registry.openshift-image-registry.svc.cluster.local:5000/custom-images/nuxeo:custom'
	// +optional
	NuxeoImage string `json:"nuxeoImage,omitempty"`

	// +kubebuilder:validation:Optional
	// Image pull policy. If not specified, then if 'nuxeoImage' is specified with the :latest tag, then this is
	// 'Always', otherwise it is 'IfNotPresent'. Note that this flows through to a Pod ultimately, and pull policy
	// is immutable in a Pod spec. Therefore if any changes are made to this value in a Nuxeo CR once the
	// Operator has generated a Deployment from the CR, subsequent Deployment reconciliations will fail.
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// +kubebuilder:validation:Optional
	// Causes a reverse proxy to be included in the Nuxeo interactive deployment. The reverse proxy will
	// receive traffic from the Route/Ingress object created by the Operator, and forward that traffic to the Nuxeo
	// Service created by the operator, which in turn will forward traffic to the Nuxeo interactive Pods. Presently,
	// Nginx is the only supported option but the structure is intended to allow other implementations in the future.
	// If omitted, then no reverse proxy is created and traffic goes directly to the Nuxeo Pods.
	// +optional
	RevProxy RevProxySpec `json:"revProxy,omitempty"`

	// +kubebuilder:validation:Optional
	// Provides the ability to minimally customize the type of Service generated by the Operator.
	// +optional
	Service ServiceSpec `json:"serviceSpec,omitempty"`

	// +kubebuilder:validation:Optional
	// Defines how Nuxeo will be accessed externally to the cluster. It results in the creation of an
	// OpenShift Route object. In the future, it will also support generation of a Kubernetes Ingress object
	// +optional
	Access NuxeoAccess `json:"access,omitempty"`

	// +kubebuilder:validation:MinItems=1
	// Each nodeSet causes a Deployment to be created with the specified number of replicas, and other
	// characteristics specified within the nodeSet spec. At least one nodeSet is required
	NodeSets []NodeSet `json:"nodeSets"`

	// +kubebuilder:validation:Optional
	// Nuxeo CLID
	// +optional
	Clid string `json:"clid,omitempty"`

	// +kubebuilder:validation:Optional
	// Backing Services are used to bind Nuxeo to cluster backing services like Kafka, MongoDB, ElasticSearch,
	// and Postgres
	// +optional
	BackingServices []BackingService `json:"backingServices,omitempty"`
}

// NuxeoStatus defines the observed state of a Nuxeo cluster. This is preliminary and will be expanded in later
// versions
type NuxeoStatus struct {
	AvailableNodes int32 `json:"availableNodes,omitempty"`
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
