package openshift

import (
	"encoding/json"

	appsv1 "github.com/openshift/api/apps/v1"
	"github.com/pkg/errors"
	errorsK8s "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/maistra/istio-workspace/api"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/reference"
	"github.com/maistra/istio-workspace/pkg/template"
)

func init() {
	api.AddToSchemes = append(api.AddToSchemes, appsv1.Install)
}

const (
	// DeploymentConfigKind is the k8s Kind for a openshift DeploymentConfig.
	DeploymentConfigKind = "DeploymentConfig"
)

var _ model.Locator = DeploymentConfigLocator
var _ model.Revertor = DeploymentConfigRevertor
var _ model.Manipulator = deploymentConfigManipulator{}

// DeploymentConfigManipulator represents a model.Manipulator implementation for handling DeploymentConfig objects.
func DeploymentConfigManipulator(engine template.Engine) model.Manipulator {
	return deploymentConfigManipulator{engine: engine}
}

type deploymentConfigManipulator struct {
	engine template.Engine
}

func (d deploymentConfigManipulator) TargetResourceType() client.Object {
	return &appsv1.DeploymentConfig{}
}
func (d deploymentConfigManipulator) Mutate() model.Mutator {
	return DeploymentConfigMutator(d.engine)
}
func (d deploymentConfigManipulator) Revert() model.Revertor {
	return DeploymentConfigRevertor
}

// DeploymentConfigLocator attempts to locate a DeploymentConfig kind based on Ref name.
func DeploymentConfigLocator(ctx model.SessionContext, ref *model.Ref) bool {
	deployment, err := getDeploymentConfig(ctx, ctx.Namespace, ref.Name)
	if err != nil {
		if errorsK8s.IsNotFound(err) { // Ref is not a DeploymentConfig type
			return false
		}
		ctx.Log.Error(err, "Could not get DeploymentConfig", "name", deployment.Name)

		return false
	}
	ref.AddTargetResource(model.NewLocatedResource(DeploymentConfigKind, deployment.Name, deployment.Spec.Template.Labels))

	return true
}

// DeploymentConfigMutator attempts to clone the located DeploymentConfig.
func DeploymentConfigMutator(engine template.Engine) model.Mutator {
	return func(ctx model.SessionContext, ref *model.Ref) error {
		targets := ref.GetTargets(model.Kind(DeploymentConfigKind))
		if len(targets) == 0 {
			return nil
		}
		target := targets[0]

		deployment, err := getDeploymentConfig(ctx, ctx.Namespace, target.Name)
		if err != nil {
			if errorsK8s.IsNotFound(err) {
				return nil
			}

			return err
		}
		ctx.Log.Info("Found DeploymentConfig", "name", deployment.Name)

		if ref.Strategy == model.StrategyExisting {
			return nil
		}

		deploymentClone, err := cloneDeployment(engine, deployment.DeepCopy(), ref, ref.GetNewVersion(ctx.Name))
		if err != nil {
			ctx.Log.Info("Failed to clone DeploymentConfig", "name", deployment.Name)

			return err
		}
		if err = reference.Add(ctx.ToNamespacedName(), deploymentClone); err != nil {
			ctx.Log.Error(err, "failed to add relation reference", "kind", deploymentClone.Kind, "name", deploymentClone.Name)
		}
		if _, err = getDeploymentConfig(ctx, deploymentClone.Namespace, deploymentClone.Name); err == nil {
			return nil
		}

		err = ctx.Client.Create(ctx, deploymentClone)
		if err != nil {
			ctx.Log.Info("Failed to create cloned DeploymentConfig", "name", deploymentClone.Name)
			ref.AddResourceStatus(model.NewFailedResource(DeploymentConfigKind, deploymentClone.Name, model.ActionCreated, err.Error()))

			return errors.Wrapf(err, "failed to create cloned DeploymentConfig %s", deploymentClone.Name)
		}
		ctx.Log.Info("Cloned DeploymentConfig", "name", deploymentClone.Name)
		ref.AddResourceStatus(model.NewSuccessResource(DeploymentConfigKind, deploymentClone.Name, model.ActionCreated))

		return nil
	}
}

// DeploymentConfigRevertor attempts to delete the cloned DeploymentConfig.
func DeploymentConfigRevertor(ctx model.SessionContext, ref *model.Ref) error {
	statuses := ref.GetResources(model.Kind(DeploymentConfigKind))
	for _, status := range statuses {
		deployment := &appsv1.DeploymentConfig{
			ObjectMeta: metav1.ObjectMeta{Name: status.Name, Namespace: ctx.Namespace},
		}
		ctx.Log.Info("Found DeploymentConfig", "name", status.Name)
		err := ctx.Client.Delete(ctx, deployment)
		if err != nil {
			if errorsK8s.IsNotFound(err) {
				return nil
			}
			ctx.Log.Info("Failed to delete DeploymentConfig", "name", status.Name)
			ref.AddResourceStatus(model.NewFailedResource(DeploymentConfigKind, status.Name, status.Action, err.Error()))

			return errors.Wrapf(err, "failed to delete DeploymentConfig %s", status.Name)
		}
		ref.RemoveResourceStatus(model.NewSuccessResource(DeploymentConfigKind, status.Name, status.Action))
	}

	return nil
}

func cloneDeployment(engine template.Engine, deployment *appsv1.DeploymentConfig, ref *model.Ref, version string) (*appsv1.DeploymentConfig, error) {
	originalDeployment, err := json.Marshal(deployment)
	if err != nil {
		return nil, errors.Wrap(err, "failed reading DeploymentConfig json")
	}

	modifiedDeployment, err := engine.Run(ref.Strategy, originalDeployment, version, ref.Args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to modify DeploymentConfig")
	}

	clone := appsv1.DeploymentConfig{}
	err = json.Unmarshal(modifiedDeployment, &clone)
	if err != nil {
		return nil, errors.Wrap(err, "failed unmarshalling json of modified DeploymentConfig")
	}

	return &clone, nil
}

func getDeploymentConfig(ctx model.SessionContext, namespace, name string) (*appsv1.DeploymentConfig, error) {
	deployment := appsv1.DeploymentConfig{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &deployment)

	return &deployment, errors.Wrapf(err, "failed finding DeploymentConfig %s in namespace %s", name, namespace)
}
