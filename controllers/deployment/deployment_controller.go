package deployment

import (
	"context"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-lib/handler"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/maistra/istio-workspace/pkg/client/instrumented"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/session"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	// Finalizer defines the Finalizer name owned by the Session reconciler.
	Finalizer  = "deployment.workspace.maistra.io/finalizer"
	IkeTarget  = "ike.target"
	IkeSession = "ike.session"
	IkeRoute   = "ike.route"
	IkeHost    = "ike.host"
)

var (
	logger = func() logr.Logger {
		return log.Log.WithValues("type", "controller")
	}
)

// Add creates a new Session Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler.
func newReconciler(mgr manager.Manager) *ReconcileDeployment {
	return &ReconcileDeployment{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler.
func add(mgr manager.Manager, r *ReconcileDeployment) error {
	// Create a new controller
	c, err := controller.New("deployment-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return errors.Wrap(err, "failed creating deployment-controller")
	}

	// Watch for changes to primary resource Deployment
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.InstrumentedEnqueueRequestForObject{}, predicate.GenerationChangedPredicate{})
	if err != nil {
		return errors.Wrap(err, "failed creating deployment-controller")
	}

	return nil
}

type data struct {
	Object appsv1.Deployment
}

func (d *data) Name() string {
	return d.Object.Name
}

func (d *data) IsIkeable() bool {
	return d.Object.Annotations[IkeTarget] != ""
}

func (d *data) Target() string {
	return d.Object.Annotations[IkeTarget]
}

func (d *data) Session() string {
	return d.Object.Annotations[IkeSession]
}

func (d *data) Namespace() string {
	return d.Object.Namespace
}

func (d *data) Route() string {
	return d.Object.Annotations[IkeRoute]
}

func (d *data) Deleted() bool {
	return d.Object.DeletionTimestamp != nil
}

var _ reconcile.Reconciler = &ReconcileDeployment{}

// ReconcileDeployment reconciles a Session object.
type ReconcileDeployment struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps,resources=deployment,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployment/finalizers,verbs=update
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create
// +kubebuilder:rbac:groups=istio.openshift.com,resources=*,verbs=*

// Reconcile reads that state of the cluster for annotated Deployments object and create Session objects
func (r *ReconcileDeployment) Reconcile(orgCtx context.Context, request reconcile.Request) (reconcile.Result, error) { //nolint:cyclop,gocyclo //reason WIP
	reqLogger := logger().WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Deployment")

	c := instrumented.New(r.client)

	deployment := &appsv1.Deployment{}
	err := c.Get(context.Background(), request.NamespacedName, deployment)
	if err != nil {
		return reconcile.Result{}, nil // assume deletion of non finalized object
	}

	d := data{Object: *deployment}
	if d.IsIkeable() {
		patch := client.MergeFrom(deployment.DeepCopy())

		if d.Deleted() {
			err := removeSession(d)
			if err != nil {
				reqLogger.Error(err, "problems removing session", "deployment", deployment.Name)
				return reconcile.Result{}, err
			}
			controllerutil.RemoveFinalizer(deployment, Finalizer)

		} else {
			state, err := createSessionAndWait(d)
			if err != nil {
				reqLogger.Error(err, "problems creating session", "deployment", deployment.Name)
				return reconcile.Result{}, err
			}

			reqLogger.Info("session created", "session", state.SessionName, "deployment", deployment.Name)

			controllerutil.AddFinalizer(deployment, Finalizer)
			deployment.Annotations[IkeSession] = state.SessionName
			deployment.Annotations[IkeRoute] = state.Route.String()
			deployment.Annotations[IkeHost] = strings.Join(state.Hosts, ",")
		}

		err := c.Patch(orgCtx, deployment, patch)
		if err != nil {
			reqLogger.Error(err, "failed to update deployment", "deployment", deployment.Name)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func createSessionAndWait(d data) (session.State, error) {
	checkDuration := time.Second * 5

	options := session.Options{
		DeploymentName: d.Target(),
		SessionName:    d.Session(),
		NamespaceName:  d.Namespace(),
		Strategy:       model.StrategyExisting,
		StrategyArgs:   map[string]string{"source": d.Name()},
		RouteExp:       d.Route(),
		Duration:       &checkDuration,
	}

	c, err := session.DefaultClient(options.NamespaceName)
	if err != nil {
		return session.State{}, err
	}

	state, remove, err := session.CreateOrJoinHandler(options, c)
	if err != nil {
		remove()
		return session.State{}, err
	}

	return state, nil
}

func removeSession(d data) error {
	options := session.Options{
		SessionName:    d.Session(),
		DeploymentName: d.Target(),
		NamespaceName:  d.Namespace(),
	}

	c, err := session.DefaultClient(options.NamespaceName)
	if err != nil {
		return err
	}

	_, remove := session.RemoveHandler(options, c)
	if err != nil {
		return err
	}
	remove()

	return nil
}
