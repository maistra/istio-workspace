package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/metrics"

	"github.com/aslakknutsen/istio-workspace/cmd/ike/config"
	"github.com/aslakknutsen/istio-workspace/pkg/apis"
	"github.com/aslakknutsen/istio-workspace/pkg/controller"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/spf13/cobra"
	k8sConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var (
	metricsHost       = "0.0.0.0"
	metricsPort int32 = 8383
)

func NewRootCmd() *cobra.Command {
	var configFile string

	rootCmd := &cobra.Command{
		Use:   "ike",
		Short: "ike lets you safely develop and test on prod without a sweat",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			return config.SetupConfigSources(configFile, cmd.Flag("config").Changed)
		},
		RunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			printVersion()
			return startOperator()
		},
	}

	rootCmd.PersistentFlags().
		StringVarP(&configFile, "config", "c", ".ike.config.yaml",
			fmt.Sprintf("config file (supported formats: %s)", strings.Join(config.SupportedExtensions(), ", ")))

	return rootCmd
}

func startOperator() error {
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Error(err, "Failed to get watch namespace")
		return err
	}

	// Get a config to talk to the apiserver
	cfg, err := k8sConfig.GetConfig()
	if err != nil {
		log.Error(err, "")
		return err
	}

	ctx := context.TODO()

	// Become the leader before proceeding
	if e := leader.Become(ctx, "istio-workspace-lock"); e != nil {
		log.Error(e, "")
		return e
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		log.Error(err, "")
		return err
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err = apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		return nil
	}

	// Setup all Controllers
	if err = controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		return err
	}

	// Create Service object to expose the metrics port.
	if _, err = metrics.ExposeMetricsPort(ctx, metricsPort); err != nil {
		log.Info(err.Error())
	}

	log.Info("Starting the operator.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		return err
	}

	return nil
}
