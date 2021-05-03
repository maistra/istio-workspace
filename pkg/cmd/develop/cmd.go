package develop

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gocmd "github.com/go-cmd/cmd"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/maistra/istio-workspace/pkg/cmd/config"
	"github.com/maistra/istio-workspace/pkg/cmd/execute"
	internal "github.com/maistra/istio-workspace/pkg/cmd/internal/session"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/shell"
	"github.com/maistra/istio-workspace/pkg/telepresence"
)

var logger = func() logr.Logger {
	return log.Log.WithValues("type", "develop")
}

var errorTpNotAvailable = errors.Errorf("unable to find %s on your $PATH", telepresence.BinaryName)

// NewCmd creates instance of "develop" Cobra Command with flags and execution logic defined.
func NewCmd() *cobra.Command {
	developCmd := &cobra.Command{
		Use:          "develop",
		Short:        "Starts the development flow",
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !telepresence.BinaryAvailable() {
				return errorTpNotAvailable
			}

			return errors.Wrap(config.SyncFullyQualifiedFlags(cmd), "failed syncing flags")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return errors.Wrapf(err, "failed executing %s command - obtaining working directory", cmd.Use)
			}
			sessionState, _, sessionClose, err := internal.Sessions(cmd)
			if err != nil {
				return errors.Wrapf(err, "failed executing %s command", cmd.Use)
			}
			defer sessionClose()

			// HACK: need contract with TP cmd?
			if err := cmd.Flags().Set("deployment", sessionState.DeploymentName); err != nil {
				return errors.Wrapf(err, "failed executing %s command", cmd.Use)
			}

			done := make(chan gocmd.Status, 1)
			defer close(done)

			arguments := createTpCommand(cmd)

			go func() {
				tp := gocmd.NewCmdOptions(shell.StreamOutput, telepresence.BinaryName, arguments...)
				tp.Dir = dir
				shell.RedirectStreams(tp, cmd.OutOrStdout(), cmd.OutOrStderr())
				shell.ShutdownHookForChildCommand(tp)
				shell.Start(tp, done)
			}()

			if hint, err := Hint(&sessionState.RefStatus, &sessionState.Route); err == nil {
				logger().Info(hint)
			}

			finalStatus := <-done

			return errors.Wrapf(finalStatus.Error, "failed executing %s command", cmd.Use)
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
	developCmd.Flags().StringSlice("watch-exclude", execute.DefaultExclusions, fmt.Sprintf("list of patterns to exclude (always excludes %v)", execute.DefaultExclusions))
	developCmd.Flags().Int64("watch-interval", 500, "watch interval (in ms)")
	if err := developCmd.Flags().MarkHidden("watch-interval"); err != nil {
		logger().Error(err, "failed while trying to hide a flag")
	}
	developCmd.Flags().Bool("offline", false, "avoid calling external sources")
	if err := developCmd.Flags().MarkHidden("offline"); err != nil {
		logger().Error(err, "failed while trying to hide a flag")
	}
	developCmd.Flags().StringP("method", "m", "inject-tcp", "telepresence proxying mode - see https://www.telepresence.io/reference/methods")
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

func createTpCommand(cmd *cobra.Command) []string {
	tpArgs := []string{
		"--deployment", cmd.Flag("deployment").Value.String(),
		"--method", cmd.Flag("method").Value.String(),
	}
	if cmd.Flags().Changed("port") {
		ports, _ := cmd.Flags().GetStringSlice("port") // ignore error, should only occur if flag does not exist. If it doesn't, it won't be Changed()
		for _, port := range ports {
			tpArgs = append(tpArgs, "--expose", port)
		}
	}

	tpArgs = append(tpArgs, "--run")
	tpCmd := append(tpArgs, createWrapperCmd(cmd)...)

	namespaceFlag := cmd.Flag("namespace")
	if namespaceFlag.Changed {
		tpCmd = append([]string{"--" + namespaceFlag.Name, namespaceFlag.Value.String()}, tpCmd...)
	}

	return tpCmd
}

func createWrapperCmd(cmd *cobra.Command) []string {
	run := cmd.Flag(execute.RunFlagName).Value.String()
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	cmdFullPath := dir + string(os.PathSeparator) + "ike"
	executeArgs := []string{
		cmdFullPath, "execute",
		"--" + execute.RunFlagName, run,
	}
	if cmd.Flag(execute.NoBuildFlagName).Changed {
		executeArgs = append(executeArgs, "--"+execute.NoBuildFlagName, cmd.Flag(execute.NoBuildFlagName).Value.String())
	}
	if cmd.Flag(execute.BuildFlagName).Changed {
		executeArgs = append(executeArgs, "--"+execute.BuildFlagName, cmd.Flag(execute.BuildFlagName).Value.String())
	}

	watch, _ := cmd.Flags().GetBool("watch")
	if watch {
		executeArgs = append(executeArgs,
			"--watch",
			"--dir", stringSliceToCSV(cmd.Flags(), "watch-include"),
			"--exclude", stringSliceToCSV(cmd.Flags(), "watch-exclude"),
			"--interval", cmd.Flag("watch-interval").Value.String(),
		)
	}

	return executeArgs
}

func stringSliceToCSV(flags *pflag.FlagSet, name string) string {
	slice, _ := flags.GetStringSlice(name)

	return fmt.Sprintf(`"%s"`, strings.Join(slice, ","))
}
