package k8s

import (
	"os"

	"github.com/maistra/istio-workspace/pkg/model"

	appsv1 "k8s.io/api/apps/v1"

	corev1 "k8s.io/api/core/v1"

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

// DeploymentLocator attempts to locate a Deployment kind based on Ref name
func DeploymentLocator(ctx model.SessionContext, ref *model.Ref) bool { //nolint[:hugeParam]
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

// DeploymentMutator attempts to clone the located Deployment
func DeploymentMutator(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	if len(ref.GetResourceStatus(DeploymentKind)) > 0 {
		return nil
	}

	deployment, err := getDeployment(ctx, ctx.Namespace, ref.Name)
	if err != nil {
		return err
	}
	ctx.Log.Info("Found Deployment", "name", ref.Name)

	deploymentClone := cloneDeployment(deployment.DeepCopy(), ctx.Name)
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

// DeploymentRevertor attempts to delete the cloned Deployment
func DeploymentRevertor(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	statuses := ref.GetResourceStatus(DeploymentKind)
	for _, status := range statuses {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: status.Name, Namespace: ctx.Namespace},
		}
		ctx.Log.Info("Found Deployment", "name", ref.Name)
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

func cloneDeployment(deployment *appsv1.Deployment, version string) *appsv1.Deployment {
	deploymentClone := deployment.DeepCopy()
	replicasClone := int32(1)
	labelsClone := deploymentClone.GetLabels()
	labelsClone["telepresence"] = "test"
	labelsClone["version-source"] = labelsClone["version"]
	labelsClone["version"] = version
	deploymentClone.SetName(deployment.GetName() + "-" + version)
	deploymentClone.SetLabels(labelsClone)
	deploymentClone.Spec.Selector.MatchLabels = labelsClone
	deploymentClone.Spec.Template.SetLabels(labelsClone)
	deploymentClone.SetResourceVersion("")
	deploymentClone.Spec.Replicas = &replicasClone

	tpVersion, found := os.LookupEnv("TELEPRESENCE_VERSION")
	if !found {
		tpVersion = "0.99"
	}

	container := deploymentClone.Spec.Template.Spec.Containers[0]
	container.Image = "datawire/telepresence-k8s:" + tpVersion
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

func getDeployment(ctx model.SessionContext, namespace, name string) (*appsv1.Deployment, error) { //nolint[:hugeParam]
	deployment := appsv1.Deployment{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &deployment)
	return &deployment, err
}
