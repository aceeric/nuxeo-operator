# envtest

See: https://sdk.operatorframework.io/docs/building-operators/golang/project_migration_guide/#migrate-your-tests

Prior to v1.0.0 the operator-sdk provided a CLI command `opertator-sdk test` for E2E testing. The prior version of the nuxeo operator supported that with a minimal e2e test. That CLI command is now deprecated in favor of *envtest* which uses *ginkgo*.

This version (0.7.1) of the operator just implements one basic test to evaluate the *envtest* framework. Thoughts so far:

1. There appear to be subtle differences in how resource Go structures are validated by the local runtime framework that differ from how the cluster validates. As a result the test code had to initialize the go structs with - for example - explicit arrays rather than nil.
2. Could not get the test framework to work with a cluster scoped CRD so - hacked the test to convert the cluster-scoped Nuxeo CRD to namespace scope but - need to resolve this.
3. Not sure about writing lots of testing code to simulate cluster use cases when tests could be scripted in an actual cluster more succinctly. Meaning: in CI mode, these `envtest` tests would be running on a faked cluster so they don't really do anything different from the existing unit tests.
4. I already had a significant investment in stretchr/testify (https://github.com/stretchr/testify) for unit tests which I've elected to preserve for now.

In summary - need to decide what the strategy should be for e2e/integration/unit testing. This current effort with *envtest* is just exploratory. Unit tests continue to be in *testify*.