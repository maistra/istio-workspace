package istio

import (
	"strings"

	istionetwork "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/networking/v1alpha3"
	"github.com/aslakknutsen/istio-workspace/pkg/model"
	v1alpha3 "istio.io/api/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
)

const (
	// DestinationRuleKind is the k8 Kind for a istio DestinationRule
	DestinationRuleKind = "DestinationRule"
)

var _ model.Mutator = DestinationRuleMutator
var _ model.Revertor = DestinationRuleRevertor

func DestinationRuleMutator(ctx model.SessionContext, ref *model.Ref) error {
	targetName := strings.Split(ref.Name, "-")[0]

	dr, err := getDestinationRuleMapped(ctx.Namespace, targetName)
	if err != nil {
		ref.AddResourceStatus(model.ResourceStatus{Kind: DestinationRuleKind, Name: targetName, Action: model.ActionFailed})
		return err
	}

	ctx.Log.Info("Found DestinationRule", "name", targetName)
	mutatedDr, err := mutateDestinationRule(*dr)
	if err != nil {
		ref.AddResourceStatus(model.ResourceStatus{Kind: DestinationRuleKind, Name: targetName, Action: model.ActionFailed})
		return err
	}
	// ctx.Client.Update
	err = setDestinationRule(ctx.Namespace, &mutatedDr)
	if err != nil {
		ref.AddResourceStatus(model.ResourceStatus{Kind: DestinationRuleKind, Name: targetName, Action: model.ActionFailed})
		return err
	}

	ref.AddResourceStatus(model.ResourceStatus{Kind: DestinationRuleKind, Name: targetName, Action: model.ActionModified})
	return nil
}

// DestinationRuleRevertor looks at the Ref.ResourceStatus and attempts to revert the state of the mutated objects
func DestinationRuleRevertor(ctx model.SessionContext, ref *model.Ref) error {
	resources := ref.GetResourceStatus(DestinationRuleKind)

	for _, resource := range resources {
		dr, err := getDestinationRuleMapped(ctx.Namespace, resource.Name)
		if err != nil {
			if errors.IsNotFound(err) { // Not found, nothing to clean
				break
			}
			ref.AddResourceStatus(model.ResourceStatus{Kind: DestinationRuleKind, Name: resource.Name, Action: model.ActionFailed})
		}

		ctx.Log.Info("Found DestinationRule", "name", resource.Name)
		mutatedDr, err := revertDestinationRule(*dr)
		if err != nil {
			ref.AddResourceStatus(model.ResourceStatus{Kind: DestinationRuleKind, Name: resource.Name, Action: model.ActionFailed})
			break
		}
		// ctx.Client.Update
		err = setDestinationRule(ctx.Namespace, &mutatedDr)
		if err != nil {
			ref.AddResourceStatus(model.ResourceStatus{Kind: DestinationRuleKind, Name: resource.Name, Action: model.ActionFailed})
			break
		}
	}

	return nil
}

func mutateDestinationRule(dr istionetwork.DestinationRule) (istionetwork.DestinationRule, error) {
	dr.Spec.Subsets = append(dr.Spec.Subsets, &v1alpha3.Subset{
		Name: "v1-test",
		Labels: map[string]string{
			"version": "v1-test",
		},
	})
	return dr, nil
}

func revertDestinationRule(dr istionetwork.DestinationRule) (istionetwork.DestinationRule, error) {
	for i := 0; i < len(dr.Spec.Subsets); i++ {
		if strings.Contains(dr.Spec.Subsets[i].Name, "-test") {
			dr.Spec.Subsets = append(dr.Spec.Subsets[:i], dr.Spec.Subsets[i+1:]...)
			break
		}
	}
	return dr, nil
}
