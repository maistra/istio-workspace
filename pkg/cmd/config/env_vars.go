package config

import (
	"github.com/maistra/istio-workspace/pkg/assets"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/openshift/parser"

	openshiftApi "github.com/openshift/api/template/v1"
)

// These consts are used across the project to load corresponding env variables
// Extend when needed.
const (
	EnvDockerRepository = "IKE_DOCKER_REPOSITORY"
	EnvDockerRegistry   = "IKE_DOCKER_REGISTRY"
	EnvDockerTestImage  = "IKE_TEST_IMAGE_NAME"
)

var (
	logger     = log.CreateOperatorAwareLogger("session").WithValues("type", "handler")
	Parameters []openshiftApi.Parameter
)

func init() {
	for _, asset := range assets.AssetNames() {
		collectTplParams(asset)
	}
}

func collectTplParams(resource string) {
	tpl, err := parser.Load(resource)
	if err != nil {
		logger.Error(err, "failed parsing "+resource+"template")
	}

	tplParams, err := parser.ParseParameters(tpl)
	if err != nil {
		logger.Error(err, "failed parsing parameters in the template "+resource)
	}
	Parameters = append(Parameters, tplParams...)
}
