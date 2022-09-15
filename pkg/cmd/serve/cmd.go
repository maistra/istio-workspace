package serve

import (
	"fmt"
	"os"
	"strings"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	k8sConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/maistra/istio-workspace/api"
	"github.com/maistra/istio-workspace/controllers"
	"github.com/maistra/istio-workspace/pkg/cmd/version"
	"github.com/maistra/istio-workspace/pkg/log"
)

const (
	watchNamespaceEnvVar       = "WATCH_NAMESPACE"
	metricsHost                = "0.0.0.0"
	metricsPort          int32 = 8080
)

var (
	errorWatchNsNotFound = fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	logger               = func() logr.Logger {
		return log.Log.WithValues("type", "serve")
	}
)

// NewCmd creates instance of "ike serve" Cobra Command which is intended to be ran in the
// cluster as it starts istio-workspace operator.
func NewCmd() *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Starts istio-workspace operator in the cluster",
		RunE:  startOperator,
	}

	return serveCmd
}

func startOperator(cmd *cobra.Command, args []string) error {
	namespace, err := getWatchNamespace()
	if err != nil {
		return errors.Wrapf(err, "failed to get watch namespace")
	}

	namespaces := strings.Split(namespace, ",")
	logger().Info("Listening for namespaces", "namespaces", namespaces)

	// Get a config to talk to the apiserver
	cfg, err := k8sConfig.GetConfig()
	if err != nil {
		return errors.Wrapf(err, "could not get kube config")
	}

	managerOptions := manager.Options{
		MetricsBindAddress:     fmt.Sprintf("%s:%d", metricsHost, metricsPort),
		HealthProbeBindAddress: "0.0.0.0:8282",
		LeaderElection:         true,
		LeaderElectionID:       "istio-workspace-lock",
	}

	if len(namespaces) == 1 {
		managerOptions.Namespace = namespaces[0]
	} else {
		managerOptions.NewCache = cache.MultiNamespacedCacheBuilder(namespaces)
	}

	// Create a new Cmd to provide shared dependencies and Start components
	mgr, err := manager.New(cfg, managerOptions)
	if err != nil {
		return errors.Wrapf(err, "failed creating manager")
	}

	logger().Info("Registering Components.")

	// Setup Scheme for all resources
	if err = api.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "failed to add scheme")
	}

	// Setup all Controllers
	if err = controllers.AddToManager(mgr); err != nil {
		return errors.Wrapf(err, "failed to add controller to manager")
	}

	// add CreateService?

	// Add readiness and health
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return errors.Wrapf(err, "could not add healthz check")
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return errors.Wrapf(err, "could not add readyz check")
	}

	logger().Info("Starting the operator.")
	version.LogVersion()

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		return errors.Wrapf(err, "manager exited non-zero")
	}

	return nil
}

// getWatchNamespace returns the namespace the operator should be watching for changes.
func getWatchNamespace() (string, error) {
	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", errorWatchNsNotFound
	}

	return ns, nil
}
