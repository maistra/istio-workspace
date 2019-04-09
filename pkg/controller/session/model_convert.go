package session

import (
	istiov1alpha1 "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/aslakknutsen/istio-workspace/pkg/model"
)

// RefToStatus appends/replaces the Ref in the provided Session.Status.Ref list
func RefToStatus(ref model.Ref, session *istiov1alpha1.Session) {
	statusRef := &istiov1alpha1.RefStatus{Name: ref.Name}
	for _, refStat := range ref.ResourceStatuses {
		rs := refStat
		action := string(rs.Action)
		statusRef.Resources = append(statusRef.Resources, &istiov1alpha1.RefResource{Name: &rs.Name, Kind: &rs.Kind, Action: &action})
	}
	var existsInStatus bool
	for i, statRef := range session.Status.Refs {
		if statRef.Name == statusRef.Name {
			if len(statusRef.Resources) == 0 { // Remove
				session.Status.Refs = append(session.Status.Refs[:i], session.Status.Refs[i+1:]...)
			} else { // Update
				session.Status.Refs[i] = statusRef
			}
			existsInStatus = true
			break
		}
	}
	if !existsInStatus {
		session.Status.Refs = append(session.Status.Refs, statusRef)
	}
}

// StatusesToRef creates a List of Refs based on the Session.Status.Refs list
func StatusesToRef(session istiov1alpha1.Session) []*model.Ref { //nolint[:hugeParam]
	refs := []*model.Ref{}
	for _, statusRef := range session.Status.Refs {
		r := &model.Ref{Name: statusRef.Name}
		StatusToRef(session, r)
		refs = append(refs, r)
	}
	return refs
}

// StatusToRef filles the ResourceStatus of a Ref based on the Session.Status.Refs with the same name
func StatusToRef(session istiov1alpha1.Session, ref *model.Ref) { //nolint[:hugeParam]
	for _, statusRef := range session.Status.Refs {
		if statusRef.Name == ref.Name {
			for _, resource := range statusRef.Resources {
				r := resource
				ref.AddResourceStatus(model.ResourceStatus{Name: *r.Name, Kind: *r.Kind, Action: model.ResourceAction(*r.Action)})
			}
		}
	}
}
