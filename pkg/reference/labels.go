package reference

import (
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/maistra/istio-workspace/pkg/naming"
)

const (
	labelPrefix = "maistra.io."
)

func CreateLabel(session, ref string) string {
	return naming.ConcatToMax(40, session, ref) + "-X"
}

// Match filters the list operation checking if the set of session specific labels
// exists without checking their values.
func Match(session string) client.ListOption {
	return client.HasLabels([]string{labelPrefix + session})
}

// AddLabel sets session specific label on a given object.
func AddLabel(object client.Object, key, value, hash string) {
	labels := object.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels[labelPrefix+key] = value + "-" + hash
	object.SetLabels(labels)
}

// GetLabel returns value of a label specific to a given session.
func GetLabel(object client.Object, key string) (value, hash string) {
	labels := object.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	vs := strings.Split(labels[labelPrefix+key], "-")
	if len(vs) == 2 {
		return vs[0], vs[1]
	}

	return vs[0], ""
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
