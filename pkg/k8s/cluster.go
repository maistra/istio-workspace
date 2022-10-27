package k8s

import (
	"context"

	"emperror.dev/errors"
	"github.com/maistra/istio-workspace/pkg/k8s/dynclient"
	errorsK8s "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type InstallationVerifier interface {
	CheckCRD() (bool, error)
}

var _ InstallationVerifier = &ClusterVerifier{}

type ClusterVerifier struct {
	client *dynclient.Client
}

func (v *ClusterVerifier) CheckCRD() (bool, error) {
	var err error
	if v.client == nil {
		v.client, err = dynclient.NewDefaultDynamicClient("", false)
		if err != nil {
			return false, errors.Wrap(err, "failed creating dynamic client for cluster verification")
		}
	}

	res, err := v.client.Dynamic().Resource(schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}).Get(context.Background(), "sessions.workspace.maistra.io", metav1.GetOptions{})

	if errorsK8s.IsNotFound(err) {
		return false, nil
	}

	return res != nil, errors.Wrap(err, "failed checking if istio-workspace operator is installed")
}

var _ InstallationVerifier = &AssumeOperatorInstalled{}

// AssumeOperatorInstalled is used for testing where we simply assume CRD is always installed.
// This is in particular useful as test double where cluster is not involved at all.
type AssumeOperatorInstalled struct{}

func (AssumeOperatorInstalled) CheckCRD() (bool, error) {
	return true, nil
}
