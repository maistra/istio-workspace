package istio

import (
	"strings"

	"github.com/maistra/istio-workspace/pkg/model/new"
	"github.com/maistra/istio-workspace/pkg/reference"
)

const (
	// LabelIkeHosts describes the labels key used on the Gateway LocatedResource of Hosts bound to this Gateway.
	LabelIkeHosts = "ike.hosts"
)

var _ new.Locator = VirtualServiceGatewayLocator

// VirtualServiceGatewayLocator locates the Gateways that are connected to VirtualServices.
func VirtualServiceGatewayLocator(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.LocatorStatusReporter) {

	switch ref.Deleted {
	case false:
		vss, err := getVirtualServices(ctx, ctx.Namespace)
		if err != nil {
			return
		}
		for _, vs := range vss.Items { //nolint:gocritic //reason for readability
			if gateways, connected := connectedToGateway(vs); connected {
				for _, gwName := range gateways {
					gw, err := getGateway(ctx, ctx.Namespace, gwName)
					if err != nil {
						continue
					}

					var existingHosts []string
					if hosts := gw.Annotations[LabelIkeHosts]; hosts != "" {
						existingHosts = strings.Split(hosts, ",") // split on empty string return empty (len(1))
					}

					var hosts []string
					for _, server := range gw.Spec.Servers {
						for _, host := range server.Hosts {
							if !existInList(existingHosts, host) {
								hosts = append(hosts, host)
							}
						}
					}

					report(new.LocatorStatus{Kind: GatewayKind, Name: gwName, Labels: map[string]string{LabelIkeHosts: strings.Join(hosts, ",")}, Action: new.ActionModify})
				}
			}
		}
	case true:
		gws, err := getGateways(ctx, ctx.Namespace, reference.Match(ctx.Name))
		if err != nil {
			// TODO: report err outside of specific resource?

			return
		}

		for _, gw := range gws.Items {
			action := new.Flip(new.StatusAction(reference.GetLabel(&gw, ctx.Name)))
			report(new.LocatorStatus{Kind: GatewayKind, Namespace: gw.Namespace, Name: gw.Name, Action: action})
		}

	}

}
