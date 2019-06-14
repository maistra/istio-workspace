package istio

import (
	"fmt"
	"strings"

	"github.com/maistra/istio-workspace/pkg/model"

	istionetwork "istio.io/api/pkg/kube/apis/networking/v1alpha3"

	"istio.io/api/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// VirtualServiceKind is the k8s Kind for a istio VirtualService
	VirtualServiceKind = "VirtualService"
)

var _ model.Mutator = VirtualServiceMutator
var _ model.Revertor = VirtualServiceRevertor

// VirtualServiceMutator attempts to create a virtual service for forked service
func VirtualServiceMutator(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	if len(ref.GetResourceStatus(VirtualServiceKind)) > 0 {
		return nil
	}

	targetName := ref.Target.GetHostName()

	vs, err := getVirtualService(ctx, ctx.Namespace, targetName)
	if err != nil {
		ref.AddResourceStatus(model.ResourceStatus{Kind: VirtualServiceKind, Name: targetName, Action: model.ActionFailed})
		return err
	}

	ctx.Log.Info("Found VirtualService", "name", targetName)
	mutatedVs, err := mutateVirtualService(ctx, ref.Target, *vs)
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
		vs, err := getVirtualService(ctx, ctx.Namespace, resource.Name)
		if err != nil {
			if errors.IsNotFound(err) { // Not found, nothing to clean
				break
			}
			ref.AddResourceStatus(model.ResourceStatus{Kind: VirtualServiceKind, Name: resource.Name, Action: model.ActionFailed})
			break
		}

		ctx.Log.Info("Found VirtualService", "name", resource.Name)
		mutatedVs, err := revertVirtualService(ctx, ref.Target.GetNewVersion(ctx.Name), *vs)
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

func mutateVirtualService(ctx model.SessionContext, sourceResource model.LocatedResourceStatus, source istionetwork.VirtualService) (istionetwork.VirtualService, error) { //nolint[:hugeParam]

	findRoute := func(vs *istionetwork.VirtualService, host, subset string) (v1alpha3.HTTPRoute, bool) {
		for _, h := range vs.Spec.Http {
			for _, r := range h.Route {
				if r.Destination.Host == host && r.Destination.Subset == subset {
					return *h, true
				}
			}
		}
		return v1alpha3.HTTPRoute{}, false
	}
	removeOtherRoutes := func(http v1alpha3.HTTPRoute, host, subset string) v1alpha3.HTTPRoute {
		for i, r := range http.Route {
			if !(r.Destination.Host == host && r.Destination.Subset == subset) {
				http.Route = append(http.Route[:i], http.Route[i+1:]...)
			}
		}
		return http
	}
	updateSubset := func(http v1alpha3.HTTPRoute, subset string) v1alpha3.HTTPRoute {
		for _, r := range http.Route {
			r.Destination.Subset = subset
		}
		return http
	}
	addHeaderMatch := func(http v1alpha3.HTTPRoute, route model.Route) v1alpha3.HTTPRoute {
		for _, m := range http.Match {
			if route.Type == "header" {
				if m.Headers == nil {
					m.Headers = map[string]*v1alpha3.StringMatch{}
				}
				m.Headers[route.Name] = &v1alpha3.StringMatch{MatchType: &v1alpha3.StringMatch_Exact{Exact: route.Value}}
			}
		}
		return http
	}
	removeWeight := func(http v1alpha3.HTTPRoute) v1alpha3.HTTPRoute {
		for _, r := range http.Route {
			r.Weight = 0
		}
		return http
	}

	target := source.DeepCopy()
	clonedSource := source.DeepCopy()

	targetHTTP, found := findRoute(clonedSource, sourceResource.GetHostName(), sourceResource.GetVersion())
	if !found {
		return istionetwork.VirtualService{}, fmt.Errorf("route not found")
	}
	targetHTTP = removeOtherRoutes(targetHTTP, sourceResource.GetHostName(), sourceResource.GetVersion())
	targetHTTP = updateSubset(targetHTTP, sourceResource.GetNewVersion(ctx.Name))
	targetHTTP = addHeaderMatch(targetHTTP, ctx.Route)
	targetHTTP = removeWeight(targetHTTP)
	targetHTTP.Mirror = nil
	targetHTTP.Redirect = nil

	target.Spec.Http = append(target.Spec.Http, &targetHTTP)
	return *target, nil
}

func revertVirtualService(ctx model.SessionContext, subsetName string, vs istionetwork.VirtualService) (istionetwork.VirtualService, error) { //nolint[:hugeParam]

	for i := 0; i < len(vs.Spec.Http); i++ {
		if strings.Contains(vs.Spec.Http[i].Route[0].Destination.Subset, subsetName) {
			vs.Spec.Http = append(vs.Spec.Http[:i], vs.Spec.Http[i+1:]...)
			break
		}
	}
	return vs, nil
}

func getVirtualService(ctx model.SessionContext, namespace, name string) (*istionetwork.VirtualService, error) { //nolint[:hugeParam]
	virtualService := istionetwork.VirtualService{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &virtualService)
	return &virtualService, err
}
