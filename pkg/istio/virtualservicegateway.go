package istio

import (
	"strings"

	"emperror.dev/errors"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/reference"
	"istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

const (
	// LabelIkeHosts describes the labels key used on the Gateway LocatedResource of Hosts bound to this Gateway.
	LabelIkeHosts = "ike.hosts"
)

var _ model.Locator = VirtualServiceGatewayLocator

// VirtualServiceGatewayLocator locates the Gateways that are connected to VirtualServices.
func VirtualServiceGatewayLocator(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.LocatorStatusReporter) error {
	var errs error

	labelKey := reference.CreateRefMarker(ctx.Name, ref.KindName.String())
	gws, err := getGateways(ctx, ctx.Namespace, reference.RefMarkerMatch(labelKey))
	if err != nil {
		return err
	}

	if !ref.Remove {
		for i := range gws.Items {
			gw := gws.Items[i]
			action, hash := reference.GetRefMarker(&gw, labelKey)
			undo := model.Flip(model.StatusAction(action))
			if ref.Hash() != hash {
				report(model.LocatorStatus{
					Resource: model.Resource{
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
					gwName := gwName // pin
					// Gateways in other namespaces may be referred to
					// by <gateway namespace>/<gateway name>;
					// specifying a gateway with no namespace qualifier
					// is the same as specifying the VirtualService’s namespace.
					gwNs := vs.Namespace // default if not specified otherwise
					gwNameNs := strings.Split(gwName, "/")
					if len(gwNameNs) > 1 {
						gwNs = gwNameNs[0]
						gwName = gwNameNs[1]
					}
					gw, err := getGateway(ctx, gwNs, gwName)
					if err != nil {
						errs = errors.Append(errs, err)

						continue
					}

					existingHosts := extractExistingHosts(gw)

					var hosts []string
					for _, server := range gw.Spec.Servers {
						hosts = findNewHosts(server, existingHosts, hosts)
					}

					report(model.LocatorStatus{
						Resource: model.Resource{
							Kind:      GatewayKind,
							Namespace: gwNs,
							Name:      gwName,
						},
						Labels: map[string]string{LabelIkeHosts: strings.Join(hosts, ",")}, Action: model.ActionModify})
				}
			}
		}
	} else {
		for i := range gws.Items {
			gw := gws.Items[i]
			action, _ := reference.GetRefMarker(&gw, labelKey)
			undo := model.Flip(model.StatusAction(action))
			report(model.LocatorStatus{
				Resource: model.Resource{
					Kind:      GatewayKind,
					Namespace: gw.Namespace,
					Name:      gw.Name,
				},
				Action: undo})
		}
	}

	return errors.WrapWithDetails(errs, "failed locating the Gateways that are connected to VirtualServices. Namespace: %s, Name: %s", ref.Namespace, ref.KindName.String())
}

func findNewHosts(server *v1alpha3.Server, existingHosts, hosts []string) []string {
	for _, host := range server.Hosts {
		if !isInSlice(existingHosts, host) {
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
