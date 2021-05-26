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

func DeploymentRegistrarCreator(engine template.Engine) new.ModificatorRegistrar {
	return func() (client.Object, new.Modificator) {
		return &appsv1.Deployment{}, DeploymentModificator(engine)
	}
}

// DeploymentLocator attempts to locate a Deployment kind based on Ref name.
func DeploymentLocator(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.LocatorStatusReporter) {
	if !ref.KindName.SupportsKind(DeploymentKind) {
		return
	}

	switch ref.Deleted {
	case false:
		deployment, err := getDeployment(ctx, ctx.Namespace, ref.KindName.Name)
		if err != nil {
			if k8sErrors.IsNotFound(err) { // Ref is not a Deployment type
				return
			}
			ctx.Log.Error(err, "Could not get Deployment", "name", deployment.Name)

			return
		}

		report(new.LocatorStatus{Kind: DeploymentKind, Name: deployment.Name, Labels: deployment.Spec.Template.Labels, Action: new.ActionCreate})
	case true:
	}
}

// DeploymentModificator attempts to clone the located Deployment.
func DeploymentModificator(engine template.Engine) new.Modificator {
	return func(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.ModificatorStatusReporter) {
		targets := store(DeploymentKind)
		if len(targets) == 0 {
			return
		}

		for _, target := range targets {

			switch target.Action {
			case new.ActionCreate:

				deployment, err := getDeployment(ctx, target.Namespace, target.Name)
				if err != nil {
					report(new.ModificatorStatus{
						LocatorStatus: target,
						Success:       false,
						Error:         errors.WrapWithDetails(err, "failed to load target Deployment", "kind", DeploymentKind, "name", target.Name)})

					continue
				}
				ctx.Log.Info("Found Deployment", "name", deployment.Name)

				if ref.Strategy == new.StrategyExisting {
					continue
				}

				deploymentClone, err := cloneDeployment(engine, deployment.DeepCopy(), ref, new.GetNewVersion(store, ctx.Name))
				if err != nil {
					ctx.Log.Info("Failed to clone Deployment", "name", deployment.Name)
					report(new.ModificatorStatus{
						LocatorStatus: target,
						Success:       false,
						Error:         errors.WrapWithDetails(err, "failed to cloned Deployment", "kind", DeploymentKind, "name", deploymentClone.Name)})

					continue
				}
				if err = reference.Add(ctx.ToNamespacedName(), deploymentClone); err != nil {
					ctx.Log.Error(err, "failed to add relation reference", "kind", deploymentClone.Kind, "name", deploymentClone.Name)
				}
				if _, err = getDeployment(ctx, deploymentClone.Namespace, deploymentClone.Name); err == nil {
					report(new.ModificatorStatus{LocatorStatus: target, Success: true})
					continue
				}

				err = ctx.Client.Create(ctx, deploymentClone)
				if err != nil {
					ctx.Log.Info("Failed to create cloned Deployment", "name", deploymentClone.Name)
					report(new.ModificatorStatus{
						LocatorStatus: target,
						Success:       false,
						Error:         errors.WrapWithDetails(err, "failed to create cloned Deployment", "kind", DeploymentKind, "name", deploymentClone.Name)})
					continue
				}
				ctx.Log.Info("Cloned Deployment", "name", deploymentClone.Name)
				report(new.ModificatorStatus{LocatorStatus: target, Success: true})

			case new.ActionDelete:
				deployment := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: target.Name, Namespace: ctx.Namespace},
				}
				ctx.Log.Info("Found Deployment", "name", target.Name)
				err := ctx.Client.Delete(ctx, deployment)
				if err != nil {
					if k8sErrors.IsNotFound(err) {
						report(new.ModificatorStatus{LocatorStatus: target, Success: true})
						continue
					}
					ctx.Log.Info("Failed to delete Deployment", "name", target.Name)
					report(new.ModificatorStatus{
						LocatorStatus: target,
						Success:       false,
						Error:         errors.WrapWithDetails(err, "failed to delete Deployment", "kind", DeploymentKind, "name", target.Name)})

					continue
				}
				report(new.ModificatorStatus{LocatorStatus: target, Success: true})

			case new.ActionModify:
			case new.ActionRevert:
			}
		}

	}
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
