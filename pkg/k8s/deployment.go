package k8s

import (
	"encoding/json"

	"emperror.dev/errors"
	appsv1 "k8s.io/api/apps/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/maistra/istio-workspace/pkg/model/new"
	"github.com/maistra/istio-workspace/pkg/reference"
	"github.com/maistra/istio-workspace/pkg/template"
)

const (
	// DeploymentKind is the k8s Kind for a Deployment.
	DeploymentKind = "Deployment"
)

var _ new.Locator = DeploymentLocator

func DeploymentRegistrar(engine template.Engine) new.ModificatorRegistrar {
	return func() (client.Object, new.Modificator) {
		return &appsv1.Deployment{}, DeploymentModificator(engine)
	}
}

// DeploymentLocator attempts to locate a Deployment kind based on Ref name.
func DeploymentLocator(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.LocatorStatusReporter) error {
	if !ref.KindName.SupportsKind(DeploymentKind) {
		return nil
	}

	labelKey := ctx.Name + "-" + ref.KindName.String()
	deployments, err := getDeployments(ctx, ctx.Namespace, reference.Match(labelKey))
	if err != nil {
		return err
	}

	if !ref.Deleted {
		for i := range deployments.Items {
			resource := deployments.Items[i]
			action, hash := reference.GetLabel(&resource, labelKey) // TODO make the name more self-explanatory - label seems to be a way of storing reference marker on a resource e.g. deployment has been created by us
			if ref.Hash() != hash {
				undo := new.Flip(new.StatusAction(action))
				report(new.LocatorStatus{Kind: DeploymentKind, Namespace: resource.Namespace, Name: resource.Name, Labels: resource.Spec.Template.Labels, Action: undo})
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

		report(new.LocatorStatus{Kind: DeploymentKind, Namespace: deployment.Namespace, Name: deployment.Name, Labels: deployment.Spec.Template.Labels, Action: new.ActionCreate})
	} else {
		for i := range deployments.Items {
			deployment := deployments.Items[i]
			action, _ := reference.GetLabel(&deployment, labelKey)
			undo := new.Flip(new.StatusAction(action))
			report(new.LocatorStatus{Kind: DeploymentKind, Namespace: deployment.Namespace, Name: deployment.Name, Labels: deployment.Spec.Template.Labels, Action: undo})
		}
	}

	return nil
}

// DeploymentModificator attempts to clone the located Deployment.
func DeploymentModificator(engine template.Engine) new.Modificator {
	return func(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.ModificatorStatusReporter) {
		for _, resource := range store(DeploymentKind) {
			switch resource.Action {
			case new.ActionCreate:
				actionCreateDeployment(ctx, ref, store, report, engine, resource)
			case new.ActionDelete:
				actionDeleteDeployment(ctx, report, resource)
			case new.ActionModify, new.ActionRevert, new.ActionLocated:
				report(new.ModificatorStatus{LocatorStatus: resource, Success: false, Error: errors.Errorf("Unknown action type for modificator: %v", resource.Action)})
			}
		}
	}
}

func actionCreateDeployment(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore,
	report new.ModificatorStatusReporter, engine template.Engine, resource new.LocatorStatus) {
	deployment, err := getDeployment(ctx, resource.Namespace, resource.Name)
	if err != nil {
		report(new.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to load target Deployment", "kind", DeploymentKind, "name", resource.Name)})

		return
	}
	ctx.Log.Info("Found Deployment", "name", deployment.Name)

	if ref.Strategy == new.StrategyExisting {
		return
	}

	deploymentClone, err := cloneDeployment(engine, deployment.DeepCopy(), ref, new.GetNewVersion(store, ctx.Name))
	if err != nil {
		ctx.Log.Info("Failed to clone Deployment", "name", deployment.Name)
		report(new.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to cloned Deployment", "kind", DeploymentKind)})

		return
	}
	if err = reference.Add(ctx.ToNamespacedName(), deploymentClone); err != nil {
		ctx.Log.Error(err, "failed to add relation reference", "kind", deploymentClone.Kind, "name", deploymentClone.Name)
	}
	reference.AddLabel(deploymentClone, ctx.Name+"-"+ref.KindName.String(), string(resource.Action), ref.Hash())

	if _, err = getDeployment(ctx, deploymentClone.Namespace, deploymentClone.Name); err == nil {
		report(new.ModificatorStatus{LocatorStatus: resource, Success: true})

		return
	}

	err = ctx.Client.Create(ctx, deploymentClone)
	if err != nil {
		ctx.Log.Info("Failed to create cloned Deployment", "name", deploymentClone.Name)
		report(new.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to create cloned Deployment", "kind", DeploymentKind, "name", deploymentClone.Name)})

		return
	}
	ctx.Log.Info("Cloned Deployment", "name", deploymentClone.Name)
	report(new.ModificatorStatus{LocatorStatus: resource, Success: true})
}

func actionDeleteDeployment(ctx new.SessionContext, report new.ModificatorStatusReporter, resource new.LocatorStatus) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: resource.Name, Namespace: ctx.Namespace},
	}
	ctx.Log.Info("Found Deployment", "name", resource.Name)
	err := ctx.Client.Delete(ctx, deployment)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			report(new.ModificatorStatus{LocatorStatus: resource, Success: true})

			return
		}
		ctx.Log.Info("Failed to delete Deployment", "name", resource.Name)
		report(new.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed to delete Deployment", "kind", DeploymentKind, "name", resource.Name)})

		return
	}
	report(new.ModificatorStatus{LocatorStatus: resource, Success: true})
}

func cloneDeployment(engine template.Engine, deployment *appsv1.Deployment, ref new.Ref, version string) (*appsv1.Deployment, error) {
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

func getDeployment(ctx new.SessionContext, namespace, name string) (*appsv1.Deployment, error) {
	deployment := appsv1.Deployment{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &deployment)

	return &deployment, errors.WrapWithDetails(err, "failed finding deployment in namespace ", "kind", DeploymentKind, "name", name, "namespace", namespace)
}

func getDeployments(ctx new.SessionContext, namespace string, opts ...client.ListOption) (*appsv1.DeploymentList, error) {
	deployments := appsv1.DeploymentList{}
	err := ctx.Client.List(ctx, &deployments, append(opts, client.InNamespace(namespace))...)

	return &deployments, errors.WrapWithDetails(err, "failed finding deployments in namespace", "namespace", namespace)
}
