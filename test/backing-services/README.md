# Backing Service Tests

This directory provides a Make file to test the Nuxeo Operator's backing service integration functionality. Most of the tests demonstrate **pre-configured** backing service integration. The idea behind a pre-configured backing service is: with a minimal bit of YAML from the configurer, the Nuxeo Operator can connect a Nuxeo cluster to a backing service cluster:

```shell
  backingServices:
  - preConfigured:
      type: ECK
      resource: elastic
```

The configuration above in a Nuxeo CR will connect Nuxeo to Elastic Search provisioned by Elastic Cloud for Kubernetes (ECK) using the ECK-provisioned built-in `elastic` user.

You're not constrained to using pre-configured backing services. Any backing service whose configuration is describable to the Operator can be integrated with Nuxeo using the more verbose backing services stanzas supported by the Nuxeo CR. (One of the tests demonstrates that.)

However, there are a few backing services that the Nuxeo Operator has pre-configured support for, and for each such backing service, there are a couple of different supported connection options. The following pre-configured backing services / options are supported by the Nuxeo Operator at present:

1. ECK Elastic Search with the built-in `elastic` user
2. ECK Elastic Search with a custom file realm user, where a cluster secret contains the user name and password of the file realm user for Nuxeo, and another secret contains that information in a specific hashed format required by Elastic Search.
3. Strimzi Kafka anonymous (no authentication, no authorization, no encryption)
4. Strimzi Kafka SASL SCRAM SHA 512 authentication with simple authorization and TLS encryption
5. Strimzi Kafka mutual TLS, simple authorization
6. Crunchy Postgres with plain username/password authentication, no encryption
6. Crunchy Postgres with plain username/password authentication and TLS encryption

## Pre-requisites

These tests support either OpenShift via Code Ready Containers (CRC) or Kubernetes via MicroK8s. Future versions of this directory will support native OpenShift and Kubernetes.

To run the tests, the following pre-requisites must be met:

- You must create a container image of Nuxeo and push it into the OpenShift integrated registry - or into the MicroK8s registry. Or, you can patch the image refs in the nuxeo manifests. If you want to build the Nuxeo image, the make target `nuxeo-lts-2019-hf29-image` is provided for this. The target creates an `images` namespace (needed for for CRC), and then runs a Docker build with a Dockerfile in the `nuxeo-build` directory. This builds an image from Nuxeo LTS-2019 from Docker Hub, plus all hot fixes in the `hf` directory under `nuxeo-build`. Then the target pushes this image into the cluster. E.g.:
  - `make nuxeo-lts-2019-hf29-image`
- Prior to that, you *should* download all the hot fixes from the Nuxeo Marketplace into the `nuxeo-build/hf` directory. That's how all the tests were run so it's unknown whether they would pass without a fully patched Nuxeo. Please note that each time Nuxeo comes up - it has to apply all the hot fixes: 29 as of this writing. Because this takes time - the Nuxeo CRs are configured with dummy liveness/readiness probes.
- You must generate the Nuxeo CRD in the cluster. The [Project Makefile](/Makefile) has a target for that: `make apply-crd` or `TARGET_CLUSTER=MICROK8S make apply-crd`
- You must have the Nuxeo Operator running, watching the `backing` namespace in the cluster. At this stage, I simply run it on the desktop. Subsequently, I will modify the Make file to install the operator. Run `make help` in the [Project Makefile](/Makefile) and look at the `operator-install` target for instructions to build and install the operator.
- If you are running on MicroK8s, please see [MicroK8s pre-requisites](docs/microk8s.md) for information on how to configure MicroK8s to support testing the Nuxeo Operator locally.

## Tests

The tests are run using a Make file in this directory. This Make file supports OpenShift CRC by default. To run on MicroK8s, run the make this way:

```shell
TARGET_CLUSTER=MICROK8S make <some target>
```

Each test does the same thing:

1. Deletes the `backing` namespace if it exists so each test starts clean.
2. Installs backing service operator(s) into the backing namespace using Kustomize and kubectl - each backing service creates the `backing` namespace if it doesn't already exist.
3. Deploys an image puller RBAC that allows the `backing` namespace to pull the Nuxeo image from the `images` namespace. (Needed for OpenShift/CRC, ignored in MicroK8s.)
4. Deploys manifests for backing service(s), and for Nuxeo, using Kustomize.
5. Waits for the Nuxeo Pod to come up, then curls Nuxeo waiting for an HTTP 200 status code, then checks for a clean start in the Nuxeo logs.

## Make Rules

The following Make rules are presently available:

| Rule                       | Tests                                                        |
| -------------------------- | ------------------------------------------------------------ |
| all                        | Runs all the tests listed below                              |
| nuxeo-embedded-test        | Runs Nuxeo with all embedded backing services                |
| elastic-builtin-test       | Nuxeo with ElasticSearch provisioned by ECK with the built-in Elastic user over TLS encryption |
| elastic-filerealm-test     | As above, except with a file realm user instead of the built in `elastic` user |
| strimzi-anonymous-test     | Strimzi with no authentication, no authorization, no encryption |
| strimzi-scram-sha-512-test | Strimzi with SASL SCRAM-SHA-512 authentication over TLS with simple authorization |
| strimzi-mutual-tls-test    | Strimzi with mutual TLS, simple authorization                |
| crunchy-plain-test         | Crunchy Postgres with plain username/password login, no encryption |
| crunchy-tls-test           | Crunchy Postgres with plain username/password login, TLS encryption |
| strimzi-eck-crunchy-test   | Strimzi Kafka, ECK Elastic Search, and Crunchy PostgreSQL with explicit backing service configuration rather than pre-configured backing services. This tests a full Nuxeo app stack, and also provide some insight into how to integrate with backing services explicitly. You can see the difference in verbosity between explicit and pre-configured backing service support. |
| percona-mongodb-test       | Percona MongoDB over TLS                                     |
| zalando-minimal-test       | Zalando PostgreSQL plain text (just to test a Postgres variant) |

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
nuxeo.appzygy.net/nuxeo created
COMPLETED STRIMZI ANONYMOUS TEST
```

