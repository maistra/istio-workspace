package istio

import (
	"strings"

	"emperror.dev/errors"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/reference"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// GatewayKind is the k8s Kind for a istio Gateway.
	GatewayKind = "Gateway"
)

var _ model.Modificator = GatewayModificator
var _ model.ModificatorRegistrar = GatewayRegistrar

func GatewayRegistrar() (client.Object, model.Modificator) {
	return &istionetwork.Gateway{}, GatewayModificator
}

// GatewayModificator attempts to expose a external host on the gateway.
func GatewayModificator(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.ModificatorStatusReporter) {
	for _, resource := range store(GatewayKind) {
		switch resource.Action {
		case model.ActionModify:
			actionModifyGateway(ctx, ref, report, resource)
		case model.ActionRevert:
			actionRevertGateway(ctx, ref, report, resource)
		case model.ActionCreate, model.ActionDelete, model.ActionLocated:
			report(model.ModificatorStatus{
				LocatorStatus: resource,
				Success:       false,
				Error:         errors.Errorf("Unknown action type for modificator: %v", resource.Action)})
		}
	}
}

func actionModifyGateway(ctx model.SessionContext, ref model.Ref, report model.ModificatorStatusReporter, resource model.LocatorStatus) {
	gw, err := getGateway(ctx, resource.Namespace, resource.Name)
	if err != nil {
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         err})

		return
	}

	ctx.Log.Info("Found Gateway", "name", gw.Name)
	patch := client.MergeFrom(gw.DeepCopy())
	mutatedGw, addedHosts := mutateGateway(ctx, *gw)

	if err = reference.Add(ctx.ToNamespacedName(), &mutatedGw); err != nil {
		ctx.Log.Error(err, "failed to add relation reference", "kind", mutatedGw.Kind, "name", mutatedGw.Name)
	}
	reference.AddRefMarker(&mutatedGw, reference.CreateRefMarker(ctx.Name, ref.KindName.String()), string(resource.Action), ref.Hash())

	err = ctx.Client.Patch(ctx, &mutatedGw, patch)
	if err != nil {
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapIfWithDetails(err, "failed updating gateway", "kind", GatewayKind, "name", mutatedGw.Name)})

		return
	}

	report(model.ModificatorStatus{
		LocatorStatus: resource,
		Success:       true,
		Prop: map[string]string{
			"hosts": strings.Join(addedHosts, ","),
		},
	})
}

func actionRevertGateway(ctx model.SessionContext, ref model.Ref, report model.ModificatorStatusReporter, resource model.LocatorStatus) {
	gw, err := getGateway(ctx, resource.Namespace, resource.Name)
	if err != nil {
		if k8sErrors.IsNotFound(err) { // Not found, nothing to clean
			report(model.ModificatorStatus{
				LocatorStatus: resource,
				Success:       true})

			return
		}
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         err})

		return
	}

	ctx.Log.Info("Found Gateway", "name", resource.Name)
	patch := client.MergeFrom(gw.DeepCopy())
	mutatedGw := revertGateway(ctx, *gw)
	if err = reference.Remove(ctx.ToNamespacedName(), &mutatedGw); err != nil {
		ctx.Log.Error(err, "failed to remove relation reference", "kind", mutatedGw.Kind, "name", mutatedGw.Name)
	}
	reference.RemoveRefMarker(&mutatedGw, reference.CreateRefMarker(ctx.Name, ref.KindName.String()))

	err = ctx.Client.Patch(ctx, &mutatedGw, patch)
	if err != nil {
		report(model.ModificatorStatus{
			LocatorStatus: resource,
			Success:       false,
			Error:         errors.WrapIfWithDetails(err, "failed updateing gateway", "kind", GatewayKind, "name", mutatedGw.Name)})

		return
	}
	// ok, removed
	report(model.ModificatorStatus{
		LocatorStatus: resource,
		Success:       true})
}

func mutateGateway(ctx model.SessionContext, source istionetwork.Gateway) (mutatedGw istionetwork.Gateway, addedHosts []string) {
	if source.Annotations == nil {
		source.Annotations = map[string]string{}
	}
	addedHosts = []string{}
	var existingHosts []string
	if hosts := source.Annotations[LabelIkeHosts]; hosts != "" {
		existingHosts = strings.Split(hosts, ",") // split on empty string return empty (len(1))
	}
	for _, server := range source.Spec.Servers {
		hosts := server.Hosts
		for _, host := range hosts {
			newHost := ctx.Name + "." + host
			if !isInSlice(existingHosts, host) && !isInSlice(existingHosts, newHost) {
				existingHosts = append(existingHosts, newHost)
				hosts = append(hosts, newHost)
			}
			if isInSlice(existingHosts, newHost) {
				addedHosts = append(addedHosts, newHost)
			}
		}
		for _, existing := range existingHosts {
			baseHost := strings.Join(strings.Split(existing, ".")[1:], ".")
			if !isInSlice(hosts, existing) && isInSlice(hosts, baseHost) {
				hosts = append(hosts, existing)
			}
		}
		server.Hosts = hosts
	}
	source.Annotations[LabelIkeHosts] = strings.Join(existingHosts, ",")

	return source, addedHosts
}

func revertGateway(ctx model.SessionContext, source istionetwork.Gateway) istionetwork.Gateway {
	if source.Annotations == nil {
		source.Annotations = map[string]string{}
	}
	var existingHosts []string
	if hosts := source.Annotations[LabelIkeHosts]; hosts != "" {
		existingHosts = strings.Split(hosts, ",") // split on empty string return empty (len(1))
	}
	var toBeRemovedHosts []string
	for _, server := range source.Spec.Servers {
		hosts := server.Hosts
		for i := 0; i < len(hosts); i++ {
			host := hosts[i]
			if isInSlice(existingHosts, host) && strings.HasPrefix(host, ctx.Name+".") {
				toBeRemovedHosts = append(toBeRemovedHosts, host)
				hosts = append(hosts[:i], hosts[i+1:]...)
				i--
			}
		}
		server.Hosts = hosts
	}
	for _, toBeRemoved := range toBeRemovedHosts {
		existingHosts = removeFromList(existingHosts, toBeRemoved)
	}
	if len(existingHosts) == 0 {
		delete(source.Annotations, LabelIkeHosts)
	} else {
		source.Annotations[LabelIkeHosts] = strings.Join(existingHosts, ",")
	}

	return source
}

func getGateway(ctx model.SessionContext, namespace, name string) (*istionetwork.Gateway, error) {
	Gateway := istionetwork.Gateway{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &Gateway)

	return &Gateway, errors.WrapWithDetails(err, "failed finding gateway in namespace", "name", name, "namespace", namespace)
}

func getGateways(ctx model.SessionContext, namespace string, opts ...client.ListOption) (*istionetwork.GatewayList, error) {
	gateways := istionetwork.GatewayList{}
	err := ctx.Client.List(ctx, &gateways, append(opts, client.InNamespace(namespace))...)

	return &gateways, errors.WrapWithDetails(err, "failed finding virtual services in namespace", "namespace", namespace)
}

func isInSlice(hosts []string, host string) bool {
	for _, eh := range hosts {
		if eh == host {
			return true
		}
	}

	return false
}

func removeFromList(hosts []string, host string) []string {
	for i, eh := range hosts {
		if eh == host {
			hosts = append(hosts[:i], hosts[i+1:]...)

			return hosts
		}
	}

	return hosts
}
