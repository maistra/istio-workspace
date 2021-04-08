package istio

import (
	istionetworkv1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/maistra/istio-workspace/pkg/model"
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
		newVersion := ref.GetNewVersion(ctx.Name)

		dr := istionetwork.DestinationRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dr-" + ref.KindName.Name + "-" + hostName.Name + "-" + ctx.Name,
				Namespace: ctx.Namespace,
			},
			Spec: istionetworkv1alpha3.DestinationRule{
				Host: hostName.Name,
				Subsets: []*istionetworkv1alpha3.Subset{
					{
						Name: newVersion,
						Labels: map[string]string{
							"version": newVersion,
						},
					},
				},
			},
		}

		err := ctx.Client.Create(ctx, &dr)
		if err != nil {
			ref.AddResourceStatus(model.NewFailedResource(DestinationRuleKind, dr.GetName(), model.ActionCreated, err.Error()))
			ctx.Log.Error(err, "failed to update DestinationRule", "name", dr.GetName())
		}

		ref.AddResourceStatus(model.NewSuccessResource(DestinationRuleKind, dr.GetName(), model.ActionCreated))
	}

	return nil
}

// DestinationRuleRevertor looks at the Ref.ResourceStatus and attempts to revert the state of the mutated objects.
func DestinationRuleRevertor(ctx model.SessionContext, ref *model.Ref) error {
	for _, hostName := range ref.GetTargetHostNames() {
		dr := istionetwork.DestinationRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dr-" + ref.KindName.Name + "-" + hostName.Name + "-" + ctx.Name,
				Namespace: ctx.Namespace,
			},
		}

		err := ctx.Client.Delete(ctx, &dr)
		if err != nil {
			if !k8sErrors.IsNotFound(err) { // Not found, nothing to clean
				ref.AddResourceStatus(model.NewFailedResource(DestinationRuleKind, dr.GetName(), model.ActionCreated, err.Error()))
			}
		}

		// ok, removed
		ref.RemoveResourceStatus(model.NewSuccessResource(DestinationRuleKind, dr.GetName(), model.ActionCreated))
	}

	return nil
}
