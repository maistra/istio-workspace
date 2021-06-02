package istio

import (
	"fmt"
	"strings"

	"emperror.dev/errors"
	"istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/maistra/istio-workspace/pkg/model/new"
	"github.com/maistra/istio-workspace/pkg/reference"
)

const (
	// VirtualServiceKind is the k8s Kind for a istio VirtualService.
	VirtualServiceKind = "VirtualService"

	// LabelIkeMutated is a bool label to indicated we own the resource.
	LabelIkeMutated = "ike.mutated"

	// LabelIkeMutatedValue is the bool value of the LabelIkeMutated label.
	LabelIkeMutatedValue = "true"
)

var (
	_                  new.Locator              = VirtualServiceLocator
	_                  new.ModificatorRegistrar = VirtualServiceRgistrar
	errorRouteNotFound                          = fmt.Errorf("route not found")
)

func VirtualServiceRgistrar() (client.Object, new.Modificator) {
	return &istionetwork.VirtualService{}, VirtualServiceModificator
}

func VirtualServiceLocator(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.LocatorStatusReporter) {
	switch ref.Deleted {
	case false:
		// TODO: expand VirtualService Tests with connected vs where not directly triggering a host route?
		// TODO: Connected GW ignores hostName during find??
		vss, err := getVirtualServices(ctx, ctx.Namespace)
		if err != nil {
			// TODO: report err outside of specific resource?

			return
		}
		targetVersion := new.GetVersion(store)

		for _, hostName := range new.GetTargetHostNames(store) {
			for _, vs := range vss.Items { //nolint:gocritic //reason for readability
				_, connected := connectedToGateway(vs)

				if !connected || vs.Labels[LabelIkeMutated] == LabelIkeMutatedValue {
					continue
				}

				report(new.LocatorStatus{Kind: VirtualServiceKind, Namespace: vs.Namespace, Name: vs.Name, Action: new.ActionCreate, Labels: map[string]string{"host": hostName.String()}})
			}
			for _, vs := range vss.Items { //nolint:gocritic //reason for readability
				if !mutationRequired(vs, hostName, targetVersion) || vsAlreadyMutated(vs, hostName, new.GetNewVersion(store, ctx.Name)) {
					continue
				}

				report(new.LocatorStatus{Kind: VirtualServiceKind, Namespace: vs.Namespace, Name: vs.Name, Action: new.ActionModify, Labels: map[string]string{"host": hostName.String()}})
			}
		}
	case true:
		vss, err := getVirtualServices(ctx, ctx.Namespace, reference.Match(ctx.Name))
		if err != nil {
			// TODO: report err outside of specific resource?

			return
		}

		for _, vs := range vss.Items {
			action := new.Flip(new.StatusAction(reference.GetLabel(&vs, ctx.Name)))
			report(new.LocatorStatus{Kind: VirtualServiceKind, Namespace: vs.Namespace, Name: vs.Name, Action: action})
		}
	}
}

// VirtualServiceModificator attempts to create a virtual service for forked service.
func VirtualServiceModificator(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.ModificatorStatusReporter) {
	for _, resource := range store(VirtualServiceKind) {
		switch resource.Action {
		case new.ActionCreate:
			actionCreateVirtualService(ctx, ref, store, report, resource)
		case new.ActionDelete:
			actionDeleteVirtualService(ctx, ref, store, report, resource)
		case new.ActionModify:
			actionModifyVirtualService(ctx, ref, store, report, resource)
		case new.ActionRevert:
			actionRevertVirtualService(ctx, ref, store, report, resource)
		}
	}
}

func actionCreateVirtualService(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.ModificatorStatusReporter, resource new.LocatorStatus) {
	vs, err := getVirtualService(ctx, resource.Namespace, resource.Name)
	if err != nil {
		report(new.ModificatorStatus{LocatorStatus: resource, Success: false, Error: err})

		return
	}

	hostName := new.ParseHostName(resource.Labels["host"])

	mutatedVs := mutateConnectedVirtualService(ctx, store, ref, hostName, *vs)

	if err = reference.Add(ctx.ToNamespacedName(), &mutatedVs); err != nil {
		ctx.Log.Error(err, "failed to add relation reference", "kind", mutatedVs.Kind, "name", mutatedVs.Name)
	}
	reference.AddLabel(&mutatedVs, ctx.Name, string(resource.Action))

	err = ctx.Client.Create(ctx, &mutatedVs)
	if err != nil && !k8sErrors.IsAlreadyExists(err) {
		report(new.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapIfWithDetails(err, "failed creating virtual service", "kind", VirtualServiceKind, "name", mutatedVs.Name, "host", hostName.String())})

		return
	}
	report(new.ModificatorStatus{LocatorStatus: resource, Success: true})
}

func actionDeleteVirtualService(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.ModificatorStatusReporter, resource new.LocatorStatus) {
	vs := istionetwork.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.Name,
			Namespace: resource.Namespace,
		},
	}

	if err := ctx.Client.Delete(ctx, &vs); err != nil {
		report(new.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed deleting VirtualService", "kind", VirtualServiceKind, "name", vs.Name)})

		return
	}
	report(new.ModificatorStatus{LocatorStatus: resource, Success: true})
}

func actionModifyVirtualService(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.ModificatorStatusReporter, resource new.LocatorStatus) {
	vs, err := getVirtualService(ctx, resource.Namespace, resource.Name)
	if err != nil {
		report(new.ModificatorStatus{LocatorStatus: resource, Success: false, Error: err})

		return
	}

	hostName := new.ParseHostName(resource.Labels["host"])

	mutatedVs, err := mutateVirtualService(ctx, store, ref, hostName, *vs)
	if err != nil {
		report(new.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapIfWithDetails(err, "failed mutating virtual service", "kind", VirtualServiceKind, "name", mutatedVs.Name, "host", hostName.String())})

		return
	}

	if err = reference.Add(ctx.ToNamespacedName(), &mutatedVs); err != nil {
		ctx.Log.Error(err, "failed to add relation reference", "kind", mutatedVs.Kind, "name", mutatedVs.Name)
	}
	reference.AddLabel(&mutatedVs, ctx.Name, string(resource.Action))

	err = ctx.Client.Update(ctx, &mutatedVs)
	if err != nil {
		report(new.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapIfWithDetails(err, "failed updating virtual service", "kind", VirtualServiceKind, "name", mutatedVs.Name, "host", hostName.String())})

		return
	}
	report(new.ModificatorStatus{LocatorStatus: resource, Success: true})
}

func actionRevertVirtualService(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.ModificatorStatusReporter, resource new.LocatorStatus) {
	vs, err := getVirtualService(ctx, resource.Namespace, resource.Name)
	if err != nil {
		report(new.ModificatorStatus{LocatorStatus: resource, Success: false, Error: err})

		return
	}

	mutatedVs := revertVirtualService(new.GetNewVersion(store, ctx.Name), *vs)
	if err = reference.Remove(ctx.ToNamespacedName(), &mutatedVs); err != nil {
		ctx.Log.Error(err, "failed to add relation reference", "kind", mutatedVs.Kind, "name", mutatedVs.Name)
	}

	reference.RemoveLabel(&mutatedVs, ctx.Name)

	err = ctx.Client.Update(ctx, &mutatedVs)
	if err != nil {
		report(new.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed updating VirtualService", "kind", VirtualServiceKind, "name", vs.Name)})

		return
	}
	report(new.ModificatorStatus{LocatorStatus: resource, Success: true})
}

func mutateVirtualService(ctx new.SessionContext, store new.LocatorStatusStore, ref new.Ref, hostName new.HostName, source istionetwork.VirtualService) (istionetwork.VirtualService, error) {
	version := new.GetVersion(store)
	newVersion := new.GetNewVersion(store, ctx.Name)
	target := source.DeepCopy()
	clonedSource := source.DeepCopy()

	targetsHTTP := findRoutes(clonedSource, hostName, version)
	if len(targetsHTTP) == 0 {
		return istionetwork.VirtualService{}, errorRouteNotFound
	}
	for _, tHTTP := range targetsHTTP {
		simplifyTargetRoute(ctx, *tHTTP, hostName, version, newVersion, target)
	}

	return *target, nil
}

func mutateConnectedVirtualService(ctx new.SessionContext, store new.LocatorStatusStore, ref new.Ref, hostName new.HostName, source istionetwork.VirtualService) istionetwork.VirtualService {
	version := new.GetVersion(store)
	newVersion := new.GetNewVersion(store, ctx.Name)
	target := source.DeepCopy()
	clonedSource := source.DeepCopy()
	gateways, _ := connectedToGateway(*target)
	hosts := getHostsFromRef(ctx, store, gateways, ref)

	target.SetName(target.Name + "-" + ctx.Name)
	target.Spec.Hosts = hosts
	target.ResourceVersion = ""
	if target.Labels == nil {
		target.Labels = map[string]string{}
	}
	target.Labels[LabelIkeMutated] = LabelIkeMutatedValue

	targetsHTTP := findRoutes(clonedSource, hostName, version)
	for _, tHTTP := range targetsHTTP {
		simplifyTargetRouteWithoutMatch(*tHTTP, hostName, version, newVersion, target)
	}
	for i := 0; i < len(target.Spec.Http); i++ {
		targetHTTP := addHeaderRequest(*target.Spec.Http[i], ctx.Route)
		target.Spec.Http[i] = &targetHTTP
	}

	return *target
}

func simplifyTargetRouteWithoutMatch(targetHTTP v1alpha3.HTTPRoute, hostName new.HostName, version, newVersion string, target *istionetwork.VirtualService) {
	targetHTTP = removeOtherRoutes(targetHTTP, hostName, version)
	targetHTTP = updateSubset(targetHTTP, newVersion)
	targetHTTP = removeWeight(targetHTTP)
	targetHTTP.Mirror = nil
	targetHTTP.Redirect = nil

	target.Spec.Http = append([]*v1alpha3.HTTPRoute{&targetHTTP}, target.Spec.Http...)
}

func simplifyTargetRoute(ctx new.SessionContext, targetHTTP v1alpha3.HTTPRoute, hostName new.HostName, version, newVersion string, target *istionetwork.VirtualService) {
	targetHTTP = removeOtherRoutes(targetHTTP, hostName, version)
	targetHTTP = updateSubset(targetHTTP, newVersion)
	targetHTTP = addHeaderMatch(targetHTTP, ctx.Route)
	targetHTTP = removeWeight(targetHTTP)
	targetHTTP.Mirror = nil
	targetHTTP.Redirect = nil

	target.Spec.Http = append([]*v1alpha3.HTTPRoute{&targetHTTP}, target.Spec.Http...)
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

func getVirtualService(ctx new.SessionContext, namespace, name string) (*istionetwork.VirtualService, error) {
	virtualService := istionetwork.VirtualService{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &virtualService)

	return &virtualService, errors.WrapWithDetails(err, "failed finding virtual service in namespace", "name", name, "namespace", namespace)
}

func getVirtualServices(ctx new.SessionContext, namespace string, opts ...client.ListOption) (*istionetwork.VirtualServiceList, error) {
	virtualServices := istionetwork.VirtualServiceList{}
	err := ctx.Client.List(ctx, &virtualServices, append(opts, client.InNamespace(namespace))...)

	return &virtualServices, errors.WrapWithDetails(err, "failed finding virtual services in namespace", "namespace", namespace)
}

func mutationRequired(vs istionetwork.VirtualService, targetHost new.HostName, targetVersion string) bool {
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

func vsAlreadyMutated(vs istionetwork.VirtualService, targetHost new.HostName, targetVersion string) bool {
	for _, http := range vs.Spec.Http {
		for _, route := range http.Route {
			if route.Destination != nil && targetHost.Match(route.Destination.Host) && route.Destination.Subset == targetVersion {
				return true
			}
		}
	}

	return false
}

func connectedToGateway(vs istionetwork.VirtualService) ([]string, bool) {
	return vs.Spec.Gateways, len(vs.Spec.Gateways) > 0
}

func findRoutes(vs *istionetwork.VirtualService, host new.HostName, subset string) []*v1alpha3.HTTPRoute {
	var routes []*v1alpha3.HTTPRoute
	for _, h := range vs.Spec.Http {
		for _, r := range h.Route {
			if r.Destination != nil && host.Match(r.Destination.Host) && (r.Destination.Subset == "" || r.Destination.Subset == subset) {
				routes = append(routes, h)
			}
		}
	}

	return routes
}

func removeOtherRoutes(http v1alpha3.HTTPRoute, host new.HostName, subset string) v1alpha3.HTTPRoute {
	for i, r := range http.Route {
		if !((r.Destination != nil && host.Match(r.Destination.Host) && r.Destination.Subset == subset) ||
			(r.Destination != nil && host.Match(r.Destination.Host) && r.Destination.Subset == "")) {
			http.Route = append(http.Route[:i], http.Route[i+1:]...)
		}
	}

	return http
}

func updateSubset(http v1alpha3.HTTPRoute, subset string) v1alpha3.HTTPRoute {
	for _, r := range http.Route {
		r.Destination.Subset = subset
	}

	return http
}

func addHeaderMatch(http v1alpha3.HTTPRoute, route new.Route) v1alpha3.HTTPRoute {
	addHeader := func(m *v1alpha3.HTTPMatchRequest, route new.Route) {
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

func addHeaderRequest(http v1alpha3.HTTPRoute, route new.Route) v1alpha3.HTTPRoute {
	if http.Headers == nil {
		http.Headers = &v1alpha3.Headers{
			Request: &v1alpha3.Headers_HeaderOperations{
				Add: map[string]string{},
			},
		}
	}
	if http.Headers.Request == nil {
		http.Headers.Request = &v1alpha3.Headers_HeaderOperations{
			Add: map[string]string{},
		}
	}
	http.Headers.Request.Add[route.Name] = route.Value

	return http
}

func removeWeight(http v1alpha3.HTTPRoute) v1alpha3.HTTPRoute {
	for _, r := range http.Route {
		r.Weight = 0
	}

	return http
}

func getHostsFromRef(ctx new.SessionContext, store new.LocatorStatusStore, gateways []string, ref new.Ref) []string {
	var hosts []string
	gwByName := func(store new.LocatorStatusStore, gatewayName string) []new.LocatorStatus {
		var f []new.LocatorStatus
		for _, g := range store(GatewayKind) {
			if g.Name == gatewayName {
				f = append(f, g)
			}
		}

		return f
	}
	for _, gateway := range gateways {
		for _, gwTarget := range gwByName(store, gateway) {
			for _, host := range strings.Split(gwTarget.Labels[LabelIkeHosts], ",") {
				hosts = append(hosts, ctx.Name+"."+host)
			}
		}
	}

	return hosts
}
