package istio

import (
	"strings"

	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/maistra/istio-workspace/pkg/model"

	"istio.io/api/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// DestinationRuleKind is the k8s Kind for a istio DestinationRule
	DestinationRuleKind = "DestinationRule"
)

var _ model.Mutator = DestinationRuleMutator
var _ model.Revertor = DestinationRuleRevertor

// DestinationRuleMutator creates destination rule mutator which is responsible for alternating the traffic for development
// of the forked service.
func DestinationRuleMutator(ctx model.SessionContext, ref *model.Ref) error {
	if len(ref.GetResourceStatus(DestinationRuleKind)) > 0 {
		return nil
	}

	for _, hostName := range ref.GetTargetHostNames() {
		drs, err := getDestinationRulesByHost(ctx, ctx.Namespace, hostName)
		if err != nil {
			return err
		}
		for _, dr := range drs {
			ctx.Log.Info("Found DestinationRule", "name", dr.GetName())
			mutatedDr := mutateDestinationRule(*dr, ref.GetNewVersion(ctx.Name))
			err = ctx.Client.Update(ctx, &mutatedDr)
			if err != nil {
				ref.AddResourceStatus(model.ResourceStatus{Kind: DestinationRuleKind, Name: dr.GetName(), Action: model.ActionFailed})
				ctx.Log.Error(err, "failed to update DestinationRule", "name", dr.GetName())
			}

			ref.AddResourceStatus(model.ResourceStatus{Kind: DestinationRuleKind, Name: dr.GetName(), Action: model.ActionModified})
		}
	}
	return nil
}

// DestinationRuleRevertor looks at the Ref.ResourceStatus and attempts to revert the state of the mutated objects.
func DestinationRuleRevertor(ctx model.SessionContext, ref *model.Ref) error {
	resources := ref.GetResourceStatus(DestinationRuleKind)

	for _, resource := range resources {
		dr, err := getDestinationRule(ctx, ctx.Namespace, resource.Name)
		if err != nil {
			if errors.IsNotFound(err) { // Not found, nothing to clean
				break
			}
			ref.AddResourceStatus(model.ResourceStatus{Kind: DestinationRuleKind, Name: resource.Name, Action: model.ActionFailed})
			break
		}

		ctx.Log.Info("Found DestinationRule", "name", resource.Name)
		mutatedDr := revertDestinationRule(*dr, ref.GetNewVersion(ctx.Name))
		err = ctx.Client.Update(ctx, &mutatedDr)
		if err != nil {
			ref.AddResourceStatus(model.ResourceStatus{Kind: DestinationRuleKind, Name: resource.Name, Action: model.ActionFailed})
			break
		}
		// ok, removed
		ref.RemoveResourceStatus(model.ResourceStatus{Kind: DestinationRuleKind, Name: resource.Name})
	}

	return nil
}

func mutateDestinationRule(dr istionetwork.DestinationRule, name string) istionetwork.DestinationRule {
	dr.Spec.Subsets = append(dr.Spec.Subsets, &v1alpha3.Subset{
		Name: name,
		Labels: map[string]string{
			"version": name,
		},
	})
	return dr
}

func revertDestinationRule(dr istionetwork.DestinationRule, name string) istionetwork.DestinationRule {
	for i := 0; i < len(dr.Spec.Subsets); i++ {
		if strings.Contains(dr.Spec.Subsets[i].Name, name) {
			dr.Spec.Subsets = append(dr.Spec.Subsets[:i], dr.Spec.Subsets[i+1:]...)
			break
		}
	}
	return dr
}

func getDestinationRule(ctx model.SessionContext, namespace, name string) (*istionetwork.DestinationRule, error) {
	destinationRule := istionetwork.DestinationRule{}
	err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &destinationRule)
	return &destinationRule, err
}

func getDestinationRulesByHost(ctx model.SessionContext, namespace string, hostName model.HostName) ([]*istionetwork.DestinationRule, error) {
	matches := []*istionetwork.DestinationRule{}

	destinationRules := istionetwork.DestinationRuleList{}
	err := ctx.Client.List(ctx, &destinationRules, client.InNamespace(namespace))
	for _, dr := range destinationRules.Items { //nolint:gocritic //reason for readability
		if hostName.Match(dr.Spec.Host) {
			match := dr
			matches = append(matches, &match)
		}
	}

	return matches, err
}
