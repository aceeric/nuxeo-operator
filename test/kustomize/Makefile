ROOT      := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
KUBECTL   := kubectl
KUBEAPPLY ?= $(KUBECTL) apply -f -
NXIMGERR  := "These tests require image 'nuxeo' with tag 'LTS-2019-HF29' in the 'images' namespace"

# at the present this Makefile only supports OpenShift

.PHONY: all
all:
	$(MAKE) elastic-builtin-test
	$(MAKE) elastic-filerealm-test
	$(MAKE) strimzi-anonymous-test
	$(MAKE) strimzi-scram-sha-512-test
	$(MAKE) strimzi-mutual-tls-test
	$(MAKE) crunchy-plain-test
	$(MAKE) crunchy-tls-test
# FAILS: $(MAKE) crunchy-mutual-tls-test
	$(MAKE) strimzi-eck-crunchy-test

# the following rules are used to do the small units of work. They are dependencies for the larger
# rules invoked in the section that follows

# remove the backing namespace
.PHONY: clean
clean:
	$(KUBECTL) get ns backing >/dev/null 2>&1 && $(KUBECTL) delete ns backing || true

# make sure the images ns exists
.PHONY: verify-images-ns
verify-images-ns:
	$(KUBECTL) get namespace images >/dev/null 2>&1 || (echo $(NXIMGERR); exit 1)

# stand up ECK Operator
.PHONY: operator-eck-1.1.2
operator-eck-1.1.2:
	kustomize build $(ROOT)/operator-eck-1.1.2 | $(KUBEAPPLY)
	$(ROOT)/scripts/podup ECK && echo ECK Operator is up || (echo ECK Operator not up; exit 1)

# stand up Strimzi Operator
.PHONY: operator-strimzi-0.18.0
operator-strimzi-0.18.0:
	kustomize build $(ROOT)/operator-strimzi-0.18.0 | $(KUBEAPPLY)
	$(ROOT)/scripts/podup STRIMZI && echo Strimzi Operator is up || (echo Strimzi Operator not up; exit 1)

# stand up Crunchy Operator
# kustomize cannot patch the 'disable_fsgroup' setting in 'pgo-deployer-cm' - which
# OpenShift can't tolerate - so sed
.PHONY: operator-crunchy-4.4.0
operator-crunchy-4.4.0:
	KUSTOMIZE_PLUGIN_HOME=$(ROOT) kustomize build --enable_alpha_plugins $(ROOT)/operator-crunchy-4.4.0\
		| sed -e 's|disable_fsgroup: "false"|disable_fsgroup: "true"|' | $(KUBEAPPLY)
	$(ROOT)/scripts/podup PGO && echo Crunchy Operator is up || (echo Crunchy Operator not up; exit 1)

# image puller role (OpenShift only)
.PHONY: security
security:
	kustomize build $(ROOT)/security | $(KUBEAPPLY)

# ElasticSearch built-in user + Nuxeo
.PHONY: elastic-builtin
elastic-builtin:
	kustomize build $(ROOT)/stacks/elastic-builtin | $(KUBEAPPLY)
	$(ROOT)/scripts/podup ELASTIC && echo Elastic Cluster is up || (echo Elastic Cluster not up; exit 1)

# ElasticSearch filerealm user + Nuxeo
.PHONY: elastic-filerealm
elastic-filerealm:
	kustomize build $(ROOT)/stacks/elastic-filerealm | $(KUBEAPPLY)
	$(ROOT)/scripts/podup ELASTIC && echo Elastic Cluster is up || (echo Elastic Cluster not up; exit 1)

# Strimzi anon + Nuxeo
.PHONY: strimzi-anonymous
strimzi-anonymous:
	kustomize build $(ROOT)/stacks/strimzi-anonymous | $(KUBEAPPLY)
	$(ROOT)/scripts/podup KAFKA && echo Kafka Broker is up || (echo Kafka Broker not up; exit 1)

# Strimzi scram sha 512 TLS encrypt + Nuxeo
.PHONY: strimzi-scram-sha-512
strimzi-scram-sha-512:
	kustomize build $(ROOT)/stacks/strimzi-scram-sha-512 | $(KUBEAPPLY)
	$(ROOT)/scripts/podup KAFKA && echo Kafka Broker is up || (echo Kafka Broker not up; exit 1)

# Strimzi mutual tls + Nuxeo
.PHONY: strimzi-mutual-tls
strimzi-mutual-tls:
	kustomize build $(ROOT)/stacks/strimzi-mutual-tls | $(KUBEAPPLY)
	$(ROOT)/scripts/podup KAFKA && echo Kafka Broker is up || (echo Kafka Broker not up; exit 1)

# Crunchy user/pass + Nuxeo
.PHONY: crunchy-plain
crunchy-plain:
	kustomize build $(ROOT)/stacks/crunchy-plain | $(KUBEAPPLY)
	$(ROOT)/scripts/podup CRUNCHY && echo Crunchy database is up || (echo Crunchy not up; exit 1)

# Crunchy user/pass over TLS + Nuxeo
.PHONY: crunchy-tls
crunchy-tls:
	kustomize build $(ROOT)/stacks/crunchy-tls | $(KUBEAPPLY)
	$(ROOT)/scripts/podup CRUNCHY && echo Crunchy database is up || (echo Crunchy not up; exit 1)

# Crunchy mutual TLS + Nuxeo (does not work)
.PHONY: crunchy-mutual-tls
crunchy-mutual-tls:
	kustomize build $(ROOT)/stacks/crunchy-mutual-tls | $(KUBEAPPLY)
	$(ROOT)/scripts/podup CRUNCHY && echo Crunchy database is up || (echo Crunchy not up; exit 1)

# Strimzi mutual TLS, ECK built-in elastic user, Crunchy user/pass + Nuxeo
.PHONY: strimzi-eck-crunchy
strimzi-eck-crunchy:
	kustomize build $(ROOT)/stacks/strimzi-eck-crunchy | $(KUBEAPPLY)
	$(ROOT)/scripts/podup ELASTIC && echo Elastic Cluster is up || (echo Elastic Cluster not up; exit 1)
	$(ROOT)/scripts/podup KAFKA && echo Kafka Broker is up || (echo Kafka Broker not up; exit 1)
	$(ROOT)/scripts/podup CRUNCHY && echo Crunchy database is up || (echo Crunchy not up; exit 1)

# Make sure Nuxeo clean start
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

# helper rules - only nuxeo-lts-2019-hf29-image must be run before running the tests because it creates an
# imagestream. The outputs of all other rules are included in the repo. (You could bypass this dependency
# by patching the Nuxeo image refs to pull directly from Docker Hub and then remove refs to the
# 'verify-images-ns' and 'security' dependencies but eventually I will make this Make file Kubernetes-
# compatible)

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

# requires you to be logged in to the OpenShift internal registry
.PHONY: nuxeo-lts-2019-hf29-image
nuxeo-lts-2019-hf29-image:
	-$(KUBECTL) create namespace images
	# in case no HF's have been downloaded ensure the directory ref'd by the Dockerfile exists
	mkdir -p $(ROOT)/nuxeo-build/hf
	docker build --tag default-route-openshift-image-registry.apps-crc.testing/images/nuxeo:LTS-2019-HF29\
		--file $(ROOT)/nuxeo-build/Dockerfile $(ROOT)/nuxeo-build/
	docker push default-route-openshift-image-registry.apps-crc.testing/images/nuxeo:LTS-2019-HF29

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