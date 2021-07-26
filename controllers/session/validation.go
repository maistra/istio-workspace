package session

import (
	"strconv"

	"emperror.dev/errors"

	istiov1alpha1 "github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/model"
)

// Validator returns a string of Type and a possible error.
type Validator func(store model.LocatorStatusStore) (string, error)

func chainValidator(ctx model.SessionContext, ref model.Ref, session *istiov1alpha1.Session, validators ...Validator) model.ModificatorController {
	return func(store model.LocatorStatusStore) bool {
		hasError := false
		for _, validator := range validators {
			var message string
			typeName, err := validator(store)
			if err != nil {
				hasError = true
				message = err.Error()
			}
			reason := "Validation"
			status := strconv.FormatBool(err == nil)
			session.AddCondition(istiov1alpha1.Condition{
				Source: istiov1alpha1.Source{
					Kind: "Session",
					Name: ctx.Name,
					Ref:  ref.KindName.String(),
				},
				Reason:  &reason,
				Type:    &typeName,
				Message: &message,
				Status:  &status,
			})
		}

		err := ctx.Client.Status().Update(ctx, session)
		if err != nil {
			ctx.Log.Error(err, "could not update session", "name", session.Name, "namespace", session.Namespace)
		}

		return !hasError
	}
}

func ResourceFound(kind string) Validator {
	return func(store model.LocatorStatusStore) (string, error) {
		if len(store(kind)) == 0 {
			return kind + "Found", errors.New("no " + kind + " found")
		}

		return kind + "Found", nil
	}
}

func TargetFound(store model.LocatorStatusStore) (string, error) {
	typeName := "TargetFound"
	if len(store("DeploymentConfig")) == 0 && len(store("Deployment")) == 0 {
		return typeName, errors.New("no target Deployment or DeploymentConfig found")
	}

	return typeName, nil
}
