package session

import (
	"strconv"

	"emperror.dev/errors"

	istiov1alpha1 "github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/model"
)

const (
	ValidationReason = "Validation"
)

// Validator returns a string of Type and a possible error.
type Validator func(store model.LocatorStatusStore) (string, error)

func chainValidator(ctx model.SessionContext, ref model.Ref, session *istiov1alpha1.Session, validators ...Validator) model.ModificatorController {
	return func(store model.LocatorStatusStore) bool {
		succeeded := true
		for _, validator := range validators {
			var message string
			typeName, err := validator(store)
			if err != nil {
				succeeded = false
				message = err.Error()
			}

			reason := ValidationReason
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

		return succeeded
	}
}

func ResourceFound(kind string) Validator {
	return func(store model.LocatorStatusStore) (string, error) {
		targetType := "Find" + kind
		if len(store(kind)) == 0 {
			return targetType, errors.New("no " + kind + " found")
		}

		return targetType, nil
	}
}

func TargetFound(store model.LocatorStatusStore) (string, error) {
	typeName := "FindTarget"
	if len(store("DeploymentConfig")) == 0 && len(store("Deployment")) == 0 {
		return typeName, errors.New("no target Deployment or DeploymentConfig found")
	}

	return typeName, nil
}
