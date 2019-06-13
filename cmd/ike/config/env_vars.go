package config

import (
	"github.com/maistra/istio-workspace/pkg/assets"
	"github.com/maistra/istio-workspace/pkg/openshift/parser"

	openshiftApi "github.com/openshift/api/template/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// These consts are used across the project to load corresponding env variables
// Extend when needed
const (
	EnvDockerRepository = "IKE_DOCKER_REPOSITORY"
	EnvDockerRegistry   = "IKE_DOCKER_REGISTRY"
)

var (
	log        = logf.Log.WithName("session_handler")
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
		log.Error(err, "failed parsing "+resource+"template")
	}

	tplParams, err := parser.ParseParameters(tpl)
	if err != nil {
		log.Error(err, "failed parsing parameters in the template "+resource)
	}
	Parameters = append(Parameters, tplParams...)
}
