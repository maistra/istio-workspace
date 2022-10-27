package k8s

import (
	"encoding/json"

	"emperror.dev/errors"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/reference"
	"github.com/maistra/istio-workspace/pkg/template"
	appsv1 "k8s.io/api/apps/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// DeploymentKind is the k8s Kind for a Deployment.
	DeploymentKind = "Deployment"
)

var _ model.Locator = DeploymentLocator

func DeploymentRegistrar(engine template.Engine) model.ModificatorRegistrar {
	return func() (client.Object, model.Modificator) {
		return &appsv1.Deployment{}, DeploymentModificator(engine)
	}
}

// DeploymentLocator attempts to locate a Deployment kind based on Ref name.
func DeploymentLocator(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.LocatorStatusReporter) error {
	if !ref.KindName.SupportsKind(DeploymentKind) {
		return nil
	}

	labelKey := reference.CreateRefMarker(ctx.Name, ref.KindName.String())
	deployments, err := getDeployments(ctx, ctx.Namespace, reference.RefMarkerMatch(labelKey))
	if err != nil {
		return err
	}

	if !ref.Remove {
		for i := range deployments.Items {
			resource := deployments.Items[i]
			action, hash := reference.GetRefMarker(&resource, labelKey)
			if ref.Hash() != hash {
				undo := model.Flip(model.StatusAction(action))
				report(model.LocatorStatus{
					Resource: model.Resource{
						Kind:      DeploymentKind,
						Namespace: resource.Namespace,
						Name:      resource.Name,
					},
					Labels: resource.Spec.Template.Labels,
					Action: undo})
			}
		}
		deployment, err := getDeployment(ctx, ref.Namespace, ref.KindName.Name)
		if err != nil {
			if k8sErrors.IsNotFound(err) { // Ref is not a Deployment type
				return nil
			}
			ctx.Log.Error(err, "Could not get Deployment", "name", deployment.Name)

			return err
		}

		report(model.LocatorStatus{
			Resource: model.Resource{
				Kind:      DeploymentKind,
				Namespace: deployment.Namespace,
				Name:      deployment.Name,
			},
			Labels: deployment.Spec.Template.Labels,
			Action: model.ActionCreate})
	} else {
		for i := range deployments.Items {
			deployment := deployments.Items[i]
			action, _ := reference.GetRefMarker(&deployment, labelKey)
			undo := model.Flip(model.StatusAction(action))
			report(model.LocatorStatus{
				Resource: model.Resource{
					Kind:      DeploymentKind,
					Namespace: deployment.Namespace,
					Name:      deployment.Name,
				},
				Labels: deployment.Spec.Template.Labels,
				Action: undo})
		}
	}

	return nil
}

// DeploymentModificator attempts to clone the located Deployment.
func DeploymentModificator(engine template.Engine) model.Modificator {
	return func(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.ModificatorStatusReporter) {
		for _, resource := range store(DeploymentKind) {
			switch resource.Action {
			case model.ActionCreate:
				actionCreateDeployment(ctx, ref, store, report, engine, resource)
			case model.ActionDelete:
				actionDeleteDeployment(ctx, report, resource)
			case model.ActionModify, model.ActionRevert, model.ActionLocated:
				report(model.ModificatorStatus{
					LocatorStatus: resource,
					Success:       false,
					Error:         errors.Errorf("Unknown action type for modificator: %v", resource.Action)})
			}
		}
	}
}

func actionCreateDeployment(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore,
	report model.ModificatorStatusReporter, engine template.Engine, resource model.LocatorStatus) {
	deployment, err := getDeployment(ctx, resource.Namespace, resource.Name)
	if err != nil {
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to load target Deployment", "kind", DeploymentKind, "name", resource.Name)})

		return
	}
	ctx.Log.Info("Found Deployment", "name", deployment.Name)

	if ref.Strategy == model.StrategyExisting {
		return
	}

	deploymentClone, err := cloneDeployment(engine, deployment.DeepCopy(), ref, model.GetCreatedVersion(store, ctx.Name))
	if err != nil {
		ctx.Log.Info("Failed to clone Deployment", "name", deployment.Name)
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to cloned Deployment", "kind", DeploymentKind)})

		return
	}
	if err = reference.Add(ctx.ToNamespacedName(), deploymentClone); err != nil {
		ctx.Log.Error(err, "failed to add relation reference", "kind", deploymentClone.Kind, "name", deploymentClone.Name)
	}
	reference.AddRefMarker(deploymentClone, reference.CreateRefMarker(ctx.Name, ref.KindName.String()), string(resource.Action), ref.Hash())

	if _, err = getDeployment(ctx, deploymentClone.Namespace, deploymentClone.Name); err == nil {
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       true,
			Target: &model.Resource{
				Namespace: deploymentClone.Namespace,
				Kind:      DeploymentKind,
				Name:      deploymentClone.Name}})

		return
	}

	err = ctx.Client.Create(ctx, deploymentClone)
	if err != nil {
		ctx.Log.Info("Failed to create cloned Deployment", "name", deploymentClone.Name)
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to create cloned Deployment", "kind", DeploymentKind, "name", deploymentClone.Name)})

		return
	}

	ctx.Log.Info("Cloned Deployment", "name", deploymentClone.Name)
	report(model.ModificatorStatus{
		LocatorStatus: resource,
		Success:       true,
		Target: &model.Resource{
			Namespace: deploymentClone.Namespace,
			Kind:      DeploymentKind,
			Name:      deploymentClone.Name}})
}

func actionDeleteDeployment(ctx model.SessionContext, report model.ModificatorStatusReporter, resource model.LocatorStatus) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: resource.Name, Namespace: ctx.Namespace},
	}
	ctx.Log.Info("Found Deployment", "name", resource.Name)
	err := ctx.Client.Delete(ctx, deployment)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			report(model.ModificatorStatus{LocatorStatus: resource, Success: true})

			return
		}
		ctx.Log.Info("Failed to delete Deployment", "name", resource.Name)
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to delete Deployment", "kind", DeploymentKind, "name", resource.Name)})

		return
	}
	report(model.ModificatorStatus{LocatorStatus: resource, Success: true})
}

func cloneDeployment(engine template.Engine, deployment *appsv1.Deployment, ref model.Ref, version string) (*appsv1.Deployment, error) {
	originalDeployment, err := json.Marshal(deployment)
	if err != nil {
		return nil, errors.Wrap(err, "failed reading deployment json")
	}

	modifiedDeployment, err := engine.Run(ref.Strategy, originalDeployment, version, ref.Args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to modify deployment")
	}

	clone := appsv1.Deployment{}
	err = json.Unmarshal(modifiedDeployment, &clone)
	if err != nil {
		return nil, errors.Wrap(err, "failed unmarshalling json of modified deployment")
	}

	return &clone, nil
}

func getDeployment(ctx model.SessionContext, namespace, name string) (*appsv1.Deployment, error) {
	deployment := appsv1.Deployment{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &deployment)

	return &deployment, errors.WrapWithDetails(err, "failed finding deployment in namespace ", "kind", DeploymentKind, "name", name, "namespace", namespace)
}

func getDeployments(ctx model.SessionContext, namespace string, opts ...client.ListOption) (*appsv1.DeploymentList, error) {
	deployments := appsv1.DeploymentList{}
	err := ctx.Client.List(ctx, &deployments, append(opts, client.InNamespace(namespace))...)

	return &deployments, errors.WrapWithDetails(err, "failed finding deployments in namespace", "namespace", namespace)
}
