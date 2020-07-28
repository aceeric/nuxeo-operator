# Backing Service Tests

This directory provides a Make file to test the backing services that the Nuxeo Operator has pre-configured support for. The idea is, with a minimal bit of configuration, you can connect Nuxeo to a backing service:

```shell
  backingServices:
  - preConfigured:
      type: ECK
      resource: elastic
```

The configuration above in a Nuxeo CR will connect Nuxeo to Elastic Search provisioned by Elastic Cloud for Kubernetes using the built-in `elastic` user. The following pre-configured backing services are supported by the Nuxeo Operator:

1. Elastic Search with the built-in `elastic` user
2. Elastic Search with a custom file realm user, where a cluster secret contains the user name and password of the file realm user
3. Strimzi anonymous (no authentication, no authorization, no encryption)
4. *Strimzi SASL SCRAM SHA 512 authentication is developed but does not work at present - believe this is a Nuxeo defect but I have a todo to verify*. This is SASL authentication, simple authorization, no encryption
5. Strimzi mutual TLS, simple authorization, TLS encryption

On deck:

- Crunchy Postgres
- Zelando Postgres

## Pre-requisites

To run the tests, the following pre-requisites must be satisfied:

- You must be running on OpenShift. Presently, the Make file only supports OpenShift - specifically regarding the Nuxeo image discussed next
- You must create a container image of Nuxeo and push it into the OpenShift integrated registry. (Or, change the image refs in the nuxeo manifests.) The make target `nuxeo-lts-2019-hf29-image` is provided for this. The target creates an `images` namespace, and then runs a Docker build with a Dockerfile the `nuxeo-build` directory. This builds an image from Nuxeo LTS-2019 from Docker Hub, plus all hot fixes in the `hf` directory under `nuxeo-build`:
  - `make nuxeo-lts-2019-hf29-image`
- Prior to that, you *should* download all the hot fixes from the Nuxeo Marketplace into the `nuxeo-build/hf` directory. That's how all the tests were run.
- To support the ElasticSearch file realm connection, you must execute a make target one time to execute the ElasticSearch CLI tool in a Docker container to generate the salted, hashed file realm credentials, and merge them into a Secret manifest:
  - `make elastic-filerealm-secret`
- Finally, you must have the Nuxeo Operator running, watching the `backing` namespace in the cluster. At this stage, I simply run it on the desktop. Subsequently I will modify the Make file to install the operator.

## Test targets

Each test target does the same thing:

1. Deletes the `backing` namespace if it exists so each test starts clean.
2. Installs backing service operator(s) into the backing namespace using Kustomize - each backing service creates the `backing` namespace if it doesn't already exist.
3. Deploys an image puller RBAC that allows the `backing` namespace to pull the Nuxeo image from the `images` namespace.
4. Deploys manifests for backing services, and for Nuxeo, using Kustomize
5. Waits for the Nuxeo Pod to come up with a clean log and then respond to a curl request with a 200 HTTP status code

The following targets are presently available. Note - these need to be run one at a time, because of the nature of the Make file:

```shell
make elastic-builtin-test
make elastic-filerealm-test
make strimzi-anonymous-test
make strimzi-mutual-tls-test
#soon?: make strimzi-scram-sha-512
```

Example:

```shell
$ make strimzi-anonymous-test
namespace "backing" deleted
namespace/backing created
customresourcedefinition.apiextensions.k8s.io/kafkabridges.kafka.strimzi.io unchanged
customresourcedefinition.apiextensions.k8s.io/kafkaconnectors.kafka.strimzi.io unchanged
customresourcedefinition.apiextensions.k8s.io/kafkaconnects.kafka.strimzi.io unchanged
customresourcedefinition.apiextensions.k8s.io/kafkaconnects2is.kafka.strimzi.io unchanged
customresourcedefinition.apiextensions.k8s.io/kafkamirrormaker2s.kafka.strimzi.io unchanged
customresourcedefinition.apiextensions.k8s.io/kafkamirrormakers.kafka.strimzi.io unchanged
customresourcedefinition.apiextensions.k8s.io/kafkarebalances.kafka.strimzi.io unchanged
customresourcedefinition.apiextensions.k8s.io/kafkas.kafka.strimzi.io unchanged
customresourcedefinition.apiextensions.k8s.io/kafkatopics.kafka.strimzi.io unchanged
customresourcedefinition.apiextensions.k8s.io/kafkausers.kafka.strimzi.io unchanged
serviceaccount/strimzi-cluster-operator created
clusterrole.rbac.authorization.k8s.io/strimzi-cluster-operator-global unchanged
clusterrole.rbac.authorization.k8s.io/strimzi-cluster-operator-namespaced unchanged
clusterrole.rbac.authorization.k8s.io/strimzi-entity-operator unchanged
clusterrole.rbac.authorization.k8s.io/strimzi-kafka-broker unchanged
clusterrole.rbac.authorization.k8s.io/strimzi-topic-operator unchanged
rolebinding.rbac.authorization.k8s.io/strimzi-cluster-operator created
rolebinding.rbac.authorization.k8s.io/strimzi-cluster-operator-entity-operator-delegation created
rolebinding.rbac.authorization.k8s.io/strimzi-cluster-operator-topic-operator-delegation created
clusterrolebinding.rbac.authorization.k8s.io/strimzi-cluster-operator unchanged
clusterrolebinding.rbac.authorization.k8s.io/strimzi-cluster-operator-kafka-broker-delegation unchanged
deployment.apps/strimzi-cluster-operator created
rolebinding.rbac.authorization.k8s.io/system:image-puller-backing unchanged
namespace/backing unchanged
kafka.kafka.strimzi.io/strimzi created
nuxeo.nuxeo.com/nuxeo created
COMPLETED STRIMZI ANONYMOUS TEST
```

