package develop

import (
	"fmt"
	"os"

	"emperror.dev/errors"
	gocmd "github.com/go-cmd/cmd"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/maistra/istio-workspace/pkg/cmd/config"
	"github.com/maistra/istio-workspace/pkg/cmd/execute"
	"github.com/maistra/istio-workspace/pkg/cmd/flag"
	internal "github.com/maistra/istio-workspace/pkg/cmd/internal/session"
	"github.com/maistra/istio-workspace/pkg/hook"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/shell"
	"github.com/maistra/istio-workspace/pkg/telepresence"
)

var (
	logger = func() logr.Logger {
		return log.Log.WithValues("type", "develop")
	}

	errorTpNotAvailable = errors.Errorf("unable to find %s on your $PATH", telepresence.BinaryName)

	// Used in the tp-wrapper to check if passed command
	// can be parsed (so has all required flags).
	tpAnnotations = map[string]string{
		"telepresence": "translatable",
	}
)

// NewCmd creates instance of "develop" Cobra Command with flags and execution logic defined.
func NewCmd() *cobra.Command {
	developCmd := &cobra.Command{
		Use:              "develop",
		Short:            "Starts the development flow",
		SilenceUsage:     true,
		TraverseChildren: true,
		Annotations:      tpAnnotations,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !telepresence.BinaryAvailable() {
				return errorTpNotAvailable
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
				return errors.Wrapf(err, "failed to set deployment flag")
			}

			arguments, err := telepresence.CreateTpCommand(cmd)
			if err != nil {
				return errors.Wrap(err, "failed translating to telepresence command")
			}

			done := make(chan gocmd.Status, 1)
			defer close(done)

			go func() {
				tp := gocmd.NewCmdOptions(shell.StreamOutput, telepresence.BinaryName, arguments...)
				tp.Dir = dir
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

	developCmd.Flags().StringP("deployment", "d", "", "name of the deployment or deployment config")
	developCmd.Flags().StringSliceP("port", "p", []string{}, "list of ports to be exposed in format local[:remote].")
	developCmd.Flags().StringP(execute.RunFlagName, "r", "", "command to run your application")
	developCmd.Flags().StringP(execute.BuildFlagName, "b", "", "command to build your application before run")
	developCmd.Flags().Bool(execute.NoBuildFlagName, false, "always skips build")
	developCmd.Flags().Bool("watch", false, "enables watch")
	developCmd.Flags().StringSliceP("watch-include", "w", []string{"."}, "list of directories to watch (relative to the one from which ike has been started)")
	developCmd.Flags().StringSlice("watch-exclude", []string{}, fmt.Sprintf("list of patterns to exclude (always excludes %v)", execute.DefaultExclusions))
	developCmd.Flags().Int64("watch-interval", 500, "watch interval (in ms)")
	if err := developCmd.Flags().MarkHidden("watch-interval"); err != nil {
		logger().Error(err, "failed while trying to hide a flag")
	}
	developCmd.Flags().Bool("offline", false, "avoid calling external sources")
	if err := developCmd.Flags().MarkHidden("offline"); err != nil {
		logger().Error(err, "failed while trying to hide a flag")
	}

	tpMethods := flag.CreateOptions("inject-tcp", "i", "vpn-tcp", "v")
	injectTCP := tpMethods[0]
	developCmd.Flags().VarP(&injectTCP, "method", "m", "telepresence proxying mode - supports inject-tcp and vpn-tcp")
	_ = developCmd.RegisterFlagCompletionFunc("method", flag.CompletionFor(tpMethods))

	developCmd.Flags().StringP("session", "s", "", "create or join an existing session")
	developCmd.Flags().StringP("route", "", "", "specifies traffic route options in the format of type:name=value. "+
		"Defaults to X-Workspace-Route header with current session name value")
	developCmd.Flags().StringP("namespace", "n", "", "target namespace to develop against "+
		"(defaults to default for the current context)")

	developCmd.Flags().VisitAll(config.BindFullyQualifiedFlag(developCmd))

	_ = developCmd.MarkFlagRequired("deployment")
	_ = developCmd.MarkFlagRequired(execute.RunFlagName)

	return developCmd
}
