package reference

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	labelPrefix = "maistra.io."
)

// Match filters the list operation checking if the set of session specific labels
// exists without checking their values.
func Match(session string) client.ListOption {
	return client.HasLabels([]string{labelPrefix + session})
}

// AddLabel sets session specific label on a given object.
func AddLabel(object client.Object, session, value string) {
	labels := object.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels[labelPrefix+session] = value
	object.SetLabels(labels)
}

// GetLabel returns value of a label specific to a given session.
func GetLabel(object client.Object, session string) string {
	labels := object.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	return labels[labelPrefix+session]
}

// RemoveLabel removes label for a specific session. If label does not exists RemoveLabel is no-op.
func RemoveLabel(object client.Object, session string) {
	labels := object.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	delete(labels, labelPrefix+session)
	object.SetLabels(labels)
}
