package istio

import (
	"github.com/maistra/istio-workspace/pkg/model"

	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// GatewayKind is the k8s Kind for a istio Gateway
	GatewayKind = "Gateway"
)

var _ model.Mutator = GatewayMutator
var _ model.Revertor = GatewayRevertor

// GatewayMutator attempts to expose a external host on the gateway
func GatewayMutator(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	if len(ref.GetResourceStatus(GatewayKind)) > 0 {
		return nil
	}

	gws, err := getGateways(ctx, ctx.Namespace)
	if err != nil {
		return err
	}

	for _, vs := range gws.Items { //nolint[:rangeValCopy]
		mutatedVs, err := mutateGateway(ctx, vs)
		if err != nil {
			ref.AddResourceStatus(model.ResourceStatus{Kind: GatewayKind, Name: vs.Name, Action: model.ActionFailed})
			return err
		}

		err = ctx.Client.Update(ctx, &mutatedVs)
		if err != nil {
			ref.AddResourceStatus(model.ResourceStatus{Kind: GatewayKind, Name: mutatedVs.Name, Action: model.ActionFailed})
			return err
		}

		ref.AddResourceStatus(model.ResourceStatus{Kind: GatewayKind, Name: mutatedVs.Name, Action: model.ActionModified})
	}
	return nil
}

// GatewayRevertor looks at the Ref.ResourceStatus and attempts to revert the state of the mutated objects
func GatewayRevertor(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	resources := ref.GetResourceStatus(GatewayKind)

	for _, resource := range resources {
		gw, err := getGateway(ctx, ctx.Namespace, resource.Name)
		if err != nil {
			if errors.IsNotFound(err) { // Not found, nothing to clean
				break
			}
			ref.AddResourceStatus(model.ResourceStatus{Kind: GatewayKind, Name: resource.Name, Action: model.ActionFailed})
			break
		}

		ctx.Log.Info("Found Gateway", "name", resource.Name)
		mutatedGw, err := revertGateway(ctx, *gw)
		if err != nil {
			ref.AddResourceStatus(model.ResourceStatus{Kind: GatewayKind, Name: resource.Name, Action: model.ActionFailed})
			break
		}
		err = ctx.Client.Update(ctx, &mutatedGw)
		if err != nil {
			ref.AddResourceStatus(model.ResourceStatus{Kind: GatewayKind, Name: resource.Name, Action: model.ActionFailed})
			break
		}
		// ok, removed
		ref.RemoveResourceStatus(model.ResourceStatus{Kind: GatewayKind, Name: resource.Name})
	}

	return nil
}

func mutateGateway(ctx model.SessionContext, source istionetwork.Gateway) (istionetwork.Gateway, error) { //nolint[:hugeParam]
	return source, nil
}

func revertGateway(ctx model.SessionContext, vs istionetwork.Gateway) (istionetwork.Gateway, error) { //nolint[:hugeParam]

	return vs, nil
}

func getGateway(ctx model.SessionContext, namespace, name string) (*istionetwork.Gateway, error) { //nolint[:hugeParam]
	Gateway := istionetwork.Gateway{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &Gateway)
	return &Gateway, err
}

func getGateways(ctx model.SessionContext, namespace string) (*istionetwork.GatewayList, error) { //nolint[:hugeParam]
	gateways := istionetwork.GatewayList{}
	err := ctx.Client.List(ctx, &gateways, client.InNamespace(namespace))
	return &gateways, err
}
