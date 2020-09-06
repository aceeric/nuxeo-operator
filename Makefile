OPERATOR_VERSION        := 0.6.2
BUNDLE_CHANNELS         ?= alpha
BUNDLE_DEFAULT_CHANNEL  ?= alpha
BUNDLE_METADATA_OPTS    ?= --channels $(BUNDLE_CHANNELS) --default-channel $(BUNDLE_DEFAULT_CHANNEL)
OPERATOR_IMAGE_REGISTRY ?= docker.io
OPERATOR_IMAGE_ORG      ?= appzygy
OPERATOR_IMAGE_NAME     ?= nuxeo-operator
OPERATOR_IMAGE          := $(OPERATOR_IMAGE_REGISTRY)/$(OPERATOR_IMAGE_ORG)/$(OPERATOR_IMAGE_NAME):$(OPERATOR_VERSION)
BUNDLE_IMAGE            := $(OPERATOR_IMAGE_REGISTRY)/$(OPERATOR_IMAGE_ORG)/$(OPERATOR_IMAGE_NAME)-bundle:$(OPERATOR_VERSION)
INDEX_IMAGE             := $(OPERATOR_IMAGE_REGISTRY)/$(OPERATOR_IMAGE_ORG)/$(OPERATOR_IMAGE_NAME)-index:$(OPERATOR_VERSION)
CRD_OPTIONS             ?= "crd:trivialVersions=true"
ROOT                    := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
UNIT_TEST_ARGS          ?= -v -coverprofile cp.out
KUBECTL                 ?= kubectl
ENVTEST_ASSETS_DIR      := $(ROOT)/testbin

# set Make variables for MicroK8s testing
ifeq ($(TARGET_CLUSTER),MICROK8S)
    KUBECTL := microk8s kubectl
endif

ifeq (, $(shell which kustomize))
    $(error "Missing required command: kustomize")
endif
ifeq (, $(shell which controller-gen))
    $(error "Missing required command: controller-gen")
endif
ifeq (, $(shell which opm))
    $(error "Missing required command: opm")
endif

.PHONY : all
all:
	echo Run 'make help' to see a list of available targets

# for desktop testing
.PHONY : operator-build
operator-build: generate fmt vet
	CGO_ENABLED=0 GO111MODULE=on go build -ldflags "-X 'main.version=$(OPERATOR_VERSION)'" -a -o $(ROOT)/manager $(ROOT)/main.go

# run the operator on the desktop for local testing using your kube config. 'WATCH_NAMESPACE=' means watch all
.PHONY : operator-run
operator-run:
	WATCH_NAMESPACE= go run ./main.go

# run operator unit tests
.PHONY : operator-unit-test
operator-unit-test:
	go test $(UNIT_TEST_ARGS) -run=UnitTestSuite $(ROOT)/controllers/nuxeo/...

# run operator CI-style integration tests using Ginkgo (replaces pre-v1.0.0 'operator-sdk test')
.PHONY : operator-envtest
operator-envtest: SHELL := bash
operator-envtest:
	mkdir -p $(ENVTEST_ASSETS_DIR)
	test -f $(ENVTEST_ASSETS_DIR)/setup-envtest.sh ||\
 		curl -sSLo $(ENVTEST_ASSETS_DIR)/setup-envtest.sh\
 		https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/master/hack/setup-envtest.sh
	source $(ENVTEST_ASSETS_DIR)/setup-envtest.sh &&\
 		fetch_envtest_tools $(ENVTEST_ASSETS_DIR) &&\
		setup_envtest_env $(ENVTEST_ASSETS_DIR) &&\
		go test $(ROOT)/test/envtest

.PHONY : operator-image-build
operator-image-build:
	docker build . -t $(OPERATOR_IMAGE) --build-arg OPERATOR_VERSION=$(OPERATOR_VERSION)

.PHONY : operator-image-push
operator-image-push:
	docker push $(OPERATOR_IMAGE)

# install the operator, CRDs, RBACs directly into the cluster. Use sed as temp work-around for:
# https://github.com/kubernetes-sigs/controller-tools/pull/480
.PHONY : operator-install
operator-install: operator-clean
	cd config/manager && kustomize edit set image controller=$(OPERATOR_IMAGE)
	kustomize build config/default | sed -e '/x-kubernetes-list-map-keys:/,+3 d' | $(KUBECTL) create -f -

.PHONY : operator-clean
operator-clean:
	-$(KUBECTL) delete clusterrole,clusterrolebinding,crd,namespace -l app=nuxeo-operator

# generate config/crd/bases
.PHONY : crd-gen
crd-gen:
	controller-gen $(CRD_OPTIONS) crd paths=$(ROOT)/api/... output:crd:dir=$(ROOT)/config/crd/bases
	# controller-gen creates this file inexplicably
	-rm $(ROOT)/config/appzygy.net_nuxeos.yaml

# Install CRD(s) into cluster. Use create/replace because apply fails if CRD size too large. Use sed as temp
# work-around for: https://github.com/kubernetes-sigs/controller-tools/pull/480.
.PHONY : crd-install
crd-install:
	$(KUBECTL) get crd/nuxeos.appzygy.net >/dev/null 2>&1 &&\
		(kustomize build config/crd | sed -e '/x-kubernetes-list-map-keys:/,+3 d' | $(KUBECTL) replace -f -) ||\
		(kustomize build config/crd | sed -e '/x-kubernetes-list-map-keys:/,+3 d' | $(KUBECTL) create -f -)

# Remove CRD(s) from cluster
.PHONY : crd-uninstall
crd-uninstall:
	$(KUBECTL) delete crd -l app=nuxeo-operator

.PHONY : fmt
fmt:
	go fmt ./...

.PHONY : vet
vet:
	go vet ./...

# generate "zz" deep copy Go code
.PHONY :
generate:
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

# generate OLM bundle. Temp work-around to remove namespace from service account yaml for now per:
# https://github.com/operator-framework/operator-sdk/issues/3809. See above for x-kubernetes... sed patch
.PHONY : olm-bundle-generate
olm-bundle-generate:
	operator-sdk generate kustomize manifests -q
	cd $(ROOT)/config/manager && kustomize edit set image controller=$(OPERATOR_IMAGE)
	kustomize build config/manifests | operator-sdk generate bundle -q --overwrite --version $(OPERATOR_VERSION) $(BUNDLE_METADATA_OPTS)
	-rm -f $(ROOT)/bundle/manifests/*serviceaccount.yaml
	sed -i '/x-kubernetes-list-map-keys:/,+3 d' $(ROOT)/bundle/manifests/appzygy.net_nuxeos.yaml
	operator-sdk bundle validate $(ROOT)/bundle

# creates nuxeo-operator-bundle in local docker cache and pushes to docker hub
.PHONY : olm-bundle-build
olm-bundle-build:
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMAGE) .
	docker push $(BUNDLE_IMAGE)

# creates the OLM index and pushes to docker hub. Note - if the index refs the bundle by tag and the bundle tag is
# cached on the machine, then it is not possible to have the index select an updated bundle.
.PHONY : olm-index-create
olm-index-create:
	$(eval SHA = $(shell docker inspect docker.io/appzygy/nuxeo-operator-bundle:$(OPERATOR_VERSION) | grep 'appzygy/nuxeo-operator-bundle@sha256:' | cut -d : -f2 | tr -d '"'))
	$(eval BUNDLE_IMAGE = $(OPERATOR_IMAGE_REGISTRY)/$(OPERATOR_IMAGE_ORG)/$(OPERATOR_IMAGE_NAME)-bundle@sha256:$(SHA))
	opm index add --bundles $(BUNDLE_IMAGE) --tag $(INDEX_IMAGE) --container-tool docker
	docker push $(INDEX_IMAGE)

# create a test catalog source to test OLM subscriptions outside of community operators etc.
.PHONY : olm-catalogsource-gen
olm-catalogsource-gen:
	-$(KUBECTL) create namespace nuxeo-test
	$(eval SHA = $(shell docker inspect docker.io/appzygy/nuxeo-operator-index:$(OPERATOR_VERSION) | grep 'appzygy/nuxeo-operator-index@sha256:' | cut -d : -f2 | tr -d '"'))
	sed -e 's/:$(OPERATOR_VERSION)/@sha256:$(SHA)/' $(ROOT)/hack/olm/nuxeo-operator-catalogsource.yaml | $(KUBECTL) apply -f -

# create an operator group and subscription to the nuxeo operator
.PHONY : olm-subscribe
olm-subscribe:
	-$(KUBECTL) create namespace nuxeo-test
	$(KUBECTL) apply -f $(ROOT)/hack/olm/nuxeo-operator-subscription.yaml

.PHONY : help
help:
	echo "$$HELPTEXT"

ifndef VERBOSE
.SILENT:
endif

.PHONY : print-%
print-%:
	$(info $($*))

export HELPTEXT
define HELPTEXT

This make file provides the following targets

Supporting desktop build/test
  operator-unit-test    Runs the operator unit tests
  operator-build        Builds the operator go binary on the local desktop, for local desktop testing
  operator-run          Runs the operator go code on the desktop using your kube config, watching all namespaces
  operator-envtest      Runs the operator CI-style integration tests using Ginkgo and a fake cluster

Preparing the operator for installation
  operator-image-build  Builds the operator docker image
  operator-image-push   Pushes the docker image built by the operator-image-build target to Docker Hub
  operator-install      Installs the operator, CRDs & RBACs directly into the cluster. Creates namespace
                        nuxeo-operator-system for the Operator Deployment. Used to test Operator functionality
                        independently of OLM
  operator-clean        Undoes operator-install

CRD-related targets
  crd-gen               Generates config/crd/bases/appzygy.net_nuxeos.yaml
  crd-install           Creates/replaces the Nuxeo CRD in cluster
  crd-uninstall         Removes the Nuxeo CRD from cluster

Low-level targets used by other targets
  fmt                   Runs go fmt
  vet                   Runs go vet
  generate              Generates "zz_..." deep copy Go code

OLM-related targets
  olm-bundle-generate   Generates an OLM bundle into the bundle directory. Note - this re-generates the CSV
                        so only needs to be done when the CRD or RBAC changes.
  olm-bundle-build      Creates nuxeo-operator-bundle in the local Docker cache and then pushes the image to
                        Docker Hub
  olm-index-create      Creates OLM index nuxeo-operator-index in the local Docker cache and then pushes the image
                        to Docker Hub
  olm-catalogsource-gen Creates namespace nuxeo-test and then creates a CatalogSource in that namespace to support
                        instantiating the Operator using an OLM subscription
  olm-subscribe         In the nuxeo-test namespace, creates an OLM OperatorGroup with target namespace 'nuxeo-test',
                        and a Subscription to the Nuxeo Operator to test OLM subscription functionality
Miscellaneous

  help                  Prints this help
  print-%               Prints the value of a Make variable. E.g. 'make print-OPERATOR_VERSION' to print the value
                        of 'OPERATOR_VERSION'
endef
