package stub

import (
	"context"

	"github.com/aslakknutsen/istio-workspace/pkg/apis/istio/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
)

func NewHandler(name string) sdk.Handler {
	return &Handler{name: name}
}

type Handler struct {
	name string
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.Session:
		if event.Deleted {
			logrus.Infof("Removing ", o.Name)
		} else {
			logrus.Infof("Adding ", o.Name)
		}
	}
	return nil
}
