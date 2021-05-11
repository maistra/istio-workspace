package session

import (
	"context"
	"os"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-lib/handler"
	errorsK8s "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	istiov1alpha1 "github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/istio"
	"github.com/maistra/istio-workspace/pkg/k8s"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/openshift"
	"github.com/maistra/istio-workspace/pkg/reference"
	"github.com/maistra/istio-workspace/pkg/template"
)

const (
	// Finalizer defines the Finalizer name owned by the Session reconciler.
	Finalizer = "finalizers.istio.workspace.session"
)

var (
	logger = func() logr.Logger {
		return log.Log.WithValues("type", "controller")
	}
)

// DefaultManipulators contains the default config for the reconciler.
func DefaultManipulators() Manipulators {
	var engine template.Engine
	if path, exists := os.LookupEnv(template.TemplatePath); exists {
		engine = template.NewDefaultPatchEngine(path)
	} else {
		engine = template.NewDefaultEngine()
	}

	return Manipulators{
		Locators: []model.Locator{
			k8s.DeploymentLocator,
			openshift.DeploymentConfigLocator,
			k8s.ServiceLocator,
			istio.VirtualServiceGatewayLocator,
		},
		Handlers: []model.Manipulator{
			k8s.DeploymentManipulator(engine),
			openshift.DeploymentConfigManipulator(engine),
			istio.DestinationRuleManipulator(),
			istio.GatewayManipulator(),
			istio.VirtualServiceManipulator(),
		},
	}
}

// Add creates a new Session Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler.
func newReconciler(mgr manager.Manager) *ReconcileSession {
	return &ReconcileSession{client: mgr.GetClient(), scheme: mgr.GetScheme(), manipulators: DefaultManipulators()}
}

// NewStandaloneReconciler returns a new reconcile.Reconciler. Primarily used for unit testing outside of the Manager.
func NewStandaloneReconciler(c client.Client, m Manipulators) *ReconcileSession {
	return &ReconcileSession{client: c, manipulators: m}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler.
func add(mgr manager.Manager, r *ReconcileSession) error {
	// Create a new controller
	c, err := controller.New("session-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return errors.Wrap(err, "failed creating session-controller")
	}

	// Watch for changes to primary resource Session
	err = c.Watch(&source.Kind{Type: &istiov1alpha1.Session{}}, &handler.InstrumentedEnqueueRequestForObject{}, predicate.GenerationChangedPredicate{})
	if err != nil {
		return errors.Wrap(err, "failed creating session-controller")
	}

	for _, object := range r.WatchTypes() {
		if _, err = mgr.GetCache().GetInformer(context.Background(), object); err != nil {
			if !meta.IsNoMatchError(err) {
				logger().Error(err, "error checking for type Kind availability")
			}

			continue
		}

		// Watch for changes to secondary resources
		err = c.Watch(&source.Kind{Type: object}, &reference.EnqueueRequestForAnnotation{
			Type: schema.GroupKind{Group: "maistra.io", Kind: "Session"},
		}, predicate.GenerationChangedPredicate{})
		if err != nil {
			logger().Error(err, "could not add watch on crd")
		}
	}

	return nil
}

// Manipulators holds the basic chain of manipulators that the ReconcileSession will use to perform it's actions.
type Manipulators struct {
	Locators []model.Locator
	Handlers []model.Manipulator
}

var _ reconcile.Reconciler = &ReconcileSession{}

// ReconcileSession reconciles a Session object.
type ReconcileSession struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client       client.Client
	scheme       *runtime.Scheme
	manipulators Manipulators
}

// WatchTypes returns a list of client.Objects to watch for changes.
func (r ReconcileSession) WatchTypes() []client.Object {
	objects := []client.Object{}
	for _, handler := range r.manipulators.Handlers {
		objects = append(objects, handler.TargetResourceType())
	}

	return objects
}

// +kubebuilder:rbac:groups=maistra.io,resources=sessions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=maistra.io,resources=sessions/finalizers,verbs=update
// +kubebuilder:rbac:groups=maistra.io,resources=sessions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get
// +kubebuilder:rbac:groups="",resources=pods;services;endpoints;persistentvolumeclaims;events;configmaps;secrets,verbs=*
// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;replicasets;statefulsets,verbs=*
// +kubebuilder:rbac:groups=apps.openshift.io,resources=deploymentconfigs,verbs=*
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create
// +kubebuilder:rbac:groups=istio.openshift.com,resources=*,verbs=*
// +kubebuilder:rbac:groups=networking.istio.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=maistra.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=apps,resourceNames=istio-workspace,resources=deployments/finalizers,verbs=update

// Reconcile reads that state of the cluster for a Session object and makes changes based on the state read
// and what is in the Session.Spec.
func (r *ReconcileSession) Reconcile(c context.Context, request reconcile.Request) (reconcile.Result, error) {
	reqLogger := logger().WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Session")

	// Fetch the Session instance
	session := &istiov1alpha1.Session{}
	err := r.client.Get(context.Background(), request.NamespacedName, session)
	if err != nil {
		if errorsK8s.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, errors.WrapWithDetails(err, "failed reconciling session", "session", request.Name)
	}

	route := ConvertAPIRouteToModelRoute(session)
	ctx := model.SessionContext{
		Context:   c,
		Name:      request.Name,
		Namespace: request.Namespace,
		Route:     route,
		Log:       reqLogger,
		Client:    r.client,
	}

	// update session.status.Route if it was not provided
	session.Status.Route = ConvertModelRouteToAPIRoute(route)
	session.Status.RouteExpression = session.Status.Route.String()
	if err := r.client.Status().Update(ctx, session); err != nil {
		ctx.Log.Error(err, "Failed to update session.status.route")
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

	refs := ConvertAPIStatusesToModelRefs(*session)
	if deleted {
		r.deleteAllRefs(ctx, session, refs)
	} else {
		r.deleteRemovedRefs(ctx, session, refs)
		if err := r.syncAllRefs(ctx, session); err != nil {
			return reconcile.Result{}, errors.WrapWithDetails(err, "failed reconciling session", "session", request.Name)
		}
	}

	if deleted {
		session.RemoveFinalizer(Finalizer)
		if err := r.client.Update(ctx, session); err != nil {
			ctx.Log.Error(err, "Failed to remove finalizer on session")
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileSession) deleteRemovedRefs(ctx model.SessionContext, session *istiov1alpha1.Session, refs []*model.Ref) {
	for _, ref := range refs {
		found := false
		for _, r := range session.Spec.Refs {
			if ref.KindName.String() == r.Name {
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

func (r *ReconcileSession) deleteAllRefs(ctx model.SessionContext, session *istiov1alpha1.Session, refs []*model.Ref) {
	for _, ref := range refs {
		if err := r.delete(ctx, session, ref); err != nil {
			ctx.Log.Error(err, "Failed to delete session ref", "ref", ref)
		}
	}
}

func (r *ReconcileSession) delete(ctx model.SessionContext, session *istiov1alpha1.Session, ref *model.Ref) error {
	ctx.Log.Info("Remove ref", "name", ref.KindName.String())

	ConvertAPIStatusToModelRef(*session, ref)
	for _, handler := range r.manipulators.Handlers {
		err := handler.Revert()(ctx, ref)
		if err != nil {
			ctx.Log.Error(err, "Revert", "name", ref.KindName.String())
		}
	}
	ConvertModelRefToAPIStatus(*ref, session)

	return errors.Wrap(ctx.Client.Status().Update(ctx, session), "failed updating session")
}

func (r *ReconcileSession) syncAllRefs(ctx model.SessionContext, session *istiov1alpha1.Session) error {
	for _, specRef := range session.Spec.Refs {
		ctx.Log.Info("Add ref", "name", specRef.Name)
		ref := ConvertAPIRefToModelRef(specRef, session.Namespace)
		err := r.sync(ctx, session, &ref)
		if err != nil {
			return err
		}
	}

	session.Status.RefNames = []string{}
	session.Status.Strategies = []string{}
	session.Status.Hosts = []string{}
	for _, statusRef := range session.Status.Refs {
		session.Status.RefNames = append(session.Status.RefNames, statusRef.Name)
		session.Status.Strategies = append(session.Status.Strategies, statusRef.Strategy)
		session.Status.Hosts = append(session.Status.Hosts, statusRef.GetHostNames()...)
	}
	session.Status.Hosts = unique(session.Status.Hosts)

	return errors.Wrap(ctx.Client.Status().Update(ctx, session), "failed syncing all refs")
}

func unique(s []string) []string {
	uniqueSlice := []string{}
	entries := make(map[string]bool)
	for _, entry := range s {
		entries[entry] = true
	}
	for k := range entries {
		uniqueSlice = append(uniqueSlice, k)
	}

	return uniqueSlice
}

func (r *ReconcileSession) sync(ctx model.SessionContext, session *istiov1alpha1.Session, ref *model.Ref) error {
	// if ref has changed, delete first
	if RefUpdated(*session, *ref) {
		err := r.delete(ctx, session, ref)
		if err != nil {
			return err
		}
		time.Sleep(10 * time.Second)
	}

	ConvertAPIStatusToModelRef(*session, ref)
	located := false
	for _, locator := range r.manipulators.Locators {
		if locator(ctx, ref) {
			located = true
		}
	}
	if located {
		for _, handler := range r.manipulators.Handlers {
			err := handler.Mutate()(ctx, ref)
			if err != nil {
				ctx.Log.Error(err, "Mutate", "name", ref.KindName.String())
			}
		}
	}

	ConvertModelRefToAPIStatus(*ref, session)

	return errors.Wrap(ctx.Client.Status().Update(ctx, session), "failed syncing ref")
}
