package istio

import (
	"fmt"
	"strings"

	"github.com/maistra/istio-workspace/pkg/model"

	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

// VirtualServiceMutator attempts to create a virtual service for forked service.
func VirtualServiceMutator(ctx model.SessionContext, ref *model.Ref) error {
	if len(ref.GetResourceStatus(VirtualServiceKind)) > 0 {
		return nil
	}

	targetVersion := ref.GetVersion()

	for _, hostName := range ref.GetTargetHostNames() {
		vss, err := getVirtualServices(ctx, ctx.Namespace)
		if err != nil {
			ref.AddResourceStatus(model.ResourceStatus{Kind: VirtualServiceKind, Name: hostName.Name, Action: model.ActionFailed})
			return err
		}

		for _, vs := range vss.Items { //nolint:gocritic //reason for readability
			if !mutationRequired(vs, hostName, targetVersion) {
				continue
			}
			ctx.Log.Info("Found VirtualService", "name", hostName)
			mutatedVs, err := mutateVirtualService(ctx, hostName, ref.GetVersion(), ref.GetNewVersion(ctx.Name), vs)
			if err != nil {
				ref.AddResourceStatus(model.ResourceStatus{Kind: VirtualServiceKind, Name: vs.Name, Action: model.ActionFailed})
				return err
			}

			err = ctx.Client.Update(ctx, &mutatedVs)
			if err != nil {
				ref.AddResourceStatus(model.ResourceStatus{Kind: VirtualServiceKind, Name: vs.Name, Action: model.ActionFailed})
				return err
			}

			ref.AddResourceStatus(model.ResourceStatus{Kind: VirtualServiceKind, Name: vs.Name, Action: model.ActionModified})
		}
	}

	return nil
}

// VirtualServiceRevertor looks at the Ref.ResourceStatus and attempts to revert the state of the mutated objects.
func VirtualServiceRevertor(ctx model.SessionContext, ref *model.Ref) error {
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
		mutatedVs := revertVirtualService(ref.GetNewVersion(ctx.Name), *vs)
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

func mutateVirtualService(ctx model.SessionContext, hostName model.HostName, version, newVersion string, source istionetwork.VirtualService) (istionetwork.VirtualService, error) { //nolint:lll,gocyclo //reason for readability
	findRoutes := func(vs *istionetwork.VirtualService, host model.HostName, subset string) []*v1alpha3.HTTPRoute {
		routes := []*v1alpha3.HTTPRoute{}
		for _, h := range vs.Spec.Http {
			for _, r := range h.Route {
				if r.Destination != nil && host.Match(r.Destination.Host) {
					if r.Destination.Subset == "" || r.Destination.Subset == subset {
						routes = append(routes, h)
					}
				}
			}
		}
		return routes
	}

	removeOtherRoutes := func(http v1alpha3.HTTPRoute, host model.HostName, subset string) v1alpha3.HTTPRoute {
		for i, r := range http.Route {
			if !((r.Destination != nil && host.Match(r.Destination.Host) && r.Destination.Subset == subset) ||
				(r.Destination != nil && host.Match(r.Destination.Host) && r.Destination.Subset == "")) {
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
		addHeader := func(m *v1alpha3.HTTPMatchRequest, route model.Route) {
			if route.Type == "header" {
				if m.Headers == nil {
					m.Headers = map[string]*v1alpha3.StringMatch{}
				}
				m.Headers[route.Name] = &v1alpha3.StringMatch{MatchType: &v1alpha3.StringMatch_Exact{Exact: route.Value}}
			}
		}
		if len(http.Match) > 0 {
			for _, m := range http.Match {
				addHeader(m, route)
			}
		} else {
			m := &v1alpha3.HTTPMatchRequest{}
			addHeader(m, route)
			http.Match = append(http.Match, m)
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

	targetsHTTP := findRoutes(clonedSource, hostName, version)
	if len(targetsHTTP) == 0 {
		return istionetwork.VirtualService{}, fmt.Errorf("route not found")
	}
	for _, tHTTP := range targetsHTTP {
		targetHTTP := *tHTTP
		targetHTTP = removeOtherRoutes(targetHTTP, hostName, version)
		targetHTTP = updateSubset(targetHTTP, newVersion)
		targetHTTP = addHeaderMatch(targetHTTP, ctx.Route)
		targetHTTP = removeWeight(targetHTTP)
		targetHTTP.Mirror = nil
		targetHTTP.Redirect = nil

		target.Spec.Http = append([]*v1alpha3.HTTPRoute{&targetHTTP}, target.Spec.Http...)
	}
	return *target, nil
}

func revertVirtualService(subsetName string, vs istionetwork.VirtualService) istionetwork.VirtualService {
	for i := 0; i < len(vs.Spec.Http); i++ {
		http := vs.Spec.Http[i]
		for n := 0; n < len(http.Route); n++ {
			if strings.Contains(http.Route[n].Destination.Subset, subsetName) {
				vs.Spec.Http = append(vs.Spec.Http[:i], vs.Spec.Http[i+1:]...)
				i--
				break
			}
		}
	}
	return vs
}

func getVirtualService(ctx model.SessionContext, namespace, name string) (*istionetwork.VirtualService, error) {
	virtualService := istionetwork.VirtualService{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &virtualService)
	return &virtualService, err
}

func getVirtualServices(ctx model.SessionContext, namespace string) (*istionetwork.VirtualServiceList, error) {
	virtualServices := istionetwork.VirtualServiceList{}
	err := ctx.Client.List(ctx, &virtualServices, client.InNamespace(namespace))
	return &virtualServices, err
}

func mutationRequired(vs istionetwork.VirtualService, targetHost model.HostName, targetVersion string) bool {
	for _, http := range vs.Spec.Http {
		for _, route := range http.Route {
			if route.Destination != nil && targetHost.Match(route.Destination.Host) {
				if route.Destination.Subset == "" || route.Destination.Subset == targetVersion {
					return true
				}
			}
		}
	}
	return false
}
