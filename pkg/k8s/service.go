package k8s

import (
	"emperror.dev/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/maistra/istio-workspace/pkg/model/new"
)

const (
	// ServiceKind is the k8s Kind for a Service.
	ServiceKind = "Service"
)

var _ new.Locator = ServiceLocator

// ServiceLocator attempts to locate the Services for the target Deployment/DeploymentConfig.
func ServiceLocator(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.LocatorStatusReporter) {
	deployments := store("Deployment", "DeploymentConfig")

	services, err := getServices(ctx, ctx.Namespace)
	if err != nil {
		ctx.Log.Error(err, "could not get Services")

		return
	}
	for _, deployment := range deployments {
		for _, service := range services.Items { //nolint:gocritic //reason for readability
			selector := labels.SelectorFromSet(service.Spec.Selector)
			if selector.Matches(labels.Set(deployment.Labels)) {
				report(new.LocatorStatus{
					Namespace: ctx.Namespace,
					Kind:      ServiceKind,
					Name:      service.Name,
					Action:    new.ActionLocated,
					Labels:    service.Labels,
				})
			}
		}
	}

	return
}

func getServices(ctx new.SessionContext, namespace string) (*corev1.ServiceList, error) {
	services := corev1.ServiceList{}
	err := ctx.Client.List(ctx, &services, client.InNamespace(namespace))

	return &services, errors.WrapWithDetails(err, "failed listing services in namespace", "namespace", namespace)
}
