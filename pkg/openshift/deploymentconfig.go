package openshift

import (
	"github.com/maistra/istio-workspace/pkg/model"
)

const (
	// DeploymentConfigKind is the k8s Kind for a openshift DeploymentConfig
	DeploymentConfigKind = "DeploymentConfig" //nolint[:unused]
)

var _ model.Locator = DeploymentConfigLocator
var _ model.Mutator = DeploymentConfigMutator
var _ model.Revertor = DeploymentConfigRevertor

// DeploymentConfigLocator attempts to locate a DeploymentConfig kind based on Ref name
func DeploymentConfigLocator(ctx model.SessionContext, ref *model.Ref) bool { //nolint[:hugeParam]
	return false
}

// DeploymentConfigMutator attempts to clone the located DeploymentConfig
func DeploymentConfigMutator(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	return nil
}

// DeploymentConfigRevertor attempts to delete the cloned DeploymentConfig
func DeploymentConfigRevertor(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	return nil
}
