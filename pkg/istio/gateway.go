package istio

import (
	"strings"

	"github.com/maistra/istio-workspace/pkg/model"

	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// GatewayKind is the k8s Kind for a istio Gateway
	GatewayKind = "Gateway"
)

var _ model.Mutator = GatewayMutator
var _ model.Revertor = GatewayRevertor

// GatewayMutator attempts to expose a external host on the gateway
func GatewayMutator(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	if len(ref.GetResources(model.Kind(GatewayKind))) > 0 {
		return nil
	}
	for _, gwName := range ref.GetTargets(model.Kind(GatewayKind)) {
		gw, err := getGateway(ctx, ctx.Namespace, gwName.Name)
		if err != nil {
			ref.AddResourceStatus(model.ResourceStatus{Kind: GatewayKind, Name: gw.Name, Action: model.ActionFailed})
			return err
		}

		mutatedGw, err := mutateGateway(ctx, *gw)
		if err != nil {
			ref.AddResourceStatus(model.ResourceStatus{Kind: GatewayKind, Name: gw.Name, Action: model.ActionFailed})
			return err
		}

		err = ctx.Client.Update(ctx, &mutatedGw)
		if err != nil {
			ref.AddResourceStatus(model.ResourceStatus{Kind: GatewayKind, Name: mutatedGw.Name, Action: model.ActionFailed})
			return err
		}

		ref.AddResourceStatus(model.ResourceStatus{Kind: GatewayKind, Name: mutatedGw.Name, Action: model.ActionModified})
	}
	return nil
}

// GatewayRevertor looks at the Ref.ResourceStatus and attempts to revert the state of the mutated objects
func GatewayRevertor(ctx model.SessionContext, ref *model.Ref) error { //nolint[:hugeParam]
	resources := ref.GetResources(model.Kind(GatewayKind))

	for _, resource := range resources {
		gw, err := getGateway(ctx, ctx.Namespace, resource.Name)
		if err != nil {
			if errors.IsNotFound(err) { // Not found, nothing to clean
				break
			}
			ref.AddResourceStatus(model.ResourceStatus{Kind: GatewayKind, Name: resource.Name, Action: model.ActionFailed})
			break
		}

		ctx.Log.Info("Found Gateway", "name", resource.Name)
		mutatedGw, err := revertGateway(ctx, *gw)
		if err != nil {
			ref.AddResourceStatus(model.ResourceStatus{Kind: GatewayKind, Name: resource.Name, Action: model.ActionFailed})
			break
		}
		err = ctx.Client.Update(ctx, &mutatedGw)
		if err != nil {
			ref.AddResourceStatus(model.ResourceStatus{Kind: GatewayKind, Name: resource.Name, Action: model.ActionFailed})
			break
		}
		// ok, removed
		ref.RemoveResourceStatus(model.ResourceStatus{Kind: GatewayKind, Name: resource.Name})
	}

	return nil
}

func mutateGateway(ctx model.SessionContext, source istionetwork.Gateway) (istionetwork.Gateway, error) { //nolint[:hugeParam]
	if source.Annotations == nil {
		source.Annotations = map[string]string{}
	}
	existingHosts := []string{}
	if hosts := source.Annotations[LabelIkeHosts]; hosts != "" {
		existingHosts = strings.Split(hosts, ",") // split on empty string return empty (len(1))
	}
	for _, server := range source.Spec.Servers {
		hosts := server.Hosts
		for _, host := range hosts {
			if !existInList(existingHosts, host) {
				newHost := ctx.Name + "." + host
				existingHosts = append(existingHosts, newHost)
				hosts = append(hosts, newHost)
			}
		}
		server.Hosts = hosts
	}
	source.Annotations[LabelIkeHosts] = strings.Join(existingHosts, ",")
	return source, nil
}

func revertGateway(ctx model.SessionContext, source istionetwork.Gateway) (istionetwork.Gateway, error) { //nolint[:hugeParam]
	if source.Annotations == nil {
		source.Annotations = map[string]string{}
	}
	existingHosts := []string{}
	if hosts := source.Annotations[LabelIkeHosts]; hosts != "" {
		existingHosts = strings.Split(hosts, ",") // split on empty string return empty (len(1))
	}
	for _, server := range source.Spec.Servers {
		hosts := server.Hosts
		for i := 0; i < len(hosts); i++ {
			host := hosts[i]
			if existInList(existingHosts, host) && strings.HasPrefix(host, ctx.Name+".") {
				existingHosts = removeFromList(existingHosts, host)
				hosts = append(hosts[:i], hosts[i+1:]...)
				i--
			}
		}
		server.Hosts = hosts
	}
	if len(existingHosts) == 0 {
		delete(source.Annotations, LabelIkeHosts)
	} else {
		source.Annotations[LabelIkeHosts] = strings.Join(existingHosts, ",")
	}

	return source, nil
}

func getGateway(ctx model.SessionContext, namespace, name string) (*istionetwork.Gateway, error) { //nolint[:hugeParam]
	Gateway := istionetwork.Gateway{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &Gateway)
	return &Gateway, err
}

func getGateways(ctx model.SessionContext, namespace string) (*istionetwork.GatewayList, error) { //nolint[:hugeParam]
	gateways := istionetwork.GatewayList{}
	err := ctx.Client.List(ctx, &gateways, client.InNamespace(namespace))
	return &gateways, err
}

func existInList(hosts []string, host string) bool {
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
