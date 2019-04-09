package session

import (
	"context"

	istiov1alpha1 "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/aslakknutsen/istio-workspace/pkg/istio"
	"github.com/aslakknutsen/istio-workspace/pkg/k8s"
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
	Finalizer = "finalizers.istio.workspace.session"
)

var (
	log = logf.Log.WithName("controller_session")
)

// defaultManipulators contains the default config for the reconciler
func defaultManipulators() Manipulators {
	return Manipulators{
		Locators: []model.Locator{
			k8s.DeploymentLocator,
			//openshift.DeploymentConfigLocator,
		},
		Mutators: []model.Mutator{
			k8s.DeploymentMutator,
			//openshift.DeploymentConfigMutator,
			istio.DestinationRuleMutator,
			istio.VirtualServiceMutator,
		},
		Revertors: []model.Revertor{
			k8s.DeploymentRevertor,
			//openshift.DeploymentConfigRevertor,
			istio.DestinationRuleRevertor,
			istio.VirtualServiceRevertor,
		},
	}
}

// Add creates a new Session Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSession{client: mgr.GetClient(), scheme: mgr.GetScheme(), manipulators: defaultManipulators()}
}

// NewStandaloneReconciler returns a new reconcile.Reconciler. Primarily used for unit testing outside of the Manager
func NewStandaloneReconciler(c client.Client, m Manipulators) reconcile.Reconciler {
	return &ReconcileSession{client: c, manipulators: m}
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

// Manipulators holds the basic chain of manipulators that the ReconcileSession will use to perform it's actions
type Manipulators struct {
	Locators  []model.Locator
	Mutators  []model.Mutator
	Revertors []model.Revertor
}

var _ reconcile.Reconciler = &ReconcileSession{}

// ReconcileSession reconciles a Session object
type ReconcileSession struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client       client.Client
	scheme       *runtime.Scheme
	manipulators Manipulators
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
		if !session.HasFinalizer(Finalizer) {
			return reconcile.Result{}, nil
		}
	} else {
		reqLogger.Info("Added session")
		if !session.HasFinalizer(Finalizer) {
			session.AddFinalizer(Finalizer)
			if err := r.client.Update(ctx, session); err != nil {
				ctx.Log.Error(err, "Failed to add finalizer on session")
			}
		}
	}

	refs := statusesToRef(*session)
	if deleted {
		r.deleteAllRefs(ctx, session, refs)
	} else {
		if err := r.syncAllRefs(ctx, session); err != nil {
			return reconcile.Result{}, err
		}
		r.deleteRemovedRefs(ctx, session, refs)
	}

	if deleted {
		session.RemoveFinalizer(Finalizer)
		if err := r.client.Update(ctx, session); err != nil {
			ctx.Log.Error(err, "Failed to remove finalizer on session")
		}

	}
	return reconcile.Result{}, nil
}

func (r *ReconcileSession) deleteRemovedRefs(ctx model.SessionContext, session *istiov1alpha1.Session, refs []*model.Ref) { //nolint[:hugeParam]
	for _, ref := range refs {
		found := false
		for _, r := range session.Spec.Refs {
			if ref.Name == r {
				found = true
				break
			}
		}
		if !found {
			if err := r.delete(ctx, session, ref); err != nil {
				ctx.Log.Error(err, "Failed to delete session ref", "ref", ref)
			}
		}
	}
}
func (r *ReconcileSession) deleteAllRefs(ctx model.SessionContext, session *istiov1alpha1.Session, refs []*model.Ref) { //nolint[:hugeParam]
	for _, ref := range refs {
		if err := r.delete(ctx, session, ref); err != nil {
			ctx.Log.Error(err, "Failed to delete session ref", "ref", ref)
		}
	}
}

func (r *ReconcileSession) delete(ctx model.SessionContext, session *istiov1alpha1.Session, ref *model.Ref) error { //nolint[:hugeParam]
	ctx.Log.Info("Remove ref", "name", ref.Name)

	statusToRef(*session, ref)
	for _, revertor := range r.manipulators.Revertors {
		err := revertor(ctx, ref)
		if err != nil {
			ctx.Log.Error(err, "Revert", "name", ref.Name)
		}
	}
	refToStatus(*ref, session)
	return ctx.Client.Status().Update(ctx, session)
}

func (r *ReconcileSession) syncAllRefs(ctx model.SessionContext, session *istiov1alpha1.Session) error { //nolint[:hugeParam]
	for _, sref := range session.Spec.Refs {
		ctx.Log.Info("Add ref", "name", sref)
		ref := model.Ref{Name: sref}
		err := r.sync(ctx, session, &ref)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileSession) sync(ctx model.SessionContext, session *istiov1alpha1.Session, ref *model.Ref) error { //nolint[:hugeParam]
	statusToRef(*session, ref)
	for _, locator := range r.manipulators.Locators {
		if locator(ctx, ref) {
			break // only use first locator
		}
	}
	for _, mutator := range r.manipulators.Mutators {
		err := mutator(ctx, ref)
		if err != nil {
			ctx.Log.Error(err, "Mutate", "name", ref.Name)
		}
	}

	refToStatus(*ref, session)
	return ctx.Client.Status().Update(ctx, session)
}

func refToStatus(ref model.Ref, session *istiov1alpha1.Session) {
	statusRef := &istiov1alpha1.RefStatus{Name: ref.Name}
	for _, refStat := range ref.ResourceStatuses {
		rs := refStat
		action := string(rs.Action)
		statusRef.Resources = append(statusRef.Resources, &istiov1alpha1.RefResource{Name: &rs.Name, Kind: &rs.Kind, Action: &action})
	}
	var existsInStatus bool
	for i, statRef := range session.Status.Refs {
		if statRef.Name == statusRef.Name {
			if len(statusRef.Resources) == 0 { // Remove
				session.Status.Refs = append(session.Status.Refs[:i], session.Status.Refs[i+1:]...)
			} else { // Update
				session.Status.Refs[i] = statusRef
			}
			existsInStatus = true
			break
		}
	}
	if !existsInStatus {
		session.Status.Refs = append(session.Status.Refs, statusRef)
	}
}

func statusesToRef(session istiov1alpha1.Session) []*model.Ref { //nolint[:hugeParam]
	refs := []*model.Ref{}
	for _, statusRef := range session.Status.Refs {
		r := &model.Ref{Name: statusRef.Name}
		statusToRef(session, r)
		refs = append(refs, r)
	}
	return refs
}

func statusToRef(session istiov1alpha1.Session, ref *model.Ref) { //nolint[:hugeParam]
	for _, statusRef := range session.Status.Refs {
		if statusRef.Name == ref.Name {
			for _, resource := range statusRef.Resources {
				r := resource
				ref.AddResourceStatus(model.ResourceStatus{Name: *r.Name, Kind: *r.Kind, Action: model.ResourceAction(*r.Action)})
			}
		}
	}
}
