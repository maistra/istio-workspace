PROJECT_NAME:=istio-workspace
PACKAGE_NAME:=github.com/aslakknutsen/istio-workspace

OPERATOR_NAMESPACE?=istio-workspace-operator
EXAMPLE_NAMESPACE?=bookinfo

CUR_DIR:=$(shell pwd)
BUILD_DIR:=$(CUR_DIR)/build
BINARY_DIR:=$(CUR_DIR)/dist
BINARY_NAME:=ike

# Call this function with $(call header,"Your message")
define header =
@echo -e "\n\e[92m\e[4m\e[1m$(1)\e[0m\n"
endef

.PHONY: all
all: deps format lint test compile ## (default) Runs 'deps format lint test compile' targets

export CUR_DIR

.PHONY: help
help:
	 @echo -e "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sort | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\x1b[36m\1\\x1b[m:\2/' | column -c2 -t -s :)"

.PHONY: deps
deps: ## Fetches all dependencies
	$(call header,"Fetching dependencies")
	dep ensure -v

.PHONY: format
format: ## Removes unneeded imports and formats source code
	$(call header,"Formatting code")
	goimports -l -w ./pkg/ ./cmd/ ./version/ ./test/ ./e2e/

.PHONY: tools
tools: ## Installs required go tools
	$(call header,"Installing required tools")
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/onsi/ginkgo/ginkgo

.PHONY: install-operator-sdk
install-operator-sdk: ## Downloads operator-sdk cli tool aligned with version defined in Gopkg
	$(call header,"Installing operator-sdk cli tool")
	mkdir -p $(CUR_DIR)/bin/
	$(eval OPERATOR_SDK_VERSION:=$(shell dep status -f='{{if eq .ProjectRoot "github.com/operator-framework/operator-sdk"}}{{.Version}}{{end}}'))
	wget -c https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk-$(OPERATOR_SDK_VERSION)-x86_64-linux-gnu -O $(CUR_DIR)/bin/operator-sdk
	chmod +x $(CUR_DIR)/bin/operator-sdk

.PHONY: lint
lint: deps ## Concurrently runs a whole bunch of static analysis tools
	$(call header,"Running a whole bunch of static analysis tools")
	golangci-lint run

GROUP_VERSIONS:="istio:v1alpha1"
.PHONY: codegen
codegen: install-operator-sdk ## Generates operator-sdk code
	$(call header,"Generates operator-sdk code")
	$(CUR_DIR)/bin/operator-sdk generate k8s
	$(call header,"Generates clientset code")
	GOPATH=$(shell echo ${GOPATH} | rev | cut -d':' -f 2 | rev) ./vendor/k8s.io/code-generator/generate-groups.sh client \
		$(PACKAGE_NAME)/pkg/client \
		$(PACKAGE_NAME)/pkg/apis \
		$(GROUP_VERSIONS)

.PHONY: compile
compile: codegen $(BINARY_DIR)/$(BINARY_NAME) ## Compiles binaries

.PHONY: test
test: codegen ## Runs tests
	$(call header,"Running tests")
	ginkgo -r -v --skipPackage=e2e ${args}

.PHONY: test-e2e
test-e2e: codegen ## Runs end-to-end tests
	$(call header,"Running end-to-end tests")
	ginkgo e2e/ -v -p ${args}

.PHONY: clean
clean: ## Removes build artifacts
	rm -rf $(BINARY_DIR) $(CUR_DIR)/bin/

# ##########################################################################
# Build configuration
# ##########################################################################

BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GITUNTRACKEDCHANGES:=$(shell git status --porcelain --untracked-files=no)
COMMIT:=$(shell git rev-parse --short HEAD)
ifneq ($(GITUNTRACKEDCHANGES),)
	COMMIT:=$(COMMIT)-dirty
endif
VERSION?=0.0.1
LDFLAGS="-w -X ${PACKAGE_NAME}/version.Version=${VERSION} -X ${PACKAGE_NAME}/version.Commit=${COMMIT} -X ${PACKAGE_NAME}/version.BuildTime=${BUILD_TIME}"

SRCS=$(shell find ./pkg -name "*.go") $(shell find ./cmd -name "*.go") $(shell find ./version -name "*.go")

$(BINARY_DIR):
	[ -d $@ ] || mkdir -p $@

$(BINARY_DIR)/$(BINARY_NAME): $(BINARY_DIR) $(SRCS)
	$(call header,"Compiling... carry on!")
	GOOS=linux CGO_ENABLED=0 go build -ldflags ${LDFLAGS} -o $@ ./cmd/$(BINARY_NAME)/

# ##########################################################################
# Docker build
# ##########################################################################

DOCKER_IMAGE?=$(PROJECT_NAME)
DOCKER_IMAGE_TAG?=$(COMMIT)
export DOCKER_IMAGE_TAG
DOCKER_REGISTRY?=docker.io
DOCKER_REPOSITORY?=aslakknutsen

.PHONY: docker-build
docker-build: ## Builds the docker image
	$(call header,"Building docker image $(DOCKER_IMAGE_CORE)")
	docker build \
		-t $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_IMAGE):$(COMMIT) \
		-f $(BUILD_DIR)/Dockerfile $(CUR_DIR)
	docker tag \
		$(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_IMAGE):$(COMMIT) \
		$(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_IMAGE):latest

.PHONY: docker-push
docker-push: ## Pushes docker image to the registry
	$(call header,"Pushing docker image $(DOCKER_IMAGE_CORE)")
	docker push $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_IMAGE):$(COMMIT)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_IMAGE):latest

# ##########################################################################
# Istio operator deployment
# ##########################################################################

define process_template # params: template location
	@oc process -f $(1) \
		-o yaml \
		--ignore-unknown-parameters=true \
		--local \
		-p DOCKER_REGISTRY=$(DOCKER_REGISTRY) \
		-p DOCKER_REPOSITORY=$(DOCKER_REPOSITORY) \
		-p IMAGE_NAME=$(DOCKER_IMAGE) \
		-p IMAGE_TAG=$(DOCKER_IMAGE_TAG) \
		-p NAMESPACE=$(OPERATOR_NAMESPACE)
endef

.PHONY: load-istio
load-istio: ## Triggers installation of istio in the cluster
	$(call header,"Deploying operator to $(OPERATOR_NAMESPACE)")
	oc create -n istio-operator -f deploy/istio/minimal-cr.yaml

.PHONY: deploy-operator
deploy-operator: ## Deploys operator resources to defined OPERATOR_NAMESPACE
	$(call header,"Deploying operator to $(OPERATOR_NAMESPACE)")
	oc new-project $(OPERATOR_NAMESPACE) || true
	oc apply -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/crds/istio_v1alpha1_session_crd.yaml
	oc apply -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/service_account.yaml
	oc apply -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/role.yaml
	$(call process_template,deploy/istio-workspace/role_binding.yaml) | oc apply -n $(OPERATOR_NAMESPACE) -f -
	$(call process_template,deploy/istio-workspace/operator.yaml) | oc apply -n $(OPERATOR_NAMESPACE) -f -

.PHONY: undeploy-operator
undeploy-operator: ## Undeploys operator resources from defined OPERATOR_NAMESPACE
	$(call header,"Undeploying operator to $(OPERATOR_NAMESPACE)")
	$(call process_template,deploy/istio-workspace/operator.yaml) | oc delete -n $(OPERATOR_NAMESPACE) -f -
	$(call process_template,deploy/istio-workspace/role_binding.yaml) | oc delete -n $(OPERATOR_NAMESPACE) -f -
	oc delete -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/role.yaml
	oc delete -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/service_account.yaml
	oc delete -n $(OPERATOR_NAMESPACE) -f deploy/istio-workspace/crds/istio_v1alpha1_session_crd.yaml

# ##########################################################################
# Istio example deployment
# ##########################################################################

.PHONY: deploy-example
deploy-example: ## Deploys istio-workspace specific resources to defined EXAMPLE_NAMESPACE
	$(call header,"Deploying session custom resource to $(EXAMPLE_NAMESPACE)")
	oc apply -n $(EXAMPLE_NAMESPACE) -f deploy/istio-workspace/crds/istio_v1alpha1_session_cr.yaml

.PHONY: undeploy-example
undeploy-example: ## Undeploys istio-workspace specific resources from defined EXAMPLE_NAMESPACE
	$(call header,"Undeploying session custom resource to $(EXAMPLE_NAMESPACE)")
	oc delete -n $(EXAMPLE_NAMESPACE) -f deploy/istio-workspace/crds/istio_v1alpha1_session_cr.yaml

# ##########################################################################
# Istio bookinfo deployment
# ##########################################################################

.PHONY: deploy-bookinfo
deploy-bookinfo: ## Deploys bookinfo app into defined EXAMPLE_NAMESPACE
	$(call header,"Deploying bookinfo app to $(EXAMPLE_NAMESPACE)")
	oc new-project $(EXAMPLE_NAMESPACE) || true
	oc adm policy add-scc-to-user anyuid -z default -n $(EXAMPLE_NAMESPACE)
	oc adm policy add-scc-to-user privileged -z default -n $(EXAMPLE_NAMESPACE)
	oc apply -n $(EXAMPLE_NAMESPACE) -f deploy/bookinfo/session_role.yaml
	oc apply -n $(EXAMPLE_NAMESPACE) -f deploy/bookinfo/session_rolebinding.yaml
	oc apply -n $(EXAMPLE_NAMESPACE) -f deploy/bookinfo/bookinfo-gateway.yaml
	oc apply -n $(EXAMPLE_NAMESPACE) -f deploy/bookinfo/destination-rule-all.yaml
	oc apply -n $(EXAMPLE_NAMESPACE) -f deploy/bookinfo/virtual-service-all-v1.yaml
	oc apply -n $(EXAMPLE_NAMESPACE) -f deploy/bookinfo/bookinfo.yaml
	# Required due to circle-ci memory limitations
	oc delete -n $(EXAMPLE_NAMESPACE) deployment reviews-v2
	oc delete -n $(EXAMPLE_NAMESPACE) deployment reviews-v3

.PHONY: undeploy-bookinfo
undeploy-bookinfo: ## Undeploys bookinfo app into defined EXAMPLE_NAMESPACE
	$(call header,"Undeploying bookinfo app to $(EXAMPLE_NAMESPACE)")
	oc delete -n $(EXAMPLE_NAMESPACE) -f deploy/bookinfo/bookinfo.yaml	
	oc delete -n $(EXAMPLE_NAMESPACE) -f deploy/bookinfo/virtual-service-all-v1.yaml
	oc delete -n $(EXAMPLE_NAMESPACE) -f deploy/bookinfo/destination-rule-all.yaml
	oc delete -n $(EXAMPLE_NAMESPACE) -f deploy/bookinfo/bookinfo-gateway.yaml
	oc delete -n $(EXAMPLE_NAMESPACE) -f deploy/bookinfo/session_rolebinding.yaml
	oc delete -n $(EXAMPLE_NAMESPACE) -f deploy/bookinfo/session_role.yaml

