package istio

import (
	"strings"

	"github.com/maistra/istio-workspace/pkg/model"
)

const (
	// LabelIkeHosts describes the labels key used on the Gateway LocatedResource of Hosts bound to this Gateway.
	LabelIkeHosts = "ike.hosts"
)

var _ model.Locator = VirtualServiceGatewayLocator

// VirtualServiceGatewayLocator locates the Gateways that are connected to VirtualServices
func VirtualServiceGatewayLocator(ctx model.SessionContext, ref *model.Ref) bool {
	located := false
	vss, err := getVirtualServices(ctx, ctx.Namespace)
	if err != nil {
		return false
	}

	for _, vs := range vss.Items { //nolint:gocritic //reason for readability
		if gateways, connected := connectedToGateway(vs); connected {
			located = true
			for _, gwName := range gateways {
				gw, err := getGateway(ctx, ctx.Namespace, gwName)
				if err != nil {
					continue
				}

				existingHosts := []string{}
				if hosts := gw.Annotations[LabelIkeHosts]; hosts != "" {
					existingHosts = strings.Split(hosts, ",") // split on empty string return empty (len(1))
				}

				hosts := []string{}
				for _, server := range gw.Spec.Servers {
					for _, host := range server.Hosts {
						if !existInList(existingHosts, host) {
							hosts = append(hosts, host)
						}
					}
				}
				ref.AddTargetResource(model.NewLocatedResource(GatewayKind, gwName, map[string]string{LabelIkeHosts: strings.Join(hosts, ",")}))
			}
		}
	}
	return located
}
