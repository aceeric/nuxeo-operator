ROOT             := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
GOROOT           := $(shell go env GOROOT)
OPERATOR_VERSION := 0.1.0
DOCKER_REGISTRY  := default-route-openshift-image-registry.apps-crc.testing
DOCKER_ORG       := images
DOCKER_IMAGE     := nuxeo-operator

.PHONY : all
all: build-operator gen-csv docker-build docker-push-imagestream

.PHONY : build-operator
build-operator:
	operator-sdk generate k8s
	operator-sdk generate crds
	go build -o $(ROOT)/build/_output/bin/nuxeo-operator $(ROOT)/cmd/manager

.PHONY : gen-csv
gen-csv:
	operator-sdk generate csv --csv-version $(OPERATOR_VERSION) --update-crds

.PHONY : docker-build
docker-build:
	docker build --tag $(DOCKER_REGISTRY)/$(DOCKER_ORG)/$(DOCKER_IMAGE):$(OPERATOR_VERSION)\
 		--file $(ROOT)/build/Dockerfile $(ROOT)/build

# This recipe pushes the docker image from the local docker cache to an imagestream in a project called $(DOCKER_ORG).
# This is to support local CRC development/testing. The Operator CSV references the image just as 'nuxeo-operator:0.1.0'
# This recipe assumes the current user/shell is logged in to the cluster and also docker logged in to the cluster.
# The .ONESHELL directive is used to enable setting and then evaluating the RESULT variable. In addition, POSIX
# redirection is used to test whether the expected namespace/project exists.
.PHONY : docker-push-imagestream
.ONESHELL:
docker-push-imagestream:
	RESULT=$(shell kubectl get namespace ${DOCKER_ORG} >/dev/null 2>&1 && echo "PASS" || echo "FAIL")
	if [ "$$RESULT" = "FAIL" ]; then echo "Missing ${DOCKER_ORG} namespace/project to push the image to"; exit 1; fi
	docker push $(DOCKER_REGISTRY)/$(DOCKER_ORG)/$(DOCKER_IMAGE):$(OPERATOR_VERSION)

# This recipe will support pushing the docker image to a named registry in a subsequent version
.PHONY : docker-push-registry
.ONESHELL:
docker-push-registry:
	echo NOT IMPLEMENTED YET

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

all                      In order, runs: build-operator gen-csv docker-build docker-push-imagestream
build-operator           Builds the operator binary
gen-csv                  Generates files under deploy/olm-catalog/nuxeo-operator to support installing
                         the Operator into OLM
docker-build             Builds a Docker image containing the Operator binary
docker-push-imagestream  Pushes the Docker image to an imagesream in the OpenShift cluster to support local
                         testing
help                     Prints this help
print-%                  Prints the value of a Make variable. E.g. 'make print-OPERATOR_VERSION' to
                         print the value of 'OPERATOR_VERSION'

The Make file runs silently unless you provide a VERBOSE arg or variable. E.g.: make VERBOSE=

endef
