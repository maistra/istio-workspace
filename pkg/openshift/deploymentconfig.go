package openshift

import (
	"github.com/aslakknutsen/istio-workspace/pkg/model"
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

func DeploymentConfigMutator(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	return nil
}

func DeploymentConfigRevertor(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	return nil
}
