package istio

import (
	"emperror.dev/errors"
	"github.com/maistra/istio-workspace/pkg/model"
	istionetworkv1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/maistra/istio-workspace/pkg/naming"
	"github.com/maistra/istio-workspace/pkg/reference"
)

const (
	// DestinationRuleKind is the k8s Kind for a istio DestinationRule.
	DestinationRuleKind = "DestinationRule"
)

var _ model.Modificator = DestinationRuleModificator
var _ model.Locator = DestinationRuleLocator
var _ model.ModificatorRegistrar = GatewayRegistrar

func DestinationRuleRegistrar() (client.Object, model.Modificator) {
	return &istionetwork.DestinationRule{}, DestinationRuleModificator
}

func DestinationRuleLocator(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.LocatorStatusReporter) error {
	var errs error

	labelKey := reference.CreateLabel(ctx.Name, ref.KindName.String())
	destinationRules, err := GetDestinationRules(ctx, ctx.Namespace, reference.Match(labelKey))
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to get all destination rules", "ref", ref.KindName.String())
	}

	if !ref.Deleted {
		for i := range destinationRules.Items {
			destinationRule := destinationRules.Items[i]
			action, hash := reference.GetLabel(&destinationRule, labelKey)
			if ref.Hash() != hash {
				undo := model.Flip(model.StatusAction(action))
				report(model.LocatorStatus{
					Resource: model.Resource{
						Kind:      DestinationRuleKind,
						Namespace: destinationRule.Namespace,
						Name:      destinationRule.Name,
					},
					Action: undo})
			}
		}

		for _, hostName := range model.GetTargetHostNames(store) {
			dr, err := locateDestinationRuleWithSubset(ctx, ctx.Namespace, hostName, model.GetVersion(store))
			if err != nil {
				errs = errors.Append(errs, err)

				continue
			}

			report(model.LocatorStatus{
				Resource: model.Resource{
					Kind:      DestinationRuleKind,
					Namespace: dr.Namespace,
					Name:      dr.Name,
				},
				Action: model.ActionCreate})
		}
	} else {
		for i := range destinationRules.Items {
			resource := destinationRules.Items[i]
			action, _ := reference.GetLabel(&resource, labelKey)
			undo := model.Flip(model.StatusAction(action))
			report(model.LocatorStatus{
				Resource: model.Resource{
					Kind:      DestinationRuleKind,
					Namespace: resource.Namespace,
					Name:      resource.Name,
				},
				Action: undo})
		}
	}

	return errors.Wrapf(errs, "failed locating destination rule %s", ref.KindName.String())
}

// DestinationRuleModificator creates destination rule mutator which is responsible for alternating the traffic for development
// of the forked service.
func DestinationRuleModificator(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.ModificatorStatusReporter) {
	for _, resource := range store(DestinationRuleKind) {
		switch resource.Action {
		case model.ActionCreate:
			actionCreateDestinationRule(ctx, ref, store, report, resource)
		case model.ActionDelete:
			actionDeleteDestinationRule(ctx, report, resource)
		case model.ActionModify, model.ActionRevert, model.ActionLocated:
			report(model.ModificatorStatus{
				LocatorStatus: resource,
				Success:       false,
				Error:         errors.Errorf("Unknown action type for modificator: %v", resource.Action)})
		}
	}
}

func actionCreateDestinationRule(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.ModificatorStatusReporter, resource model.LocatorStatus) {
	dr, err := getDestinationRule(ctx, resource.Namespace, resource.Name)
	if err != nil {
		report(model.ModificatorStatus{LocatorStatus: resource, Success: false, Error: err})

		return
	}

	newVersion := model.GetCreatedVersion(store, ctx.Name)

	subset := locateSubset(dr, model.GetVersion(store))
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
	reference.AddLabel(&destinationRule, reference.CreateLabel(ctx.Name, ref.KindName.String()), string(resource.Action), ref.Hash())

	if err := ctx.Client.Create(ctx, &destinationRule); err != nil {
		if !k8sErrors.IsAlreadyExists(err) {
			report(model.ModificatorStatus{
				LocatorStatus: resource,
				Success:       false,
				Error: errors.WrapWithDetails(
					err, "failed to create DestinationRule", "kind", DestinationRuleKind, "name", destinationRule.Name, "host", destinationRule.Spec.Host)})

			return
		}
	}

	report(model.ModificatorStatus{
		LocatorStatus: resource,
		Success:       true,
		Target: &model.Resource{
			Namespace: destinationRule.Namespace,
			Kind:      DestinationRuleKind,
			Name:      destinationRule.Name}})
}

func actionDeleteDestinationRule(ctx model.SessionContext, report model.ModificatorStatusReporter, resource model.LocatorStatus) {
	dr := istionetwork.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.Name,
			Namespace: resource.Namespace,
		},
	}

	if err := ctx.Client.Delete(ctx, &dr); err != nil {
		if !k8sErrors.IsNotFound(err) { // Not found, nothing to clean
			report(model.ModificatorStatus{
				LocatorStatus: resource,
				Success:       false,
				Error:         errors.WrapWithDetails(err, "failed to delete DestinationRule", "kind", DestinationRuleKind, "name", dr.Name)})

			return
		}
	}

	// ok, removed
	report(model.ModificatorStatus{
		LocatorStatus: resource,
		Success:       true})
}

func locateDestinationRuleWithSubset(ctx model.SessionContext, namespace string, hostName model.HostName, targetVersion string) (*istionetwork.DestinationRule, error) {
	destinationRules := istionetwork.DestinationRuleList{}
	err := ctx.Client.List(ctx, &destinationRules, client.InNamespace(namespace))
	if err != nil {
		return nil, errors.WrapWithDetails(err, "failed to get destinationrules in namespace", "namespace", namespace)
	}
	for _, dr := range destinationRules.Items { //nolint:gocritic //reason for readability
		dr := dr
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

func getDestinationRule(ctx model.SessionContext, namespace, name string) (*istionetwork.DestinationRule, error) {
	destinationRule := istionetwork.DestinationRule{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &destinationRule)

	return &destinationRule, errors.WrapWithDetails(err, "failed finding destinationrule in namespace", "name", name, "namespace", namespace)
}

func GetDestinationRules(ctx model.SessionContext, namespace string, opts ...client.ListOption) (*istionetwork.DestinationRuleList, error) {
	destinationRules := istionetwork.DestinationRuleList{}
	err := ctx.Client.List(ctx, &destinationRules, append(opts, client.InNamespace(namespace))...)

	return &destinationRules, errors.WrapWithDetails(err, "failed finding destination rules in namespace", "namespace", namespace)
}
