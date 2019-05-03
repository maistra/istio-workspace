package openshift

import (
	"github.com/maistra/istio-workspace/pkg/apis"
	"github.com/maistra/istio-workspace/pkg/model"

	appsv1 "github.com/openshift/api/apps/v1"

	corev1 "k8s.io/api/core/v1"
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
		if errors.IsNotFound(err) { // Ref is not a Deployment type
			return false
		}
		ctx.Log.Error(nil, "Could not get DeploymentConfig", "name", deployment.Name)
		return false
	}
	return true
}

// DeploymentConfigMutator attempts to clone the located DeploymentConfig
func DeploymentConfigMutator(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	if len(ref.GetResourceStatus(DeploymentConfigKind)) > 0 {
		return nil
	}

	deployment, err := getDeploymentConfig(ctx, ctx.Namespace, ref.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	ctx.Log.Info("Found DeploymentConfig", "name", ref.Name)

	deploymentClone := cloneDeployment(deployment.DeepCopy())
	err = ctx.Client.Create(ctx, deploymentClone)
	if err != nil {
		ctx.Log.Info("Failed to clone DeploymentConfig", "name", deploymentClone.Name)
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
		ctx.Log.Info("Found DeploymentConfig", "name", ref.Name)
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

func cloneDeployment(deployment *appsv1.DeploymentConfig) *appsv1.DeploymentConfig {
	deploymentClone := deployment.DeepCopy()
	labelsClone := deploymentClone.Spec.Selector
	labelsClone["version"] += "-test"
	labelsClone["telepresence"] = "test"
	deploymentClone.SetName(deployment.GetName() + "-test")
	deploymentClone.SetLabels(labelsClone)
	deploymentClone.Spec.Selector = labelsClone
	deploymentClone.Spec.Template.SetLabels(labelsClone)
	deploymentClone.SetResourceVersion("")
	deploymentClone.Spec.Replicas = 1

	container := deploymentClone.Spec.Template.Spec.Containers[0]
	container.Image = "datawire/telepresence-k8s:0.99"
	container.Env = append(container.Env, corev1.EnvVar{
		Name: "TELEPRESENCE_CONTAINER_NAMESPACE",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.namespace",
			},
		},
	})
	deploymentClone.Spec.Template.Spec.Containers[0] = container
	return deploymentClone
}

func getDeploymentConfig(ctx model.SessionContext, namespace, name string) (*appsv1.DeploymentConfig, error) { //nolint[:hugeParam]
	deployment := appsv1.DeploymentConfig{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &deployment)
	return &deployment, err
}
