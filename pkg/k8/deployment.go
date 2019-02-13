package k8

import (
	"github.com/aslakknutsen/istio-workspace/pkg/model"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// DeploymentKind is the k8 Kind for a Deployment
	DeploymentKind = "Deployment"
)

var _ model.Locator = DeploymentLocator
var _ model.Mutator = DeploymentMutator
var _ model.Revertor = DeploymentRevertor

// DeploymentLocator attempts to locate a Deployment kind based on Ref name
func DeploymentLocator(ctx model.SessionContext, ref *model.Ref) bool {
	deployment, err := getDeployment(ctx, ctx.Namespace, ref.Name)
	if err != nil {
		if errors.IsNotFound(err) { // Ref is not a Deployment type
			return false
		}
		ctx.Log.Error(nil, "Could not get Deployment", "name", deployment.Name)
		return false
	}
	return true
}

func DeploymentMutator(ctx model.SessionContext, ref *model.Ref) error {
	deployment, err := getDeployment(ctx, ctx.Namespace, ref.Name)
	if err != nil {
		return err
	}
	deploymentClone := cloneDeployment(deployment.DeepCopy())

	err = ctx.Client.Create(ctx, deploymentClone)
	if err != nil {
		ctx.Log.Info("Failed to clone Deployment", "name", deploymentClone.Name)
		ref.AddResourceStatus(model.ResourceStatus{Kind: DeploymentKind, Name: deploymentClone.Name, Action: model.ActionFailed})
		return err
	}
	ctx.Log.Info("Cloned Deployment", "name", deploymentClone.Name)
	ref.AddResourceStatus(model.ResourceStatus{Kind: DeploymentKind, Name: deploymentClone.Name, Action: model.ActionCreated})
	return nil
}

func DeploymentRevertor(ctx model.SessionContext, ref *model.Ref) error {
	statuses := ref.GetResourceStatus(DeploymentKind)
	for _, status := range statuses {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: status.Name, Namespace: ctx.Namespace},
		}
		err := ctx.Client.Delete(ctx, deployment)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			ctx.Log.Info("Failed to delete Deployment", "name", deployment.Name)
			ref.AddResourceStatus(model.ResourceStatus{Kind: DeploymentKind, Name: deployment.Name, Action: model.ActionFailed})
			return err
		}
	}
	return nil
}

func cloneDeployment(deployment *appsv1.Deployment) *appsv1.Deployment {
	deploymentClone := deployment.DeepCopy()
	replicasClone := int32(0)
	labelsClone := deploymentClone.GetLabels()
	labelsClone["version"] = labelsClone["version"] + "-test"
	deploymentClone.SetName(deployment.GetName() + "-test")
	deploymentClone.SetLabels(labelsClone)
	deploymentClone.Spec.Selector.MatchLabels = labelsClone
	deploymentClone.Spec.Template.SetLabels(labelsClone)
	deploymentClone.SetResourceVersion("")
	deploymentClone.Spec.Replicas = &replicasClone
	return deploymentClone
}

func getDeployment(ctx model.SessionContext, namespace, name string) (*appsv1.Deployment, error) {
	deployment := appsv1.Deployment{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &deployment)
	return &deployment, err
}
