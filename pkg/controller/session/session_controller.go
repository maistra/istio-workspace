package session

import (
	"context"

	istiov1alpha1 "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/aslakknutsen/istio-workspace/pkg/istio"
	"github.com/aslakknutsen/istio-workspace/pkg/k8"
	"github.com/aslakknutsen/istio-workspace/pkg/model"

	"github.com/operator-framework/operator-sdk/pkg/predicate"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	finalizer = "finalizers.istio.workspace.session"
)

var (
	log = logf.Log.WithName("controller_session")

	locators = []model.Locator{
		k8.DeploymentLocator,
		//openshift.DeploymentConfigLocator,
	}
	mutators = []model.Mutator{
		k8.DeploymentMutator,
		//openshift.DeploymentConfigMutator,
		istio.DestinationRuleMutator,
		//istio.VirtualServiceMutator,
	}
	revertors = []model.Revertor{
		k8.DeploymentRevertor,
		//openshift.DeploymentConfigRevertor,
		istio.DestinationRuleRevertor,
		//istio.VirtualServiceRevertor,
	}
)

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

	ctx := model.SessionContext{
		Context:   context.TODO(),
		Name:      request.Name,
		Namespace: request.Namespace,
		Log:       reqLogger,
		Client:    r.client,
	}

	// Fetch the Session instance
	session := &istiov1alpha1.Session{}
	err := r.client.Get(ctx, request.NamespacedName, session)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	deleted := session.DeletionTimestamp != nil
	if deleted {
		reqLogger.Info("Deleted session")
		if !session.HasFinalizer(finalizer) {
			return reconcile.Result{}, nil
		}
	} else {
		reqLogger.Info("Added session")
		if !session.HasFinalizer(finalizer) {
			session.AddFinalizer(finalizer)
			r.client.Update(ctx, session)
		}
	}

	if deleted {
		for _, r := range session.Spec.Refs {
			reqLogger.Info("Remove ref", "name", r)
			ref := model.Ref{Name: r}

			statusToRef(*session, &ref)
			for _, revertor := range revertors {
				err := revertor(ctx, &ref)
				if err != nil {
					reqLogger.Error(err, "Revert", "name", r)
				}
			}
		}
	} else {
		for _, r := range session.Spec.Refs {
			reqLogger.Info("Add ref", "name", r)
			ref := model.Ref{Name: r}

			statusToRef(*session, &ref)
			for _, locator := range locators {
				if locator(ctx, &ref) {
					break // only use first locator
				}
			}
			for _, mutator := range mutators {
				err := mutator(ctx, &ref)
				if err != nil {
					reqLogger.Error(err, "Mutate", "name", r)
				}
			}

			refToStatus(ref, session)
			err := ctx.Client.Status().Update(ctx, session)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	}

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

	if deleted {
		session.RemoveFinalizer(finalizer)
		r.client.Update(ctx, session)
	}
	return reconcile.Result{}, nil
}

func refToStatus(ref model.Ref, session *istiov1alpha1.Session) {
	statusRef := &istiov1alpha1.RefStatus{Name: ref.Name}
	for _, refStat := range ref.ResourceStatuses {
		rs := refStat
		statusRef.Resources = append(statusRef.Resources, &istiov1alpha1.RefResource{Name: &rs.Name, Kind: &rs.Kind})
	}
	// TODO: replace the Ref by name, not just append. Assume the new list contain all ResourceStatus
	session.Status.Refs = append(session.Status.Refs, statusRef)
}

func statusToRef(session istiov1alpha1.Session, ref *model.Ref) {
	for _, statusRef := range session.Status.Refs {
		if statusRef.Name == ref.Name {
			for _, resource := range statusRef.Resources {
				r := resource
				ref.AddResourceStatus(model.ResourceStatus{Name: *r.Name, Kind: *r.Kind})
			}
		}
	}
}
