PROJECT_NAME:=istio-workspace
PACKAGE_NAME:=github.com/maistra/istio-workspace

OPERATOR_NAMESPACE?=istio-workspace-operator
OPERATOR_WATCH_NAMESPACE=""
TEST_NAMESPACE?=bookinfo

PROJECT_DIR:=$(shell pwd)
export PROJECT_DIR
BUILD_DIR:=$(PROJECT_DIR)/build
BINARY_DIR:=$(PROJECT_DIR)/dist
BINARY_NAME:=ike
TEST_BINARY_NAME:=test-service
ASSETS:=pkg/assets/isto-workspace-deploy.go
ASSET_SRCS=$(shell find ./deploy/istio-workspace -name "*.yaml")

TELEPRESENCE_VERSION?=$(shell telepresence --version)

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
all: deps format lint compile test ## Runs 'deps format lint test compile' targets

##@ Build

.PHONY: build-ci
build-ci: deps format compile test # Like 'all', but without linter which is executed as separated PR check

.PHONY: compile
compile: operator-codegen $(BINARY_DIR)/$(BINARY_NAME) ## Compiles binaries

.PHONY: test
test: operator-codegen ## Runs tests
	$(call header,"Running tests")
	ginkgo -r -v --skipPackage=e2e ${args}

.PHONY: test-e2e
test-e2e: compile ## Runs end-to-end tests
	$(call header,"Running end-to-end tests")
	ginkgo e2e/ -v -p ${args}

.PHONY: clean
clean: ## Removes build artifacts
	rm -rf $(BINARY_DIR) $(PROJECT_DIR)/bin/ vendor/ .vendor-new/ Gopkg.lock

.PHONY: deps
deps: check-tools ## Fetches all dependencies
	$(call header,"Fetching dependencies")
	dep ensure -v

.PHONY: format
format: ## Removes unneeded imports and formats source code
	$(call header,"Formatting code")
	goimports -l -w -e ./pkg/ ./cmd/ ./version/ ./test/ ./e2e/

.PHONY: lint-prepare
lint-prepare: deps operator-codegen

.PHONY: lint
lint: lint-prepare ## Concurrently runs a whole bunch of static analysis tools
	$(call header,"Running a whole bunch of static analysis tools")
	golangci-lint run

GOPATH_1:=$(shell echo ${GOPATH} | cut -d':' -f 1)
.PHONY: operator-codegen
operator-codegen: $(PROJECT_DIR)/bin/operator-sdk $(PROJECT_DIR)/$(ASSETS) ## Generates operator-sdk code and bundles packages using go-bindata
	$(call header,"Generates operator-sdk code")
	GOPATH=$(GOPATH_1) $(PROJECT_DIR)/bin/operator-sdk generate k8s
	$(call header,"Generates clientset code")
	GOPATH=$(GOPATH_1) ./vendor/k8s.io/code-generator/generate-groups.sh client \
		$(PACKAGE_NAME)/pkg/client \
		$(PACKAGE_NAME)/pkg/apis \
		"istio:v1alpha1" \
		--go-header-file ./scripts/boilerplate.txt

# ##########################################################################
# Build configuration
# ##########################################################################

OS:=$(shell uname -s)
export OS
GOOS?=$(shell echo $(OS) | awk '{print tolower($$0)}')
GOARCH:= amd64

BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GITUNTRACKEDCHANGES:=$(shell git status --porcelain --untracked-files=no)
COMMIT:=$(shell git rev-parse --short HEAD)
ifneq ($(GITUNTRACKEDCHANGES),)
	COMMIT:=$(COMMIT)-dirty
endif

IKE_VERSION?=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
GIT_TAG:=$(shell git describe --tags --abbrev=0 --exact-match > /dev/null 2>&1; echo $$?)
ifneq ($(GIT_TAG),0)
	IKE_VERSION:=$(IKE_VERSION)-next-$(COMMIT)
else ifneq ($(GITUNTRACKEDCHANGES),)
	IKE_VERSION:=$(IKE_VERSION)-dirty
endif

.PHONY: version
version:
	@echo $(IKE_VERSION)

GOBUILD:=GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0
RELEASE?=false
LDFLAGS="-w -X ${PACKAGE_NAME}/version.Release=${RELEASE} -X ${PACKAGE_NAME}/version.Version=${IKE_VERSION} -X ${PACKAGE_NAME}/version.Commit=${COMMIT} -X ${PACKAGE_NAME}/version.BuildTime=${BUILD_TIME}"
SRCS=$(shell find ./pkg -name "*.go") $(shell find ./cmd -name "*.go") $(shell find ./version -name "*.go") $(shell find ./test -name "*.go")

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

##@ Setup

.PHONY: install-dep
install-dep:
	go get -u github.com/golang/dep/cmd/dep

.PHONY: tools
tools: install-dep ## Installs required go tools
	$(call header,"Installing required tools")
	GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0
	go get -u golang.org/x/tools/cmd/goimports
	$(eval GINKGO_VERSION:=$(shell dep status -f='{{if eq .ProjectRoot "github.com/onsi/ginkgo"}}{{.Version}}{{end}}'))
	GO111MODULE=on go get -u github.com/onsi/ginkgo/ginkgo@$(GINKGO_VERSION)
	GO111MODULE=on go get -u github.com/go-bindata/go-bindata/...@v3.1.2
	GO111MODULE=on go get -u github.com/golang/protobuf/protoc-gen-go

EXECUTABLES:=dep golangci-lint goimports ginkgo go-bindata protoc-gen-go
CHECK:=$(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec) 2>/dev/null),,"install"))
.PHONY: check-tools
check-tools:
	$(call header,"Checking required tools")
	@$(if $(strip $(CHECK)),$(MAKE) -f $(THIS_MAKEFILE) tools,echo "'$(EXECUTABLES)' are installed")

OPERATOR_OS=linux-gnu
ifeq ($(OS), Darwin)
	OPERATOR_OS=apple-darwin
endif
OPERATOR_ARCH:=$(shell uname -m)

$(PROJECT_DIR)/bin/operator-sdk:
	$(call header,"Installing operator-sdk cli")
	mkdir -p $(PROJECT_DIR)/bin/
	$(eval OPERATOR_SDK_VERSION:=$(shell dep status -f='{{if eq .ProjectRoot "github.com/operator-framework/operator-sdk"}}{{.Version}}{{end}}'))
	wget -q --show-progress -c https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk-$(OPERATOR_SDK_VERSION)-$(OPERATOR_ARCH)-$(OPERATOR_OS) -O $(PROJECT_DIR)/bin/operator-sdk
	chmod +x $(PROJECT_DIR)/bin/operator-sdk

$(PROJECT_DIR)/bin/protoc:
	$(call header,"Installing protoc")
	mkdir -p $(PROJECT_DIR)/bin/
	$(PROJECT_DIR)/scripts/dev/get-protobuf.sh
	chmod +x $(PROJECT_DIR)/bin/protoc

$(PROJECT_DIR)/$(ASSETS): $(ASSET_SRCS)
	$(call header,"Adds assets to the binary")
	go-bindata -o $(ASSETS) -pkg assets -ignore 'example.yaml' $(ASSET_SRCS)

# ##########################################################################
##@ Docker
# ##########################################################################

IKE_IMAGE_NAME?=$(PROJECT_NAME)
IKE_TEST_IMAGE_NAME?=$(IKE_IMAGE_NAME)-test
IKE_TEST_PREPARED_IMAGE_NAME?=$(IKE_TEST_IMAGE_NAME)-prepared
IKE_TEST_PREPARED_NAME?=prepared-image
IKE_IMAGE_TAG?=$(IKE_VERSION)
IKE_DOCKER_REGISTRY?=quay.io
IKE_DOCKER_REPOSITORY?=maistra
export IKE_IMAGE_TAG
export IKE_VERSION

.PHONY: docker-build
docker-build: GOOS=linux
docker-build: compile ## Builds the docker image
	$(call header,"Building docker image $(IKE_IMAGE_NAME)")
	docker build \
		--label "org.opencontainers.image.title=$(IKE_IMAGE_NAME)" \
		--label "org.opencontainers.image.description=Tool enabling developers to safely develop and test on any kubernetes cluster without distracting others." \
		--label "org.opencontainers.image.source=https://$(PACKAGE_NAME)" \
		--label "org.opencontainers.image.documentation=https://istio-workspace-docs.netlify.com/istio-workspace" \
		--label "org.opencontainers.image.licenses=Apache-2.0" \
		--label "org.opencontainers.image.authors=Aslak Knutsen, Bartosz Majsak" \
		--label "org.opencontainers.image.vendor=Red Hat, Inc." \
		--label "org.opencontainers.image.revision=$(COMMIT)" \
		--label "org.opencontainers.image.created=$(shell date -u +%F\ %T%z)" \
		-t $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_IMAGE_NAME):$(IKE_IMAGE_TAG) \
		-f $(BUILD_DIR)/Dockerfile $(PROJECT_DIR)
	docker tag \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_IMAGE_NAME):$(IKE_IMAGE_TAG) \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_IMAGE_NAME):latest

.PHONY: docker-push
docker-push: docker-push--latest docker-push-versioned ## Pushes docker images to the registry

docker-push-versioned: docker-push--$(IKE_IMAGE_TAG)

docker-push--%:
	$(eval image_tag:=$(subst docker-push--,,$@))
	$(call header,"Pushing docker image $(image_tag)")
	docker push $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_IMAGE_NAME):$(image_tag)

.PHONY: docker-build-test
docker-build-test: $(BINARY_DIR)/$(TEST_BINARY_NAME)
	$(call header,"Building docker image $(IKE_TEST_IMAGE_NAME)")
	docker build \
		--label "org.opencontainers.image.title=$(IKE_TEST_IMAGE_NAME)" \
		--label "org.opencontainers.image.description=Test Services for end-to-end testing of the $(IKE_IMAGE_NAME)" \
		--label "org.opencontainers.image.source=https://$(PACKAGE_NAME)" \
		--label "org.opencontainers.image.documentation=https://istio-workspace-docs.netlify.com/istio-workspace" \
		--label "org.opencontainers.image.licenses=Apache-2.0" \
		--label "org.opencontainers.image.authors=Aslak Knutsen, Bartosz Majsak" \
		--label "org.opencontainers.image.vendor=Red Hat, Inc." \
		--label "org.opencontainers.image.revision=$(COMMIT)" \
		--label "org.opencontainers.image.created=$(shell date -u +%F\ %T%z)" \
		-t $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):$(IKE_IMAGE_TAG) \
		-f $(BUILD_DIR)/DockerfileTest $(PROJECT_DIR)

	docker tag \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):$(IKE_IMAGE_TAG) \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):latest

.PHONY: docker-push-test
docker-push-test:
	$(call header,"Pushing docker image $(IKE_TEST_IMAGE_NAME)")
	docker push $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):$(IKE_IMAGE_TAG)
	docker push $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_IMAGE_NAME):latest

.PHONY: docker-build-test-prepared
docker-build-test-prepared:
	$(call header,"Building docker image $(IKE_TEST_PREPARED_IMAGE_NAME)")
	docker build \
		--label "org.opencontainers.image.title=$(IKE_TEST_PREPARED_IMAGE_NAME)" \
		--label "org.opencontainers.image.description=Test Prepared Services for end-to-end testing of the $(IKE_IMAGE_NAME)" \
		--label "org.opencontainers.image.source=https://$(PACKAGE_NAME)" \
		--label "org.opencontainers.image.documentation=https://istio-workspace-docs.netlify.com/istio-workspace" \
		--label "org.opencontainers.image.licenses=Apache-2.0" \
		--label "org.opencontainers.image.authors=Aslak Knutsen, Bartosz Majsak" \
		--label "org.opencontainers.image.vendor=Red Hat, Inc." \
		--label "org.opencontainers.image.revision=$(COMMIT)" \
		--label "org.opencontainers.image.created=$(shell date -u +%F\ %T%z)" \
		--build-arg name=$(IKE_TEST_PREPARED_NAME) \
		-t $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):$(IKE_IMAGE_TAG) \
		-f $(BUILD_DIR)/DockerfileTestPrepared $(PROJECT_DIR)

	docker tag \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):$(IKE_IMAGE_TAG) \
		$(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):latest

.PHONY: docker-push-test-prepared
docker-push-test-prepared:
	$(call header,"Pushing docker image $(IKE_TEST_PREPARED_IMAGE_NAME)")
	docker push $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):$(IKE_IMAGE_TAG)
	docker push $(IKE_DOCKER_REGISTRY)/$(IKE_DOCKER_REPOSITORY)/$(IKE_TEST_PREPARED_IMAGE_NAME)-$(IKE_TEST_PREPARED_NAME):latest

# ##########################################################################
##@ Istio-workspace operator deployment
# ##########################################################################

define process_template # params: template location
	@oc process -f $(1) \
		-o yaml \
		--ignore-unknown-parameters=true \
		--local \
		-p IKE_DOCKER_REGISTRY=$(IKE_DOCKER_REGISTRY) \
		-p IKE_DOCKER_REPOSITORY=$(IKE_DOCKER_REPOSITORY) \
		-p IKE_IMAGE_NAME=$(IKE_IMAGE_NAME) \
		-p IKE_IMAGE_TAG=$(IKE_IMAGE_TAG) \
		-p NAMESPACE=$(OPERATOR_NAMESPACE) \
		-p TELEPRESENCE_VERSION=$(TELEPRESENCE_VERSION) \
		-p WATCH_NAMESPACE=$(OPERATOR_WATCH_NAMESPACE)
endef

.PHONY: deploy-operator
deploy-operator: ## Deploys istio-workspace operator resources to defined OPERATOR_NAMESPACE
	$(call header,"Deploying operator to $(OPERATOR_NAMESPACE)")
	oc new-project $(OPERATOR_NAMESPACE) || true
	oc apply -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/crds/istio_v1alpha1_session_crd.yaml
	oc apply -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/service_account.yaml
	oc apply -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/role.yaml
	$(call process_template,deploy/istio-workspace/role_binding.yaml) | oc apply -n $(OPERATOR_NAMESPACE) -f -
	$(call process_template,deploy/istio-workspace/operator.yaml) | oc apply -n $(OPERATOR_NAMESPACE) -f -

.PHONY: undeploy-operator
undeploy-operator: ## Undeploys istio-workspace operator resources from defined OPERATOR_NAMESPACE
	$(call header,"Undeploying operator from $(OPERATOR_NAMESPACE)")
	$(call process_template,deploy/istio-workspace/operator.yaml) | oc delete -n $(OPERATOR_NAMESPACE) -f -
	$(call process_template,deploy/istio-workspace/role_binding.yaml) | oc delete -n $(OPERATOR_NAMESPACE) -f -
	oc delete -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/role.yaml
	oc delete -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/service_account.yaml
	oc delete -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/crds/istio_v1alpha1_session_crd.yaml

.PHONY: deploy-operator-local
deploy-operator-local: export OPERATOR_WATCH_NAMESPACE=$(OPERATOR_NAMESPACE)
deploy-operator-local: ## Deploys istio-workspace operator resources to a single Namespace defined by OPERATOR_NAMESPACE
	$(call header,"Deploying local operator to $(OPERATOR_NAMESPACE)")
	oc new-project $(OPERATOR_NAMESPACE) || true
	oc apply -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/crds/istio_v1alpha1_session_crd.yaml
	oc apply -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/service_account.yaml
	oc apply -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/role_local.yaml
	$(call process_template,deploy/istio-workspace/role_binding_local.yaml) | oc apply -n $(OPERATOR_NAMESPACE) -f -
	$(call process_template,deploy/istio-workspace/operator.yaml) | oc apply -n $(OPERATOR_NAMESPACE) -f -

.PHONY: undeploy-operator-local
undeploy-operator-local: export OPERATOR_WATCH_NAMESPACE=$(OPERATOR_NAMESPACE)
undeploy-operator-local: ## Undeploys istio-workspace operator resources from a single Namespace defined by OPERATOR_NAMESPACE
	$(call header,"Undeploying local operator from $(OPERATOR_NAMESPACE)")
	$(call process_template,deploy/istio-workspace/operator.yaml) | oc delete -n $(OPERATOR_NAMESPACE) -f -
	$(call process_template,deploy/istio-workspace/role_binding_local.yaml) | oc delete -n $(OPERATOR_NAMESPACE) -f -
	oc delete -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/role_local.yaml
	oc delete -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/service_account.yaml

# ##########################################################################
##@ Istio-workspace example deployment
# ##########################################################################

.PHONY: deploy-example
deploy-example: ## Deploys istio-workspace specific resources to defined TEST_NAMESPACE
	$(call header,"Deploying session custom resource to $(TEST_NAMESPACE)")
	oc apply -n $(TEST_NAMESPACE) -f deploy/istio-workspace/crds/istio_v1alpha1_session_cr.example.yaml

.PHONY: undeploy-example
undeploy-example: ## Undeploys istio-workspace specific resources from defined TEST_NAMESPACE
	$(call header,"Undeploying session custom resource to $(TEST_NAMESPACE)")
	oc delete -n $(TEST_NAMESPACE) -f deploy/istio-workspace/crds/istio_v1alpha1_session_cr.example.yaml

# ##########################################################################
# Istio test application deployment
# ##########################################################################

deploy-test-%:
	$(eval scenario:=$(subst deploy-test-,,$@))
	$(call header,"Deploying bookinfo $(scenario) app to $(TEST_NAMESPACE)")

	oc new-project $(TEST_NAMESPACE) || true
	oc adm policy add-scc-to-user anyuid -z default -n $(TEST_NAMESPACE)
	oc adm policy add-scc-to-user privileged -z default -n $(TEST_NAMESPACE)
	oc apply -n $(TEST_NAMESPACE) -f deploy/bookinfo/session_role.yaml
	oc apply -n $(TEST_NAMESPACE) -f deploy/bookinfo/session_rolebinding.yaml

	go run ./test/cmd/test-scenario/ $(scenario) | oc apply -n $(TEST_NAMESPACE) -f -

undeploy-test-%:
	$(eval scenario:=$(subst undeploy-test-,,$@))
	$(call header,"Undeploying bookinfo $(scenario) app from $(TEST_NAMESPACE)")

	go run ./test/cmd/test-scenario/ $(scenario) | oc delete -n $(TEST_NAMESPACE) -f -
	oc delete -n $(TEST_NAMESPACE) -f deploy/bookinfo/session_rolebinding.yaml
	oc delete -n $(TEST_NAMESPACE) -f deploy/bookinfo/session_role.yaml

##@ Helpers

.PHONY: help
help:  ## Displays this help \o/
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m\033[2m %s\033[0m\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@cat $(MAKEFILE_LIST) | grep "^[A-Za-z_]*.?=" | sort | awk 'BEGIN {FS="?="; printf "\n\n\033[1mEnvironment variables\033[0m\n"} {printf "  \033[36m%-25s\033[0m\033[2m %s\033[0m\n", $$1, $$2}'
