ROOT      := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
KUBECTL   := kubectl
KUBEAPPLY ?= $(KUBECTL) apply -f -
NUXHFERR  := "These tests require image 'nuxeo' with tag 'LTS-2019-HF29' in the 'images' namespace"

# at the present this Makefile only supports OpenShift

.PHONY: all
all:
	echo This Makefile does not provide an 'all' target

.PHONY: all-tests
all-tests:
	$(MAKE) elastic-builtin-test
	$(MAKE) elastic-filerealm-test
	$(MAKE) strimzi-anonymous-test
	$(MAKE) strimzi-scram-sha-512-test
	$(MAKE) strimzi-mutual-tls-test
	$(MAKE) crunchy-plain-test
	$(MAKE) crunchy-mutual-tls-test

# the following rules are used to do the small units of work. They are dependencies for the larger
# rules invoked in the section following this

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

# kustomize refuses to patch the 'disable_fsgroup' setting in 'pgo-deployer-cm' - which
# OpenShift can't tolerate - so sed
.PHONY: operator-crunchy-4.4.0
operator-crunchy-4.4.0:
	KUSTOMIZE_PLUGIN_HOME=$(ROOT) kustomize build --enable_alpha_plugins $(ROOT)/operator-crunchy-4.4.0\
		| sed -e 's|disable_fsgroup: "false"|disable_fsgroup: "true"|' | $(KUBEAPPLY)

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

.PHONY: crunchy-plain
crunchy-plain:
	kustomize build $(ROOT)/stacks/crunchy-plain | $(KUBEAPPLY)

.PHONY: verify-nuxeo
verify-nuxeo:
	$(eval result = $(shell bash -c "$(ROOT)/scripts/verify-nuxeo"))
	if [ FAIL = "$(result)" ]; then exit 1; else exit 0; fi

# these rules each test one connectivity configuration

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

.PHONY: crunchy-plain-test
crunchy-plain-test: verify-images-ns clean operator-crunchy-4.4.0 security crunchy-plain verify-nuxeo
	echo COMPLETED CRUNCHY PLAIN TEST

.PHONY: crunchy-mutual-tls-test
crunchy-mutual-tls-test:
	echo CRUNCHY MUTUAL TLS NOT IMPLEMENTED YET

# helper rules - only need to be run once to set up the tests

.PHONY: elastic-filerealm-secret
elastic-filerealm-secret:
	$(eval result = $(shell bash -c "$(ROOT)/scripts/crt-filerealm-secret"))
	if [ FAIL = "$(result)" ]; then exit 1; else exit 0; fi

.PHONY: crunchy-backrest-secret
crunchy-backrest-secret:
	$(eval result = $(shell bash -c "$(ROOT)/scripts/crt-backrest-secret"))
	if [ FAIL = "$(result)" ]; then exit 1; else exit 0; fi

# requires you to be logged in to the OpenShift internal registry
.PHONY: nuxeo-lts-2019-hf29-image
nuxeo-lts-2019-hf29-image:
	-$(KUBECTL) create namespace images
	# in case no HF's have been downloaded ensure the directory ref'd by the Dockerfile exists
	mkdir -p $(ROOT)/nuxeo-build/hf
	docker build --tag default-route-openshift-image-registry.apps-crc.testing/images/nuxeo:LTS-2019-HF29\
		--file $(ROOT)/nuxeo-build/Dockerfile $(ROOT)/nuxeo-build/
	docker push default-route-openshift-image-registry.apps-crc.testing/images/nuxeo:LTS-2019-HF29

ifndef VERBOSE
.SILENT:
endif
