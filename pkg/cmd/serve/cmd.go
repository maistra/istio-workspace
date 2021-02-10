package serve

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-lib/leader"

	"github.com/maistra/istio-workspace/api"
	"github.com/maistra/istio-workspace/controllers"
	"github.com/maistra/istio-workspace/pkg/cmd/version"
	"github.com/maistra/istio-workspace/pkg/log"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	k8sConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

const (
	watchNamespaceEnvVar = "WATCH_NAMESPACE"
)

var logger = func() logr.Logger {
	return log.Log.WithValues("type", "serve")
}

var (
	metricsHost       = "0.0.0.0"
	metricsPort int32 = 8080
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
		logger().Error(err, "Failed to get watch namespace")
		return err
	}

	namespaces := strings.Split(namespace, ",")
	logger().Info("Listening for namespaces", "namespaces", namespaces)

	// Get a config to talk to the apiserver
	cfg, err := k8sConfig.GetConfig()
	if err != nil {
		logger().Error(err, "")
		return err
	}

	ctx := context.Background()

	// Become the leader before proceeding
	if e := leader.Become(ctx, "istio-workspace-lock"); e != nil {
		logger().Error(e, "")
		return e
	}

	// Create a new Cmd to provide shared dependencies and Start components
	mgr, err := manager.New(cfg, manager.Options{
		MetricsBindAddress:     fmt.Sprintf("%s:%d", metricsHost, metricsPort),
		HealthProbeBindAddress: "0.0.0.0:8282",
		NewCache:               cache.MultiNamespacedCacheBuilder(namespaces),
	})
	if err != nil {
		logger().Error(err, "")
		return err
	}

	logger().Info("Registering Components.")

	// Setup Scheme for all resources
	if err = api.AddToScheme(mgr.GetScheme()); err != nil {
		logger().Error(err, "")
		return nil
	}

	// Setup all Controllers
	if err = controllers.AddToManager(mgr); err != nil {
		logger().Error(err, "")
		return err
	}

	// add CreateService?

	// Add readiness and health

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logger().Error(err, "Could not add healthz check")
		return err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		logger().Error(err, "Could not add readyz check")
		return err
	}

	logger().Info("Starting the operator.")
	version.LogVersion()

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logger().Error(err, "Manager exited non-zero")
		return err
	}

	return nil
}

// getWatchNamespace returns the namespace the operator should be watching for changes.
func getWatchNamespace() (string, error) {
	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}
	return ns, nil
}
