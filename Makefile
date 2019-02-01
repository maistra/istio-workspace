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
DOCKER_REPO?=docker.io/aslakknutsen

docker-build: ## Builds the docker image
	@echo "Building docker image $(DOCKER_IMAGE_CORE)"
	$(DOCKER) build \
		-t $(DOCKER_REPO)/$(DOCKER_IMAGE):$(COMMIT) \
		-f $(BUILD_DIR)/Dockerfile $(CUR_DIR)
	$(DOCKER) tag \
		$(DOCKER_REPO)/$(DOCKER_IMAGE):$(COMMIT) \
		$(DOCKER_REPO)/$(DOCKER_IMAGE):latest

# istio example deployment
.PHONY:
deploy-operator:
	@echo "Deploying operator to $(OPERATOR_NAMESPACE)"
	oc apply -f deploy/crds/istio_v1alpha1_session_crd.yaml -n $(OPERATOR_NAMESPACE)
	oc apply -f deploy/service_account.yaml -n $(OPERATOR_NAMESPACE)
	oc apply -f deploy/role.yaml -n $(OPERATOR_NAMESPACE)
	oc apply -f deploy/role_binding.yaml -n $(OPERATOR_NAMESPACE)
	oc apply -f deploy/operator.yaml -n $(OPERATOR_NAMESPACE)

.PHONY:
undeploy-operator:
	@echo "UnDeploying operator to $(OPERATOR_NAMESPACE)"
	oc delete -f deploy/operator.yaml -n $(OPERATOR_NAMESPACE)
	oc delete -f deploy/role_binding.yaml -n $(OPERATOR_NAMESPACE)
	oc delete -f deploy/role.yaml -n $(OPERATOR_NAMESPACE)
	oc delete -f deploy/service_account.yaml -n $(OPERATOR_NAMESPACE)
	oc delete -f deploy/crds/istio_v1alpha1_session_crd.yaml -n $(OPERATOR_NAMESPACE)

.PHONY:
deploy-example:
	@echo "Deploying operator to $(EXAMPLE_NAMESPACE)"
	oc apply -f deploy/crds/istio_v1alpha1_session_cr.yaml -n $(EXAMPLE_NAMESPACE)

.PHONY:
undeploy-example:
	@echo "UnDeploying operator to $(EXAMPLE_NAMESPACE)"
	oc delete -f deploy/crds/istio_v1alpha1_session_cr.yaml -n $(EXAMPLE_NAMESPACE)
