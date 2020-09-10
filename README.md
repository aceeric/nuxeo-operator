# Nuxeo Operator

The Nuxeo operator is a Kubernetes/OpenShift Operator written in Go to manage the state of a *Nuxeo* cluster. Nuxeo is an open source digital asset management system. (See https://www.nuxeo.com/).

This project is under development. The current version is 0.6.2. Testing is performed with OpenShift Code Ready Containers (https://github.com/code-ready/crc) and MicroK8s (https://microk8s.io).

### Current Feature Set (as of 0.6.2)

| Feature                                                      |
| ------------------------------------------------------------ |
| Generate and reconcile a Deployment from a Nuxeo Custom Resource (CR) to represent the desired state of a Nuxeo cluster |
| Run Nuxeo in development mode with all embedded services, or in full production mode integrated with backing services like Kafka, Elastic Search, and PostgreSQL |
| Support an optional Nginx TLS reverse proxy sidecar container in each Nuxeo Pod to support TLS |
| Create and reconcile a Route resource for access outside of the cluster |
| Create and reconcile a Service resource for the Route to use, and for potential use within the cluster. The service will communicate with Nuxeo on 8080, or Nginx on 8443 |
| Create and reconcile a Service Account to limit the operator to only the resources it needs to manage |
| Create all resources that originate from a Nuxeo CR with `ownerReferences` that reference the Nuxeo CR - so that deletion of the Nuxeo CR will result in recursive removal of all generated resources for proper clean up |
| Support custom Nuxeo images, with a default of `nuxeo:latest` if no custom image is provided in the Nuxeo CR |
| Implement a Status field of the Nuxeo CR for visual and scripted health check |
| Include the elements (CSV, RBACs, bundling, etc.) to support packaging the Operator as a community Operator |
| Support the ability to deploy the Operator from an internal Operator registry in the cluster via OLM subscription |
| Automate all build / test activities with *GNU Make*         |
| Incorporate unit testing into the operator build using https://github.com/stretchr/testify |
| Incorporate end-to-end testing using the `envtest` scaffolding provided by the Operator SDK v1.0.0 |
| Implement the ability to detect whether the Operator is running in a Kubernetes cluster vs. an OpenShift cluster |
| Create an *Ingress* resource for access outside of the Kubernetes cluster and test with HTTP as well as TLS passthrough |
| Documents Kubernetes testing using MicroK8s (https://microk8s.io/) |
| Support configurable readiness and liveness probes for the Nuxeo pods |
| Support storage configuration for Nuxeo binaries, the transient store, etc. |
| Support explicit definition of nuxeo.conf properties in the Nuxeo CR |
| Support additional fields in the Nuxeo CR to configure Nuxeo:  Java opts, templates, packages, nuxeo URL, nuxeo name |
| Support passthrough and edge termination in Kubernetes Ingress, and all Route termination types for OpenShift |
| Support the ability to terminate TLS directly in Nuxeo, rather than requiring a sidecar. See `test-nuxeo-tls.md` in the docs folder |
| Support a secret for JVM-wide PKI configuration in the Nuxeo Pod - in order to support cases where Nuxeo is running in a PKI-enabled enterprise and is interacting with internal PKI-enabled corporate services that use an internal corporate CA. See `test-jvm-pki.md` in the docs folder. |
| Support installing marketplace packages in disconnected mode if no Internet connection is available in-cluster. See `test-offline-packages.md` in the docs folder. |
| Ability to configure *Interactive* nodes and *Worker* nodes differently by configuring contributions via cluster resources. The objective is to support compute-intensive back-end processing on a set of nodes having a greater resource share in the cluster then the interactive nodes that serve the Nuxeo GUI. See `test-contribution.md` in the docs folder. |
| Support clustering - use Pod UID as `nuxeo.cluster.nodeid` via the downward API |
| Support defining resource request/limit in the Nuxeo CR      |
| Support Nuxeo CLID. See `nuxeo-cr-clid-ex.yaml` in hack/examples |
| Support flexible integration with backing services by virtue of the `backingService` resource in the Nuxeo CR. Validate that with specific integrations (see below) |
| Integrate with Elastic Cloud on Kubernetes (https://github.com/elastic/cloud-on-k8s) for ElasticSearch support |
| Integrate with Strimzi (https://strimzi.io/) for Nuxeo Stream support |
| Integrate with Crunchy PostgreSQL (https://www.crunchydata.com/products/crunchy-postgresql-for-kubernetes/) for database support |
| Integrate with Percona MongoDB backing service (https://www.percona.com/doc/kubernetes-operator-for-psmongodb/index.html) |
| Integrate with Zalando Postgres (https://github.com/zalando/postgres-operator) as an alternative to Crunchy |
| The project includes a test/kustomize directory to support automated testing of all backing service integrations |
| Support rolling deployment updates: `kubectl rollout restart deployment nuxeo-cluster` |
| Provide a sidecar array, init container array, and volumes array to support flexible configuration |
| The Operator can watch a single namespace, multiple namespaces, or all namespaces. If subscribing the Operator using OLM, this is specified in the `OperatorGroup`. If manually installing, you can patch the Operator's deployment - specifically the `WATCH_NAMESPACE` environment variable. This can be in the format *""* - meaning watch all, or *"my-namespace"*, meaning one namespace, or *"namespace-1,namespace-2"* meaning the specified namespaces. |

### Planned Work

#### Version 0.6.3 *(in progress)*

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Test in a full production-grade OpenShift cluster, and a full production-grade Kubernetes cluster to ensure compatibility with production environments |        |
| Verify Prometheus monitoring support                         |        |



#### Version 0.7.x.y...

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Support day 2 operations: backing service password change, cert expiration |  |
| Implement deployment annotations (`nuxeoConfHash`, `clidHash`, backing service credential hashes?) to roll the deployment if CLID or nuxeo.conf or backing service credentials change |  |
| Consider a validating webhook |  |
| Build out unit tests for close to 100% coverage. Extend unit tests to cover more scenarios associated with various mutations of the Nuxeo CR - adding then removing then adding, etc. to ensure the reconciliation logic is robust |  |
| Develop and test the elements needed to qualify the Operator for evaluation as a community Operator. Submit the operator for evaluation. Iterate |   |
| Build on kustomize testing from 0.6.x to provide exemplars for bringing up Nuxeo Clusters using kustomize (https://kubectl.docs.kubernetes.io/pages/examples/kustomize.html) |   |
| kpt (https://googlecontainertools.github.io/kpt/) + kustomize? | |
| Review and augment envtest tests                 |   |
| Support multi-architecture build. Incorporate lint (https://golangci.com?) into the build process |   |
| GitHub build & test automation |   |
| Review the license |   |
| Refactor all the documentation into a user guide | |
| Find someone else to work on this with... | |
| Make the Operator available as a community Operator (https://github.com/operator-framework/community-operators) |   |



#### Other...

These have not been prioritized yet.

| Feature                         | Status |
| ------------------------------- | ------ |
| Integrate with the Service Binding Operator (https://github.com/k8s-service-bindings/spec) as soon as a reasonably table implementation is available |          |
| Phase V Operator Maturity Model |   |
| OperatorHub availability |   |
| Deploy a cluster as a Stateful Set or Deployment |   |
| JetStack Cert Manager integration |   |
| Horizontal Pod Auto-scaling |   |
| cert-utils support? (https://github.com/redhat-cop/cert-utils-operator) | |
| Other?... |   |

## Quick Start

To evaluate the Nuxeo Operator, follow these steps:

1. Install  the Operator
2. Create a Nuxeo CR
3. Access the resulting Nuxeo cluster

#### Install the Nuxeo Operator

This step installs the CRD, creates a namespace *nuxeo-operator-system*, and installs an Operator Deployment and associated RBACs into that namespace to enable the Operator to watch all namespaces:

```
kubectl create -f https://raw.githubusercontent.com/aceeric/nuxeo-operator/master/config/all-in-one/nuxeo-operator-all-in-one.yaml
```

#### Create a Nuxeo CR

Next, deploy a Nuxeo CR. This example is an ephemeral instance of Nuxeo, with all embedded backing services, and one replica:

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: appzygy.net/v1alpha1
kind: Nuxeo
metadata:
  name: nuxeo-server
spec:
  nuxeoImage: nuxeo:LTS-2019
  version: "10.10"
  access:
    hostname: nuxeo-server.apps-crc.testing
  nodeSets:
  - name: cluster
    replicas: 1
    interactive: true
    nuxeoConfig:
      nuxeoPackages:
      - nuxeo-web-ui
EOF
```

Note - you will have to pick a host name for `access/hostname` that your DNS resolves to your Kubernetes cluster. The example above is for Code Ready Containers. The quick-start CR above configures the following items in the `spec`:

1. *nuxeoImage* - the Nuxeo image from Docker Hub
2. *version* - the Nuxeo version - in this case 10.10
3. *access/hostname* - creates an OpenShift Route or Kubernetes Ingress depending on cluster type
4. *nodeSets* - each nodeSet translates to a Deployment object in the cluster with the specified number of replicas. Interactive `true` means to generate a Route/Ingress to the Pods associated with this Deployment
5. *nodeSet/nuxeoConfig/nuxeoPackages* - selects a set of Marketplace packages to install at startup. The `nuxeo-web-ui` package is included in the Nuxeo Docker image and so does not require Nuxeo Marketplace connectivity

After a moment, the Nuxeo Pod should come up:

```shell
$ kubectl get pod
NAME                                    READY     STATUS    RESTARTS   AGE
nuxeo-server-cluster-64dcbb8c89-pbxh9   1/1       Running   0          37s

$ kubectl get nuxeo
NAME           VERSION   HEALTH    AVAILABLE   DESIRED
nuxeo-server   10.10     healthy   1           1

$ kubectl logs nuxeo-server-cluster-64dcbb8c89-pbxh9
/docker-entrypoint.sh: ignoring /docker-entrypoint-initnuxeo.d/*
Nuxeo home:          /opt/nuxeo/server
Nuxeo configuration: /etc/nuxeo/nuxeo.conf
Include template: /opt/nuxeo/server/templates/common-base
Include template: /opt/nuxeo/server/templates/common
Include template: /opt/nuxeo/server/templates/default
Include template: /opt/nuxeo/server/templates/docker
<deleted for brevity>
======================================================================
= Component Loading Status: Pending: 0 / Missing: 0 / Unstarted: 0 / Total: 499
======================================================================
2020-08-31 23:55:48.621 INFO [main] org.apache.catalina.startup.Catalina.start Server startup in [19,127] milliseconds
```

Then, from your browser, access the host name you specified in `access/hostname` and log in to this development instance with Administrator/Administrator:

![](resources/images/nuxeo-ui.jpg)



### Nuxeo Operator and Nuxeo Custom Resource features

The Nuxeo Custom Resource (CR) enables you to configure a Nuxeo Cluster. The Nuxeo Operator watches the Nuxeo CR and reconciles Kubernetes objects to keep the Nuxeo Cluster running and accessible. These objects consist of: Deployments, Pods, Routes, Ingresses, Services, Service Accounts, ConfigMaps, Secrets, and Persistent Volume Claims.

#### The Basics

A `NodeSet` creates a Nuxeo cluster. (Actually, it creates a Deployment, which creates a cluster.) Along with a `nuxeoImage` and a `version`, this establishes the basics of the Nuxeo cluster. The `NodeSets` stanza is a list of `NodeSet`. Each `NodeSet` is a Kubernetes Deployment. The number of `replicas` in the `NodeSet` is the number of Nuxeo Pods. Each replica runs the Nuxeo Docker image specified in the `nuxeoImage` setting. You can omit `nuxeoImage` in which case the Operator defaults to `nuxeo:latest`.

```shell
apiVersion: appzygy.net/v1alpha1
kind: Nuxeo
metadata:
  name: nuxeo-server
spec:
  nuxeoImage: nuxeo:LTS-2019
  version: "10.10"
  nodeSets:
  - name: cluster
    replicas: 1
    interactive: true
    nuxeoConfig:
      nuxeoPackages:
      - nuxeo-web-ui
```

You can optionally specify the list of packages to install via the `nodeSet.nuxeoConfig.nuxeoPackages` list. The *nuxeo-web-ui* package comes pre-loaded with the Nuxeo 10.10 image so you can specify this package without Marketplace connectivity. Other packages require marketplace connectivity. Specify `interactive: true` to make this Nuxeo cluster accessible outside the Kubernetes cluster. (More on this below.)

#### Accessing the Nuxeo cluster from outside the Kubernetes cluster

To access Nuxeo outside of the Kubernetes cluster, in addition to marking one of the node sets as interactive as shown above, you also need to define the `spec/access`. This causes the Nuxeo Operator to create a Kubernetes Ingress or OpenShift Route. The only requirement is a `hostname`:

```shell
apiVersion: appzygy.net/v1alpha1
kind: Nuxeo
metadata:
  name: nuxeo-server
spec:
  access:
    hostname: nuxeo-server.apps-crc.testing
```

The host name must be resolvable by DNS, which may require a system administrator in your organization to set up. The example `hostname` above is for Code Ready Containers. An Ingress/Route will be configured by the Operator to route to a Service (also created by the Operator) and from there to the Pods associated with the `interactive` node set. The Operator will configure *passthrough* termination by default.

The `access` field supports some other settings which are documented in the Nuxeo CRD.

#### Nginx reverse proxy

You can specify that you want Nuxeo accessed via a reverse proxy. The minimal configuration is to add `revProxy.nginx`:

```shell
apiVersion: appzygy.net/v1alpha1
kind: Nuxeo
metadata:
  name: nuxeo-server
spec:
  revProxy:
    nginx:
      secret: tls-secret
  access:
    hostname: nuxeo-server.apps-crc.testing
  ...
```

Currently, Nginx is the only supported reverse proxy. The only required configuration is the `secret` that contains keys `tls.key`, `tls.cert`, and `dhparam` which are used to terminate the Nginx TLS connection. The Operator will supply an Nginx configuration. As a result of the Nuxeo CR configuration shown. the Operator will create a sidecar to run Nginx in the Nuxeo Pod, and automatically configure the Route/Ingress/Service for TLS. Nginx will listen on 8443, and forward to Nuxeo on 8080. In addition, the specified secret will be mounted into the sidecar under `/etc/secrets/`.

 You can supply your own configuration by defining a ConfigMap with keys `nginx.conf` and `proxy.conf`. Then you provide the name of this ConfigMap in the `nginx` stanza:

```shell
spec:
  revProxy:
    nginx:
      secret: tls-secret
      configMap: my-externally-provisioned-nginx-cfgmap
```

Then, something like:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-externally-provisioned-nginx-cfgmap
data:
  nginx.conf: |
    <your unique configuration>
  proxy.conf: |
    <your unique configuration>
EOF
```

Note, however, that you can't change the sidecar mount paths for the TLS secret. You can also specify the `image` and the `imagePullPolicy` for the `nginx` reverse proxy:

```shell
spec:
  revProxy:
    nginx:
      image: my-custom-nginx:1.2.3
      imagePullPolicy: Always
```

### Other NodeSet configuration options

#### Clustering

To enable all of the nodes in a Deployment to participate in a Nuxeo cluster, you enable clustering in the Node Set with the `clusterEnabled` configuration. You must also configure nuxeo binaries in the `storage` stanza such that the binary store can be shared across the Nuxeo cluster. This example allocates the binary store to a (presumably) read/write many PVC and Volume.

```shell
apiVersion: appzygy.net/v1alpha1
kind: Nuxeo
metadata:
  name: nuxeo-server
spec:
  nuxeoImage: nuxeo:LTS-2019
  version: "10.10"
  nodeSets:
  - name: my-cluster
    replicas: 3
    clusterEnabled: true
    interactive: true
    storage:
    - storageType: Binaries
      size: 10Gi
      volumeSource:
        persistentVolumeClaim:
          name: my-externally-provisioned-rw-many-pvc
```

#### Environment Variables

The Nuxeo CR supports direct configuration of environment variables in a way that is consistent with a Pod's environment variable definition:

```shell
spec:
  nodeSets:
  - name: my-cluster
    env:
    - name: MYENV
      value: 100
    - name: MYOTHERENV
      valueFrom:
        secretKeyRef:
          name: my-externally-provisioned-secret
          key: some-secret-key
```

#### Resource Limits

The Nuxeo CR supports direct configuration of resource requests and limits in a way that is consistent with a Pod's resource constraints:

```shell
spec:
  nodeSets:
  - name: my-cluster
    resources:
      limits:
        cpu: 100m
        memory: 300Mi
      requests:
        cpu: 100m
        memory: 300Mi
```

#### Probes

The Nuxeo CR supports direct configuration of Readiness and Liveness probes in a way that is consistent with a Pod's probe configuration:

```shell
spec:
  nodeSets:
  - name: my-cluster
    livenessProbe:
      exec:
        command:
        - "true"
```

The example above creates a liveness probe that just always reports "live", which can be useful for troubleshooting issues in the Nuxeo Pods. By default, the Operator configures probes as shown below:

```shell
httpGet:
  path: /nuxeo/runningstatus
  port: 8080 (or 8443 if Nuxeo is terminating TLS)
  scheme: HTTP (or HTTPS if Nuxeo is terminating TLS)
initialDelaySeconds: 15
timeoutSeconds: 3
periodSeconds: 10
successThreshold: 1
failureThreshold: 3
```

#### General Nuxeo Configuration

The Nuxeo CR supports a stanza called `nuxeoConfig`. This was discussed above in configuring packages to install at startup time. This stanza supports a handful of other general-use Nuxeo configuration settings:

```shell
spec:
  nodeSets:
  - name: my-cluster
    nuxeoConfig:
      javaOpts: "-Xms8m"
      nuxeoTemplates:
      - postgresql
      - s3binaries
      nuxeoPackages:
      - nuxeo-web-ui
      - nuxeo-drive
      nuxeoUrl: https://my-nuxeo.corp.io/nuxeo
      nuxeoName: my-nuxeo
      nuxeoConf:
        inline: |
          my.custom.setting=100
          some.other.setting=false
      tlsSecret: this-is-discussed-below
      jvmPKISecret: so-is-this
      offlinePackages: and-this-also
```

The `javaOpts` configuration supports re-defining the Operator-generated Java Opts, which are: `"-XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap -XX:MaxRAMFraction=1"`. The `nuxeoTemplates` configuration supports installing a set of templates to load at Nuxeo startup. The `nuxeoPackages` we've already seen. The `nuxeoUrl` setting is TODO. The `nuxeoName` setting is TODO. The `nuxeoConf` setting allows definition of nuxeo.conf settings. These can be inline as shown above, or referenced from a Secret or ConfigMap.

#### Configuring Nuxeo to terminate TLS

Nuxeo can be configured to terminate TLS, dispensing with the need for an Nginx sidecar. This is accomplished with the `nodeSet.nuxeoConfig.tlsSecret` configuration:

```shell
spec:
  access:
    hostname: nuxeo-server.apps-crc.testing
  nodeSets:
  - name: cluster
    nuxeoConfig:
      tlsSecret: my-externally-provisioned-tls-secret
```

This causes the Operator to configure Nuxeo to terminate TLS without the need for a sidecar. The referenced secret **must** contain keys: `keystore.jks` and `keystorePass`. The keystore must in fact be a JKS store. (This is a Nuxeo constraint - PKCS12 is not supported.)

A more in-depth presentation is in [Nuxeo TLS termination](docs/test-nuxeo-tls.md) in the docs directory.

#### Configuring a JVM-wide keystore/truststore

It is possible to configure a JVM-wide keystore/truststore using `nodeSet.nuxeoConfig.jvmPKISecret`:

```shell
spec:
  nodeSets:
  - name: cluster
    nuxeoConfig:
      jvmPKISecret: my-externally-provisioned-secret
```

This requires that the referenced secret contains **all** of the following keys: `keyStore`, `keyStoreType`, `keyStorePassword`, `trustStore`, `trustStoreType`, and `trustStorePassword`. With this setting, the Operator mounts secret and defines JVM properties to start the JVM with. (E.g.: `-Djavax.net.ssl.keyStore=/etc/pki/jvm/<the keystore from the secret>`)

A more in-depth presentation is in [JVM PKI](docs/test-jvm-pki.md) in the docs directory.

#### Installing packages in off-line mode

If you're running a Kubernetes cluster that does not have Nuxeo Marketplace connectivity, it is possible to install packages from Secrets and ConfigMaps. This approach is therefore subject to the Kubernetes limitations on the size of ConfigMap/Secret data elements. According to https://kubernetesbyexample.com/secrets/: "A per-secret size limit of 1MB exists". This also applies to config maps.

As of the current Nuxeo version supported by the Operator, the Nuxeo Docker entrypoint script does not readily support mounting a persistent volume into the package installation directory. So for the present, only ConfigMaps and Secrets are supported. Hopefully, a change will be made to the Nuxeo Docker entrypoint that will facilitate using PVs and PVCs. If and when that happens, the Operator will be updated to support this.

Assuming you have marketplace packages in a Secret or ConfigMap, the following example shows how to configure the Nuxeo CR to install them:

```shell
spec:
  nodeSets:
  - name: cluster
    nuxeoConfig:
      offlinePackages:
      - packageName: nuxeo-sample-2.5.3.zip
        valueFrom:
          configMap:
            name: nuxeo-sample-marketplace-package
      - packageName: nuxeo-tree-snapshot-1.2.3.zip
        valueFrom:
          configMap:
            name: nuxeo-tree-snapshot-marketplace-package
```

A more in-depth presentation is in [offline packages](docs/test-offline-packages.md) in the docs directory.

#### Adding custom contributions to a Nuxeo cluster

This feature allows you to configure the Nuxeo CR to reference a Kubernetes resource - such as a ConfigMap or Persistent Volume - that holds a custom Nuxeo contribution. For example, if you had a ConfigMap like this:

```shell
apiVersion: v1
kind: ConfigMap
metadata:
  name: custom-automation-chain-configmap
data:
  nuxeo.defaults: |
    custom-automation-chain.target=.
    custom.automation.chain.property=123
  custom-automation-chain-config.xml:
    <?xml version="1.0"?>
    <component name="custom.automation.chain">
      <extension target="org.nuxeo.ecm.core.operation.OperationServiceComponent"
                 point="chains">
        <chain id="Document.Query.Extension">
          <operation id="Document.Query">
            <param type="string" name="query">SELECT * FROM Document</param>
            <param type="integer" name="pageSize">1</param>
          </operation>
        </chain>
      </extension>
    </component>
```

Then you would use the `contribs` feature in the Nuxeo CR as shown below to load the contribution when the Nuxeo cluster starts:

```shell
spec:
  nodeSets:
  - name: cluster
    contribs:
    - volumeSource:
        configMap:
          name: custom-automation-chain-configmap
      templates:
      - custom-automation-chain
```

Above, you can see that the *contribs* element specifies two things. First, the Kubernetes `VolumeSource` that contains the contribution items. In this case, a ConfigMap. Second, the *templates* element establishes a name for the contribution. It results in the creation of a directory `/etc/nuxeo/nuxeo-operator-config/custom-automation-chain`. And it causes the Operator  to configure the `NUXEO_TEMPLATES` environment var variable with an absolute path reference to that directory.

A more in-depth presentation is in [contributions](docs/test-contribution.md) in the docs directory.

#### Adding your CLID to the CR

The following example shows how to add your CLID to the Nuxeo CR. This format quote-encloses the CLID string and escapes the line endings so the string is provided to the Nuxeo Operator as single line, but the YAML is more readable.

Of course it's not not *necessary* to format it this way - you can just paste the CLID into the Nuxeo CR as is, but that makes it a little harder to read as a YAML file. Note the CLID below is completely fictional.

```shell
spec:
  clid: "12345678-1234-1234-1234-123456789012.9999999999.MV8wYUlL6DoyjhDPagrvh/
    /gzHwfdIaeeaJJBmyuOa1YsYjIxv4HVq6R/5zqW9A24BA89095zf1lPYt3O9ZqHhtg1Uz/
    Wzg87hEAGwKD0QhZVVYHZ5YwbkkGl3sXA45u/jlTrnRsTxBE/K79fO5BDqactRBv86vFm/
    i2e2Zj2MfAVg1WHqAf4zDit0gn/RM19NJE1MtH2v2ukbfY9w2O0dquABCdE84qE90JtnD/
    8CqepiHxwmZe7ajhyPBNaFdNLAZmrkrfM5Ygem/RHMjzgzTEF7uhit0hflJD23Opi9PQD/
    xPFZkJgIqzB1RhhEPy5GifKtvpD==--12345678-1234-1234-1234-123456789012"
```

What's important to understand is that the CLID should be added to the CR just like you would copy it from the Nuxeo Registration website. Specifically: it **must** be a single line - no newlines - and contain exactly one double-dash character sequence ("--") as the line separator. The Operator validates this and injects the CLID into the Nuxeo container as a two-line file, split on that separator. With a valid CLID in the Nuxeo CR, you can install Hot Fixes and Marketplace packages that require a subscription.

### Init containers

The Nuxeo CR Supports custom init containers, containers, and volumes, as illustrated by the following trivial example:

```shell
spec:
  initContainers:
  - name: my-init-container
    image: alpine:latest
    volumeMounts:
    - name: my-volume
      mountPath: /data/test
    command:
    - sh
    - -c
    - echo foo > /data/test/my-file
  containers:
  - name: my-container
    image: alpine:latest
    volumeMounts:
    - name: my-volume
      mountPath: /data/test
    command:
    - sh
    - -c
    - cat /data/test/my-file
  volumes:
  - name: my-volume
    emptyDir: {}  
```

### Backing Services

The Nuxeo CR supports integrating the Nuxeo cluster with backing services. Two examples are presented here as an overview.

The first example shows something called a *pre-configured* backing service. The Nuxeo Operator has pre-configured support for Strimzi Kafka, Elastic Cloud on Kubernetes, and Crunchy Postgres. This means that you can connect Nuxeo to these backing services with minimal YAML. The backing service configuration is in the `backingServices` stanza:

```shell
apiVersion: appzygy.net/v1alpha1
kind: Nuxeo
metadata:
  name: my-nuxeo
spec:
  nuxeoImage: nuxeo:LTS-2019
  version: "10.10"
  access:
    hostname: nuxeo-server.apps-crc.testing
  nodeSets:
  - name: cluster
    replicas: 1
    interactive: true
    nuxeoConfig:
      nuxeoPackages:
      - nuxeo-web-ui
  backingServices:
  - preConfigured:
      type: ECK
      resource: elastic
```

The example above will start a Nuxeo cluster and connect it to Elastic Cloud for Kubernetes (preconfigured type = ECK) using the built-in elastic search user, and TLS. The `resource` key, which specifies that in the namespace that the Nuxeo cluster is running in, there is instance of a `elasticsearch.k8s.elastic.co` resource named `elastic`.

The second example shows an explicit configuration, demonstrating the Operator's support for general-purpose backing service integration. This example connects Nuxeo to a Strimzi-provisioned Kafka cluster. This example assumes that the Strimzi Operator is running, and you've already provisioned a `Kafka` CR named `strimzi`, and a `KafkaUser` CR named `nxkafka` that you want Nuxeo to use in the Kafka broker connection:

```shell
apiVersion: appzygy.net/v1alpha1
kind: Nuxeo
metadata:
  name: my-nuxeo
spec:
  nuxeoImage: nuxeo:LTS-2019
  version: "10.10"
  access:
    hostname: nuxeo-server.apps-crc.testing
  nodeSets:
  - name: cluster
    replicas: 1
    interactive: true
    nuxeoConfig:
      nuxeoPackages:
      - nuxeo-web-ui
  backingServices:
  - name: mykafka
    resources:
    - group: ""
      version: v1
      kind: Secret
      name: strimzi-cluster-ca-cert
      projections:
      - from: ca.password
        env: KAFKA_TRUSTSTORE_PASS
      - from: ca.p12
        mount: truststore.p12
    - group: ""
      version: v1
      kind: Secret
      # Strimzi operator creates a secret with the same name as the KafkaUser
      # CR name, which is 'nxkafka'
      name: nxkafka
      projections:
      - from: user.password
        env: KAFKA_KEYSTORE_PASS
      - from: user.p12
        mount: keystore.p12
    nuxeoConf: |
      kafka.enabled=true
      kafka.ssl=true
      # 'strimzi' is the Kafka cluster name so the Strimzi operator creates a service
      # named [strimzi]-kafka-bootstrap. This is TLS so we know Strimzi will configure the
      # bootstrap service to listen on 9093
      kafka.bootstrap.servers=strimzi-kafka-bootstrap:9093
      kafka.truststore.type=PKCS12
      # here, mykafka is a subdirectory mounted by the Nuxeo operator because
      # this backing service is named mykafka
      kafka.truststore.path=/etc/nuxeo-operator/binding/mykafka/truststore.p12
      kafka.truststore.password=${env:KAFKA_TRUSTSTORE_PASS}
      kafka.keystore.type=PKCS12
      kafka.keystore.path=/etc/nuxeo-operator/binding/mykafka/keystore.p12
      kafka.keystore.password=${env:KAFKA_KEYSTORE_PASS}  
```

This example shows how cluster resources (secrets in this case) are projected into the Nuxeo Pod, and then nuxeo.conf entries are inlined that reference the projected resources as environment variables and filesystem objects. The Nuxeo Operator will add these nuxeo.conf settings to the system-wide nuxeo.conf that it mounts into the Nuxeo container at startup.

The directory `test/backing-services/stacks` has YAML configuring Nuxeo to integrate with a variety of backing services. There are plenty of examples there to draw on.

A more in-depth presentation is in [configuring backing services](docs/backing-services.md) in the docs directory. 

## Developer Quick Start

To bring up a Nuxeo Cluster as a developer, the simplest way is to generate the Nuxeo CRD into your Kubernetes cluster, and then just run the operator on your desktop. As a developer, you're probably running as kube admin, so with your kube config the Operator should have the permission to do what it needs without bothering with a full install.

First, get the repo. Then, there are two Make targets for installing the CRD and running the operator:

```shell
$ make crd-install operator-run &
```

The `crd-install` target creates the Nuxeo CRD in the cluster. The `operator-run` target runs the operator on your desktop, using your kube config, and watching all namespaces in the cluster.

With the Operator running, deploy a Nuxeo CR and start using Nuxeo. Refer to the **Quick Start** above for those steps.

Run `make help` to get a list of all the Make targets and what they do. The goal of the Makefile is to ensure that it can do everything needed to build, test, install, and run the Operator.

