package k8s

import (
	"github.com/maistra/istio-workspace/pkg/model"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ServiceKind is the k8s Kind for a Service
	ServiceKind = "Service"
)

var _ model.Locator = ServiceLocator

// ServiceLocator attempts to locate the Services for the target Deployment/DeploymentConfig.
func ServiceLocator(ctx model.SessionContext, ref *model.Ref) bool {
	deployments := ref.GetTargetsByKind("Deployment", "DeploymentConfig")

	services, err := getServices(ctx, ctx.Namespace)
	if err != nil {
		ctx.Log.Error(err, "Could not get Services")
		return false
	}
	found := false
	for _, deployment := range deployments {
		for _, service := range services.Items { //nolint:gocritic //reason for readability
			selector := labels.SelectorFromSet(service.Spec.Selector)
			if selector.Matches(labels.Set(deployment.Labels)) {
				found = true
				ref.AddTargetResource(model.NewLocatedResource(ServiceKind, service.Name, service.Labels))
			}
		}
	}
	return found
}

func getServices(ctx model.SessionContext, namespace string) (*corev1.ServiceList, error) {
	services := corev1.ServiceList{}
	err := ctx.Client.List(ctx, &services, client.InNamespace(namespace))
	return &services, err
}
