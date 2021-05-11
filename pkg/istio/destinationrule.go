package istio

import (
	"emperror.dev/errors"
	istionetworkv1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/naming"
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
	var errs []error
	for _, hostName := range ref.GetTargetHostNames() {
		newVersion := ref.GetNewVersion(ctx.Name)

		subset, err := getTargetSubset(ctx, ctx.Namespace, hostName, ref.GetVersion())
		if err != nil {
			errs = append(errs, errors.WrapIfWithDetails(err, "failed to find Subset", "version", ref.GetVersion(), "host", hostName.String()))

			continue
		}

		destinationRule := istionetwork.DestinationRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      naming.ConcatToMax(63, "dr", ref.KindName.Name, hostName.Name, ctx.Name),
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
						TrafficPolicy: subset.TrafficPolicy,
					},
				},
			},
		}

		if err := reference.Add(ctx.ToNamespacedName(), &destinationRule); err != nil {
			ctx.Log.Error(err, "failed to add relation reference", "kind", destinationRule.Kind, "name", destinationRule.Name, "host", hostName.String())
		}

		if err := ctx.Client.Create(ctx, &destinationRule); err != nil {
			if !k8sErrors.IsAlreadyExists(err) {
				ref.AddResourceStatus(model.NewFailedResource(DestinationRuleKind, destinationRule.GetName(), model.ActionCreated, err.Error()))
				errs = append(errs, errors.WrapWithDetails(
					err, "failed to create DestinationRule",
					"kind", DestinationRuleKind, "name", destinationRule.Name, "host", hostName.String()))

				continue
			}
		}

		ref.AddResourceStatus(model.NewSuccessResource(DestinationRuleKind, destinationRule.GetName(), model.ActionCreated))
	}

	return errors.WrapIfWithDetails(
		errors.Combine(errs...),
		"failed to manipulate destination rules for session", "session", ctx.Name, "namespace", ctx.Namespace, "ref", ref.KindName.Name)
}

// DestinationRuleRevertor looks at the Ref.ResourceStatus and attempts to revert the state of the mutated objects.
func DestinationRuleRevertor(ctx model.SessionContext, ref *model.Ref) error {
	var errs []error
	for _, hostName := range ref.GetTargetHostNames() {
		dr := istionetwork.DestinationRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      naming.ConcatToMax(63, "dr", ref.KindName.Name, hostName.Name, ctx.Name),
				Namespace: ctx.Namespace,
			},
		}

		if err := ctx.Client.Delete(ctx, &dr); err != nil {
			if !k8sErrors.IsNotFound(err) { // Not found, nothing to clean
				ref.AddResourceStatus(model.NewFailedResource(DestinationRuleKind, dr.GetName(), model.ActionCreated, err.Error()))
				errs = append(errs, errors.WrapWithDetails(err, "failed to delete DestinationRule", "kind", DestinationRuleKind, "name", dr.Name, "host", hostName.String()))

				continue
			}
		}

		// ok, removed
		ref.RemoveResourceStatus(model.NewSuccessResource(DestinationRuleKind, dr.GetName(), model.ActionCreated))
	}

	return errors.WrapIfWithDetails(
		errors.Combine(errs...),
		"failed to revert destination rules for session", "session", ctx.Name, "namespace", ctx.Namespace, "ref", ref.KindName.Name)
}

func getTargetSubset(ctx model.SessionContext, namespace string, hostName model.HostName, targetVersion string) (*istionetworkv1alpha3.Subset, error) {
	destinationRules := istionetwork.DestinationRuleList{}
	err := ctx.Client.List(ctx, &destinationRules, client.InNamespace(namespace))
	if err != nil {
		return nil, errors.WrapWithDetails(err, "failed to get destinationrules in namespace", "namespace", namespace)
	}
	for _, dr := range destinationRules.Items { //nolint:gocritic //reason for readability
		if hostName.Match(dr.Spec.Host) {
			for _, subset := range dr.Spec.Subsets {
				if subset.Labels["version"] == targetVersion {
					return subset, nil
				}
			}
		}
	}

	return nil, errors.NewWithDetails("failed finding subset with given host and version", "host", hostName.String(), "version", targetVersion, "namespace", namespace)
}
