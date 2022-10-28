# Determines this makefile's path.
# Be sure to place this BEFORE `include` directives, if any.
THIS_MAKEFILE:=$(lastword $(MAKEFILE_LIST))

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
DIST_DIR:=$(PROJECT_DIR)/dist
BINARY_NAME:=ike
TEST_BINARY_NAME:=test-service
TPL_BINARY_NAME:=tpl
ASSETS:=pkg/assets/isto-workspace-deploy.go
ASSET_SRCS=$(shell find ./template -name "*.yaml" -o -name "*.tpl" -o -name "*.var" | sort)
MANIFEST_DIR:=$(PROJECT_DIR)/deploy

GOPATH_1:=$(shell echo ${GOPATH} | cut -d':' -f 1)
GOBIN=$(GOPATH_1)/bin
PATH:=${GOBIN}/bin:$(PROJECT_DIR)/bin:$(PATH)

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
GIT_TAG?=$(shell git describe --tags --abbrev=0 --exact-match > /dev/null 2>&1; echo $$?)
ifneq ($(GIT_TAG),0)
	ifeq ($(origin IKE_VERSION),file)
		IKE_VERSION:=$(IKE_VERSION)-next
		OPERATOR_VERSION:=$(OPERATOR_VERSION)-next
	endif
endif
export OPERATOR_VERSION

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

# Prefer to use podman if not explicitly set
ifneq (, $(shell which podman))
	IMG_BUILDER?=podman
else
	IMG_BUILDER?=docker
endif

###########################################################################
##@ Build
###########################################################################

.PHONY: build-ci
build-ci: deps tools format compile test # Like 'all', but without linter which is executed as separated PR check

.PHONY: compile
compile: deps generate format $(DIST_DIR)/$(BINARY_NAME) ## Compiles binaries

.PHONY: test
test: generate ## Runs tests
	$(call header,"Running tests")
	ginkgo -r -v -progress -vet=off -trace --skip-package=e2e --junit-report=ginkgo-test-results.xml ${args}

.PHONY: test-e2e
test-e2e: compile ## Runs end-to-end tests
	$(call header,"Running end-to-end tests")
	ginkgo e2e/ -r -v -progress -vet=off -trace --junit-report=ginkgo-test-results.xml ${args}

.PHONY: clean
clean: ## Removes build artifacts
	rm -rf $(DIST_DIR) $(PROJECT_DIR)/bin/ vendor/ bundle/ bundle-*/

.PHONY: deps
deps: ## Fetches all dependencies
	$(call header,"Fetching dependencies")
	go mod tidy
	go mod download
	go mod vendor

.PHONY: format
format: $(SRCS) ## Removes unneeded imports and formats source code
	$(call header,"Formatting code")
	goimports -l -w -e $(SRC_DIRS) $(TEST_DIRS)

.PHONY: lint-prepare
lint-prepare: deps tools generate

.PHONY: release-prepare
release-prepare: deps tools generate format

.PHONY: lint
lint: lint-prepare ## Concurrently runs a whole bunch of static analysis tools
	$(call header,"Running a whole bunch of static analysis tools")
	golangci-lint run --fix --sort-results

CRD_OPTIONS ?= crd
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

$(DIST_DIR):
	[ -d $@ ] || mkdir -p $@

$(DIST_DIR)/$(BINARY_NAME): $(DIST_DIR) $(SRCS)
	$(call header,"Compiling... carry on!")
	${GOBUILD} go build -v -ldflags ${LDFLAGS} -o $@ ./cmd/$(BINARY_NAME)/

$(DIST_DIR)/$(TEST_BINARY_NAME): $(DIST_DIR) $(SRCS) test/cmd/test-service/html.go test/cmd/test-service/main.pb.go
	$(call header,"Compiling test service... carry on!")
	${GOBUILD} go build -ldflags ${LDFLAGS} -o $@ ./test/cmd/$(TEST_BINARY_NAME)/

test/cmd/test-service/main.pb.go: $(PROJECT_DIR)/bin/protoc test/cmd/test-service/main.proto
	$(call header,"Compiling test proto... carry on!")
	$(PROJECT_DIR)/bin/protoc -I test/cmd/test-service/ test/cmd/test-service/main.proto --go_out=plugins=grpc:test/cmd/test-service --go_opt=paths=source_relative

test/cmd/test-service/html.go: test/cmd/test-service/assets/index.html
	$(call header,"Compiling test assets... carry on!")
	go-bindata -o test/cmd/test-service/html.go -pkg main -prefix test/cmd/test-service/assets test/cmd/test-service/assets/*

.PHONY: compile-test-service
compile-test-service: test/cmd/test-service/html.go test/cmd/test-service/main.pb.go $(DIST_DIR)/$(TEST_BINARY_NAME)

$(DIST_DIR)/$(TPL_BINARY_NAME): $(DIST_DIR) $(SRCS)
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
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

.PHONY: tools
tools: $(PROJECT_DIR)/bin/operator-sdk $(PROJECT_DIR)/bin/controller-gen $(PROJECT_DIR)/bin/kustomize ## Installs required tools
tools: $(PROJECT_DIR)/bin/golangci-lint $(PROJECT_DIR)/bin/goimports $(PROJECT_DIR)/bin/go-bindata
tools: $(PROJECT_DIR)/bin/protoc-gen-go $(PROJECT_DIR)/bin/yq $(PROJECT_DIR)/bin/ginkgo

$(PROJECT_DIR)/bin/yq:
	$(call header,"Installing yq")
	GOBIN=$(PROJECT_DIR)/bin go install -mod=readonly github.com/mikefarah/yq/v4

$(PROJECT_DIR)/bin/protoc-gen-go:
	$(call header,"Installing protoc-gen-go")
	GOBIN=$(PROJECT_DIR)/bin go install -mod=readonly github.com/golang/protobuf/protoc-gen-go

$(PROJECT_DIR)/bin/go-bindata:
	$(call header,"Installing go-bindata")
	GOBIN=$(PROJECT_DIR)/bin go install -mod=readonly github.com/go-bindata/go-bindata/v3/...

$(PROJECT_DIR)/bin/ginkgo:
	$(call header,"Installing ginkgo")
	GOBIN=$(PROJECT_DIR)/bin go install -mod=readonly github.com/onsi/ginkgo/v2/ginkgo

$(PROJECT_DIR)/bin/goimports:
	$(call header,"Installing goimports")
	GOBIN=$(PROJECT_DIR)/bin go install -mod=readonly golang.org/x/tools/cmd/goimports

GOLANGCI_LINT_VERSION?=v1.50.1
$(PROJECT_DIR)/bin/golangci-lint:
	$(call header,"Installing golangci-lint")
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PROJECT_DIR)/bin $(GOLANGCI_LINT_VERSION)

$(PROJECT_DIR)/bin/controller-gen:
	$(call header,"Installing controller-gen")
	$(call go-get-tool,$(PROJECT_DIR)/bin/controller-gen,sigs.k8s.io/controller-tools/cmd/controller-gen@$(shell go mod graph | grep controller-tools | head -n 1 | cut -d'@' -f 2))

KUSTOMIZE_VERSION?=v4.2.0
$(PROJECT_DIR)/bin/kustomize:
	$(call header,"Installing kustomize")
	wget -q -c https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F$(KUSTOMIZE_VERSION)/kustomize_$(KUSTOMIZE_VERSION)_$(GOOS)_$(GOARCH).tar.gz -O /tmp/kustomize.tar.gz
	tar xzvf /tmp/kustomize.tar.gz -C $(PROJECT_DIR)/bin/
	chmod +x $(PROJECT_DIR)/bin/kustomize

$(PROJECT_DIR)/bin/protoc:
	$(call header,"Installing")
	mkdir -p $(PROJECT_DIR)/bin/
	$(PROJECT_DIR)/scripts/dev/get-protobuf.sh
	chmod +x $(PROJECT_DIR)/bin/protoc

OPERATOR_SDK_VERSION=v1.22.1
$(PROJECT_DIR)/bin/operator-sdk:
	$(call header,"Installing operator-sdk cli")
	wget -q -c https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$(GOOS)_$(GOARCH) -O $(PROJECT_DIR)/bin/operator-sdk
	chmod +x $(PROJECT_DIR)/bin/operator-sdk

###########################################################################
##@ Image builds
###########################################################################

IKE_CONTAINER_REGISTRY?=quay.io
IKE_CONTAINER_REPOSITORY?=maistra
IKE_CONTAINER_DEV_REPOSITORY?=maistra-dev
IKE_IMAGE_NAME?=$(PROJECT_NAME)
IKE_IMAGE_TAG?=$(IKE_VERSION)
IKE_TEST_IMAGE_NAME?=$(IKE_IMAGE_NAME)-test
IKE_TEST_PREPARED_IMAGE_NAME?=$(IKE_TEST_IMAGE_NAME)-prepared
IKE_TEST_PREPARED_NAME?=prepared-image
IKE_IMAGE=${IKE_CONTAINER_REGISTRY}\/${IKE_CONTAINER_REPOSITORY}\/${IKE_IMAGE_NAME}:${IKE_IMAGE_TAG}

export IKE_CONTAINER_REGISTRY
export IKE_CONTAINER_REPOSITORY
export IKE_CONTAINER_DEV_REPOSITORY
export IKE_IMAGE_NAME
export IKE_IMAGE_TAG
export IKE_VERSION
export IKE_IMAGE

.PHONY: container-image
container-image: GOOS=linux
container-image: compile ## Builds the container image
	$(call header,"Building container image $(IKE_IMAGE_NAME)")
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
		-t $(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_REPOSITORY)/$(IKE_IMAGE_NAME):$(IKE_IMAGE_TAG) \
		-f $(BUILD_DIR)/Containerfile $(DIST_DIR)
	$(IMG_BUILDER) tag \
		$(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_REPOSITORY)/$(IKE_IMAGE_NAME):$(IKE_IMAGE_TAG) \
		$(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_REPOSITORY)/$(IKE_IMAGE_NAME):latest

.PHONY: container-push
container-push: container-push--latest container-push-versioned ## Pushes container images to the registry (latest and versioned)

container-push-versioned: container-push--$(IKE_IMAGE_TAG)

container-push--%:
	$(eval image_tag:=$(subst container-push--,,$@))
	$(call header,"Pushing container image $(image_tag)")
	$(IMG_BUILDER) push $(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_REPOSITORY)/$(IKE_IMAGE_NAME):$(image_tag)

.PHONY: container-image-test
container-image-test: $(DIST_DIR)/$(TEST_BINARY_NAME)
	$(call header,"Building container image $(IKE_TEST_IMAGE_NAME)")
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
		--tag $(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_DEV_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):$(IKE_IMAGE_TAG) \
		-f $(BUILD_DIR)/ContainerfileTest $(DIST_DIR)

	$(IMG_BUILDER) tag \
		$(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_DEV_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):$(IKE_IMAGE_TAG) \
		$(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_DEV_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):latest

.PHONY: container-push-test
container-push-test:
	$(call header,"Pushing container image $(IKE_TEST_IMAGE_NAME)")
	$(IMG_BUILDER) push $(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_DEV_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):$(IKE_IMAGE_TAG)
	$(IMG_BUILDER) push $(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_DEV_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):latest

.PHONY: container-image-test-prepared
container-image-test-prepared:
	$(call header,"Building container image $(IKE_TEST_PREPARED_IMAGE_NAME)")
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
		--tag $(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_DEV_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):$(IKE_IMAGE_TAG) \
		-f $(BUILD_DIR)/ContainerfileTestPrepared $(DIST_DIR)

	$(IMG_BUILDER) tag \
		$(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_DEV_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):$(IKE_IMAGE_TAG) \
		$(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_DEV_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):latest

.PHONY: container-push-test-prepared
container-push-test-prepared:
	$(call header,"Pushing container image $(IKE_TEST_PREPARED_IMAGE_NAME)")
	$(IMG_BUILDER) push $(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_DEV_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):$(IKE_IMAGE_TAG)
	$(IMG_BUILDER) push $(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_DEV_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):latest

###########################################################################
##@ Operator SDK bundle
###########################################################################

CHANNELS?="alpha"
DEFAULT_CHANNEL?="alpha"
BUNDLE_CHANNELS:=--channels=$(CHANNELS)
BUNDLE_DEFAULT_CHANNEL:=--default-channel=$(DEFAULT_CHANNEL)
BUNDLE_METADATA_OPTS?=$(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

BUNDLE_IMG?=$(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_REPOSITORY)/istio-workspace-operator-bundle:$(IKE_IMAGE_TAG)
DESC_FILE:=dist/operatorhub_description.md
CSV_FILE:=bundle/manifests/istio-workspace-operator.clusterserviceversion.yaml

.PHONY: bundle
bundle: $(PROJECT_DIR)/bin/operator-sdk $(PROJECT_DIR)/bin/kustomize $(DIST_DIR) ## Generate bundle manifests and metadata, then validate generated files
	$(call header,"Generate bundle manifests and metadata")
	operator-sdk generate kustomize manifests -q
	cd config/manager && kustomize edit set image controller=$(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_REPOSITORY)/$(IKE_IMAGE_NAME):$(IKE_IMAGE_TAG)
	kustomize build config/manifests | operator-sdk generate bundle -q --overwrite --version $(OPERATOR_VERSION) $(BUNDLE_METADATA_OPTS)
	mv -f bundle.Dockerfile build/bundle.Containerfile
	sed -i 's/COPY bundle\//COPY /g' build/bundle.Containerfile
	sed -i 's/containerImage: controller:latest/containerImage: $(IKE_CONTAINER_REGISTRY)\/$(IKE_CONTAINER_REPOSITORY)\/$(IKE_IMAGE_NAME):$(IKE_IMAGE_TAG)/' $(PROJECT_DIR)/$(CSV_FILE)
	sed -i 's/createdAt: "1970-01-01 00:00:0"/createdAt: $(shell date -u +%Y-%m-%dT%H:%M:%SZ)/' $(PROJECT_DIR)/$(CSV_FILE)
	cat $(PROJECT_DIR)/README.md | awk '/tag::description/{flag=1;next}/end::description/{flag=0}flag' > $(PROJECT_DIR)/$(DESC_FILE)
	sed -i 's/^/    /g' $(PROJECT_DIR)/$(DESC_FILE) # to make YAML happy we have to indent each line for the description field
	sed -i -e '/insert::description-from-readme/{r $(PROJECT_DIR)/$(DESC_FILE)' -e 'd}' $(PROJECT_DIR)/$(CSV_FILE)
	rm $(DESC_FILE)
	$(call header,"Validate bundle")
	operator-sdk bundle validate ./bundle

.PHONY: bundle-image
bundle-image:	## Build the bundle image
	$(call header,"Building bundle image")
	$(IMG_BUILDER) build -f build/bundle.Containerfile -t $(BUNDLE_IMG) bundle/

.PHONY: bundle-push
bundle-push:	## Push the bundle image
	$(call header,"Pushing bundle image")
	$(IMG_BUILDER) push $(BUNDLE_IMG)

BUNDLE_TIMEOUT?=5m
bundle-run:=operator-sdk run bundle $(BUNDLE_IMG) -n $(OPERATOR_NAMESPACE) --timeout $(BUNDLE_TIMEOUT) --index-image quay.io/operator-framework/opm:$(OPERATOR_SDK_VERSION)

.PHONY: bundle-run
bundle-run:		## Run the bundle image in OwnNamespace(OPERATOR_NAMESPACE) install mode
	$(k8s) create namespace $(OPERATOR_NAMESPACE) || true
	$(bundle-run) --install-mode OwnNamespace

.PHONY: bundle-run-single
bundle-run-single:		## Run the bundle image in SingleNamespace(OPERATOR_NAMESPACE) install mode targeting OPERATOR_WATCH_NAMESPACE
	$(k8s) create namespace $(OPERATOR_NAMESPACE) || true
	$(bundle-run) --install-mode MultiNamespace=${OPERATOR_WATCH_NAMESPACE}

.PHONY: bundle-run-multi
bundle-run-multi:		## Run the bundle image in MultiNamespace(OPERATOR_NAMESPACE) install mode targeting OPERATOR_WATCH_NAMESPACE
	$(k8s) create namespace $(OPERATOR_NAMESPACE) || true
	$(bundle-run) --install-mode MultiNamespace=${OPERATOR_WATCH_NAMESPACE}

.PHONY: bundle-run-all
bundle-run-all:		## Run the bundle image in AllNamespace(OPERATOR_NAMESPACE) install mode
	$(k8s) create namespace $(OPERATOR_NAMESPACE) || true
	$(bundle-run) --install-mode AllNamespaces

.PHONY: bundle-clean
bundle-clean:	## Clean the bundle image
	operator-sdk cleanup istio-workspace-operator -n $(OPERATOR_NAMESPACE)

.PHONY: bundle-test
bundle-test: bundle	## Run the Operator Hub test suite
	./scripts/release/operatorhub/test.sh

.PHONY: bundle-publish
bundle-publish:	## Open up a PR to the Operator Hub community catalog
	./scripts/release/operatorhub/publish.sh

# ##########################################################################
##@ Tekton tasks
# ##########################################################################
TEKTON_TASK_VERSION:=$(shell echo "${OPERATOR_VERSION}" | cut -d'.' -f 1,2)

.PHONY: tekton-deploy
tekton-deploy: ## Deploy the Tekton tasks
	./scripts/release/tektonhub/prepare_task.sh replace_placeholders "$(PROJECT_DIR)/integration/tekton/tasks/ike-create/ike-create.yaml" "${TEKTON_TASK_VERSION}" "${IKE_IMAGE}" | $(k8s) apply -f - -n $(TEST_NAMESPACE)
	./scripts/release/tektonhub/prepare_task.sh replace_placeholders "$(PROJECT_DIR)/integration/tekton/tasks/ike-session-url/ike-session-url.yaml" "${TEKTON_TASK_VERSION}" "${IKE_IMAGE}" | $(k8s) apply -f - -n $(TEST_NAMESPACE)
	./scripts/release/tektonhub/prepare_task.sh replace_placeholders "$(PROJECT_DIR)/integration/tekton/tasks/ike-delete/ike-delete.yaml" "${TEKTON_TASK_VERSION}" "${IKE_IMAGE}" | $(k8s) apply -f - -n $(TEST_NAMESPACE)

.PHONY: tekton-undeploy
tekton-undeploy: ## UnDeploy the Tekton tasks
	$(k8s) delete -n $(TEST_NAMESPACE) -f "$(PROJECT_DIR)/integration/tekton/tasks/ike-create/ike-create.yaml" || true
	$(k8s) delete -n $(TEST_NAMESPACE) -f "$(PROJECT_DIR)/integration/tekton/tasks/ike-session-url/ike-session-url.yaml" || true
	$(k8s) delete -n $(TEST_NAMESPACE) -f "$(PROJECT_DIR)/integration/tekton/tasks/ike-delete/ike-delete.yaml" || true

TEST_SESSION_NAME?=test-session
IKE_TEST_PREPARED_IMG:=$(IKE_CONTAINER_REGISTRY)/$(IKE_CONTAINER_DEV_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):$(IKE_IMAGE_TAG)

tekton-test-%: $(PROJECT_DIR)/bin/yq ## Run a Tekton tasks for test purpose
	$(eval task:=$(subst tekton-test-,,$@))
	@yq e '.spec.params[] | select(.name=="session") | .value="${TEST_SESSION_NAME}", .spec.params[] | select(.name=="route") | .value="header:x-test-suite=smoke", .spec.params[] | select(.name=="image") | .value="${IKE_TEST_PREPARED_IMG}", . ' $(PROJECT_DIR)/integration/tekton/tasks/$(task)/samples/$(task).yaml \
		| awk '/apiVersion:/,0  {print $1}' | $(k8s) apply -f - -n ${TEST_NAMESPACE}

.PHONY: tekton-publish
tekton-publish: ## Prepares Tekton tasks for release and opens a PR on the Tekton Hub
	./scripts/release/tektonhub/publish.sh


# ##########################################################################
## Istio-workspace sample project deployment
# ##########################################################################

deploy-test-%:
	$(eval scenario:=$(subst deploy-test-,,$@))
	$(call header,"Deploying test $(scenario) app to $(TEST_NAMESPACE)")

	$(k8s) create namespace $(TEST_NAMESPACE) || true

	# Do not remove line breaks as they're intentionally set for docs toolchain to always get right snippet
	# tag::anyuid[]
	oc adm policy add-scc-to-user anyuid -z default -n $(TEST_NAMESPACE) || true
	# end::anyuid[]

	# Do not remove line breaks as they're intentionally set for docs toolchain to always get right snippet
	# tag::privileged[]
	oc adm policy add-scc-to-user privileged -z default -n $(TEST_NAMESPACE) || true
	# end::privileged[]

	go run ./test/cmd/test-scenario/ $(scenario) | $(k8s) apply -f -

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
	@./scripts/release/create_release_notes.sh --version=$(VERSION)

.PHONY: help
help:  ## Displays this help \o/
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m\033[2m %s\033[0m\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@cat $(MAKEFILE_LIST) | grep "^[A-Za-z_]*.?=" | sort | awk 'BEGIN {FS="?="; printf "\n\n\033[1mEnvironment variables\033[0m\n"} {printf "  \033[36m%-25s\033[0m\033[2m %s\033[0m\n", $$1, $$2}'
