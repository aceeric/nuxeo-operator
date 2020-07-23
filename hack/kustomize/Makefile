ROOT      := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
KUBEAPPLY ?= cat

.PHONY: all
all:
	echo There is no 'all' target

.PHONY: eck-operator
eck-operator:
	kustomize build eck-operator | $(KUBEAPPLY)
	echo TODO-ME kubectl wait for operator

.PHONY: strimzi-operator
strimzi-operator:
	kustomize build strimzi-operator | $(KUBEAPPLY)
	echo TODO-ME kubectl wait for operator

.PHONY: nuxeo-operator
nuxeo-operator:
	echo use the main Makefile for this or run on the desktop...

.PHONY: strimzi-2waytls
strimzi-2waytls:
	kubectl wait pod -l control-plane=elastic-operator -n elastic-system  --for condition=ready --timeout=1m
	kubectl wait deployment strimzi-cluster-operator -n backing --for condition=available --timeout=1m
	kustomize build $(ROOT)/components/stacks/strimzi-2waytls | $(KUBEAPPLY)
	kubectl wait pod -l common.k8s.elastic.co/type=elasticsearch -n backing --for condition=ready --timeout=1m
	kubectl wait kafka/strimzi -n backing --for condition=ready --timeout=1m
	kubectl wait pod -l app=nuxeo -n backing --for condition=ready --timeout=1m

ifndef VERBOSE
.SILENT:
endif
