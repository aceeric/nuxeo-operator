ROOT      := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
KUBECTL   := kubectl
KUBEAPPLY ?= $(KUBECTL) apply -f -
NUXHFERR  := "These tests require image 'nuxeo' with tag 'LTS-2019-HF29' in the 'images' namespace"

# at the present this Makefile only supports OpenShift

# the following targets are used to do the small units of work. They are dependencies for the larger
# targets invoked in the section below

.PHONY: clean
clean:
	$(KUBECTL) get ns backing >/dev/null 2>&1 && $(KUBECTL) delete ns backing || true

.PHONY: verify-images-ns
verify-images-ns:
	$(KUBECTL) get namespace images >/dev/null 2>&1 || (echo $(NUXHFERR); exit 1)

.PHONY: operator-eck-1.1.2
operator-eck-1.1.2:
	kustomize build $(ROOT)/operator-eck-1.1.2 | $(KUBEAPPLY)

.PHONY: operator-strimzi-0.18.0
operator-strimzi-0.18.0:
	kustomize build $(ROOT)/operator-strimzi-0.18.0 | $(KUBEAPPLY)

.PHONY: security
security:
	kustomize build $(ROOT)/security | $(KUBEAPPLY)

.PHONY: elastic-builtin
elastic-builtin:
	kustomize build $(ROOT)/stacks/elastic-builtin | $(KUBEAPPLY)

.PHONY: elastic-filerealm
elastic-filerealm:
	kustomize build $(ROOT)/stacks/elastic-filerealm | $(KUBEAPPLY)

.PHONY: strimzi-anonymous
strimzi-anonymous:
	kustomize build $(ROOT)/stacks/strimzi-anonymous | $(KUBEAPPLY)

.PHONY: strimzi-scram-sha-512
strimzi-scram-sha-512:
	kustomize build $(ROOT)/stacks/strimzi-scram-sha-512 | $(KUBEAPPLY)

.PHONY: strimzi-mutual-tls
strimzi-mutual-tls:
	kustomize build $(ROOT)/stacks/strimzi-mutual-tls | $(KUBEAPPLY)

.PHONY: verify-nuxeo
verify-nuxeo:
	$(eval result = $(shell bash -c "$(ROOT)/scripts/verify-nuxeo"))
	if [ FAIL = "$(result)" ]; then exit 1; else exit 0; fi

# these targets each test one connectivity configuration

.PHONY: elastic-builtin-test
elastic-builtin-test: verify-images-ns clean operator-eck-1.1.2 security elastic-builtin verify-nuxeo
	echo COMPLETED ELASTIC BUILTIN TEST

.PHONY: elastic-filerealm-test
elastic-filerealm-test: verify-images-ns clean operator-eck-1.1.2 security elastic-filerealm verify-nuxeo
	echo COMPLETED ELASTIC FILEREALM TEST

.PHONY: strimzi-anonymous-test
strimzi-anonymous-test: verify-images-ns clean operator-strimzi-0.18.0 security strimzi-anonymous verify-nuxeo
	echo COMPLETED STRIMZI ANONYMOUS TEST

.PHONY: strimzi-scram-sha-512-test
strimzi-scram-sha-512-test: verify-images-ns clean operator-strimzi-0.18.0 security strimzi-scram-sha-512 verify-nuxeo
	echo COMPLETED STRIMZI SCRAM-SHA-512 TEST

.PHONY: strimzi-mutual-tls-test
strimzi-mutual-tls-test: verify-images-ns clean operator-strimzi-0.18.0 security strimzi-mutual-tls verify-nuxeo
	echo COMPLETED STRIMZI MUTUAL TLS TEST

# a couple helper targets - only need to be run once to set up the tests

.PHONY: elastic-filerealm-secret
elastic-filerealm-secret:
	$(eval result = $(shell bash -c "$(ROOT)/scripts/crt-filerealm-secret"))
	if [ FAIL = "$(result)" ]; then exit 1; else exit 0; fi

# requires you to be logged in to the OpenShift internal registry
.PHONY: nuxeo-lts-2019-hf29-image
nuxeo-lts-2019-hf29-image:
	-$(KUBECTL) create namespace images
	docker build --tag default-route-openshift-image-registry.apps-crc.testing/images/nuxeo:LTS-2019-HF29\
		--file $(ROOT)/nuxeo-build/Dockerfile $(ROOT)/nuxeo-build/
	docker push default-route-openshift-image-registry.apps-crc.testing/images/nuxeo:LTS-2019-HF29

ifndef VERBOSE
.SILENT:
endif
