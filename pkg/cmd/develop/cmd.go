package develop

import (
	"fmt"
	"os"

	"emperror.dev/errors"
	gocmd "github.com/go-cmd/cmd"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/maistra/istio-workspace/pkg/cmd/config"
	"github.com/maistra/istio-workspace/pkg/cmd/execute"
	internal "github.com/maistra/istio-workspace/pkg/cmd/internal/session"
	"github.com/maistra/istio-workspace/pkg/generator"
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
				shell.ShutdownHookForChildCommand(tp)
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
	developCmd.PersistentFlags().StringSlice("watch-exclude", execute.DefaultExclusions, fmt.Sprintf("list of patterns to exclude (always excludes %v)", execute.DefaultExclusions))
	developCmd.PersistentFlags().Int64("watch-interval", 500, "watch interval (in ms)")
	if err := developCmd.PersistentFlags().MarkHidden("watch-interval"); err != nil {
		logger().Error(err, "failed while trying to hide a flag")
	}
	developCmd.PersistentFlags().Bool("offline", false, "avoid calling external sources")
	if err := developCmd.PersistentFlags().MarkHidden("offline"); err != nil {
		logger().Error(err, "failed while trying to hide a flag")
	}
	developCmd.PersistentFlags().StringP("method", "m", "inject-tcp", "telepresence proxying mode - see https://www.telepresence.io/reference/methods")
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
	newCmd := &cobra.Command{
		Use:         "new",
		Short:       "Enables development flow for non-existing service.",
		Annotations: tpAnnotations,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			name := cmd.Flag("name").Value.String()
			e := cmd.Parent().PersistentFlags().Set("deployment", name+"-v1")

			return errors.Wrapf(e, "Failed populating flags")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// service-name (provided by --name flag or autogenerated?)
			// --type deployment|deploymentconfig
			ns := cmd.Flag("namespace").Value.String()
			client, err := dynclient.NewDefaultDynamicClient(ns, true)
			if err != nil {
				return errors.Wrap(err, "Failed creating dynamic client")
			}

			serviceName := cmd.Flag("name").Value.String()

			var collectedErrors error
			basicNewService(serviceName, ns, func(object runtime.Object) {
				creationErr := client.Create(object) // Create k8s objects on the fly
				collectedErrors = errors.Append(collectedErrors, creationErr)
			})

			if collectedErrors != nil {
				fmt.Println(">>>> WE HAVE ERRORS")
				// FIX should return collectedErrors here but fails with admission webhook - investigate
			}

			return errors.Wrapf(cmd.Parent().RunE(cmd, args), "failed executing `ike develop` command from `ike develop new`")
		},
	}

	newCmd.Flags().String("name", "", "defines service/deployment name")
	newCmd.Flags().String("type", "deployment", "deployment/deploymentconfig")

	return newCmd
}

func basicNewService(name, ns string, printer generator.Printer) {
	newService := generator.NewServiceEntry(name, ns,
		"Deployment",
		"quay.io/maistra-dev/istio-workspace-test-prepared-prepared-image")

	generator.Generate(
		printer,
		[]generator.ServiceEntry{newService},
		generator.AllSubGenerators,
		generator.WithVersion("v1"),
		generator.ForService(newService, generator.ConnectToGateway(generator.GatewayHost)),
	)
}
