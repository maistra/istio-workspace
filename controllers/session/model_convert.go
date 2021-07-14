package session

import (
	istiov1alpha1 "github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	n "github.com/maistra/istio-workspace/pkg/model/new"
)

const (
	// DefaultRouteHeaderName holds the name of the Header used to route traffic if no Route is provided.
	DefaultRouteHeaderName = "x-workspace-route"

	// RouteStrategyHeader holds the Route Type keyword for a Header based Route strategy.
	RouteStrategyHeader = "header"
)

// ConvertAPIRefToModelRef converts a Session.Spec.Ref to a model.Ref.
func ConvertAPIRefToModelRef(ref istiov1alpha1.Ref, namespace string) n.Ref {
	return n.Ref{KindName: n.ParseRefKindName(ref.Name), Namespace: namespace, Strategy: ref.Strategy, Args: ref.Args}
}

// ConvertModelRouteToAPIRoute returns Model route as a session Route.
func ConvertModelRouteToAPIRoute(route n.Route) *istiov1alpha1.Route {
	return &istiov1alpha1.Route{
		Type:  route.Type,
		Name:  route.Name,
		Value: route.Value,
	}
}

// ConvertAPIRouteToModelRoute returns the defined route from the session or the Default.
func ConvertAPIRouteToModelRoute(session *istiov1alpha1.Session) n.Route {
	if session.Spec.Route.Type == "" {
		return n.Route{
			Type:  RouteStrategyHeader,
			Name:  DefaultRouteHeaderName,
			Value: session.Name,
		}
	}

	return n.Route{
		Type:  session.Spec.Route.Type,
		Name:  session.Spec.Route.Name,
		Value: session.Spec.Route.Value,
	}
}
