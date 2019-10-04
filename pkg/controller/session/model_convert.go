package session

import (
	"reflect"

	istiov1alpha1 "github.com/maistra/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/model"
)

const (
	// DefaultRouteHeaderName holds the name of the Header used to route traffic if no Route is provided
	DefaultRouteHeaderName = "x-workspace-route"

	// RouteStrategyHeader holds the Route Type keyword for a Header based Route strategy
	RouteStrategyHeader = "header"
)

// RefToStatus appends/replaces the Ref in the provided Session.Status.Ref list
func RefToStatus(ref model.Ref, session *istiov1alpha1.Session) { //nolint[:hugeParam]
	statusRef := &istiov1alpha1.RefStatus{Ref: istiov1alpha1.Ref{Name: ref.Name, Strategy: ref.Strategy, Args: ref.Args}}
	if ref.Target.Name != "" {
		action := string(ref.Target.Action)
		statusRef.Target = &istiov1alpha1.LabeledRefResource{
			RefResource: istiov1alpha1.RefResource{Kind: &ref.Target.Kind, Name: &ref.Target.Name, Action: &action},
			Labels:      ref.Target.Labels,
		}
	}
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
		r := &model.Ref{Name: statusRef.Name, Strategy: statusRef.Strategy, Args: statusRef.Args}
		StatusToRef(session, r)
		refs = append(refs, r)
	}
	return refs
}

// StatusToRef fills the ResourceStatus of a Ref based on the Session.Status.Refs with the same name
func StatusToRef(session istiov1alpha1.Session, ref *model.Ref) { //nolint[:hugeParam]
	for _, statusRef := range session.Status.Refs {
		if statusRef.Name == ref.Name {
			if statusRef.Target != nil {
				ref.Target = model.LocatedResourceStatus{
					ResourceStatus: model.ResourceStatus{Kind: *statusRef.Target.Kind, Name: *statusRef.Target.Name, Action: model.ResourceAction(*statusRef.Target.Action)},
					Labels:         statusRef.Target.Labels,
				}
			}
			for _, resource := range statusRef.Resources {
				r := resource
				ref.AddResourceStatus(model.ResourceStatus{Name: *r.Name, Kind: *r.Kind, Action: model.ResourceAction(*r.Action)})
			}
		}
	}
}

// RefToRef converts a Session.Spec.Ref to a model.Ref
func RefToRef(ref istiov1alpha1.Ref) model.Ref {
	return model.Ref{Name: ref.Name, Strategy: ref.Strategy, Args: ref.Args}
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

// RefUpdated check if a Ref has been updated compared to current status
func RefUpdated(session istiov1alpha1.Session, ref model.Ref) bool { //nolint[:hugeParam]
	for _, statusRef := range session.Status.Refs {
		if statusRef.Name == ref.Name {
			if statusRef.Strategy != ref.Strategy || !reflect.DeepEqual(statusRef.Args, ref.Args) {
				return true
			}
		}
	}
	return false
}
