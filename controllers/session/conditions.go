package session

import (
	"strconv"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	istiov1alpha1 "github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/model"
)

func createConditionForLocatedRef(ref model.Ref, located model.LocatorStatus) istiov1alpha1.Condition {
	message := located.Kind + "/" + located.Name + " status " + ref.KindName.String() + ": "
	reason := "Scheduled"
	typeStr := createType(located.Action, located.Kind)
	status := "true"

	return istiov1alpha1.Condition{
		Source: istiov1alpha1.Source{
			Kind: located.Kind,
			Name: located.Name,
			Ref:  ref.KindName.String(),
		},
		Message: &message,
		Reason:  &reason,
		Status:  &status,
		Type:    &typeStr,
	}
}

func createConditionForModifiedRef(ref model.Ref, modified model.ModificatorStatus) istiov1alpha1.Condition {
	message := modified.Kind + "/" + modified.Name + " modified to satisfy " + ref.KindName.String() + ": "
	if modified.Error != nil {
		message += modified.Error.Error()
	} else {
		message += "ok"
	}
	var target *istiov1alpha1.Target
	if modified.Target != nil {
		target = &istiov1alpha1.Target{
			Kind: modified.Target.Kind,
			Name: modified.Target.Name,
		}
	}
	status := strconv.FormatBool(modified.Success)

	reason := "Applied"
	typeStr := createType(modified.Action, modified.Kind)

	return istiov1alpha1.Condition{
		Source: istiov1alpha1.Source{
			Kind: modified.Kind,
			Name: modified.Name,
			Ref:  ref.KindName.String(),
		},
		Target:  target,
		Message: &message,
		Reason:  &reason,
		Status:  &status,
		Type:    &typeStr,
	}
}

func createType(action model.StatusAction, kindName string) string {
	title := cases.Title(language.English)

	return title.String(string(action)) + title.String(kindName)
}

func cleanupRelatedConditionsOnRemoval(ref model.Ref, session *istiov1alpha1.Session) {
	if ref.Remove && refSuccessful(ref, session.Status.Conditions) {
		var otherConditions []*istiov1alpha1.Condition
		for i := range session.Status.Conditions {
			condition := session.Status.Conditions[i]
			if condition.Source.Ref != ref.KindName.String() {
				otherConditions = append(otherConditions, condition)
			}
		}
		session.Status.Conditions = otherConditions
	}
}
