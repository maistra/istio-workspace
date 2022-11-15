package istio

import (
	"fmt"
	"strings"

	"emperror.dev/errors"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/reference"
	"istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	_                model.Locator              = VirtualServiceLocator
	_                model.ModificatorRegistrar = VirtualServiceRegistrar
	errRouteNotFound                            = fmt.Errorf("route not found")
)

func VirtualServiceRegistrar() (client.Object, model.Modificator) {
	return &istionetwork.VirtualService{}, VirtualServiceModificator
}

func VirtualServiceLocator(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.LocatorStatusReporter) error {
	labelKey := reference.CreateRefMarker(ctx.Name, ref.KindName.String())
	vss, err := getVirtualServices(ctx, ctx.Namespace, reference.RefMarkerMatch(labelKey))
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to get all virtual services", "ref", ref.KindName.String())
	}

	if !ref.Remove {
		for i := range vss.Items {
			vs := vss.Items[i]
			action, hash := reference.GetRefMarker(&vs, labelKey)
			undo := model.Flip(model.StatusAction(action))
			if ref.Hash() != hash {
				report(model.LocatorStatus{
					Resource: model.Resource{
						Kind:      VirtualServiceKind,
						Namespace: vs.Namespace,
						Name:      vs.Name,
					},
					Action: undo})
			}
		}

		virtualServices, err := getVirtualServices(ctx, ctx.Namespace)
		if err != nil {
			return err
		}
		targetVersion := model.GetVersion(store)

		for _, hostName := range model.GetTargetHostNames(store) {
			reportVsToBeCreated(virtualServices, hostName, report)
			reportVsToBeModified(virtualServices, hostName, targetVersion, report)
		}
	} else {
		for i := range vss.Items {
			vs := vss.Items[i]
			action, _ := reference.GetRefMarker(&vs, labelKey)
			undo := model.Flip(model.StatusAction(action))
			report(model.LocatorStatus{
				Resource: model.Resource{
					Kind:      VirtualServiceKind,
					Namespace: vs.Namespace,
					Name:      vs.Name,
				},
				Action: undo})
		}
	}

	return nil
}

func reportVsToBeCreated(vss *istionetwork.VirtualServiceList, hostName model.HostName, report model.LocatorStatusReporter) {
	for i := range vss.Items {
		vs := vss.Items[i]
		_, connected := connectedToGateway(vs)

		if !connected || vs.Labels[LabelIkeMutated] == LabelIkeMutatedValue {
			continue
		}

		report(model.LocatorStatus{
			Resource: model.Resource{
				Kind:      VirtualServiceKind,
				Namespace: vs.Namespace,
				Name:      vs.Name,
			},
			Action: model.ActionCreate,
			Labels: map[string]string{"host": hostName.String()}})
	}
}

func reportVsToBeModified(vss *istionetwork.VirtualServiceList, hostName model.HostName, targetVersion string, report model.LocatorStatusReporter) {
	for i := range vss.Items {
		vs := vss.Items[i]
		if !mutationRequired(vs, hostName, targetVersion) {
			continue
		}

		report(model.LocatorStatus{
			Resource: model.Resource{
				Kind:      VirtualServiceKind,
				Namespace: vs.Namespace,
				Name:      vs.Name,
			},
			Action: model.ActionModify,
			Labels: map[string]string{"host": hostName.String()}})
	}
}

// VirtualServiceModificator attempts to create a virtual service for forked service.
func VirtualServiceModificator(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.ModificatorStatusReporter) {
	for _, resource := range store(VirtualServiceKind) {
		switch resource.Action {
		case model.ActionCreate:
			actionCreateVirtualService(ctx, ref, store, report, resource)
		case model.ActionDelete:
			actionDeleteVirtualService(ctx, report, resource)
		case model.ActionModify:
			actionModifyVirtualService(ctx, ref, store, report, resource)
		case model.ActionRevert:
			actionRevertVirtualService(ctx, ref, store, report, resource)
		case model.ActionLocated:
			report(model.ModificatorStatus{
				LocatorStatus: resource,
				Success:       false,
				Error:         errors.Errorf("Unknown action type for modificator: %v", resource.Action)})
		}
	}
}

func actionCreateVirtualService(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.ModificatorStatusReporter, resource model.LocatorStatus) {
	vs, err := getVirtualService(ctx, resource.Namespace, resource.Name)
	if err != nil {
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         err})

		return
	}

	hostName := model.NewHostName(resource.Labels["host"])

	mutatedVs := mutateConnectedVirtualService(ctx, store, hostName, *vs)

	if err = reference.Add(ctx.ToNamespacedName(), &mutatedVs); err != nil {
		ctx.Log.Error(err, "failed to add relation reference", "kind", mutatedVs.Kind, "name", mutatedVs.Name)
	}
	reference.AddRefMarker(&mutatedVs, reference.CreateRefMarker(ctx.Name, ref.KindName.String()), string(resource.Action), ref.Hash())

	err = ctx.Client.Create(ctx, &mutatedVs)
	if err != nil && !k8sErrors.IsAlreadyExists(err) {
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapIfWithDetails(err, "failed creating virtual service", "kind", VirtualServiceKind, "name", mutatedVs.Name, "host", hostName.String())})

		return
	}
	report(model.ModificatorStatus{
		LocatorStatus: resource,
		Success:       true,
		Target: &model.Resource{
			Namespace: mutatedVs.Namespace,
			Kind:      VirtualServiceKind,
			Name:      mutatedVs.Name}})
}

func actionDeleteVirtualService(ctx model.SessionContext, report model.ModificatorStatusReporter, resource model.LocatorStatus) {
	vs := istionetwork.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.Name,
			Namespace: resource.Namespace,
		},
	}

	if err := ctx.Client.Delete(ctx, &vs); err != nil {
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed deleting VirtualService", "kind", VirtualServiceKind, "name", vs.Name)})

		return
	}
	report(model.ModificatorStatus{
		LocatorStatus: resource,
		Success:       true})
}

func actionModifyVirtualService(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.ModificatorStatusReporter, resource model.LocatorStatus) {
	vs, err := getVirtualService(ctx, resource.Namespace, resource.Name)
	if err != nil {
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         err})

		return
	}

	hostName := model.NewHostName(resource.Labels["host"])
	if vsAlreadyMutated(*vs, hostName, model.GetCreatedVersion(store, ctx.Name)) {
		report(model.ModificatorStatus{LocatorStatus: resource, Success: true})

		return
	}
	patch := client.MergeFrom(vs.DeepCopy())
	mutatedVs, err := mutateVirtualService(ctx, store, hostName, *vs)
	if err != nil {
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapIfWithDetails(err, "failed mutating virtual service", "kind", VirtualServiceKind, "name", resource.Name, "host", hostName.String())})
	}

	if err = reference.Add(ctx.ToNamespacedName(), &mutatedVs); err != nil {
		ctx.Log.Error(err, "failed to add relation reference", "kind", mutatedVs.Kind, "name", mutatedVs.Name)
	}
	reference.AddRefMarker(&mutatedVs, reference.CreateRefMarker(ctx.Name, ref.KindName.String()), string(resource.Action), ref.Hash())

	err = ctx.Client.Patch(ctx, &mutatedVs, patch)
	if err != nil {
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapIfWithDetails(err, "failed updating virtual service", "kind", VirtualServiceKind, "name", mutatedVs.Name, "host", hostName.String())})

		return
	}
	report(model.ModificatorStatus{LocatorStatus: resource, Success: true})
}

func actionRevertVirtualService(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.ModificatorStatusReporter, resource model.LocatorStatus) {
	vs, err := getVirtualService(ctx, resource.Namespace, resource.Name)
	if err != nil {
		report(model.ModificatorStatus{LocatorStatus: resource, Success: false, Error: err})

		return
	}

	patch := client.MergeFrom(vs.DeepCopy())
	mutatedVs := revertVirtualService(model.GetDeletedVersion(store), *vs)
	if err = reference.Remove(ctx.ToNamespacedName(), &mutatedVs); err != nil {
		ctx.Log.Error(err, "failed to add relation reference", "kind", mutatedVs.Kind, "name", mutatedVs.Name)
	}
	reference.RemoveRefMarker(&mutatedVs, reference.CreateRefMarker(ctx.Name, ref.KindName.String()))

	err = ctx.Client.Patch(ctx, &mutatedVs, patch)
	if err != nil {
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapWithDetails(err, "failed updating VirtualService", "kind", VirtualServiceKind, "name", vs.Name)})

		return
	}
	report(model.ModificatorStatus{LocatorStatus: resource, Success: true})
}

func mutateVirtualService(ctx model.SessionContext, store model.LocatorStatusStore,
	hostName model.HostName, source istionetwork.VirtualService) (istionetwork.VirtualService, error) {
	version := model.GetVersion(store)
	newVersion := model.GetCreatedVersion(store, ctx.Name)
	target := source.DeepCopy()
	clonedSource := source.DeepCopy()

	targetsHTTP := findRoutes(clonedSource, hostName, version)
	if len(targetsHTTP) == 0 {
		return istionetwork.VirtualService{}, errRouteNotFound
	}
	for _, tHTTP := range targetsHTTP {
		simplifyTargetRoute(ctx, *tHTTP, hostName, version, newVersion, target)
	}

	return *target, nil
}

func mutateConnectedVirtualService(ctx model.SessionContext, store model.LocatorStatusStore,
	hostName model.HostName, source istionetwork.VirtualService) istionetwork.VirtualService {
	version := model.GetVersion(store)
	newVersion := model.GetCreatedVersion(store, ctx.Name)
	target := source.DeepCopy()
	clonedSource := source.DeepCopy()
	gateways, _ := connectedToGateway(*target)
	hosts := getHostsFromGateway(ctx, store, gateways)

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

func simplifyTargetRouteWithoutMatch(targetHTTP v1alpha3.HTTPRoute, hostName model.HostName, version, newVersion string, target *istionetwork.VirtualService) {
	targetHTTP = removeOtherRoutes(targetHTTP, hostName, version)
	targetHTTP = updateSubset(targetHTTP, newVersion)
	targetHTTP = removeWeight(targetHTTP)
	targetHTTP.Mirror = nil
	targetHTTP.Redirect = nil

	target.Spec.Http = append([]*v1alpha3.HTTPRoute{&targetHTTP}, target.Spec.Http...)
}

func simplifyTargetRoute(ctx model.SessionContext, targetHTTP v1alpha3.HTTPRoute, hostName model.HostName, version, newVersion string, target *istionetwork.VirtualService) {
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

func getVirtualService(ctx model.SessionContext, namespace, name string) (*istionetwork.VirtualService, error) {
	virtualService := istionetwork.VirtualService{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &virtualService)

	return &virtualService, errors.WrapWithDetails(err, "failed finding virtual service in namespace", "name", name, "namespace", namespace)
}

func getVirtualServices(ctx model.SessionContext, namespace string, opts ...client.ListOption) (*istionetwork.VirtualServiceList, error) {
	virtualServices := istionetwork.VirtualServiceList{}
	err := ctx.Client.List(ctx, &virtualServices, append(opts, client.InNamespace(namespace))...)

	return &virtualServices, errors.WrapWithDetails(err, "failed finding virtual services in namespace", "namespace", namespace)
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

func vsAlreadyMutated(vs istionetwork.VirtualService, targetHost model.HostName, targetVersion string) bool {
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

func findRoutes(vs *istionetwork.VirtualService, host model.HostName, subset string) []*v1alpha3.HTTPRoute {
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

func removeOtherRoutes(http v1alpha3.HTTPRoute, host model.HostName, subset string) v1alpha3.HTTPRoute {
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

func addHeaderMatch(http v1alpha3.HTTPRoute, route model.Route) v1alpha3.HTTPRoute {
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

func addHeaderRequest(http v1alpha3.HTTPRoute, route model.Route) v1alpha3.HTTPRoute {
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

func getHostsFromGateway(ctx model.SessionContext, store model.LocatorStatusStore, gateways []string) []string {
	var hosts []string
	gwByName := func(store model.LocatorStatusStore, gatewayName string) []model.LocatorStatus {
		var f []model.LocatorStatus
		for _, g := range store(GatewayKind) {
			if g.GetNamespaceName() == gatewayName || g.Name == gatewayName {
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
