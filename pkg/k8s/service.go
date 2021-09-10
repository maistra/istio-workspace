package k8s

import (
	"emperror.dev/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/maistra/istio-workspace/pkg/model"
)

const (
	// ServiceKind is the k8s Kind for a Service.
	ServiceKind = "Service"
)

var (
	_ model.Locator              = ServiceLocator
	_ model.ModificatorRegistrar = ServiceRegistrar
)

func ServiceRegistrar() (client.Object, model.Modificator) {
	return &corev1.Service{}, ServiceModificator
}

// ServiceLocator attempts to locate the Services for the target Deployment/DeploymentConfig.
func ServiceLocator(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.LocatorStatusReporter) error {
	deployments := store("Deployment", "DeploymentConfig")

	services, err := getServices(ctx, ctx.Namespace)
	if err != nil {
		ctx.Log.Error(err, "could not get Services")

		return err
	}
	for _, deployment := range deployments {
		for _, service := range services.Items { //nolint:gocritic //reason for readability
			selector := labels.SelectorFromSet(service.Spec.Selector)
			if selector.Matches(labels.Set(deployment.Labels)) {
				report(model.LocatorStatus{
					Resource: model.Resource{
						Namespace: ctx.Namespace,
						Kind:      ServiceKind,
						Name:      service.Name,
					},
					Action: model.ActionLocated,
					Labels: service.Labels,
				})
			}
		}
	}

	return nil
}

// ServiceModificator will set a located service to modification status true.
func ServiceModificator(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.ModificatorStatusReporter) {
	for _, resource := range store(ServiceKind) {
		switch resource.Action {
		case model.ActionLocated:
			actionLocatedService(report, resource)
		case model.ActionModify, model.ActionRevert, model.ActionCreate, model.ActionDelete:
			report(model.ModificatorStatus{
				LocatorStatus: resource,
				Success:       false,
				Error:         errors.Errorf("Unknown action type for modificator: %v", resource.Action)})
		}
	}
}

func actionLocatedService(report model.ModificatorStatusReporter, resource model.LocatorStatus) {
	report(model.ModificatorStatus{
		LocatorStatus: resource,
		Success:       true})
}

func getServices(ctx model.SessionContext, namespace string) (*corev1.ServiceList, error) {
	services := corev1.ServiceList{}
	err := ctx.Client.List(ctx, &services, client.InNamespace(namespace))

	return &services, errors.WrapWithDetails(err, "failed listing services in namespace", "namespace", namespace)
}
