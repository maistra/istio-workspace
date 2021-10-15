package deployment

import (
	"context"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-lib/handler"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/maistra/istio-workspace/pkg/log"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	// Finalizer defines the Finalizer name owned by the Session reconciler.
	Finalizer = "finalizers.istio.workspace.deployment"
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

	return reconcile.Result{}, nil
}
