package session

import (
	"context"
	"fmt"
	"strings"

	istionetwork "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/networking/v1alpha3"
	istiov1alpha1 "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/v1alpha1"
	yaml "gopkg.in/yaml.v2"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-sdk/pkg/predicate"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_session")

// Add creates a new Session Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSession{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("session-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Session
	err = c.Watch(&source.Kind{Type: &istiov1alpha1.Session{}}, &handler.EnqueueRequestForObject{}, predicate.GenerationChangedPredicate{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileSession{}

// ReconcileSession reconciles a Session object
type ReconcileSession struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Session object and makes changes based on the state read
// and what is in the Session.Spec
func (r *ReconcileSession) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Session")

	ctx := context.TODO()

	// Fetch the Session instance
	instance := &istiov1alpha1.Session{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	reqLogger.Info("Added session", "name", request.Name, "namespace", request.Namespace)

	// locateDeployment()
	// locateDestinationRule()
	// locateVirtualService()
	// cloneDeployment()
	// mutateDestinationRule()
	// mutateVirtualService()
	// updateStatus
	deployment := appsv1.Deployment{}

	err = r.client.Get(ctx, types.NamespacedName{Namespace: request.Namespace, Name: instance.Spec.Ref}, &deployment)
	if err != nil {
		updateStatus(ctx, reqLogger, r.client, setStatus(instance, fmt.Sprintf("%v", err)))
		return reconcile.Result{Requeue: false}, err
	}

	reqLogger.Info("Found Deployment", "image", deployment.Spec.Template.Spec.Containers[0].Image)

	// HACK: resolve the chain somehow

	targetName := strings.Split(deployment.Name, "-")[0]

	/*
		serviceList := corev1.ServiceList{}
		err = r.client.List(ctx, &client.ListOptions{Namespace: request.Namespace}, &serviceList)
		if err != nil {
			updateStatus(reqLogger, ctx, r.client, instance, fmt.Sprintf("%v", err))
			return reconcile.Result{Requeue: false}, err
		}
		for _, vs := range serviceList.Items {
			reqLogger.Info("Found Service", "name", vs.ObjectMeta.Name, "labels", vs.Labels)
		}
	*/

	destRule := istionetwork.DestinationRule{}
	err = r.client.Get(ctx, types.NamespacedName{Namespace: request.Namespace, Name: targetName}, &destRule)
	if err != nil {
		updateStatus(ctx, reqLogger, r.client, setStatus(instance, fmt.Sprintf("%v", err)))
		return reconcile.Result{Requeue: false}, err
	}
	reqLogger.Info("Found DestinationRule", "name", destRule.Name)
	{
		x, _ := yaml.Marshal(destRule)
		fmt.Println(string(x))
	}

	virtService := istionetwork.VirtualService{}
	err = r.client.Get(ctx, types.NamespacedName{Namespace: request.Namespace, Name: targetName}, &virtService)
	if err != nil {
		updateStatus(ctx, reqLogger, r.client, setStatus(instance, fmt.Sprintf("%v", err)))
		return reconcile.Result{Requeue: false}, err
	}
	reqLogger.Info("Found VirtualService", "name", virtService.Name)

	destMutator := DestinationRuleMutator{}
	mutatedDestRule, err := destMutator.Add(destRule)
	if err != nil {
		updateStatus(ctx, reqLogger, r.client, setStatus(instance, fmt.Sprintf("%v", err)))
		return reconcile.Result{Requeue: false}, err
	}
	/*
		virtServiceMutator := VirtualServiceMutator{}
		mutatedVirtService, err := virtServiceMutator.Add(virtService)
		if err != nil {
			updateStatus(reqLogger, ctx, r.client, setStatus(instance, fmt.Sprintf("%v", err)))
			return reconcile.Result{Requeue: false}, err
		}
	*/

	refStatus := &istiov1alpha1.RefStatus{
		Params: map[string]string{
			"end-user": "jason",
		},
	}
	instance.Status.Refs = append(instance.Status.Refs, refStatus)

	err = r.client.Update(ctx, &mutatedDestRule)
	if err != nil {
		updateStatus(ctx, reqLogger, r.client, setStatus(instance, fmt.Sprintf("%v", err)))
		return reconcile.Result{Requeue: false}, err
	}

	refStatus.Resources = append(refStatus.Resources, &istiov1alpha1.RefResource{Kind: &mutatedDestRule.TypeMeta.Kind, Name: &mutatedDestRule.Name})
	updateStatus(ctx, reqLogger, r.client, instance)
	/*
		err = r.client.Update(ctx, &mutatedVirtService)
		if err != nil {
			updateStatus(reqLogger, ctx, r.client, setStatus(instance, fmt.Sprintf("%v", err)))
			return reconcile.Result{Requeue: false}, err
		}
	*/
	updateStatus(ctx, reqLogger, r.client, setStatus(instance, "success"))

	return reconcile.Result{}, nil
}

func setStatus(session *istiov1alpha1.Session, status string) *istiov1alpha1.Session {
	session.Status.State = &status
	return session
}

func updateStatus(ctx context.Context, log logr.Logger, c client.Client, session *istiov1alpha1.Session) {
	log.Info("Updating status", "state", session.Status.State)
	err := c.Status().Update(ctx, session)
	if err != nil {
		log.Error(err, "failed to update status of session")
	}
}
