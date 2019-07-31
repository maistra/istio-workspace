package openshift

import (
	"encoding/json"
	"os"

	"github.com/maistra/istio-workspace/pkg/apis"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/template"

	appsv1 "github.com/openshift/api/apps/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func init() {
	apis.AddToSchemes = append(apis.AddToSchemes, appsv1.SchemeBuilder.AddToScheme)
}

const (
	// DeploymentConfigKind is the k8s Kind for a openshift DeploymentConfig
	DeploymentConfigKind = "DeploymentConfig"
)

var _ model.Locator = DeploymentConfigLocator
var _ model.Mutator = DeploymentConfigMutator
var _ model.Revertor = DeploymentConfigRevertor

// DeploymentConfigLocator attempts to locate a DeploymentConfig kind based on Ref name
func DeploymentConfigLocator(ctx model.SessionContext, ref *model.Ref) bool { //nolint[:hugeParam]
	deployment, err := getDeploymentConfig(ctx, ctx.Namespace, ref.Name)
	if err != nil {
		if errors.IsNotFound(err) { // Ref is not a DeploymentConfig type
			return false
		}
		ctx.Log.Error(err, "Could not get DeploymentConfig", "name", deployment.Name)
		return false
	}
	ref.Target = model.NewLocatedResource(DeploymentConfigKind, deployment.Name, deployment.Spec.Template.Labels)
	return true
}

// DeploymentConfigMutator attempts to clone the located DeploymentConfig
func DeploymentConfigMutator(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	if len(ref.GetResourceStatus(DeploymentConfigKind)) > 0 {
		return nil
	}
	if !ref.HasTarget(DeploymentConfigKind) {
		return nil
	}
	deployment, err := getDeploymentConfig(ctx, ctx.Namespace, ref.Target.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	ctx.Log.Info("Found DeploymentConfig", "name", ref.Target.Name)

	deploymentClone, err := cloneDeployment(deployment.DeepCopy(), ref.Target.GetNewVersion(ctx.Name))
	if err != nil {
		ctx.Log.Info("Failed to clone DeploymentConfig", "name", deployment.Name)
		return err
	}
	err = ctx.Client.Create(ctx, deploymentClone)
	if err != nil {
		ctx.Log.Info("Failed to create cloned DeploymentConfig", "name", deploymentClone.Name)
		ref.AddResourceStatus(model.ResourceStatus{Kind: DeploymentConfigKind, Name: deploymentClone.Name, Action: model.ActionFailed})
		return err
	}
	ctx.Log.Info("Cloned DeploymentConfig", "name", deploymentClone.Name)
	ref.AddResourceStatus(model.ResourceStatus{Kind: DeploymentConfigKind, Name: deploymentClone.Name, Action: model.ActionCreated})
	return nil
}

// DeploymentConfigRevertor attempts to delete the cloned DeploymentConfig
func DeploymentConfigRevertor(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	statuses := ref.GetResourceStatus(DeploymentConfigKind)
	for _, status := range statuses {
		deployment := &appsv1.DeploymentConfig{
			ObjectMeta: metav1.ObjectMeta{Name: status.Name, Namespace: ctx.Namespace},
		}
		ctx.Log.Info("Found DeploymentConfig", "name", ref.Target.Name)
		err := ctx.Client.Delete(ctx, deployment)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			ctx.Log.Info("Failed to delete DeploymentConfig", "name", deployment.Name)
			ref.AddResourceStatus(model.ResourceStatus{Kind: DeploymentConfigKind, Name: deployment.Name, Action: model.ActionFailed})
			return err
		}
		ref.RemoveResourceStatus(model.ResourceStatus{Kind: DeploymentConfigKind, Name: deployment.Name})
	}
	return nil
}

func cloneDeployment(deployment *appsv1.DeploymentConfig, version string) (*appsv1.DeploymentConfig, error) {
	tpVersion, found := os.LookupEnv("TELEPRESENCE_VERSION")
	if !found {
		tpVersion = "0.101"
	}

	originalDeployment, err := json.Marshal(deployment)
	if err != nil {
		return nil, err
	}

	e := template.NewDefaultEngine()
	modifiedDeployment, err := e.Run("telepresence", originalDeployment, version, map[string]string{
		"TelepresenceVersion": tpVersion,
	})
	if err != nil {
		return nil, err
	}

	clone := appsv1.DeploymentConfig{}
	err = json.Unmarshal(modifiedDeployment, &clone)
	if err != nil {
		return nil, err
	}
	return &clone, nil
}

func getDeploymentConfig(ctx model.SessionContext, namespace, name string) (*appsv1.DeploymentConfig, error) { //nolint[:hugeParam]
	deployment := appsv1.DeploymentConfig{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &deployment)
	return &deployment, err
}
