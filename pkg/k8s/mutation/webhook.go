package mutation

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-logr/logr"

	istiov1alpha1 "github.com/maistra/istio-workspace/pkg/apis/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/internal/session"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	logger = func() logr.Logger {
		return log.Log.WithValues("type", "controller")
	}
)

type data struct {
	Object appsv1.Deployment
}

func (d *data) IsIkeable() bool {
	//return d.Object.Labels["ike.target"] != ""
	return d.Object.Spec.Selector.MatchLabels["deployment"] == "workspace"
}

func (d *data) Target() string {
	t := d.Object.Labels["ike.target"]
	if t != "" {
		return t
	}
	return "preference-v1"
}

func (d *data) Session() string {
	s := d.Object.Labels["ike.session"]
	if s != "" {
		return s
	}
	//return strings.ReplaceAll(d.Object.Name, ".", "-")
	return "test"
}

func (d *data) Namespace() string {
	return d.Object.Namespace
}

func (d *data) Route() string {
	r := d.Object.Labels["ike.route"]
	if r != "" {
		return r
	}
	return "header:ike-session-id=feature-y"
}

// Webhook to mutate Deployments with ike.target annotations to setup routing to existing pods.
type Webhook struct {
	Client  client.Client
	decoder *admission.Decoder
}

var _ admission.DecoderInjector = &Webhook{}
var _ admission.Handler = &Webhook{}

// Handle will create a Session with a "existing" strategy to setup a route to the upcoming deployment.
// The deployment will be injected the correct labels to get the prod route.
func (w *Webhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	// if review.Request.DryRun don't do stuff with sideeffects....

	deployment := &appsv1.Deployment{}

	if req.Operation == admissionv1beta1.Delete {
		err := w.decoder.DecodeRaw(req.OldObject, deployment)
		if err != nil {
			logger().Error(err, "problems decoding delete request", "deployment", deployment.Name)
			return admission.Allowed(err.Error())
		}
		d := data{Object: *deployment}
		if d.IsIkeable() {
			logger().Info("Removing session", "deployment", req.Name)
			err := removeSession(d)
			if err != nil {
				logger().Error(err, "problems removing session", "deployment", deployment.Name)
				return admission.Allowed(err.Error())
			}
		}
		return admission.Allowed("") // TODO: impl delete behavior
	}

	err := w.decoder.Decode(req, deployment)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	d := data{Object: *deployment}
	if !d.IsIkeable() {
		return admission.Allowed("not ikable, move on")
	}

	logger().Info("Creating session", "deployment", req.Name)
	refStatus, err := createSessionAndWait(d)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	logger().Info("Session created", "deployment", req.Name)

	lables := findLables(refStatus)
	for k, v := range lables {
		logger().Info("Label added", "deployemnt", req.Name, k, v)
		deployment.Spec.Template.Labels[k] = v
	}

	targetHost := findGwHost(refStatus)
	if targetHost != "" {
		for i := 0; i < len(deployment.Spec.Template.Spec.Containers); i++ {
			c := deployment.Spec.Template.Spec.Containers[i]
			logger().Info("Env added", "deployemnt", req.Name, "container", c.Name, "IKE_HOST", targetHost)
			c.Env = append(c.Env, corev1.EnvVar{Name: "IKE_HOST", Value: targetHost})
			deployment.Spec.Template.Spec.Containers[i] = c
		}
	}

	marshaledDeployment, err := json.Marshal(deployment)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	logger().Info("Patch response sent", "deployemnt", req.Name)
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
		RouteExp:       d.Route(),
		Duration:       &checkDuration,
		WaitCondition: func(ref *istiov1alpha1.RefResource) bool {
			return true // TODO: valid wait
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

func removeSession(d data) error {
	options := session.Options{
		DeploymentName: d.Target(),
		SessionName:    d.Session(),
		NamespaceName:  d.Namespace(),
	}

	c, err := session.DefaultClient(options.NamespaceName)
	if err != nil {
		return err
	}

	_, remove, err := session.RemoveHandler(options, c)
	if err != nil {
		return err
	}
	remove()

	return nil
}

func findLables(ref *istiov1alpha1.RefStatus) map[string]string {
	for _, target := range ref.Targets {
		if *target.Kind == "Deployment" || *target.Kind == "DeploymentConfig" {
			lables := target.Labels
			lables["version"] += "-test" // TODO: dynamically lookup all target labels
			return lables
		}
	}
	return map[string]string{}
}

func findGwHost(ref *istiov1alpha1.RefStatus) string {
	for _, target := range ref.Resources {
		if *target.Kind == "Gateway" {
			return target.Prop["hosts"]
		}
	}
	return ""
}
