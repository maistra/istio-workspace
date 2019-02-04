PROJECT_NAME:=istio-workspace
PACKAGE_NAME:=github.com/aslakknutsen/istio-workspace

OPERATOR_NAMESPACE?=istio-system
EXAMPLE_NAMESPACE?=bookinfo

CUR_DIR = $(shell pwd)
BUILD_DIR:=${PWD}/build
BINARY_DIR:=${PWD}/dist
BINARY_NAME:=ike

.PHONY: all
all: tools deps format lint compile ## (default)

.PHONY: help
help:
	 @echo -e "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sort | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\x1b[36m\1\\x1b[m:\2/' | column -c2 -t -s :)"

.PHONY: deps
deps: ## Fetches all dependencies using dep
	dep ensure -v

.PHONY: format
format: ## Removes unneeded imports and formats source code
	@goimports -l -w pkg/ cmd/ version/

.PHONY: tools
tools: ## Installs required go tools
	@go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	@go get -u golang.org/x/tools/cmd/goimports
	@mkdir -p bin/
	@wget https://github.com/operator-framework/operator-sdk/releases/download/v0.4.0/operator-sdk-v0.4.0-x86_64-linux-gnu -O ./bin/operator-sdk
	@chmod +x ./bin/operator-sdk


.PHONY: lint
lint: deps ## Concurrently runs a whole bunch of static analysis tools
	@golangci-lint run

.PHONY: codegen
codegen:
	./bin/operator-sdk generate k8s

.PHONY: compile
compile: codegen $(BINARY_DIR)/$(BINARY_NAME)

.PHONY: clean
clean:
	rm -rf $(BINARY_DIR)

# Build configuration
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GITUNTRACKEDCHANGES:=$(shell git status --porcelain --untracked-files=no)
COMMIT:=$(shell git rev-parse --short HEAD)
ifneq ($(GITUNTRACKEDCHANGES),)
	COMMIT := $(COMMIT)-dirty
endif
VERSION?=0.0.1
LDFLAGS="-w -X ${PACKAGE_NAME}/version.Version=${VERSION} -X ${PACKAGE_NAME}/version.Commit=${COMMIT} -X ${PACKAGE_NAME}/version.BuildTime=${BUILD_TIME}"

SRCS=$(shell find ./pkg -name "*.go") $(shell find ./cmd -name "*.go") $(shell find ./version -name "*.go")

$(BINARY_DIR):
	[ -d $@ ] || mkdir -p $@

$(BINARY_DIR)/$(BINARY_NAME): $(BINARY_DIR) $(SRCS)
	GOOS=linux CGO_ENABLED=0 go build -ldflags ${LDFLAGS} -o $@ ./cmd/$(BINARY_NAME)/

# Docker build

DOCKER?=$(if $(or $(in_docker_group),$(is_root)),docker,sudo docker)
DOCKER_IMAGE?=$(PROJECT_NAME)
DOCKER_REGISTRY?=docker.io
DOCKER_REPOSITORY?=aslakknutsen

.PHONY: deploy-operator
docker-build: ## Builds the docker image
	@echo "Building docker image $(DOCKER_IMAGE_CORE)"
	$(DOCKER) build \
		-t $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_IMAGE):$(COMMIT) \
		-f $(BUILD_DIR)/Dockerfile $(CUR_DIR)
	$(DOCKER) tag \
		$(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_IMAGE):$(COMMIT) \
		$(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_IMAGE):latest

# istio example deployment

define process_template # params: template location
	@oc process -f $(1) \
		-o yaml \
		--local \
		-p DOCKER_REGISTRY=$(DOCKER_REGISTRY) \
		-p DOCKER_REPOSITORY=$(DOCKER_REPOSITORY) \
		-p IMAGE_NAME=$(DOCKER_IMAGE) \
		-p IMAGE_TAG=$(COMMIT)
endef

.PHONY: deploy-operator
deploy-operator:
	@echo "Deploying operator to $(OPERATOR_NAMESPACE)"
	oc apply -n $(OPERATOR_NAMESPACE) -f deploy/crds/istio_v1alpha1_session_crd.yaml
	oc apply -n $(OPERATOR_NAMESPACE) -f deploy/service_account.yaml
	oc apply -n $(OPERATOR_NAMESPACE) -f deploy/role.yaml
	oc apply -n $(OPERATOR_NAMESPACE) -f deploy/role_binding.yaml
	$(call process_template,deploy/operator.yaml) | oc apply -n $(OPERATOR_NAMESPACE) -f -

.PHONY: undeploy-operator
undeploy-operator:
	@echo "UnDeploying operator to $(OPERATOR_NAMESPACE)"
	$(call process_template,deploy/operator.yaml) | oc delete -n $(OPERATOR_NAMESPACE) -f -
	oc delete -n $(OPERATOR_NAMESPACE) -f deploy/role_binding.yaml
	oc delete -n $(OPERATOR_NAMESPACE) -f deploy/role.yaml
	oc delete -n $(OPERATOR_NAMESPACE) -f deploy/service_account.yaml
	oc delete -n $(OPERATOR_NAMESPACE) -f deploy/crds/istio_v1alpha1_session_crd.yaml

.PHONY: deploy-example
deploy-example:
	@echo "Deploying operator to $(EXAMPLE_NAMESPACE)"
	oc apply -n $(EXAMPLE_NAMESPACE) -f deploy/crds/istio_v1alpha1_session_cr.yaml

.PHONY: undeploy-example
undeploy-example:
	@echo "UnDeploying operator to $(EXAMPLE_NAMESPACE)"
	oc delete -n $(EXAMPLE_NAMESPACE) -f deploy/crds/istio_v1alpha1_session_cr.yaml
