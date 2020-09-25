package mutation

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/maistra/istio-workspace/pkg/internal/session"
	"github.com/maistra/istio-workspace/pkg/model"

	istiov1alpha1 "github.com/maistra/istio-workspace/pkg/apis/maistra/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type data struct {
	Object appsv1.Deployment
}

func (d *data) IsIkeable() bool {
	return d.Object.Labels["ike.target"] != ""
}

func (d *data) Join() bool {
	_, ok := d.Object.Labels["ike.session"]
	return ok
}

func (d *data) Target() string {
	return d.Object.Labels["ike.target"]
}

func (d *data) Session() string {
	return d.Object.Labels["ike.session"]
}

func (d *data) Namespace() string {
	return d.Object.Namespace
}

// Webhook to mutate Deployments with ike.target annotations to setup routing to existing pods
type Webhook struct {
	Client  client.Client
	decoder *admission.Decoder
}

var _ admission.DecoderInjector = &Webhook{}

// Handle will create a Session with a "existing" strategy to setup a route to the upcoming deployment.
// The deployment will be injected the correct labels to get the prod route
func (w *Webhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	// if review.Request.DryRun don't do stuff with sideeffects....

	deployment := &appsv1.Deployment{}
	err := w.decoder.Decode(req, deployment)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	d := data{Object: *deployment}
	if !d.IsIkeable() {
		return admission.Allowed("not ikable, move on")
	}

	sess, err := createSessionAndWait(d)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	lables := findLables(sess)
	for k, v := range lables {
		deployment.Labels[k] = v
	}

	marshaledDeployment, err := json.Marshal(deployment)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledDeployment)
}

// InjectDecoder injects the decoder.
func (w *Webhook) InjectDecoder(d *admission.Decoder) error {
	w.decoder = d
	return nil
}

func createSessionAndWait(d data) (*istiov1alpha1.RefStatus, error) {
	checkDuration := time.Millisecond * 100
	options := session.Options{
		DeploymentName: d.Target(),
		SessionName:    d.Session(),
		NamespaceName:  d.Namespace(),
		Strategy:       model.StrategyExisting,
		Duration:       &checkDuration,
		WaitCondition: func(ref *istiov1alpha1.RefResource) bool {
			return false
		},
	}

	c, err := session.DefaultClient(options.NamespaceName)
	if err != nil {
		return nil, err
	}

	state, remove, err := session.CreateOrJoinHandler(options, c)
	if err != nil {
		remove()
		return nil, err
	}

	return &state.RefStatus, nil
}

func findLables(ref *istiov1alpha1.RefStatus) map[string]string {
	// for each Resource, if Deployment, Make label patch
	return map[string]string{}
}
