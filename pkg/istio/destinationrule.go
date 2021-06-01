package istio

import (
	"emperror.dev/errors"
	istionetworkv1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/maistra/istio-workspace/pkg/model/new"
	"github.com/maistra/istio-workspace/pkg/naming"
	"github.com/maistra/istio-workspace/pkg/reference"
)

const (
	// DestinationRuleKind is the k8s Kind for a istio DestinationRule.
	DestinationRuleKind = "DestinationRule"
)

var _ new.Modificator = DestinationRuleModificator
var _ new.Locator = DestinationRuleLocator
var _ new.ModificatorRegistrar = GatewayRegistrar

func DestinationRuleRegistrar() (client.Object, new.Modificator) {
	return &istionetwork.DestinationRule{}, DestinationRuleModificator
}

func DestinationRuleLocator(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.LocatorStatusReporter) {
	for _, hostName := range new.GetTargetHostNames(store) {
		switch ref.Deleted {
		case false:
			dr, err := locateDestinationRuleWithSubset(ctx, ctx.Namespace, hostName, new.GetVersion(store))
			if err != nil {
				// TODO: report non found subset as a Locator Status??

				continue
			}

			report(new.LocatorStatus{Kind: DestinationRuleKind, Name: dr.Name, Action: new.ActionCreate})
		case true:
			// TODO shall we use labeling to know if the given resource should be handled by us?
		}
	}
}

// DestinationRuleModificator creates destination rule mutator which is responsible for alternating the traffic for development
// of the forked service.
func DestinationRuleModificator(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.ModificatorStatusReporter) {
	for _, resource := range store(DestinationRuleKind) {
		switch resource.Action {
		case new.ActionCreate:
			actionCreateDestinationRule(ctx, ref, store, report, resource)
		case new.ActionDelete:
			actionDeleteDestinationRule(ctx, ref, store, report, resource)
		default:
			report(new.ModificatorStatus{LocatorStatus: resource, Success: false, Error: errors.Errorf("Unknown action type for modificator: %v", resource.Action)})
		}
	}
}

func actionCreateDestinationRule(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.ModificatorStatusReporter, resource new.LocatorStatus) {
	dr, err := getDestinationRule(ctx, resource.Namespace, resource.Name)
	if err != nil {
		report(new.ModificatorStatus{LocatorStatus: resource, Success: false, Error: err})

		return
	}

	newVersion := new.GetNewVersion(store, ctx.Name)

	subset := locateSubset(dr, new.GetVersion(store))
	destinationRule := istionetwork.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      naming.ConcatToMax(63, "dr", ref.KindName.Name, dr.Spec.Host, ctx.Name),
			Namespace: ctx.Namespace,
		},
		Spec: istionetworkv1alpha3.DestinationRule{
			Host: dr.Spec.Host,
			Subsets: []*istionetworkv1alpha3.Subset{
				{
					Name: newVersion,
					Labels: map[string]string{
						"version": newVersion,
					},
					TrafficPolicy: subset.TrafficPolicy,
				},
			},
		},
	}

	if err := reference.Add(ctx.ToNamespacedName(), &destinationRule); err != nil {
		ctx.Log.Error(err, "failed to add relation reference", "kind", destinationRule.Kind, "name", destinationRule.Name, "host", dr.Spec.Host)
	}

	if err := ctx.Client.Create(ctx, &destinationRule); err != nil {
		if !k8sErrors.IsAlreadyExists(err) {
			report(new.ModificatorStatus{
				LocatorStatus: resource,
				Success:       false,
				Error: errors.WrapWithDetails(
					err, "failed to create DestinationRule", "kind", DestinationRuleKind, "name", destinationRule.Name, "host", destinationRule.Spec.Host)})

			return
		}
	}

	report(new.ModificatorStatus{LocatorStatus: resource, Success: true})
}

func actionDeleteDestinationRule(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.ModificatorStatusReporter, resource new.LocatorStatus) {
	dr := istionetwork.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.Name,
			Namespace: resource.Namespace,
		},
	}

	if err := ctx.Client.Delete(ctx, &dr); err != nil {
		if !k8sErrors.IsNotFound(err) { // Not found, nothing to clean
			report(new.ModificatorStatus{
				LocatorStatus: resource,
				Success:       false,
				Error:         errors.WrapWithDetails(err, "failed to delete DestinationRule", "kind", DestinationRuleKind, "name", dr.Name)})

			return
		}
	}

	// ok, removed
	report(new.ModificatorStatus{LocatorStatus: resource, Success: true})
}

func locateDestinationRuleWithSubset(ctx new.SessionContext, namespace string, hostName new.HostName, targetVersion string) (*istionetwork.DestinationRule, error) {
	destinationRules := istionetwork.DestinationRuleList{}
	err := ctx.Client.List(ctx, &destinationRules, client.InNamespace(namespace))
	if err != nil {
		return nil, errors.WrapWithDetails(err, "failed to get destinationrules in namespace", "namespace", namespace)
	}
	for _, dr := range destinationRules.Items { //nolint:gocritic //reason for readability
		if hostName.Match(dr.Spec.Host) {
			subset := locateSubset(&dr, targetVersion)
			if subset != nil {
				return &dr, nil
			}
		}
	}

	return nil, errors.NewWithDetails("failed finding subset with given host and version", "host", hostName.String(), "version", targetVersion, "namespace", namespace)
}

func locateSubset(dr *istionetwork.DestinationRule, targetVersion string) *istionetworkv1alpha3.Subset {
	for _, subset := range dr.Spec.Subsets {
		if subset.Labels["version"] == targetVersion {
			return subset
		}
	}
	return nil
}

func getDestinationRule(ctx new.SessionContext, namespace, name string) (*istionetwork.DestinationRule, error) {
	destinationRule := istionetwork.DestinationRule{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &destinationRule)

	return &destinationRule, errors.WrapWithDetails(err, "failed finding destinationrule in namespace", "name", name, "namespace", namespace)
}
