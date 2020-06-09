ROOT             := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
GOROOT           := $(shell go env GOROOT)
OPERATOR_VERSION := 0.1.0
DOCKER_REGISTRY  := default-route-openshift-image-registry.apps-crc.testing
DOCKER_ORG       := images
DOCKER_IMAGE     := nuxeo-operator

.PHONY : all
all: build-operator olm-generate build-operator-image push-operator-image

.PHONY : build-operator
build-operator:
	operator-sdk generate k8s
	operator-sdk generate crds
	go build -o $(ROOT)/build/_output/bin/nuxeo-operator $(ROOT)/cmd/manager

.PHONY : olm-generate
olm-generate:
	operator-sdk generate csv --csv-version $(OPERATOR_VERSION) --update-crds

.PHONY : build-operator-image
build-operator-image:
	docker build --tag $(DOCKER_REGISTRY)/$(DOCKER_ORG)/$(DOCKER_IMAGE):$(OPERATOR_VERSION)\
 		--file $(ROOT)/build/Dockerfile $(ROOT)/build

.PHONY : push-operator-image
push-operator-image:
	docker push $(DOCKER_REGISTRY)/$(DOCKER_ORG)/$(DOCKER_IMAGE):$(OPERATOR_VERSION)

.PHONY : help
help:
	echo "$$HELPTEXT"

ifndef VERBOSE
.SILENT:
endif

.PHONY : print-%
print-%: ; $(info $* is a $(flavor $*) variable set to [$($*)]) @true

export HELPTEXT
define HELPTEXT

This Make file builds the Nuxeo Operator. Options are a) run from within project root, or, b) use the -C
make arg if running from outside project root. Why? 'go build' - as of 1.14 - seems to have module-based
behavior that is current-working-dir dependent. So since this project uses Go modules, the Make needs to
run in the project root directory. This Make file assumes the necessary dependencies (go etc.) are already
installed. The Make file doesn't do any dependency checking, it just runs the full build each time.

Targets:

all                  In order, runs: build-operator olm-generate build-operator-image push-operator-image
build-operator       Builds the operator binary from Go sources.
olm-generate         Generates files under deploy/olm-catalog/nuxeo-operator to support creating an installable
                     Operator that integrates with OLM.
build-operator-image Builds a container image containing the Operator Go binary that was built by the
                     'build-operator' target.
push-operator-image  Pushes the Operator container image to a registry identified by the DOCKER_REGISTRY and
                     DOCKER_ORG varaibles. This supports pushing to a public/private registry, as well as an
                     OpenShift imagesream in the OpenShift cluster. The current version of the Makefile defaults
                     to $(DOCKER_REGISTRY)/$(DOCKER_ORG)/$(DOCKER_IMAGE):$(OPERATOR_VERSION) since this
                     version of the project is targeted at local CRC testing. A future version will improve this.
help                 Prints this help.
print-%              Prints the value of a Make variable. E.g. 'make print-OPERATOR_VERSION' to
                     print the value of 'OPERATOR_VERSION'.

The Make file runs silently unless you provide a VERBOSE arg or variable. E.g.: make VERBOSE=

endef
