PROJECT_NAME:=istio-workspace
PACKAGE_NAME:=github.com/maistra/istio-workspace

OPERATOR_NAMESPACE?=istio-workspace-operator
OPERATOR_WATCH_NAMESPACE?=""
TEST_NAMESPACE?=bookinfo

PROJECT_DIR:=$(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
$(shell mkdir -p $(PROJECT_DIR)/bin/)
export PROJECT_DIR
export GO111MODULE:=on
BUILD_DIR:=$(PROJECT_DIR)/build
BINARY_DIR:=$(PROJECT_DIR)/dist
BINARY_NAME:=ike
TEST_BINARY_NAME:=test-service
TPL_BINARY_NAME:=tpl
ASSETS:=pkg/assets/isto-workspace-deploy.go
ASSET_SRCS=$(shell find ./template -name "*.yaml" -o -name "*.tpl" -o -name "*.var" | sort)
MANIFEST_DIR:=$(PROJECT_DIR)/deploy

GOPATH_1:=$(shell echo ${GOPATH} | cut -d':' -f 1)
GOBIN=$(GOPATH_1)/bin
PATH:=${GOBIN}/bin:$(PROJECT_DIR)/bin:$(PATH)

# Determine this makefile's path.
# Be sure to place this BEFORE `include` directives, if any.
THIS_MAKEFILE:=$(lastword $(MAKEFILE_LIST))

CHANNELS?="alpha"
DEFAULT_CHANNEL?="alpha"
BUNDLE_CHANNELS:=--channels=$(CHANNELS)
BUNDLE_DEFAULT_CHANNEL:=--default-channel=$(DEFAULT_CHANNEL)
BUNDLE_METADATA_OPTS?=$(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Call this function with $(call header,"Your message") to see underscored green text
define header =
@echo -e "\n\e[92m\e[4m\e[1m$(1)\e[0m\n"
endef

##@ Default target (all you need - just run "make")
.DEFAULT_GOAL:=all
.PHONY: all
all: deps tools generate format lint compile test ## Runs 'deps operator-codegen format lint compile test' targets

###########################################################################
# Build configuration
###########################################################################

OS:=$(shell uname -s)
export OS
GOOS?=$(shell echo $(OS) | awk '{print tolower($$0)}')
GOARCH:=amd64

BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GITUNTRACKEDCHANGES:=$(shell git status --porcelain --untracked-files=no)
COMMIT:=$(shell git rev-parse --short HEAD)
ifneq ($(GITUNTRACKEDCHANGES),)
	COMMIT:=$(COMMIT)-dirty
endif

IKE_VERSION?=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
OPERATOR_VERSION:=$(IKE_VERSION:v%=%)
export OPERATOR_VERSION
GIT_TAG?=$(shell git describe --tags --abbrev=0 --exact-match > /dev/null 2>&1; echo $$?)
ifneq ($(GIT_TAG),0)
	ifeq ($(origin IKE_VERSION),file)
		IKE_VERSION:=$(IKE_VERSION)-next
		OPERATOR_VERSION:=$(OPERATOR_VERSION)-next
	endif
endif

GOBUILD:=GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0
RELEASE?=false
LDFLAGS="-w -X ${PACKAGE_NAME}/version.Release=${RELEASE} -X ${PACKAGE_NAME}/version.Version=${IKE_VERSION} -X ${PACKAGE_NAME}/version.Commit=${COMMIT} -X ${PACKAGE_NAME}/version.BuildTime=${BUILD_TIME}"
SRC_DIRS:=./api ./controllers ./pkg ./cmd ./version ./test
TEST_DIRS:=./e2e
SRCS:=$(shell find ${SRC_DIRS} -name "*.go")

k8s:=kubectl
ifneq (, $(shell which oc))
	k8s=oc
endif

IMG_BUILDER:=docker
## Prefer to use podman
ifneq (, $(shell which podman))
	IMG_BUILDER=podman
endif

###########################################################################
##@ Build
###########################################################################

.PHONY: build-ci
build-ci: deps tools format compile test # Like 'all', but without linter which is executed as separated PR check

.PHONY: compile
compile: deps generate format $(BINARY_DIR)/$(BINARY_NAME) ## Compiles binaries

.PHONY: test
test: generate ## Runs tests
	$(call header,"Running tests")
	ginkgo -r -v --skipPackage=e2e ${args}

.PHONY: test-e2e
test-e2e: compile ## Runs end-to-end tests
	$(call header,"Running end-to-end tests")
	ginkgo e2e/ -v -p ${args}

.PHONY: clean
clean: ## Removes build artifacts
	rm -rf $(BINARY_DIR) $(PROJECT_DIR)/bin/ vendor/ bundle/

.PHONY: deps
deps: ## Fetches all dependencies
	$(call header,"Fetching dependencies")
	go mod download
	go mod vendor

.PHONY: format
format: $(SRCS) ## Removes unneeded imports and formats source code
	$(call header,"Formatting code")
	goimports -l -w -e $(SRC_DIRS) $(TEST_DIRS)

.PHONY: lint-prepare
lint-prepare: deps tools generate compile

.PHONY: release-prepare
release-prepare: deps tools generate format

.PHONY: lint
lint: lint-prepare ## Concurrently runs a whole bunch of static analysis tools
	$(call header,"Running a whole bunch of static analysis tools")
	golangci-lint run

.PHONY: generate
generate: tools $(PROJECT_DIR)/$(ASSETS) $(PROJECT_DIR)/api ## Generates k8s manifests and srcs
	$(call header,"Generates CRDs et al")
	controller-gen $(CRD_OPTIONS) rbac:roleName=istio-workspace paths="./..." output:crd:artifacts:config=config/crd/bases
	controller-gen object:headerFile="scripts/boilerplate.txt" paths="./..."
	$(call header,"Generates clientset code")
	chmod +x ./vendor/k8s.io/code-generator/generate-groups.sh
	GOPATH=$(GOPATH_1) ./vendor/k8s.io/code-generator/generate-groups.sh client \
		$(PACKAGE_NAME)/pkg/client \
		$(PACKAGE_NAME)/api \
		"maistra:v1alpha1" \
		--go-header-file ./scripts/boilerplate.txt

.PHONY: version
version:
	@echo $(IKE_VERSION)

$(BINARY_DIR):
	[ -d $@ ] || mkdir -p $@

$(BINARY_DIR)/$(BINARY_NAME): $(BINARY_DIR) $(SRCS)
	$(call header,"Compiling... carry on!")
	${GOBUILD} go build -ldflags ${LDFLAGS} -o $@ ./cmd/$(BINARY_NAME)/

$(BINARY_DIR)/$(TEST_BINARY_NAME): $(BINARY_DIR) $(SRCS) test/cmd/test-service/html.go
	$(call header,"Compiling test service... carry on!")
	${GOBUILD} go build -ldflags ${LDFLAGS} -o $@ ./test/cmd/$(TEST_BINARY_NAME)/

test/cmd/test-service/main.pb.go: $(PROJECT_DIR)/bin/protoc test/cmd/test-service/main.proto
	$(call header,"Compiling test proto... carry on!")
	$(PROJECT_DIR)/bin/protoc -I test/cmd/test-service/ test/cmd/test-service/main.proto --go_out=plugins=grpc:test/cmd/test-service

test/cmd/test-service/html.go: test/cmd/test-service/assets/index.html
	$(call header,"Compiling test assets... carry on!")
	go-bindata -o test/cmd/test-service/html.go -pkg main -prefix test/cmd/test-service/assets test/cmd/test-service/assets/*

.PHONY: compile-test-service
compile-test-service: test/cmd/test-service/html.go test/cmd/test-service/main.pb.go $(BINARY_DIR)/$(TEST_BINARY_NAME)

$(BINARY_DIR)/$(TPL_BINARY_NAME): $(BINARY_DIR) $(SRCS)
	$(call header,"Compiling tpl processor... carry on!")
	${GOBUILD} go build -ldflags ${LDFLAGS} -o $@ ./cmd/$(TPL_BINARY_NAME)/

$(PROJECT_DIR)/$(ASSETS): $(ASSET_SRCS)
	$(call header,"Adds assets to the binary")
	go-bindata -o $(ASSETS) -nometadata -pkg assets -ignore 'examples/' $(ASSET_SRCS)

###########################################################################
## Setup
###########################################################################

# go-get-tool will 'go get' any package $2 and install it to $1.
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

.PHONY: tools
tools: $(PROJECT_DIR)/bin/operator-sdk $(PROJECT_DIR)/bin/controller-gen $(PROJECT_DIR)/bin/kustomize
tools: $(PROJECT_DIR)/bin/golangci-lint $(PROJECT_DIR)/bin/goimports $(PROJECT_DIR)/bin/go-bindata
tools: $(PROJECT_DIR)/bin/protoc-gen-go $(PROJECT_DIR)/bin/yq $(PROJECT_DIR)/bin/ginkgo

$(PROJECT_DIR)/bin/yq:
	$(call header,"Installing")
	GOBIN=$(PROJECT_DIR)/bin go install -mod=readonly github.com/mikefarah/yq/v4

$(PROJECT_DIR)/bin/protoc-gen-go:
	$(call header,"Installing")
	GOBIN=$(PROJECT_DIR)/bin go install -mod=readonly github.com/golang/protobuf/protoc-gen-go

$(PROJECT_DIR)/bin/go-bindata:
	$(call header,"Installing")
	GOBIN=$(PROJECT_DIR)/bin go install -mod=readonly github.com/go-bindata/go-bindata/v3/...

$(PROJECT_DIR)/bin/ginkgo:
	$(call header,"Installing")
	GOBIN=$(PROJECT_DIR)/bin go install -mod=readonly github.com/onsi/ginkgo/ginkgo

$(PROJECT_DIR)/bin/goimports:
	$(call header,"Installing")
	GOBIN=$(PROJECT_DIR)/bin go install -mod=readonly golang.org/x/tools/cmd/goimports

$(PROJECT_DIR)/bin/golangci-lint:
	$(call header,"Installing")
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PROJECT_DIR)/bin v1.28.3

$(PROJECT_DIR)/bin/controller-gen:
	$(call header,"Installing")
	$(call go-get-tool,$(PROJECT_DIR)/bin/controller-gen,sigs.k8s.io/controller-tools/cmd/controller-gen@$(shell go mod graph | grep controller-tools | head -n 1 | cut -d'@' -f 2))

KUSTOMIZE_VERSION?=v3.9.3
$(PROJECT_DIR)/bin/kustomize:
	$(call header,"Installing")
	wget -q -c https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F$(KUSTOMIZE_VERSION)/kustomize_$(KUSTOMIZE_VERSION)_$(GOOS)_$(GOARCH).tar.gz -O /tmp/kustomize.tar.gz
	tar xzvf /tmp/kustomize.tar.gz -C $(PROJECT_DIR)/bin/
	chmod +x $(PROJECT_DIR)/bin/kustomize

$(PROJECT_DIR)/bin/protoc:
	$(call header,"Installing")
	mkdir -p $(PROJECT_DIR)/bin/
	$(PROJECT_DIR)/scripts/dev/get-protobuf.sh
	chmod +x $(PROJECT_DIR)/bin/protoc

OPERATOR_SDK_VERSION=v1.3.0
$(PROJECT_DIR)/bin/operator-sdk:
	$(call header,"Installing operator-sdk cli")
	wget -q -c https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$(GOOS)_$(GOARCH) -O $(PROJECT_DIR)/bin/operator-sdk
	chmod +x $(PROJECT_DIR)/bin/operator-sdk

###########################################################################
##@ Image builds
###########################################################################

IKE_DOCKER_REGISTRY?=quay.io
IKE_DOCKER_REPOSITORY?=maistra
IKE_DOCKER_DEV_REPOSITORY?=maistra-dev
IKE_IMAGE_NAME?=$(PROJECT_NAME)
IKE_IMAGE_TAG?=$(IKE_VERSION)
IKE_TEST_IMAGE_NAME?=$(IKE_IMAGE_NAME)-test
IKE_TEST_PREPARED_IMAGE_NAME?=$(IKE_TEST_IMAGE_NAME)-prepared
IKE_TEST_PREPARED_NAME?=prepared-image

export IKE_DOCKER_REGISTRY
export IKE_DOCKER_REPOSITORY
export IKE_DOCKER_DEV_REPOSITORY
export IKE_IMAGE_NAME
export IKE_IMAGE_TAG
export IKE_VERSION

.PHONY: docker-build
docker-build: GOOS=linux
docker-build: compile ## Builds the docker image
	$(call header,"Building docker image $(IKE_IMAGE_NAME)")
	$(IMG_BUILDER) build \
		--label "org.opencontainers.image.title=$(IKE_IMAGE_NAME)" \
		--label "org.opencontainers.image.description=Tool enabling developers to safely develop and test on any kubernetes cluster without distracting others." \
		--label "org.opencontainers.image.source=https://$(PACKAGE_NAME)" \
		--label "org.opencontainers.image.documentation=https://istio-workspace-docs.netlify.com/istio-workspace" \
		--label "org.opencontainers.image.licenses=Apache-2.0" \
		--label "org.opencontainers.image.authors=Aslak Knutsen, Bartosz Majsak" \
		--label "org.opencontainers.image.vendor=Red Hat, Inc." \
		--label "org.opencontainers.image.revision=$(COMMIT)" \
		--label "org.opencontainers.image.created=$(shell date -u +%F\ %T%z)" \
		--network=host \
		-t $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_IMAGE_NAME):$(IKE_IMAGE_TAG) \
		-f $(BUILD_DIR)/Dockerfile $(BINARY_DIR)
	$(IMG_BUILDER) tag \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_IMAGE_NAME):$(IKE_IMAGE_TAG) \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_IMAGE_NAME):latest

.PHONY: docker-push
docker-push: docker-push--latest docker-push-versioned ## Pushes docker images to the registry (latest and versioned)

docker-push-versioned: docker-push--$(IKE_IMAGE_TAG)

docker-push--%:
	$(eval image_tag:=$(subst docker-push--,,$@))
	$(call header,"Pushing docker image $(image_tag)")
	$(IMG_BUILDER) push $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_IMAGE_NAME):$(image_tag)

.PHONY: docker-build-test
docker-build-test: $(BINARY_DIR)/$(TEST_BINARY_NAME)
	$(call header,"Building docker image $(IKE_TEST_IMAGE_NAME)")
	$(IMG_BUILDER) build \
		--no-cache \
		--label "org.opencontainers.image.title=$(IKE_TEST_IMAGE_NAME)" \
		--label "org.opencontainers.image.description=Test Services for end-to-end testing of the $(IKE_IMAGE_NAME)" \
		--label "org.opencontainers.image.source=https://$(PACKAGE_NAME)" \
		--label "org.opencontainers.image.documentation=https://istio-workspace-docs.netlify.com/istio-workspace" \
		--label "org.opencontainers.image.licenses=Apache-2.0" \
		--label "org.opencontainers.image.authors=Aslak Knutsen, Bartosz Majsak" \
		--label "org.opencontainers.image.vendor=Red Hat, Inc." \
		--label "org.opencontainers.image.revision=$(COMMIT)" \
		--label "org.opencontainers.image.created=$(shell date -u +%F\ %T%z)" \
		--network=host \
		--tag $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_DEV_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):$(IKE_IMAGE_TAG) \
		-f $(BUILD_DIR)/DockerfileTest $(BINARY_DIR)

	$(IMG_BUILDER) tag \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_DEV_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):$(IKE_IMAGE_TAG) \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_DEV_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):latest

.PHONY: docker-push-test
docker-push-test:
	$(call header,"Pushing docker image $(IKE_TEST_IMAGE_NAME)")
	$(IMG_BUILDER) push $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_DEV_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):$(IKE_IMAGE_TAG)
	$(IMG_BUILDER) push $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_DEV_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):latest

.PHONY: docker-build-test-prepared
docker-build-test-prepared:
	$(call header,"Building docker image $(IKE_TEST_PREPARED_IMAGE_NAME)")
	$(IMG_BUILDER) build \
		--no-cache \
		--build-arg=name=$(IKE_TEST_PREPARED_NAME) \
		--label "org.opencontainers.image.title=$(IKE_TEST_PREPARED_IMAGE_NAME)" \
		--label "org.opencontainers.image.description=Test Prepared Services for end-to-end testing of the $(IKE_IMAGE_NAME)" \
		--label "org.opencontainers.image.source=https://$(PACKAGE_NAME)" \
		--label "org.opencontainers.image.documentation=https://istio-workspace-docs.netlify.com/istio-workspace" \
		--label "org.opencontainers.image.licenses=Apache-2.0" \
		--label "org.opencontainers.image.authors=Aslak Knutsen, Bartosz Majsak" \
		--label "org.opencontainers.image.vendor=Red Hat, Inc." \
		--label "org.opencontainers.image.revision=$(COMMIT)" \
		--label "org.opencontainers.image.created=$(shell date -u +%F\ %T%z)" \
		--network=host \
		--tag $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_DEV_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):$(IKE_IMAGE_TAG) \
		-f $(BUILD_DIR)/DockerfileTestPrepared $(BINARY_DIR)

	$(IMG_BUILDER) tag \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_DEV_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):$(IKE_IMAGE_TAG) \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_DEV_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):latest

.PHONY: docker-push-test-prepared
docker-push-test-prepared:
	$(call header,"Pushing docker image $(IKE_TEST_PREPARED_IMAGE_NAME)")
	$(IMG_BUILDER) push $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_DEV_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):$(IKE_IMAGE_TAG)
	$(IMG_BUILDER) push $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_DEV_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):latest

###########################################################################
##@ Operator SDK bundle
###########################################################################

BUNDLE_IMG?=$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/istio-workspace-operator-bundle:$(IKE_IMAGE_TAG)
DESC_FILE:=dist/operatorhub_description.md
CSV_FILE:=bundle/manifests/istio-workspace-operator.clusterserviceversion.yaml

.PHONY: bundle
bundle: $(PROJECT_DIR)/bin/operator-sdk $(PROJECT_DIR)/bin/kustomize	## Generate bundle manifests and metadata, then validate generated files
	operator-sdk generate kustomize manifests -q
	cd config/manager && kustomize edit set image controller=$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_IMAGE_NAME):$(IKE_IMAGE_TAG)
	kustomize build config/manifests | operator-sdk generate bundle -q --overwrite --version $(OPERATOR_VERSION) $(BUNDLE_METADATA_OPTS)
	mv bundle.Dockerfile build
	sed -i 's/COPY bundle\//COPY /g' build/bundle.Dockerfile
	sed -i 's/containerImage: controller:latest/containerImage: $(IKE_DOCKER_REGISTRY)\/$(IKE_DOCKER_REPOSITORY)\/$(IKE_IMAGE_NAME):$(IKE_IMAGE_TAG)/' $(PROJECT_DIR)/$(CSV_FILE)
	sed -i 's/createdAt: "1970-01-01 00:00:0"/createdAt: $(shell date -u +%Y-%m-%dT%H:%M:%SZ)/' $(PROJECT_DIR)/$(CSV_FILE)
	cat $(PROJECT_DIR)/README.md | awk '/tag::description/{flag=1;next}/end::description/{flag=0}flag' > $(PROJECT_DIR)/$(DESC_FILE)
	sed -i 's/^/    /g' $(PROJECT_DIR)/$(DESC_FILE) # to make YAML happy we have to indent each line for the description field
	sed -i -e '/insert::description-from-readme/{r $(PROJECT_DIR)/$(DESC_FILE)' -e 'd}' $(PROJECT_DIR)/$(CSV_FILE)
	rm $(DESC_FILE)

	operator-sdk bundle validate ./bundle

.PHONY: bundle-build
bundle-build:	## Build the bundle image
	$(IMG_BUILDER) build -f build/bundle.Dockerfile -t $(BUNDLE_IMG) bundle/

.PHONY: bundle-push
bundle-push:	## Push the bundle image
	$(IMG_BUILDER) push $(BUNDLE_IMG)

.PHONY: bundle-run
bundle-run:		## Run the bundle image in OwnNamespace(OPERATOR_NAMESPACE) install mode
	$(k8s) create namespace $(OPERATOR_NAMESPACE) || true
	operator-sdk run bundle $(BUNDLE_IMG) -n $(OPERATOR_NAMESPACE) --install-mode OwnNamespace

.PHONY: bundle-run-single
bundle-run-single:		## Run the bundle image in SingleNamespace(OPERATOR_NAMESPACE) install mode targeting OPERATOR_WATCH_NAMESPACE
	$(k8s) create namespace $(OPERATOR_NAMESPACE) || true
	operator-sdk run bundle $(BUNDLE_IMG) -n $(OPERATOR_NAMESPACE) --install-mode SingleNamespace=${OPERATOR_WATCH_NAMESPACE}

.PHONY: bundle-run-multi
bundle-run-multi:		## Run the bundle image in MultiNamespace(OPERATOR_NAMESPACE) install mode targeting OPERATOR_WATCH_NAMESPACE
	$(k8s) create namespace $(OPERATOR_NAMESPACE) || true
	operator-sdk run bundle $(BUNDLE_IMG) -n $(OPERATOR_NAMESPACE) --install-mode MultiNamespace=${OPERATOR_WATCH_NAMESPACE}

.PHONY: bundle-run-all
bundle-run-all:		## Run the bundle image in AllNamespace(OPERATOR_NAMESPACE) install mode
	$(k8s) create namespace $(OPERATOR_NAMESPACE) || true
	operator-sdk run bundle $(BUNDLE_IMG) -n $(OPERATOR_NAMESPACE) --install-mode AllNamespaces


.PHONY: bundle-clean
bundle-clean:	## Clean the bundle image
	operator-sdk cleanup istio-workspace-operator -n $(OPERATOR_NAMESPACE)

.PHONY: bundle-publish
bundle-publish:	## Open up a PR to the Operator Hub community catalog
	./scripts/release/operatorhub.sh

# ##########################################################################
##@ Tekton tasks
# ##########################################################################

.PHONY: tekton-publish
tekton-publish: ## Prepares Tekton tasks for release and opens a PR on the Tekton Hub
	./scripts/release/tektoncatalog.sh

# ##########################################################################
## Istio-workspace sample project deployment
# ##########################################################################

deploy-test-%:
	$(eval scenario:=$(subst deploy-test-,,$@))
	$(call header,"Deploying test $(scenario) app to $(TEST_NAMESPACE)")

	$(k8s) create namespace $(TEST_NAMESPACE) || true
	oc adm policy add-scc-to-user anyuid -z default -n $(TEST_NAMESPACE) || true
	oc adm policy add-scc-to-user privileged -z default -n $(TEST_NAMESPACE) || true
	go run ./test/cmd/test-scenario/ $(scenario) | $(k8s) apply -n $(TEST_NAMESPACE) -f -

undeploy-test-%:
	$(eval scenario:=$(subst undeploy-test-,,$@))
	$(call header,"Undeploying test $(scenario) app from $(TEST_NAMESPACE)")

	go run ./test/cmd/test-scenario/ $(scenario) | $(k8s) delete -n $(TEST_NAMESPACE) -f -

# ##########################################################################
##@ Helpers
# ##########################################################################

VERSION?=x
export VERSION

.PHONY: release-notes-draft
release-notes-draft: ## Prepares release notes based on template. e.g. VERSION=v1.0.0 make release-notes-draft
	@if [ "$(VERSION)" = "x" ]; then\
		echo "missing version: VERSION=v1.0.0 make prepare-release" && exit -1;\
	else\
		./scripts/release/validate.sh $(VERSION) --skip-release-notes-check && \
		git checkout -b release_$(VERSION) && \
		cp docs/modules/ROOT/pages/release_notes/release_notes_template.adoc docs/modules/ROOT/pages/release_notes/$(VERSION).adoc && \
		sed -i -e "s/vX.Y.Z/${VERSION}/" docs/modules/ROOT/pages/release_notes/$(VERSION).adoc;\
	fi

.PHONY: help
help:  ## Displays this help \o/
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m\033[2m %s\033[0m\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@cat $(MAKEFILE_LIST) | grep "^[A-Za-z_]*.?=" | sort | awk 'BEGIN {FS="?="; printf "\n\n\033[1mEnvironment variables\033[0m\n"} {printf "  \033[36m%-25s\033[0m\033[2m %s\033[0m\n", $$1, $$2}'
