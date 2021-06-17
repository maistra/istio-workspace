package openshift

import (
	"encoding/json"

	"emperror.dev/errors"
	appsv1 "github.com/openshift/api/apps/v1"
	errorsK8s "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/maistra/istio-workspace/api"
	"github.com/maistra/istio-workspace/pkg/model/new"
	"github.com/maistra/istio-workspace/pkg/reference"
	"github.com/maistra/istio-workspace/pkg/template"
)

func init() {
	api.AddToSchemes = append(api.AddToSchemes, appsv1.Install)
}

const (
	// DeploymentConfigKind is the k8s Kind for a openshift DeploymentConfig.
	DeploymentConfigKind       = "DeploymentConfig"
	deploymentConfigAbbrevKind = "dc"
)

var _ new.Locator = DeploymentConfigLocator

func DeploymentConfigRegistrar(engine template.Engine) new.ModificatorRegistrar {
	return func() (client.Object, new.Modificator) {
		return &appsv1.DeploymentConfig{}, DeploymentConfigModificator(engine)
	}
}

// DeploymentConfigLocator attempts to locate a DeploymentConfig kind based on Ref name.
func DeploymentConfigLocator(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.LocatorStatusReporter) error {
	if !ref.KindName.SupportsKind(DeploymentConfigKind) && !ref.KindName.SupportsKind(deploymentConfigAbbrevKind) {
		return nil
	}

	labelKey := ctx.Name + "-" + ref.KindName.String()
	deploymentConfigs, err := getDeploymentConfigs(ctx, ctx.Namespace, reference.Match(labelKey))
	if err != nil {
		return err
	}

	if !ref.Deleted {
		for i := range deploymentConfigs.Items {
			deploymentConfig := deploymentConfigs.Items[i]
			action, hash := reference.GetLabel(&deploymentConfig, labelKey)
			if ref.Hash() != hash {
				undo := new.Flip(new.StatusAction(action))
				report(new.LocatorStatus{Kind: DeploymentConfigKind, Namespace: deploymentConfig.Namespace, Name: deploymentConfig.Name, Labels: deploymentConfig.Spec.Template.Labels, Action: undo})
			}
		}

		deployment, err := getDeploymentConfig(ctx, ctx.Namespace, ref.KindName.Name)
		if err != nil {
			if errorsK8s.IsNotFound(err) { // Ref is not a DeploymentConfig type
				return nil
			}

			return errors.WrapIfWithDetails(err, "Could not get DeploymentConfig", "name", deployment.Name, "ref", ref.KindName.String())
		}
		report(new.LocatorStatus{Kind: DeploymentConfigKind, Namespace: deployment.Namespace, Name: deployment.Name, Labels: deployment.Spec.Template.Labels, Action: new.ActionCreate})
	} else {
		for i := range deploymentConfigs.Items {
			deploymentConfig := deploymentConfigs.Items[i]
			action, _ := reference.GetLabel(&deploymentConfig, labelKey)
			undo := new.Flip(new.StatusAction(action))
			report(new.LocatorStatus{Kind: DeploymentConfigKind, Namespace: deploymentConfig.Namespace, Name: deploymentConfig.Name, Labels: deploymentConfig.Spec.Template.Labels, Action: undo})
		}
	}

	return nil
}

// DeploymentConfigModificator attempts to clone the located Deployment.
func DeploymentConfigModificator(engine template.Engine) new.Modificator {
	return func(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.ModificatorStatusReporter) {
		for _, resource := range store(DeploymentConfigKind) {
			switch resource.Action {
			case new.ActionCreate:
				actionCreateDeploymentConfig(ctx, ref, store, report, engine, resource)
			case new.ActionDelete:
				actionDeleteDeploymentConfig(ctx, report, resource)
			case new.ActionModify, new.ActionRevert, new.ActionLocated:
				report(new.ModificatorStatus{LocatorStatus: resource, Success: false, Error: errors.Errorf("Unknown action type for modificator: %v", resource.Action)})
			}
		}
	}
}

func actionCreateDeploymentConfig(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore,
	report new.ModificatorStatusReporter, engine template.Engine, resource new.LocatorStatus) {
	deployment, err := getDeploymentConfig(ctx, resource.Namespace, resource.Name)
	if err != nil {
		report(new.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to load target DeploymentConfig", "kind", DeploymentConfigKind, "name", resource.Name)})

		return
	}
	ctx.Log.Info("Found DeploymentConfig", "name", deployment.Name)

	if ref.Strategy == new.StrategyExisting {
		return
	}

	deploymentClone, err := cloneDeployment(engine, deployment.DeepCopy(), ref, new.GetNewVersion(store, ctx.Name))
	if err != nil {
		ctx.Log.Info("Failed to clone DeploymentConfig", "name", deployment.Name)
		report(new.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to cloned DeploymentConfig", "kind", DeploymentConfigKind)})

		return
	}
	if err = reference.Add(ctx.ToNamespacedName(), deploymentClone); err != nil {
		ctx.Log.Error(err, "failed to add relation reference", "kind", deploymentClone.Kind, "name", deploymentClone.Name)
	}
	reference.AddLabel(deploymentClone, ctx.Name+"-"+ref.KindName.String(), string(resource.Action), ref.Hash())

	if _, err = getDeploymentConfig(ctx, deploymentClone.Namespace, deploymentClone.Name); err == nil {
		report(new.ModificatorStatus{LocatorStatus: resource, Success: true})

		return
	}

	err = ctx.Client.Create(ctx, deploymentClone)
	if err != nil {
		ctx.Log.Info("Failed to create cloned DeploymentConfig", "name", deploymentClone.Name)
		report(new.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to create cloned DeploymentConfig", "kind", DeploymentConfigKind, "name", deploymentClone.Name)})

		return
	}
	ctx.Log.Info("Cloned Deployment", "name", deploymentClone.Name)
	report(new.ModificatorStatus{LocatorStatus: resource, Success: true})
}

func actionDeleteDeploymentConfig(ctx new.SessionContext, report new.ModificatorStatusReporter, resource new.LocatorStatus) {
	deployment := &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: resource.Name, Namespace: ctx.Namespace},
	}
	ctx.Log.Info("Found DeploymentConfig", "name", resource.Name)
	err := ctx.Client.Delete(ctx, deployment)
	if err != nil {
		if errorsK8s.IsNotFound(err) {
			report(new.ModificatorStatus{LocatorStatus: resource, Success: true})

			return
		}
		ctx.Log.Info("Failed to delete DeploymentConfig", "name", resource.Name)
		report(new.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to delete DeploymentConfig", "kind", DeploymentConfigKind, "name", resource.Name)})

		return
	}
	report(new.ModificatorStatus{LocatorStatus: resource, Success: true})
}

func cloneDeployment(engine template.Engine, deployment *appsv1.DeploymentConfig, ref new.Ref, version string) (*appsv1.DeploymentConfig, error) {
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

func getDeploymentConfig(ctx new.SessionContext, namespace, name string) (*appsv1.DeploymentConfig, error) {
	deployment := appsv1.DeploymentConfig{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &deployment)

	return &deployment, errors.WrapWithDetails(err, "failed finding DeploymentConfig", "kind", DeploymentConfigKind, "name", name, "namespace", namespace)
}

func getDeploymentConfigs(ctx new.SessionContext, namespace string, opts ...client.ListOption) (*appsv1.DeploymentConfigList, error) {
	deployments := appsv1.DeploymentConfigList{}
	err := ctx.Client.List(ctx, &deployments, append(opts, client.InNamespace(namespace))...)

	return &deployments, errors.WrapWithDetails(err, "failed finding deploymentconfig in namespace", "namespace", namespace)
}
