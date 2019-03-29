package istio

import (
	"strings"

	"github.com/aslakknutsen/istio-workspace/pkg/model"

	istionetwork "istio.io/api/pkg/kube/apis/networking/v1alpha3"

	"istio.io/api/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// VirtualServiceKind is the k8 Kind for a istio VirtualService
	VirtualServiceKind = "VirtualService"
)

var _ model.Mutator = VirtualServiceMutator
var _ model.Revertor = VirtualServiceRevertor

func VirtualServiceMutator(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	if len(ref.GetResourceStatus(VirtualServiceKind)) > 0 {
		return nil
	}

	targetName := strings.Split(ref.Name, "-")[0]

	vs, err := getVirtualService2(ctx, ctx.Namespace, targetName)
	if err != nil {
		ref.AddResourceStatus(model.ResourceStatus{Kind: VirtualServiceKind, Name: targetName, Action: model.ActionFailed})
		return err
	}

	ctx.Log.Info("Found VirtualService", "name", targetName)
	mutatedVs, err := mutateVirtualService(*vs)
	if err != nil {
		ref.AddResourceStatus(model.ResourceStatus{Kind: VirtualServiceKind, Name: targetName, Action: model.ActionFailed})
		return err
	}

	err = ctx.Client.Update(ctx, &mutatedVs)
	if err != nil {
		ref.AddResourceStatus(model.ResourceStatus{Kind: VirtualServiceKind, Name: targetName, Action: model.ActionFailed})
		return err
	}

	ref.AddResourceStatus(model.ResourceStatus{Kind: VirtualServiceKind, Name: targetName, Action: model.ActionModified})
	return nil
}

// VirtualServiceRevertor looks at the Ref.ResourceStatus and attempts to revert the state of the mutated objects
func VirtualServiceRevertor(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	resources := ref.GetResourceStatus(VirtualServiceKind)

	for _, resource := range resources {
		vs, err := getVirtualService2(ctx, ctx.Namespace, resource.Name)
		if err != nil {
			if errors.IsNotFound(err) { // Not found, nothing to clean
				break
			}
			ref.AddResourceStatus(model.ResourceStatus{Kind: VirtualServiceKind, Name: resource.Name, Action: model.ActionFailed})
			break
		}

		ctx.Log.Info("Found VirtualService", "name", resource.Name)
		mutatedVs, err := revertVirtualService(*vs)
		if err != nil {
			ref.AddResourceStatus(model.ResourceStatus{Kind: VirtualServiceKind, Name: resource.Name, Action: model.ActionFailed})
			break
		}
		err = ctx.Client.Update(ctx, &mutatedVs)
		if err != nil {
			ref.AddResourceStatus(model.ResourceStatus{Kind: VirtualServiceKind, Name: resource.Name, Action: model.ActionFailed})
			break
		}
		// ok, removed
		ref.RemoveResourceStatus(model.ResourceStatus{Kind: VirtualServiceKind, Name: resource.Name})
	}

	return nil
}

func mutateVirtualService(vs istionetwork.VirtualService) (istionetwork.VirtualService, error) { //nolint[:hugeParam]

	source := vs.Spec.Http[0]

	sourceRoutes := []*v1alpha3.HTTPRouteDestination{}
	for _, r := range source.Route {
		sourceRoute := v1alpha3.HTTPRouteDestination{
			Destination: &v1alpha3.Destination{
				Host:   r.Destination.Host,
				Port:   r.Destination.Port,
				Subset: r.Destination.Subset + "-test",
			},
		}
		sourceRoutes = append(sourceRoutes, &sourceRoute)
	}

	route := &v1alpha3.HTTPRoute{
		Match: []*v1alpha3.HTTPMatchRequest{
			{
				Headers: map[string]*v1alpha3.StringMatch{
					"end-user": {MatchType: &v1alpha3.StringMatch_Exact{Exact: "jason"}},
				},
			},
		},
		Route:                 sourceRoutes,
		Redirect:              source.Redirect,
		AppendHeaders:         source.AppendHeaders,
		AppendResponseHeaders: source.AppendResponseHeaders,
		RemoveRequestHeaders:  source.RemoveRequestHeaders,
		RemoveResponseHeaders: source.RemoveResponseHeaders,
		CorsPolicy:            source.CorsPolicy,
		Fault:                 source.Fault,
		Headers:               source.Headers,
		Mirror:                source.Mirror,
		Retries:               source.Retries,
		Rewrite:               source.Rewrite,
		Timeout:               source.Timeout,
		WebsocketUpgrade:      source.WebsocketUpgrade,
	}
	vs.Spec.Http = append([]*v1alpha3.HTTPRoute{route}, vs.Spec.Http...)

	return vs, nil
}

func revertVirtualService(vs istionetwork.VirtualService) (istionetwork.VirtualService, error) { //nolint[:hugeParam]

	for i := 0; i < len(vs.Spec.Http); i++ {
		if strings.Contains(vs.Spec.Http[i].Route[0].Destination.Subset, "-test") {
			vs.Spec.Http = append(vs.Spec.Http[:i], vs.Spec.Http[i+1:]...)
			break
		}
	}
	return vs, nil
}

func getVirtualService2(ctx model.SessionContext, namespace, name string) (*istionetwork.VirtualService, error) { //nolint[:hugeParam]
	virtualService := istionetwork.VirtualService{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &virtualService)
	return &virtualService, err
}
