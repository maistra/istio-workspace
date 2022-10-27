package openshift

import (
	"encoding/json"

	"emperror.dev/errors"
	"github.com/maistra/istio-workspace/api"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/reference"
	"github.com/maistra/istio-workspace/pkg/template"
	appsv1 "github.com/openshift/api/apps/v1"
	errorsK8s "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	api.AddToSchemes = append(api.AddToSchemes, appsv1.Install)
}

const (
	// DeploymentConfigKind is the k8s Kind for a openshift DeploymentConfig.
	DeploymentConfigKind       = "DeploymentConfig"
	deploymentConfigAbbrevKind = "dc"
)

var _ model.Locator = DeploymentConfigLocator

func DeploymentConfigRegistrar(engine template.Engine) model.ModificatorRegistrar {
	return func() (client.Object, model.Modificator) {
		return &appsv1.DeploymentConfig{}, DeploymentConfigModificator(engine)
	}
}

// DeploymentConfigLocator attempts to locate a DeploymentConfig kind based on Ref name.
func DeploymentConfigLocator(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.LocatorStatusReporter) error {
	if !ref.KindName.SupportsKind(DeploymentConfigKind) && !ref.KindName.SupportsKind(deploymentConfigAbbrevKind) {
		return nil
	}

	labelKey := reference.CreateRefMarker(ctx.Name, ref.KindName.String())
	deploymentConfigs, err := getDeploymentConfigs(ctx, ctx.Namespace, reference.RefMarkerMatch(labelKey))
	if err != nil {
		return err
	}

	if !ref.Remove {
		for i := range deploymentConfigs.Items {
			deploymentConfig := deploymentConfigs.Items[i]
			action, hash := reference.GetRefMarker(&deploymentConfig, labelKey)
			if ref.Hash() != hash {
				undo := model.Flip(model.StatusAction(action))
				report(model.LocatorStatus{
					Resource: model.Resource{
						Kind:      DeploymentConfigKind,
						Namespace: deploymentConfig.Namespace,
						Name:      deploymentConfig.Name,
					},
					Labels: deploymentConfig.Spec.Template.Labels,
					Action: undo})
			}
		}

		deployment, err := getDeploymentConfig(ctx, ctx.Namespace, ref.KindName.Name)
		if err != nil {
			if errorsK8s.IsNotFound(err) { // Ref is not a DeploymentConfig type
				return nil
			}

			return errors.WrapIfWithDetails(err, "Could not get DeploymentConfig", "name", deployment.Name, "ref", ref.KindName.String())
		}
		report(model.LocatorStatus{
			Resource: model.Resource{
				Kind:      DeploymentConfigKind,
				Namespace: deployment.Namespace,
				Name:      deployment.Name,
			},
			Labels: deployment.Spec.Template.Labels,
			Action: model.ActionCreate})
	} else {
		for i := range deploymentConfigs.Items {
			deploymentConfig := deploymentConfigs.Items[i]
			action, _ := reference.GetRefMarker(&deploymentConfig, labelKey)
			undo := model.Flip(model.StatusAction(action))
			report(model.LocatorStatus{
				Resource: model.Resource{
					Kind:      DeploymentConfigKind,
					Namespace: deploymentConfig.Namespace,
					Name:      deploymentConfig.Name,
				},
				Labels: deploymentConfig.Spec.Template.Labels,
				Action: undo})
		}
	}

	return nil
}

// DeploymentConfigModificator attempts to clone the located Deployment.
func DeploymentConfigModificator(engine template.Engine) model.Modificator {
	return func(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.ModificatorStatusReporter) {
		for _, resource := range store(DeploymentConfigKind) {
			switch resource.Action {
			case model.ActionCreate:
				actionCreateDeploymentConfig(ctx, ref, store, report, engine, resource)
			case model.ActionDelete:
				actionDeleteDeploymentConfig(ctx, report, resource)
			case model.ActionModify, model.ActionRevert, model.ActionLocated:
				report(model.ModificatorStatus{
					LocatorStatus: resource,
					Success:       false,
					Error:         errors.Errorf("Unknown action type for modificator: %v", resource.Action)})
			}
		}
	}
}

func actionCreateDeploymentConfig(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore,
	report model.ModificatorStatusReporter, engine template.Engine, resource model.LocatorStatus) {
	deployment, err := getDeploymentConfig(ctx, resource.Namespace, resource.Name)
	if err != nil {
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to load target DeploymentConfig", "kind", DeploymentConfigKind, "name", resource.Name)})

		return
	}
	ctx.Log.Info("Found DeploymentConfig", "name", deployment.Name)

	if ref.Strategy == model.StrategyExisting {
		return
	}

	deploymentClone, err := cloneDeployment(engine, deployment.DeepCopy(), ref, model.GetCreatedVersion(store, ctx.Name))
	if err != nil {
		ctx.Log.Info("Failed to clone DeploymentConfig", "name", deployment.Name)
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to cloned DeploymentConfig", "kind", DeploymentConfigKind)})

		return
	}
	if err = reference.Add(ctx.ToNamespacedName(), deploymentClone); err != nil {
		ctx.Log.Error(err, "failed to add relation reference", "kind", deploymentClone.Kind, "name", deploymentClone.Name)
	}
	reference.AddRefMarker(deploymentClone, reference.CreateRefMarker(ctx.Name, ref.KindName.String()), string(resource.Action), ref.Hash())

	if _, err = getDeploymentConfig(ctx, deploymentClone.Namespace, deploymentClone.Name); err == nil {
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       true,
			Target: &model.Resource{
				Namespace: deploymentClone.Namespace,
				Kind:      DeploymentConfigKind,
				Name:      deploymentClone.Name}})

		return
	}

	err = ctx.Client.Create(ctx, deploymentClone)
	if err != nil {
		ctx.Log.Info("Failed to create cloned DeploymentConfig", "name", deploymentClone.Name)
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to create cloned DeploymentConfig", "kind", DeploymentConfigKind, "name", deploymentClone.Name)})

		return
	}

	ctx.Log.Info("Cloned Deployment", "name", deploymentClone.Name)
	report(model.ModificatorStatus{
		LocatorStatus: resource,
		Success:       true,
		Target: &model.Resource{
			Namespace: deploymentClone.Namespace,
			Kind:      DeploymentConfigKind,
			Name:      deploymentClone.Name}})
}

func actionDeleteDeploymentConfig(ctx model.SessionContext, report model.ModificatorStatusReporter, resource model.LocatorStatus) {
	deployment := &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: resource.Name, Namespace: ctx.Namespace},
	}
	ctx.Log.Info("Found DeploymentConfig", "name", resource.Name)
	err := ctx.Client.Delete(ctx, deployment)
	if err != nil {
		if errorsK8s.IsNotFound(err) {
			report(model.ModificatorStatus{LocatorStatus: resource, Success: true})

			return
		}
		ctx.Log.Info("Failed to delete DeploymentConfig", "name", resource.Name)
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to delete DeploymentConfig", "kind", DeploymentConfigKind, "name", resource.Name)})

		return
	}
	report(model.ModificatorStatus{LocatorStatus: resource, Success: true})
}

func cloneDeployment(engine template.Engine, deployment *appsv1.DeploymentConfig, ref model.Ref, version string) (*appsv1.DeploymentConfig, error) {
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

	return &deployment, errors.WrapWithDetails(err, "failed finding DeploymentConfig", "kind", DeploymentConfigKind, "name", name, "namespace", namespace)
}

func getDeploymentConfigs(ctx model.SessionContext, namespace string, opts ...client.ListOption) (*appsv1.DeploymentConfigList, error) {
	deployments := appsv1.DeploymentConfigList{}
	err := ctx.Client.List(ctx, &deployments, append(opts, client.InNamespace(namespace))...)

	return &deployments, errors.WrapWithDetails(err, "failed finding deploymentconfig in namespace", "namespace", namespace)
}
