PROJECT_NAME:=istio-workspace
PACKAGE_NAME:=github.com/maistra/istio-workspace

OPERATOR_NAMESPACE?=istio-workspace-operator
OPERATOR_WATCH_NAMESPACE=""
TEST_NAMESPACE?=bookinfo

PROJECT_DIR:=$(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
export PROJECT_DIR
export GO111MODULE:=on
BUILD_DIR:=$(PROJECT_DIR)/build
BINARY_DIR:=$(PROJECT_DIR)/dist
BINARY_NAME:=ike
TEST_BINARY_NAME:=test-service
TPL_BINARY_NAME:=tpl
ASSETS:=pkg/assets/isto-workspace-deploy.go
ASSET_SRCS=$(shell find ./deploy ./template -name "*.yaml" -o -name "*.tpl" -o -name "*.var" | sort)
MANIFEST_DIR:=$(PROJECT_DIR)/deploy

TELEPRESENCE_VERSION?=$(shell telepresence --version)

GOPATH_1:=$(shell echo ${GOPATH} | cut -d':' -f 1)
GOBIN=$(GOPATH_1)/bin
PATH:=${GOBIN}/bin:$(PROJECT_DIR)/bin:$(PATH)

# Determine this makefile's path.
# Be sure to place this BEFORE `include` directives, if any.
THIS_MAKEFILE:=$(lastword $(MAKEFILE_LIST))

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
	IKE_VERSION:=$(IKE_VERSION)-next-$(COMMIT)
	OPERATOR_VERSION:=$(OPERATOR_VERSION)-next-$(COMMIT)
else ifneq ($(GITUNTRACKEDCHANGES),)
	IKE_VERSION:=$(IKE_VERSION)-dirty
endif

GOBUILD:=GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0
RELEASE?=false
LDFLAGS="-w -X ${PACKAGE_NAME}/version.Release=${RELEASE} -X ${PACKAGE_NAME}/version.Version=${IKE_VERSION} -X ${PACKAGE_NAME}/version.Commit=${COMMIT} -X ${PACKAGE_NAME}/version.BuildTime=${BUILD_TIME}"
SRC_DIRS:=./api ./controller ./pkg ./cmd ./version ./test
TEST_DIRS:=./e2e
SRCS:=$(shell find ${SRC_DIRS} -name "*.go")

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
	rm -rf $(BINARY_DIR) $(PROJECT_DIR)/bin/ vendor/

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
	controller-gen crd paths=./api/... output:crd:dir=./deploy/crds
	controller-gen object paths=./api/...
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

###########################################################################
##@ Setup
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

OPERATOR_SDK_VERSION=v1.3.0
$(PROJECT_DIR)/bin/operator-sdk:
	$(call header,"Installing operator-sdk cli")
	mkdir -p $(PROJECT_DIR)/bin/
	wget -q -c https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$(GOOS)_$(GOARCH) -O $(PROJECT_DIR)/bin/operator-sdk
	chmod +x $(PROJECT_DIR)/bin/operator-sdk

.PHONY: tools
install-tools:  $(PROJECT_DIR)/bin/operator-sdk ## Installs required go tools
	$(call header,"Installing required tools")
	go install -mod=readonly golang.org/x/tools/cmd/goimports
	go install -mod=readonly github.com/golang/protobuf/protoc-gen-go
	go install -mod=readonly github.com/onsi/ginkgo/ginkgo
	go install -mod=readonly github.com/mikefarah/yq/v3
	go install -mod=readonly github.com/go-bindata/go-bindata/v3/...
	# go get causes problems and is not recommended by the creators. installing binary instead
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH_1)/bin v1.28.3
	$(call go-get-tool,$(PROJECT_DIR)/bin/controller-gen,sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0)
	$(call go-get-tool,$(PROJECT_DIR)/bin/kustomize,sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)
	operator-sdk version

EXECUTABLES:=operator-sdk controller-gen kustomize golangci-lint goimports ginkgo go-bindata protoc-gen-go yq
CHECK:=$(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec) 2>/dev/null),,"install"))
.PHONY: tools
tools:
	$(call header,"Checking required tools")
	@$(if $(strip $(CHECK)),$(MAKE) -f $(THIS_MAKEFILE) install-tools,echo "'$(EXECUTABLES)' are installed")


$(PROJECT_DIR)/bin/protoc:
	$(call header,"Installing protoc")
	mkdir -p $(PROJECT_DIR)/bin/
	$(PROJECT_DIR)/scripts/dev/get-protobuf.sh
	chmod +x $(PROJECT_DIR)/bin/protoc

$(PROJECT_DIR)/$(ASSETS): $(ASSET_SRCS)
	$(call header,"Adds assets to the binary")
	go-bindata -o $(ASSETS) -nometadata -pkg assets -ignore 'examples/' $(ASSET_SRCS)

.PHONY: operator-tpl
operator-tpl: $(BINARY_DIR)/$(TPL_BINARY_NAME)
	$(call header,"Updates operator.yaml")
	@printf "## Autogenerated from operator.tpl.yaml at $(shell date)\n" > $(MANIFEST_DIR)/operator.yaml
	@printf "## DO NOT MODIFY THIS FILE. Please change operator.tpl.yaml instead.\n\n" >> $(MANIFEST_DIR)/operator.yaml
	@IKE_VERSION=$(IKE_VERSION) \
   	IKE_DOCKER_REGISTRY=$(IKE_DOCKER_REGISTRY) \
   	IKE_DOCKER_REPOSITORY=$(IKE_DOCKER_REPOSITORY) \
   	IKE_IMAGE_NAME=$(IKE_IMAGE_NAME) \
   	IKE_IMAGE_TAG=$(IKE_IMAGE_TAG) \
   	WATCH_NAMESPACE=$(OPERATOR_WATCH_NAMESPACE) \
	$(BINARY_DIR)/$(TPL_BINARY_NAME) \
	$(MANIFEST_DIR)/operator.tpl.yaml >> $(MANIFEST_DIR)/operator.yaml

###########################################################################
##@ Docker
###########################################################################

IKE_IMAGE_NAME?=$(PROJECT_NAME)
IKE_TEST_IMAGE_NAME?=$(IKE_IMAGE_NAME)-test
IKE_TEST_PREPARED_IMAGE_NAME?=$(IKE_TEST_IMAGE_NAME)-prepared
IKE_TEST_PREPARED_NAME?=prepared-image
IKE_IMAGE_TAG?=$(IKE_VERSION)
IKE_DOCKER_REGISTRY?=quay.io
IKE_DOCKER_REPOSITORY?=maistra
export IKE_IMAGE_TAG
export IKE_VERSION

IMG_BUILDER:=docker

## Prefer to use podman
ifneq (, $(shell which podman))
	IMG_BUILDER=podman
endif

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
		-f $(BUILD_DIR)/Dockerfile $(PROJECT_DIR)
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
		--tag $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):$(IKE_IMAGE_TAG) \
		-f $(BUILD_DIR)/DockerfileTest $(PROJECT_DIR)

	$(IMG_BUILDER) tag \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):$(IKE_IMAGE_TAG) \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):latest

.PHONY: docker-push-test
docker-push-test:
	$(call header,"Pushing docker image $(IKE_TEST_IMAGE_NAME)")
	$(IMG_BUILDER) push $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):$(IKE_IMAGE_TAG)
	$(IMG_BUILDER) push $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):latest

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
		--tag $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):$(IKE_IMAGE_TAG) \
		-f $(BUILD_DIR)/DockerfileTestPrepared $(PROJECT_DIR)

	$(IMG_BUILDER) tag \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):$(IKE_IMAGE_TAG) \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):latest

.PHONY: docker-push-test-prepared
docker-push-test-prepared:
	$(call header,"Pushing docker image $(IKE_TEST_PREPARED_IMAGE_NAME)")
	$(IMG_BUILDER) push $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):$(IKE_IMAGE_TAG)
	$(IMG_BUILDER) push $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):latest

# ##########################################################################
##@ Istio-workspace sample project deployment
# ##########################################################################

k8s:=kubectl

ifneq (, $(shell which oc))
	k8s=oc
endif

deploy-test-%:
	$(eval scenario:=$(subst deploy-test-,,$@))
	$(call header,"Deploying test $(scenario) app to $(TEST_NAMESPACE)")

	$(k8s) create namespace $(TEST_NAMESPACE) || true
	oc adm policy add-scc-to-user anyuid -z default -n $(TEST_NAMESPACE) || true
	oc adm policy add-scc-to-user privileged -z default -n $(TEST_NAMESPACE) || true
	$(k8s) apply -n $(TEST_NAMESPACE) -f deploy/examples/session_role.yaml
	$(k8s) apply -n $(TEST_NAMESPACE) -f deploy/examples/session_rolebinding.yaml

	go run ./test/cmd/test-scenario/ $(scenario) | $(k8s) apply -n $(TEST_NAMESPACE) -f -

undeploy-test-%:
	$(eval scenario:=$(subst undeploy-test-,,$@))
	$(call header,"Undeploying test $(scenario) app from $(TEST_NAMESPACE)")

	go run ./test/cmd/test-scenario/ $(scenario) | $(k8s) delete -n $(TEST_NAMESPACE) -f -
	$(k8s) delete -n $(TEST_NAMESPACE) -f deploy/examples/session_rolebinding.yaml
	$(k8s) delete -n $(TEST_NAMESPACE) -f deploy/examples/session_role.yaml

##@ Helpers

.PHONY: help
help:  ## Displays this help \o/
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m\033[2m %s\033[0m\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@cat $(MAKEFILE_LIST) | grep "^[A-Za-z_]*.?=" | sort | awk 'BEGIN {FS="?="; printf "\n\n\033[1mEnvironment variables\033[0m\n"} {printf "  \033[36m%-25s\033[0m\033[2m %s\033[0m\n", $$1, $$2}'
