package istio

import (
	"strings"

	"emperror.dev/errors"
	"istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"

	"github.com/maistra/istio-workspace/pkg/model/new"
	"github.com/maistra/istio-workspace/pkg/reference"
)

const (
	// LabelIkeHosts describes the labels key used on the Gateway LocatedResource of Hosts bound to this Gateway.
	LabelIkeHosts = "ike.hosts"
)

var _ new.Locator = VirtualServiceGatewayLocator

// VirtualServiceGatewayLocator locates the Gateways that are connected to VirtualServices.
func VirtualServiceGatewayLocator(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.LocatorStatusReporter) error {
	var errs error

	labelKey := reference.CreateLabel(ctx.Name, ref.KindName.String())
	gws, err := getGateways(ctx, ctx.Namespace, reference.Match(labelKey))
	if err != nil {
		return err
	}

	if !ref.Deleted {
		for i := range gws.Items {
			gw := gws.Items[i]
			action, hash := reference.GetLabel(&gw, labelKey)
			undo := new.Flip(new.StatusAction(action))
			if ref.Hash() != hash {
				report(new.LocatorStatus{
					Resource: new.Resource{
						Kind:      GatewayKind,
						Namespace: gw.Namespace,
						Name:      gw.Name,
					},
					Action: undo})
			}
		}

		vss, err := getVirtualServices(ctx, ctx.Namespace)
		if err != nil {
			return err
		}
		for i := range vss.Items {
			vs := vss.Items[i]
			if gateways, connected := connectedToGateway(vs); connected {
				for _, gwName := range gateways {
					gw, err := getGateway(ctx, ctx.Namespace, gwName)
					if err != nil {
						errs = errors.Append(errs, err)

						continue
					}

					existingHosts := extractExistingHosts(gw)

					var hosts []string
					for _, server := range gw.Spec.Servers {
						hosts = findNewHosts(server, existingHosts, hosts)
					}

					report(new.LocatorStatus{
						Resource: new.Resource{
							Kind:      GatewayKind,
							Namespace: gw.Namespace,
							Name:      gwName,
						},
						Labels: map[string]string{LabelIkeHosts: strings.Join(hosts, ",")}, Action: new.ActionModify})
				}
			}
		}
	} else {
		for i := range gws.Items {
			gw := gws.Items[i]
			action, _ := reference.GetLabel(&gw, labelKey)
			undo := new.Flip(new.StatusAction(action))
			report(new.LocatorStatus{
				Resource: new.Resource{
					Kind:      GatewayKind,
					Namespace: gw.Namespace,
					Name:      gw.Name,
				},
				Action: undo})
		}
	}

	return errors.Wrapf(errs, "failed locating the Gateways that are connected to VirtualServices %s", ref.KindName.String())
}

func findNewHosts(server *v1alpha3.Server, existingHosts, hosts []string) []string {
	for _, host := range server.Hosts {
		if !existInList(existingHosts, host) {
			hosts = append(hosts, host)
		}
	}

	return hosts
}

func extractExistingHosts(gw *istionetwork.Gateway) []string {
	var existingHosts []string
	if hosts := gw.Annotations[LabelIkeHosts]; hosts != "" {
		existingHosts = strings.Split(hosts, ",") // split on empty string return empty (len(1))
	}

	return existingHosts
}
