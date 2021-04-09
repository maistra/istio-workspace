package istio

import (
	"math"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	istionetworkv1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	var accErrors *multierror.Error
	for _, hostName := range ref.GetTargetHostNames() {
		newVersion := ref.GetNewVersion(ctx.Name)

		destinationRule := istionetwork.DestinationRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      concatToMax(63, "dr", ref.KindName.Name, hostName.Name, ctx.Name),
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

		if err := reference.Add(ctx.ToNamespacedName(), &destinationRule); err != nil {
			ctx.Log.Error(err, "failed to add relation reference", "kind", destinationRule.Kind, "name", destinationRule.Name)
		}

		if err := ctx.Client.Create(ctx, &destinationRule); err != nil {
			if !k8sErrors.IsAlreadyExists(err) {
				ref.AddResourceStatus(model.NewFailedResource(DestinationRuleKind, destinationRule.GetName(), model.ActionCreated, err.Error()))
				ctx.Log.Error(err, "failed to create DestinationRule", "name", destinationRule.GetName())
				accErrors = multierror.Append(accErrors, errors.Wrap(err, "failed to create DestinationRule"))

				continue
			}
		}

		ref.AddResourceStatus(model.NewSuccessResource(DestinationRuleKind, destinationRule.GetName(), model.ActionCreated))
	}

	return errors.Wrapf(accErrors.ErrorOrNil(), "failed to manipulate destination rules for session %s in namespace %s", ctx.Name, ctx.Namespace)
}

// DestinationRuleRevertor looks at the Ref.ResourceStatus and attempts to revert the state of the mutated objects.
func DestinationRuleRevertor(ctx model.SessionContext, ref *model.Ref) error {
	var accErrors *multierror.Error
	for _, hostName := range ref.GetTargetHostNames() {
		dr := istionetwork.DestinationRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      concatToMax(63, "dr", ref.KindName.Name, hostName.Name, ctx.Name),
				Namespace: ctx.Namespace,
			},
		}

		if err := ctx.Client.Delete(ctx, &dr); err != nil {
			if !k8sErrors.IsNotFound(err) { // Not found, nothing to clean
				ref.AddResourceStatus(model.NewFailedResource(DestinationRuleKind, dr.GetName(), model.ActionCreated, err.Error()))
				accErrors = multierror.Append(accErrors, errors.Wrap(err, "failed to delete DestinationRule"))

				continue
			}
		}

		// ok, removed
		ref.RemoveResourceStatus(model.NewSuccessResource(DestinationRuleKind, dr.GetName(), model.ActionCreated))
	}

	return errors.Wrapf(accErrors.ErrorOrNil(), "failed to revert destination rules for session %s in namespace %s", ctx.Name, ctx.Namespace)
}

// concatToMax will cut each section to length based on number of sections to not go beyond max and separate the sections with -.
func concatToMax(max int, sections ...string) string {
	sectionLength := (max - len(sections) - 1) / len(sections)
	name := ""

	for _, section := range sections {
		name = name + "-" + section[:int32(math.Min(float64(len(section)), float64(sectionLength)))]
	}

	return name[1:]
}
