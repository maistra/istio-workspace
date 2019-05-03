package session

import (
	istiov1alpha1 "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/aslakknutsen/istio-workspace/pkg/model"
)

const (
	// DefaultRouteHeaderName holds the name of the Header used to route traffic if no Route is provided
	DefaultRouteHeaderName = "x-workspace-route"

	// RouteStrategyHeader holds the Route Type keyword for a Header based Route strategy
	RouteStrategyHeader = "header"
)

// RefToStatus appends/replaces the Ref in the provided Session.Status.Ref list
func RefToStatus(ref model.Ref, session *istiov1alpha1.Session) {
	statusRef := &istiov1alpha1.RefStatus{Name: ref.Name}
	for _, refStat := range ref.ResourceStatuses {
		rs := refStat
		action := string(rs.Action)
		statusRef.Resources = append(statusRef.Resources, &istiov1alpha1.RefResource{Name: &rs.Name, Kind: &rs.Kind, Action: &action})
	}
	var existsInStatus bool
	for i, statRef := range session.Status.Refs {
		if statRef.Name == statusRef.Name {
			if len(statusRef.Resources) == 0 { // Remove
				session.Status.Refs = append(session.Status.Refs[:i], session.Status.Refs[i+1:]...)
			} else { // Update
				session.Status.Refs[i] = statusRef
			}
			existsInStatus = true
			break
		}
	}
	if !existsInStatus {
		session.Status.Refs = append(session.Status.Refs, statusRef)
	}
}

// StatusesToRef creates a List of Refs based on the Session.Status.Refs list
func StatusesToRef(session istiov1alpha1.Session) []*model.Ref { //nolint[:hugeParam]
	refs := []*model.Ref{}
	for _, statusRef := range session.Status.Refs {
		r := &model.Ref{Name: statusRef.Name}
		StatusToRef(session, r)
		refs = append(refs, r)
	}
	return refs
}

// StatusToRef fills the ResourceStatus of a Ref based on the Session.Status.Refs with the same name
func StatusToRef(session istiov1alpha1.Session, ref *model.Ref) { //nolint[:hugeParam]
	for _, statusRef := range session.Status.Refs {
		if statusRef.Name == ref.Name {
			for _, resource := range statusRef.Resources {
				r := resource
				ref.AddResourceStatus(model.ResourceStatus{Name: *r.Name, Kind: *r.Kind, Action: model.ResourceAction(*r.Action)})
			}
		}
	}
}

// RouteToRoute returns the defined route from the session or the Default
func RouteToRoute(session *istiov1alpha1.Session) model.Route {
	if session.Spec.Route.Type == "" {
		return model.Route{
			Type:  RouteStrategyHeader,
			Name:  DefaultRouteHeaderName,
			Value: session.Name,
		}
	}
	return model.Route{
		Type:  session.Spec.Route.Type,
		Name:  session.Spec.Route.Name,
		Value: session.Spec.Route.Value,
	}
}
