ROOT           := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
KUBECTL        := kubectl
KUBEAPPLY      ?= $(KUBECTL) apply -f -
NXIMGERR       := "These tests require image 'nuxeo' with tag 'LTS-2019-HF29' in the 'images' namespace"
IMAGE_REGISTRY := default-route-openshift-image-registry.apps-crc.testing
IMAGE_ORG      := images
STACKS         ?= stacks
FSGROUP_PATCH  := sed -e 's|disable_fsgroup: "false"|disable_fsgroup: "true"|'

# set Make variables for MicroK8s
ifeq ($(TARGET_CLUSTER),MICROK8S)
    IMAGE_REGISTRY := localhost:32000
    MICROK8S       := $(shell which microk8s)
    KUBECTL        := $(MICROK8S) kubectl
    STACKS         := stacks-mk8s
    # don't patch for Kubernetes/MicroK8s
    FSGROUP_PATCH  := cat
endif

.PHONY: all
all:
	$(MAKE) nuxeo-embedded-test
	$(MAKE) elastic-builtin-test
	$(MAKE) elastic-filerealm-test
	$(MAKE) strimzi-anonymous-test
	$(MAKE) strimzi-scram-sha-512-test
	$(MAKE) strimzi-mutual-tls-test
	$(MAKE) crunchy-plain-test
	$(MAKE) crunchy-tls-test
# FAILS: $(MAKE) crunchy-mutual-tls-test
	$(MAKE) strimzi-eck-crunchy-test
	$(MAKE) percona-mongodb-test

# the following rules are used to do the small units of work. They are dependencies for the larger
# rules invoked in the section that follows

# remove the backing namespace
.PHONY: clean
clean:
	$(KUBECTL) get ns backing >/dev/null 2>&1 && $(KUBECTL) delete ns backing || true

# make sure the images ns exists if this is OpenShift/CRC - NOP for MicroK8s
.PHONY: verify-images-ns
verify-images-ns:
ifneq ($(TARGET_CLUSTER),MICROK8S)
	$(KUBECTL) get namespace images >/dev/null 2>&1 || (echo $(NXIMGERR); exit 1)
else
	true
endif

# stand up ECK Operator
.PHONY: operator-eck-1.1.2
operator-eck-1.1.2:
	kustomize build $(ROOT)/operator-eck-1.1.2 | $(KUBEAPPLY)
	$(ROOT)/scripts/podup ECK "$(KUBECTL)" && echo ECK Operator is up || (echo ECK Operator not up; exit 1)

# stand up Strimzi Operator
.PHONY: operator-strimzi-0.18.0
operator-strimzi-0.18.0:
	kustomize build $(ROOT)/operator-strimzi-0.18.0 | $(KUBEAPPLY)
	$(ROOT)/scripts/podup STRIMZI "$(KUBECTL)" && echo Strimzi Operator is up || (echo Strimzi Operator not up; exit 1)

# stand up Percona MongoDB Operator
.PHONY: operator-percona-mongodb-1.4.0
operator-percona-mongodb-1.4.0:
	kustomize build $(ROOT)/operator-percona-mongodb-1.4.0 | $(KUBEAPPLY)
	$(ROOT)/scripts/podup PERCONA "$(KUBECTL)" && echo Percona Operator is up || (echo Percona Operator not up; exit 1)

# stand up Crunchy Operator
# kustomize cannot patch the 'disable_fsgroup' setting in 'pgo-deployer-cm' - which OpenShift can't tolerate - so if
# OpenShift/CRC, it is patched here with sed. If Kubernetes/MK8s - no patch
.PHONY: operator-crunchy-4.4.0
operator-crunchy-4.4.0:
	KUSTOMIZE_PLUGIN_HOME=$(ROOT) kustomize build --enable_alpha_plugins $(ROOT)/operator-crunchy-4.4.0\
		| $(FSGROUP_PATCH) | $(KUBEAPPLY)
	$(ROOT)/scripts/podup PGO "$(KUBECTL)" && echo Crunchy Operator is up || (echo Crunchy Operator not up; exit 1)

# image puller role - required for OpenShift/CRC - benign for Kubernetes/MicroK8s
.PHONY: security
security:
	kustomize build $(ROOT)/security | $(KUBEAPPLY)

# Nuxeo all embedded backing services (just a scaffolding sanity check)
.PHONY: nuxeo-embedded
nuxeo-embedded:
	kustomize build $(ROOT)/$(STACKS)/nuxeo-embedded | $(KUBEAPPLY)

# ElasticSearch built-in user + Nuxeo
.PHONY: elastic-builtin
elastic-builtin:
	kustomize build $(ROOT)/$(STACKS)/elastic-builtin | $(KUBEAPPLY)
	$(ROOT)/scripts/podup ELASTIC "$(KUBECTL)" && echo Elastic Cluster is up || (echo Elastic Cluster not up; exit 1)

# ElasticSearch filerealm user + Nuxeo
.PHONY: elastic-filerealm
elastic-filerealm:
	kustomize build $(ROOT)/$(STACKS)/elastic-filerealm | $(KUBEAPPLY)
	$(ROOT)/scripts/podup ELASTIC "$(KUBECTL)" && echo Elastic Cluster is up || (echo Elastic Cluster not up; exit 1)

# Strimzi anon + Nuxeo
.PHONY: strimzi-anonymous
strimzi-anonymous:
	kustomize build $(ROOT)/$(STACKS)/strimzi-anonymous | $(KUBEAPPLY)
	$(ROOT)/scripts/podup KAFKA "$(KUBECTL)" && echo Kafka Broker is up || (echo Kafka Broker not up; exit 1)

# Strimzi scram sha 512 TLS encrypt + Nuxeo
.PHONY: strimzi-scram-sha-512
strimzi-scram-sha-512:
	kustomize build $(ROOT)/$(STACKS)/strimzi-scram-sha-512 | $(KUBEAPPLY)
	$(ROOT)/scripts/podup KAFKA "$(KUBECTL)" && echo Kafka Broker is up || (echo Kafka Broker not up; exit 1)

# Strimzi mutual tls + Nuxeo
.PHONY: strimzi-mutual-tls
strimzi-mutual-tls:
	kustomize build $(ROOT)/$(STACKS)/strimzi-mutual-tls | $(KUBEAPPLY)
	$(ROOT)/scripts/podup KAFKA "$(KUBECTL)" && echo Kafka Broker is up || (echo Kafka Broker not up; exit 1)

# Crunchy user/pass plain text + Nuxeo
.PHONY: crunchy-plain
crunchy-plain:
	kustomize build $(ROOT)/$(STACKS)/crunchy-plain | $(KUBEAPPLY)
	$(ROOT)/scripts/podup CRUNCHY "$(KUBECTL)" && echo Crunchy database is up || (echo Crunchy not up; exit 1)

# Crunchy user/pass over TLS + Nuxeo
.PHONY: crunchy-tls
crunchy-tls:
	kustomize build $(ROOT)/$(STACKS)/crunchy-tls | $(KUBEAPPLY)
	$(ROOT)/scripts/podup CRUNCHY "$(KUBECTL)" && echo Crunchy database is up || (echo Crunchy not up; exit 1)

# Crunchy mutual TLS + Nuxeo (does not work)
.PHONY: crunchy-mutual-tls
crunchy-mutual-tls:
	kustomize build $(ROOT)/$(STACKS)/crunchy-mutual-tls | $(KUBEAPPLY)
	$(ROOT)/scripts/podup CRUNCHY "$(KUBECTL)" && echo Crunchy database is up || (echo Crunchy not up; exit 1)

# Strimzi mutual TLS, ECK built-in elastic user, Crunchy user/pass plain text + Nuxeo
.PHONY: strimzi-eck-crunchy
strimzi-eck-crunchy:
	kustomize build $(ROOT)/$(STACKS)/strimzi-eck-crunchy | $(KUBEAPPLY)
	$(ROOT)/scripts/podup ELASTIC "$(KUBECTL)" && echo Elastic Cluster is up || (echo Elastic Cluster not up; exit 1)
	$(ROOT)/scripts/podup KAFKA "$(KUBECTL)" && echo Kafka Broker is up || (echo Kafka Broker not up; exit 1)
	$(ROOT)/scripts/podup CRUNCHY "$(KUBECTL)" && echo Crunchy database is up || (echo Crunchy not up; exit 1)

# Percona user/pass over TLS + Nuxeo
.PHONY: percona-tls
percona-tls:
	kustomize build $(ROOT)/$(STACKS)/percona-mongo-explicit | $(KUBEAPPLY)
	$(ROOT)/scripts/podup PMONGO "$(KUBECTL)" && echo Mongo Cluster is up || (echo Mongo Cluster not up; exit 1)

# Make sure Nuxeo clean start
.PHONY: verify-nuxeo
verify-nuxeo:
	$(eval result = $(shell bash -c "$(ROOT)/scripts/verify-nuxeo \"$(KUBECTL)\""))
	if [ FAIL = "$(result)" ]; then exit 1; else exit 0; fi

# these rules each test one connectivity configuration

.PHONY: nuxeo-embedded-test
nuxeo-embedded-test: verify-images-ns clean nuxeo-embedded verify-nuxeo
	echo COMPLETED NUXEO EMBEDDED TEST

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

.PHONY: crunchy-tls-test
crunchy-tls-test: verify-images-ns clean operator-crunchy-4.4.0 security crunchy-tls verify-nuxeo
	echo COMPLETED CRUNCHY TLS TEST

# FAILS
.PHONY: crunchy-mutual-tls-test
crunchy-mutual-tls-test: verify-images-ns clean operator-crunchy-4.4.0 security crunchy-mutual-tls verify-nuxeo
	echo COMPLETED CRUNCHY MUTUAL TLS TEST

.PHONY: strimzi-eck-crunchy-test
strimzi-eck-crunchy-test: verify-images-ns clean operator-crunchy-4.4.0 operator-strimzi-0.18.0 operator-eck-1.1.2 security strimzi-eck-crunchy verify-nuxeo
	echo COMPLETED STRIMZI/ECK/CRUNCHY TEST

.PHONY: percona-mongodb-test
percona-mongodb-test: verify-images-ns clean operator-percona-mongodb-1.4.0 security percona-tls verify-nuxeo
	echo COMPLETED PERCONA MONGODB TEST

# helper rules - only nuxeo-lts-2019-hf29-image must be run before running the tests because it creates an
# imagestream in OpenShift/CRC, or pushes an image into the registry for Kubernetes/MicroK8s. This image is
# referenced in the Nuxeo yaml. The outputs of all other rules are included in the repo. (You could bypass
# this dependency by patching the Nuxeo image in Kustomize to pull Nuxeo directly from Docker Hub and then
# remove refs to the 'verify-images-ns' and 'security' dependencies - though - I don't believe the un-hotfixed
# Nuxeo 2019 LTS image supports the Strimzi SCRAM-SHA-512/TLS test.)

.PHONY: elastic-filerealm-secret
elastic-filerealm-secret:
	$(eval result = $(shell bash -c "$(ROOT)/scripts/crt-filerealm-secret"))
	if [ FAIL = "$(result)" ]; then exit 1; else exit 0; fi

.PHONY: crunchy-backrest-secret
crunchy-backrest-secret:
	$(eval result = $(shell bash -c "$(ROOT)/scripts/crt-backrest-secret"))
	if [ FAIL = "$(result)" ]; then exit 1; else exit 0; fi

.PHONY: crunchy-tls-secrets
crunchy-tls-secrets:
	$(eval result = $(shell bash -c "$(ROOT)/scripts/crt-crunchy-tls-secrets"))
	if [ FAIL = "$(result)" ]; then exit 1; else exit 0; fi

# if OpenShift - requires you to be logged in to the OpenShift internal registry
.PHONY: nuxeo-lts-2019-hf29-image
nuxeo-lts-2019-hf29-image:
	# required for OpenShift/CRC, benign for Kubernetes/MicroK8s
	-$(KUBECTL) create namespace images
	# in case no HF's have been downloaded ensure the directory ref'd by the Dockerfile exists
	mkdir -p $(ROOT)/nuxeo-build/hf
	docker build --tag $(IMAGE_REGISTRY)/$(IMAGE_ORG)/nuxeo:LTS-2019-HF29\
		--file $(ROOT)/nuxeo-build/Dockerfile $(ROOT)/nuxeo-build/
	docker push $(IMAGE_REGISTRY)/$(IMAGE_ORG)/nuxeo:LTS-2019-HF29


.PHONY: help
help:
	echo "$$HELPTEXT"

ifndef VERBOSE
.SILENT:
endif

export HELPTEXT
define HELPTEXT

Some useful notes:

1) To prove that Nuxeo is connecting to PostgreSQL using TLS, once the crunchy TLS
   test has completed successfully then:

   kubectl get pod -l pgo-pg-database=true -o name -n backing | xargs -I% kubectl exec % --\\
    psql -c "SELECT ssl.pid, usename, datname, ssl, client_addr, backend_type, wait_event\\
    FROM pg_catalog.pg_stat_ssl ssl, pg_catalog.pg_stat_activity a\\
    WHERE ssl.pid = a.pid;"

   If you get a result set like the following, then Nuxeo connected via SSL (column ssl=t):

    pid  | usename | datname | ssl | client_addr  | backend_type   | wait_event
   ------+---------+---------+-----+--------------+----------------+-----------
   ...
    3524 | nuxeo   | nuxeo   | t   | 10.116.0.237 | client backend | ClientRead
    3525 | nuxeo   | nuxeo   | t   | 10.116.0.237 | client backend | ClientRead
     439 | nuxeo   | nuxeo   | t   | 10.116.0.237 | client backend | ClientRead
    3526 | nuxeo   | nuxeo   | t   | 10.116.0.237 | client backend | ClientRead
     441 | nuxeo   | nuxeo   | t   | 10.116.0.237 | client backend | ClientRead
   ...
endef