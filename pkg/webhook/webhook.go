package webhook

import (
	"encoding/json"
	"time"

	istiov1alpha1 "github.com/maistra/istio-workspace/pkg/apis/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/internal/session"
	"github.com/maistra/istio-workspace/pkg/model"
	admission "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
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

func Hook(review admission.AdmissionReview) admission.AdmissionReview {

	// if review.Request.DryRun don't do stuff with sideeffects....

	data, err := parseRequest(review.Request)
	if err != nil {
		// TODO: createErrorReview
	}

	if !data.IsIkeable() {
		review.Response = &admission.AdmissionResponse{
			UID:     review.Request.UID,
			Allowed: true,
		}
		return review
	}

	session, err := createSessionAndWait(data)
	if err != nil {
		// TODO: createErrorReview
	}

	patch := createLabelsPatch(session)

	patchType := admission.PatchTypeJSONPatch
	review.Response = &admission.AdmissionResponse{
		UID:       review.Request.UID,
		Allowed:   true,
		PatchType: &patchType,
		Patch:     []byte(patch),
	}
	return review
}

func parseRequest(request *admission.AdmissionRequest) (data, error) {
	var dep appsv1.Deployment

	err := json.Unmarshal(request.Object.Raw, &dep)
	if err != nil {
		return data{}, err
	}
	return data{
		Object: dep,
	}, nil
}

func createSessionAndWait(d data) (*istiov1alpha1.RefStatus, error) {
	checkDuration := time.Millisecond * 100
	options := session.Options{
		DeploymentName: d.Target(),
		SessionName:    d.Session(),
		NamespaceName:  d.Namespace(),
		Strategy:       model.StrategyExisting,
		Duration:       &checkDuration,
		// TODO: add support for what to wait for WaitExpression: FoundDeployment
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

func createLabelsPatch(ref *istiov1alpha1.RefStatus) string {
	// for each Resource, if Deployment, Make label path
	return ""
}
