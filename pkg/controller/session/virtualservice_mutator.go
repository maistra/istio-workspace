package session

import (
	"strings"

	istionetwork "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/networking/v1alpha3"
	v1alpha3 "istio.io/api/networking/v1alpha3"
)

// VirtualServiceMutator mutates a VirtualService by adding the required routes
type VirtualServiceMutator struct{}

// Add adds a mutation
func (v *VirtualServiceMutator) Add(vs istionetwork.VirtualService) (istionetwork.VirtualService, error) {

	source := vs.Spec.Http[0]

	var sourceRoutes []*v1alpha3.HTTPRouteDestination
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
			&v1alpha3.HTTPMatchRequest{
				Headers: map[string]*v1alpha3.StringMatch{
					"end-user": &v1alpha3.StringMatch{MatchType: &v1alpha3.StringMatch_Exact{Exact: "jason"}},
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

// Remove removes a mutation
func (v *VirtualServiceMutator) Remove(vs istionetwork.VirtualService) (istionetwork.VirtualService, error) {

	for i := 0; i < len(vs.Spec.Http); i++ {
		if strings.Contains(vs.Spec.Http[i].Route[0].Destination.Subset, "-test") {
			vs.Spec.Http = append(vs.Spec.Http[:i], vs.Spec.Http[i+1:]...)
			break
		}
	}
	return vs, nil
}
