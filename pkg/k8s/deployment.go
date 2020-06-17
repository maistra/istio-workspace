package k8s

import (
	"encoding/json"

	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/template"

	appsv1 "k8s.io/api/apps/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// DeploymentKind is the k8s Kind for a Deployment
	DeploymentKind = "Deployment"
)

var _ model.Locator = DeploymentLocator
var _ model.Mutator = DeploymentMutator
var _ model.Revertor = DeploymentRevertor

// DeploymentLocator attempts to locate a Deployment kind based on Ref name.
func DeploymentLocator(ctx model.SessionContext, ref *model.Ref) bool {
	deployment, err := getDeployment(ctx, ctx.Namespace, ref.Name)
	if err != nil {
		if errors.IsNotFound(err) { // Ref is not a Deployment type
			return false
		}
		ctx.Log.Error(err, "Could not get Deployment", "name", deployment.Name)
		return false
	}
	ref.AddTargetResource(model.NewLocatedResource(DeploymentKind, deployment.Name, deployment.Spec.Template.Labels))
	return true
}

// DeploymentMutator attempts to clone the located Deployment.
func DeploymentMutator(ctx model.SessionContext, ref *model.Ref) error {
	if len(ref.GetResourceStatus(DeploymentKind)) > 0 {
		return nil
	}
	targets := ref.GetTargetsByKind(DeploymentKind)
	if len(targets) == 0 {
		return nil
	}
	target := targets[0]

	deployment, err := getDeployment(ctx, ctx.Namespace, target.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	ctx.Log.Info("Found Deployment", "name", deployment.Name)

	deploymentClone, err := cloneDeployment(deployment.DeepCopy(), ref, ref.GetNewVersion(ctx.Name))
	if err != nil {
		ctx.Log.Info("Failed to clone Deployment", "name", deployment.Name)
		return err
	}
	err = ctx.Client.Create(ctx, deploymentClone)
	if err != nil {
		ctx.Log.Info("Failed to create cloned Deployment", "name", deploymentClone.Name)
		ref.AddResourceStatus(model.ResourceStatus{Kind: DeploymentKind, Name: deploymentClone.Name, Action: model.ActionFailed})
		return err
	}
	ctx.Log.Info("Cloned Deployment", "name", deploymentClone.Name)
	ref.AddResourceStatus(model.ResourceStatus{Kind: DeploymentKind, Name: deploymentClone.Name, Action: model.ActionCreated})
	return nil
}

// DeploymentRevertor attempts to delete the cloned Deployment.
func DeploymentRevertor(ctx model.SessionContext, ref *model.Ref) error {
	statuses := ref.GetResourceStatus(DeploymentKind)
	for _, status := range statuses {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: status.Name, Namespace: ctx.Namespace},
		}
		ctx.Log.Info("Found Deployment", "name", status.Name)
		err := ctx.Client.Delete(ctx, deployment)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			ctx.Log.Info("Failed to delete Deployment", "name", deployment.Name)
			ref.AddResourceStatus(model.ResourceStatus{Kind: DeploymentKind, Name: deployment.Name, Action: model.ActionFailed})
			return err
		}
		ref.RemoveResourceStatus(model.ResourceStatus{Kind: DeploymentKind, Name: deployment.Name})
	}
	return nil
}

func cloneDeployment(deployment *appsv1.Deployment, ref *model.Ref, version string) (*appsv1.Deployment, error) {
	originalDeployment, err := json.Marshal(deployment)
	if err != nil {
		return nil, err
	}

	e := template.NewDefaultEngine()
	modifiedDeployment, err := e.Run(ref.Strategy, originalDeployment, version, ref.Args)
	if err != nil {
		return nil, err
	}

	clone := appsv1.Deployment{}
	err = json.Unmarshal(modifiedDeployment, &clone)
	if err != nil {
		return nil, err
	}
	return &clone, nil
}

func getDeployment(ctx model.SessionContext, namespace, name string) (*appsv1.Deployment, error) {
	deployment := appsv1.Deployment{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &deployment)
	return &deployment, err
}
