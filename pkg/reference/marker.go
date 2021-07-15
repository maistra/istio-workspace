package reference

import (
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/maistra/istio-workspace/pkg/naming"
)

const (
	prefix = "maistra.io."
)

func CreateRefMarker(session, ref string) string {
	return naming.ConcatToMax(40, session, ref) + "-X"
}

// RefMarkerMatch filters the list operation checking if the set of session specific labels
// exists without checking their values.
func RefMarkerMatch(session string) client.ListOption {
	return client.HasLabels([]string{prefix + session})
}

// AddRefMarker sets session specific label on a given object.
func AddRefMarker(object client.Object, key, value, hash string) {
	labels := object.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels[prefix+key] = value + "-" + hash
	object.SetLabels(labels)
}

// GetRefMarker returns value of a label specific to a given session.
func GetRefMarker(object client.Object, key string) (value, hash string) {
	labels := object.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	vs := strings.Split(labels[prefix+key], "-")
	if len(vs) == 2 {
		return vs[0], vs[1]
	}

	return vs[0], ""
}

// RemoveRefMarker removes label for a specific session. If label does not exists RemoveRefMarker is no-op.
func RemoveRefMarker(object client.Object, session string) {
	labels := object.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	delete(labels, prefix+session)
	object.SetLabels(labels)
}
