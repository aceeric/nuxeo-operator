# Pre-Configured Backing Service Tests

This directory provides a Make file to test the backing services that the Nuxeo Operator has **pre-configured** support for. The idea is, with a minimal bit of configuration, you can connect Nuxeo to a backing service:

```shell
  backingServices:
  - preConfigured:
      type: ECK
      resource: elastic
```

The configuration above in a Nuxeo CR will connect Nuxeo to Elastic Search provisioned by Elastic Cloud for Kubernetes (ECK) using the built-in `elastic` user.  You're not constrained to using pre-configured backing services. Any backing service whose configuration is discoverable in the cluster can be integrated with Nuxeo using the more verbose backing services stanzas supported by the Nuxeo CR.

However, there are a few backing services that the Nuxeo Operator has pre-configured support for. For each backing service, there are a couple of different connection options.

This directory and Make file test those pre-configured backing services and connection options. The following pre-configured backing services / options are supported by the Nuxeo Operator at present:

1. ECK Elastic Search with the built-in `elastic` user
2. ECK Elastic Search with a custom file realm user, where a cluster secret contains the user name and password of the file realm user for Nuxeo, and another secret contains that information in a specific hashed format required by Elastic Search.
3. Strimzi Kafka anonymous (no authentication, no authorization, no encryption)
4. Strimzi Kafka SASL SCRAM SHA 512 authentication with simple authorization and TLS encryption
5. Strimzi Kafka mutual TLS, simple authorization
6. Crunchy Postgres with plain username/password authentication, no encryption
6. Crunchy Postgres with plain username/password authentication and TLS encryption

## Pre-requisites

To run the tests, the following pre-requisites must be met:

- You must be running on OpenShift. Presently, the Backing Service Test Make file only supports OpenShift - specifically regarding the Nuxeo image discussed next. The Make file will be improved in a later version to support native Kubernetes.
- You must create a container image of Nuxeo and push it into the OpenShift integrated registry. Or, you must change the image refs in the nuxeo manifests. If you want to build the image, the make target `nuxeo-lts-2019-hf29-image` is provided for this. The target creates an `images` namespace, and then runs a Docker build with a Dockerfile the `nuxeo-build` directory. This builds an image from Nuxeo LTS-2019 from Docker Hub, plus all hot fixes in the `hf` directory under `nuxeo-build`:
  - `make nuxeo-lts-2019-hf29-image`.
- Prior to that, you *should* download all the hot fixes from the Nuxeo Marketplace into the `nuxeo-build/hf` directory. That's how all the tests were run so it's unknown whether they would pass without a fully patched Nuxeo.
- Finally, you must have the Nuxeo Operator running, watching the `backing` namespace in the cluster. At this stage, I simply run it on the desktop. Subsequently, I will modify the Make file to install the operator. Run `make help` and look at the `operator-install` target in the [Project Makefile](/Makefile) for instructions to build and install the operator.

## Tests

The tests are run using a Make file in this directory. Each test does the same thing:

1. Deletes the `backing` namespace if it exists so each test starts clean.
2. Installs backing service operator(s) into the backing namespace using Kustomize and kubectl - each backing service creates the `backing` namespace if it doesn't already exist.
3. Deploys an image puller RBAC that allows the `backing` namespace to pull the Nuxeo image from the `images` namespace. (Hence OpenShift)
4. Deploys manifests for backing services, and for Nuxeo, using Kustomize.
5. Waits for the Nuxeo Pod to come up, then curls Nuxeo waiting for an HTTP 200 status code, then checks for a clean start in the Nuxeo logs.

## Make Rules

The following Make rules are presently available:

| Rule                       | Tests                                                        |
| -------------------------- | ------------------------------------------------------------ |
| all                        | Runs all the tests listed below                              |
| elastic-builtin-test       | Nuxeo with ElasticSearch provisioned by ECK with the built-in Elastic user over TLS encryption |
| elastic-filerealm-test     | As above, except with a file realm user instead of the built in `elastic` user |
| strimzi-anonymous-test     | Strimzi with no authentication, no authorization, no encryption |
| strimzi-scram-sha-512-test | Strimzi with SASL SCRAM-SHA-512 authentication over TLS with simple authorization |
| strimzi-mutual-tls-test    | Strimzi with mutual TLS, simple authorization                |
| crunchy-plain-test         | Crunchy Postgres with plain username/password login, no encryption |
| crunchy-tls-test           | Crunchy Postgres with plain username/password login, TLS encryption |

## Example

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

