package reference

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	lablePrefix = "maistra.io."
)

func Match(session string) client.ListOption {
	return client.HasLabels([]string{lablePrefix + session})
}

func AddLabel(object client.Object, session string, value string) {
	labels := object.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels[lablePrefix+session] = string(value)
	object.SetLabels(labels)
}

func GetLabel(object client.Object, session string) string {
	labels := object.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	return labels[lablePrefix+session]
}

func RemoveLabel(object client.Object, session string) {
	labels := object.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	delete(labels, lablePrefix+session)
	object.SetLabels(labels)
}
