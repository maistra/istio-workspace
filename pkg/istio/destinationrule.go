package istio

import (
	"strings"

	"github.com/pkg/errors"
	"istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/reference"
)

const (
	// DestinationRuleKind is the k8s Kind for a istio DestinationRule.
	DestinationRuleKind = "DestinationRule"
)

var _ model.Mutator = DestinationRuleMutator
var _ model.Revertor = DestinationRuleRevertor
var _ model.Manipulator = destinationRuleManipulator{}

// DestinationRuleManipulator represents a model.Manipulator implementation for handling DestinationRule objects.
func DestinationRuleManipulator() model.Manipulator {
	return destinationRuleManipulator{}
}

type destinationRuleManipulator struct {
}

func (d destinationRuleManipulator) TargetResourceType() client.Object {
	return &istionetwork.DestinationRule{}
}
func (d destinationRuleManipulator) Mutate() model.Mutator {
	return DestinationRuleMutator
}
func (d destinationRuleManipulator) Revert() model.Revertor {
	return DestinationRuleRevertor
}

// DestinationRuleMutator creates destination rule mutator which is responsible for alternating the traffic for development
// of the forked service.
func DestinationRuleMutator(ctx model.SessionContext, ref *model.Ref) error {
	for _, hostName := range ref.GetTargetHostNames() {
		drs, err := getDestinationRulesByHost(ctx, ctx.Namespace, hostName)
		if err != nil {
			return err
		}
		for _, dr := range drs {
			newVersion := ref.GetNewVersion(ctx.Name)
			if alreadyMutated(*dr, newVersion) {
				continue
			}
			ctx.Log.Info("Found DestinationRule", "name", dr.GetName())
			mutatedDr := mutateDestinationRule(*dr, newVersion)
			if err = reference.Add(ctx.ToNamespacedName(), &mutatedDr); err != nil {
				ctx.Log.Error(err, "failed to add relation reference", "kind", mutatedDr.Kind, "name", mutatedDr.Name)
			}
			err = ctx.Client.Update(ctx, &mutatedDr)
			if err != nil {
				ref.AddResourceStatus(model.NewFailedResource(DestinationRuleKind, dr.GetName(), model.ActionModified, err.Error()))
				ctx.Log.Error(err, "failed to update DestinationRule", "name", dr.GetName())
			}

			ref.AddResourceStatus(model.NewSuccessResource(DestinationRuleKind, dr.GetName(), model.ActionModified))
		}
	}

	return nil
}

// DestinationRuleRevertor looks at the Ref.ResourceStatus and attempts to revert the state of the mutated objects.
func DestinationRuleRevertor(ctx model.SessionContext, ref *model.Ref) error {
	resources := ref.GetResources(model.Kind(DestinationRuleKind))

	for _, resource := range resources {
		dr, err := getDestinationRule(ctx, ctx.Namespace, resource.Name)
		if err != nil {
			if k8sErrors.IsNotFound(err) { // Not found, nothing to clean
				break
			}
			ref.AddResourceStatus(model.NewFailedResource(DestinationRuleKind, resource.Name, resource.Action, err.Error()))

			break
		}

		ctx.Log.Info("Found DestinationRule", "name", resource.Name)
		mutatedDr := revertDestinationRule(*dr, ref.GetNewVersion(ctx.Name))
		if err = reference.Remove(ctx.ToNamespacedName(), &mutatedDr); err != nil {
			ctx.Log.Error(err, "failed to remove relation reference", "kind", mutatedDr.Kind, "name", mutatedDr.Name)
		}
		err = ctx.Client.Update(ctx, &mutatedDr)
		if err != nil {
			ref.AddResourceStatus(model.NewFailedResource(DestinationRuleKind, resource.Name, resource.Action, err.Error()))

			break
		}
		// ok, removed
		ref.RemoveResourceStatus(model.NewSuccessResource(DestinationRuleKind, resource.Name, resource.Action))
	}

	return nil
}

func alreadyMutated(dr istionetwork.DestinationRule, name string) bool {
	for _, sub := range dr.Spec.Subsets {
		if sub.Name == name {
			return true
		}
	}

	return false
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

	return &destinationRule, errors.Wrapf(err, "failed obtaining destinationrule %s in namespace %s", name, namespace)
}

func getDestinationRulesByHost(ctx model.SessionContext, namespace string, hostName model.HostName) ([]*istionetwork.DestinationRule, error) {
	var matches []*istionetwork.DestinationRule

	destinationRules := istionetwork.DestinationRuleList{}
	err := ctx.Client.List(ctx, &destinationRules, client.InNamespace(namespace))
	for _, dr := range destinationRules.Items { //nolint:gocritic //reason for readability
		if hostName.Match(dr.Spec.Host) {
			match := dr
			matches = append(matches, &match)
		}
	}

	return matches, errors.Wrapf(err, "failed finding destinationrule in namespace %s matching hostname %v", namespace, hostName)
}
