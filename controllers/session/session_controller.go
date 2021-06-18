package session

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
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
	n "github.com/maistra/istio-workspace/pkg/model/new"
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
		Locators: []n.Locator{
			k8s.DeploymentLocator,
			openshift.DeploymentConfigLocator,
			k8s.ServiceLocator,
			istio.VirtualServiceLocator,
			istio.DestinationRuleLocator,
			istio.VirtualServiceGatewayLocator,
		},
		Handlers: []n.ModificatorRegistrar{
			k8s.DeploymentRegistrar(engine),
			openshift.DeploymentConfigRegistrar(engine),
			istio.DestinationRuleRegistrar,
			istio.GatewayRegistrar,
			istio.VirtualServiceRegistrar,
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
	Locators []n.Locator
	Handlers []n.ModificatorRegistrar
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
	for _, h := range r.manipulators.Handlers {
		object, _ := h()
		objects = append(objects, object)
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
func (r *ReconcileSession) Reconcile(c context.Context, request reconcile.Request) (reconcile.Result, error) { //nolint:cyclop,gocyclo //reason WIP
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
	ctx := n.SessionContext{
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
	processing := istiov1alpha1.StateProcessing
	session.Status.State = &processing
	err = r.client.Status().Update(ctx, session)
	if err != nil {
		ctx.Log.Error(err, "Failed to update session.status.route")
	}

	extractModificators := func(registrars []n.ModificatorRegistrar) []n.Modificator {
		var mods []n.Modificator
		for _, reg := range registrars {
			_, mod := reg()
			mods = append(mods, mod)
		}

		return mods
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
			err = r.client.Update(ctx, session)
			if err != nil {
				ctx.Log.Error(err, "Failed to add finalizer on session")
			}
		}
	}

	refs := calculateReferences(ctx, session)
	sync := n.EngineImpl(r.manipulators.Locators, extractModificators(r.manipulators.Handlers))
	session.Status.Conditions = []*istiov1alpha1.Condition{}

	for _, ref := range refs {
		ref := ref // pin
		sync(ctx, ref,
			func(located n.LocatorStatusStore) bool {
				// TODO validate stuff
				return true
			},
			func(located n.LocatorStatusStore) {
				/* updateComponents(session.components + unique(located)) */
				for _, stored := range located() {
					fmt.Println("located: ", stored)
				}
			},
			func(modified n.ModificatorStatus) {
				/* updateComponent() && callEventAPI() */

				addCondition(session, ref, &modified)
				err = ctx.Client.Status().Update(ctx, session)
				if err != nil {
					ctx.Log.Error(err, "could not update session", "name", session.Name, "namespace", session.Namespace)
				}
			})
		cleanupRelatedConditionsOnRemoval(ref, session)
	}
	session.Status.State = calculateSessionState(session)
	err = ctx.Client.Status().Update(ctx, session)
	if err != nil {
		ctx.Log.Error(err, "could not update session", "name", session.Name, "namespace", session.Namespace)
	}

	if deleted {
		if *session.Status.State == istiov1alpha1.StateSuccess {
			session.RemoveFinalizer(Finalizer)
			if err := r.client.Update(ctx, session); err != nil {
				ctx.Log.Error(err, "Failed to remove finalizer on session")
			}
		}

		return reconcile.Result{RequeueAfter: 1 * time.Second}, nil
	}

	return reconcile.Result{}, nil
}

func cleanupRelatedConditionsOnRemoval(ref n.Ref, session *istiov1alpha1.Session) {
	if ref.Deleted && refSuccessful(ref, session.Status.Conditions) {
		var otherConditions []*istiov1alpha1.Condition
		for i := range session.Status.Conditions {
			condition := session.Status.Conditions[i]
			if condition.Source.Ref != ref.KindName.String() {
				otherConditions = append(otherConditions, condition)
			}
		}
		session.Status.Conditions = otherConditions
	}
}

func refSuccessful(ref n.Ref, conditions []*istiov1alpha1.Condition) bool {
	for i := range conditions {
		condition := conditions[i]
		conditionFailed := condition.Status != nil && *condition.Status == istiov1alpha1.StatusFailed
		if condition.Source.Ref == ref.KindName.String() && conditionFailed {
			return false
		}
	}

	return true
}

func calculateSessionState(session *istiov1alpha1.Session) *istiov1alpha1.SessionState {
	state := istiov1alpha1.StateSuccess
	for _, con := range session.Status.Conditions {
		if con.Status != nil && *con.Status == istiov1alpha1.StatusFailed {
			state = istiov1alpha1.StateFailed

			break
		}
	}

	return &state
}

func addCondition(session *istiov1alpha1.Session, ref n.Ref, modified *n.ModificatorStatus) {
	getReason := func(a n.StatusAction) string {
		switch a {
		case n.ActionCreate, n.ActionDelete, n.ActionLocated:

			return "Handled"
		case n.ActionModify, n.ActionRevert:

			return "Configured"
		}

		return ""
	}

	message := modified.Kind + "/" + modified.Name + " modified to satisfy " + ref.KindName.String() + ": "
	if modified.Error != nil {
		message += modified.Error.Error()
	} else {
		message += "ok"
	}
	var target *istiov1alpha1.Target
	if modified.Target != nil {
		target = &istiov1alpha1.Target{
			Kind: modified.Target.Kind,
			Name: modified.Target.Name,
		}
	}

	reason := getReason(modified.Action)
	status := strconv.FormatBool(modified.Success)
	typeStr := string(modified.Action)
	session.AddCondition(istiov1alpha1.Condition{
		Source: istiov1alpha1.Source{
			Kind: modified.Kind,
			Name: modified.Name,
			Ref:  ref.KindName.String(),
		},
		Target:  target,
		Message: &message,
		Reason:  &reason,
		Status:  &status,
		Type:    &typeStr,
	})
}

func calculateReferences(ctx n.SessionContext, session *istiov1alpha1.Session) []n.Ref {
	refs := []n.Ref{}
	for _, ref := range session.Spec.Refs {
		modelRef := ConvertAPIRefToModelRef(ref, ctx.Namespace)
		modelRef.Deleted = session.DeletionTimestamp != nil
		refs = append(refs, modelRef)
	}

	uniqueOldRefs := make(map[string]bool, 2)
	for _, condition := range session.Status.Conditions {
		uniqueOldRefs[condition.Source.Ref] = true
	}
	for key := range uniqueOldRefs {
		found := false
		for _, ref := range refs {
			if ref.KindName.String() == key {
				found = true

				break
			}
		}
		if !found {
			deletedRef := n.Ref{KindName: n.ParseRefKindName(key)}
			deletedRef.Deleted = true
			refs = append(refs, deletedRef)
		}
	}

	sort.SliceStable(refs, func(i, j int) bool {
		if refs[i].Deleted && !refs[j].Deleted {
			return true
		}
		if !refs[i].Deleted && refs[j].Deleted {
			return false
		}

		return true
	})

	return refs
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
