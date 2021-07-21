package session

import (
	istiov1alpha1 "github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/model"
	"strconv"
)

func addConditionForLocatedRef(session *istiov1alpha1.Session, ref model.Ref, modified *model.LocatorStatus) {
	message := modified.Kind + "/" + modified.Name + " located " + ref.KindName.String() + ": "
	reason := toReason(modified.Action)
	typeStr := string(modified.Action)
	located := "true"
	session.AddCondition(istiov1alpha1.Condition{
		Source: istiov1alpha1.Source{
			Kind: modified.Kind,
			Name: modified.Name,
			Ref:  ref.KindName.String(),
		},
		Message: &message,
		Reason:  &reason,
		Status:  &located,
		Type:    &typeStr,
	})
}

func addConditionForModifiedRef(session *istiov1alpha1.Session, ref model.Ref, modified *model.ModificatorStatus) {
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

	reason := toReason(modified.Action)
	typeStr := string(modified.Action)
	session.AddCondition(istiov1alpha1.Condition{
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
	})
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

func toReason(a model.StatusAction) string {
	switch a {
	case model.ActionCreate, model.ActionDelete, model.ActionLocated:

		return "Handled"
	case model.ActionModify, model.ActionRevert:

		return "Configured"
	}

	return ""
}
