package develop

import (
	"fmt"
	"os"

	"emperror.dev/errors"
	gocmd "github.com/go-cmd/cmd"
	"github.com/go-logr/logr"
	"github.com/lucasepe/codename"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/maistra/istio-workspace/pkg/cmd/config"
	"github.com/maistra/istio-workspace/pkg/cmd/execute"
	"github.com/maistra/istio-workspace/pkg/cmd/flag"
	internal "github.com/maistra/istio-workspace/pkg/cmd/internal/session"
	"github.com/maistra/istio-workspace/pkg/generator"
	"github.com/maistra/istio-workspace/pkg/hook"
	"github.com/maistra/istio-workspace/pkg/k8s/dynclient"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/shell"
	"github.com/maistra/istio-workspace/pkg/telepresence"
)

var (
	logger = func() logr.Logger {
		return log.Log.WithValues("type", "develop")
	}

	// Used in the tp-wrapper to check if passed command
	// can be parsed (so has all required flags).
	tpAnnotations = map[string]string{
		"telepresence": "translatable",
	}
)

// NewCmd creates instance of "develop" command (and its children) with flags and execution logic defined.
func NewCmd() *cobra.Command {
	developCmd := createDevelopCmd()
	newCmd := createDevelopNewCmd()

	developCmd.AddCommand(newCmd)

	return developCmd
}

func createDevelopCmd() *cobra.Command {
	developCmd := &cobra.Command{
		Use:              "develop",
		Short:            "Starts local development flow by acting like your services runs in the cluster.",
		SilenceUsage:     true,
		TraverseChildren: true,
		Annotations:      tpAnnotations,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := telepresence.BinaryAvailable(); err != nil {
				return errors.Wrapf(err, "Failed starting %s command", cmd.Name())
			}

			return errors.Wrap(config.SyncFullyQualifiedFlags(cmd), "Failed syncing flags")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return errors.Wrap(err, "failed obtaining working directory")
			}
			sessionState, _, sessionClose, err := internal.Sessions(cmd)
			if sessionClose != nil {
				defer sessionClose()
			}
			if err != nil {
				return errors.Wrap(err, "failed setting up session")
			}

			// HACK: need contract with TP cmd?
			if err = cmd.Flags().Set("deployment", sessionState.DeploymentName); err != nil {
				return errors.Wrapf(err, "Failed to set deployment flag")
			}

			arguments, err := telepresence.ParseTpArgs(cmd)
			if err != nil {
				return errors.Wrap(err, "Failed translating to telepresence command")
			}

			done := make(chan gocmd.Status, 1)
			defer close(done)

			go func() {
				tp := telepresence.NewCmdWithOptions(dir, arguments...)
				shell.RedirectStreams(tp, cmd.OutOrStdout(), cmd.OutOrStderr())
				hook.Register(func() error {
					err := tp.Stop()
					if err == nil {
						<-tp.Done()
					}

					return errors.Wrap(err, "failed on telepresence shutdown hook")
				})
				shell.Start(tp, done)
			}()

			if hint, err := Hint(&sessionState); err == nil {
				logger().Info(hint)
			}

			finalStatus := <-done

			return errors.WrapIf(finalStatus.Error, "Failed executing sub command")
		},
	}

	if developCmd.Annotations == nil {
		developCmd.Annotations = map[string]string{}
	}
	developCmd.Annotations[internal.AnnotationRevert] = "true"

	developCmd.PersistentFlags().StringP("deployment", "d", "", "name of the deployment or deployment config")
	developCmd.PersistentFlags().StringSliceP("port", "p", []string{}, "list of ports to be exposed in format local[:remote].")
	developCmd.PersistentFlags().StringP(execute.RunFlagName, "r", "", "command to run your application")
	developCmd.PersistentFlags().StringP(execute.BuildFlagName, "b", "", "command to build your application before run")
	developCmd.PersistentFlags().Bool(execute.NoBuildFlagName, false, "always skips build")
	developCmd.PersistentFlags().Bool("watch", false, "enables watch")
	developCmd.PersistentFlags().StringSliceP("watch-include", "w", []string{"."}, "list of directories to watch (relative to the one from which ike has been started)")
	developCmd.PersistentFlags().StringSlice("watch-exclude", []string{}, fmt.Sprintf("list of patterns to exclude (always excludes %v)", execute.DefaultExclusions))
	developCmd.PersistentFlags().Int64("watch-interval", 500, "watch interval (in ms)")
	if err := developCmd.PersistentFlags().MarkHidden("watch-interval"); err != nil {
		logger().Error(err, "failed while trying to hide a flag")
	}
	developCmd.PersistentFlags().Bool("offline", false, "avoid calling external sources")
	if err := developCmd.PersistentFlags().MarkHidden("offline"); err != nil {
		logger().Error(err, "failed while trying to hide a flag")
	}

	tpMethods := flag.CreateOptions("inject-tcp", "i", "vpn-tcp", "v")
	injectTCP := tpMethods[0]
	developCmd.PersistentFlags().VarP(&injectTCP, "method", "m", "telepresence proxying mode - supports inject-tcp and vpn-tcp")
	_ = developCmd.RegisterFlagCompletionFunc("method", flag.CompletionFor(tpMethods))

	developCmd.PersistentFlags().StringP("session", "s", "", "create or join an existing session")
	developCmd.PersistentFlags().StringP("route", "", "", "specifies traffic route options in the format of type:name=value. "+
		"Defaults to X-Workspace-Route header with current session name value")
	developCmd.PersistentFlags().StringP("namespace", "n", "", "target namespace to develop against "+
		"(defaults to default for the current context)")

	developCmd.Flags().VisitAll(config.BindFullyQualifiedFlag(developCmd))

	_ = developCmd.MarkPersistentFlagRequired("deployment")
	_ = developCmd.MarkPersistentFlagRequired(execute.RunFlagName)

	return developCmd
}

func createDevelopNewCmd() *cobra.Command {
	deploymentTypes := flag.CreateOptions("Deployment", "d", "DeploymentConfig", "dc")
	deploymentType := deploymentTypes[0]

	var createdObj []runtime.Object
	deploymentCleanup := func(client *dynclient.Client) func() error {
		return func() error {
			var err error
			for _, object := range createdObj {
				err = errors.Append(err, client.Delete(object))
			}

			return errors.Wrap(err, "failed cleaning up namespace")
		}
	}

	newCmd := &cobra.Command{
		Use:          "new",
		Short:        "Enables development flow for non-existing service.",
		SilenceUsage: true,
		Annotations:  tpAnnotations,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			name := cmd.Flag("name").Value.String()
			e := cmd.Parent().PersistentFlags().Set("deployment", name+"-v1")

			return errors.Wrapf(e, "Failed populating flags")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ns := cmd.Flag("namespace").Value.String()
			client, err := dynclient.NewDefaultDynamicClient(ns, true)
			if err != nil {
				return errors.Wrap(err, "Failed creating dynamic client")
			}

			var serviceName string
			if cmd.Flag("name").Changed {
				serviceName = cmd.Flag("name").Value.String()
			} else {
				rng, err := codename.DefaultRNG()
				if err != nil {
					panic(err)
				}

				serviceName = codename.Generate(rng, 0)
				fmt.Printf("generated name %s\n", serviceName)
			}

			gateway := cmd.Flag("gateway").Value.String()

			hook.Register(deploymentCleanup(client))

			var collectedErrors error
			basicNewService(serviceName, deploymentType.String(), gateway, ns, func(object runtime.Object) {
				creationErr := client.Create(object) // Create k8s objects on the fly
				createdObj = append(createdObj, object)
				collectedErrors = errors.Append(collectedErrors, creationErr)
			})

			if collectedErrors != nil {
				return errors.Wrap(collectedErrors, "failed creating new service")
			}

			return errors.Wrapf(cmd.Parent().RunE(cmd, args), "failed executing `ike develop` command from `ike develop new`")
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			ns := cmd.Flag("namespace").Value.String()
			client, err := dynclient.NewDefaultDynamicClient(ns, true)
			if err != nil {
				return errors.Wrap(err, "Failed creating dynamic client")
			}

			return deploymentCleanup(client)()
		},
	}

	newCmd.Flags().String("name", "", "defines service/deployment name. if none specified it will be autogenerated.")
	newCmd.Flags().Var(&deploymentType, "type", "defines deployment type, available options are: "+deploymentType.Hint())
	newCmd.Flags().String("gateway", "", "defines gateway name to be created. if none specified it will be skipped")

	_ = newCmd.RegisterFlagCompletionFunc("type", flag.CompletionFor(deploymentTypes))

	return newCmd
}

func basicNewService(name, deploymentType, gateway, ns string, printer generator.Printer) {
	newService := generator.ServiceEntry{Name: name,
		Namespace:      ns,
		DeploymentType: deploymentType,
		Image:          "quay.io/maistra-dev/istio-workspace-test-prepared-prepared-image",
		Gateway:        gateway,
		HTTPPort:       9080,
		GRPCPort:       9081}

	var nsGenerators []generator.SubGenerator
	if gateway != "" {
		nsGenerators = append(nsGenerators, generator.Gateway)
	} else {
		newService.Gateway = "test-gateway" // TMP HACK we assume GW exists
	}

	generator.Generate(
		printer,
		[]generator.ServiceEntry{newService},
		nsGenerators,
		generator.AllSubGenerators,
		generator.WithVersion("v1"),
		generator.GatewayOnHost(generator.GatewayHostFromEnv()),
		generator.ForService(newService, generator.ConnectToGateway(generator.GatewayHostFromEnv())),
	)
}
