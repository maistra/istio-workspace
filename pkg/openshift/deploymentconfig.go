package openshift

import (
	"github.com/aslakknutsen/istio-workspace/pkg/model"
)

const (
	// DeploymentConfigKind is the k8 Kind for a openshift DeploymentConfig
	DeploymentConfigKind = "DeploymentConfig"
)

var _ model.Locator = DeploymentConfigLocator
var _ model.Mutator = DeploymentConfigMutator
var _ model.Revertor = DeploymentConfigRevertor

// DeploymentConfigLocator attempts to locate a DeploymentConfig kind based on Ref name
func DeploymentConfigLocator(ctx model.SessionContext, ref *model.Ref) bool {
	return false
}

func DeploymentConfigMutator(ctx model.SessionContext, ref *model.Ref) error {
	return nil
}

func DeploymentConfigRevertor(ctx model.SessionContext, ref *model.Ref) error {
	return nil
}
