# Nuxeo Operator
# Nuxeo Operator

The Nuxeo operator is an OpenShift/Kubernetes Operator written in Go to manage the state of a *Nuxeo* cluster. Nuxeo is an open source digital asset management system. (See https://www.nuxeo.com/). The Operator scaffolding was generated using the Operator SDK v1.0.0.

This project is under development. The current version is 0.6.1. Testing is performed with OpenShift Code Ready Containers (https://github.com/code-ready/crc) and MicroK8s (https://microk8s.io).

### Current Feature Set (as of 0.6.1)

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
| Automate all build / test activities with *GNU Make* |
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
| Support defining resource request/limit in the Nuxeo CR |
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



#### Version 0.6.2 *(in progress)*

| Feature                                                      | Status   |
| ------------------------------------------------------------ | -------- |
| Test the operator in a single namespace, multi-namespaces, and with cluster scope | in-progress |
| Test in a full production-grade OpenShift cluster, and a full production-grade Kubernetes cluster to ensure compatibility with production environments (all work so far has been in CRC and MicroK8s) |   |



#### Version 0.7.x.y...

This iteration makes the Operator available as a Community Operator. This will get chunked into multiple smaller units.

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Support day 2 operations: backing service password change, cert expiration |  |
| Consider a validating webhook |  |
| Verify Prometheus monitoring support (Prometheus not available in CRC out of the box) |  |
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



------

## Building and Running the Operator

TODO