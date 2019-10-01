package develop

import (
	"fmt"
	"strings"

	"github.com/maistra/istio-workspace/pkg/cmd/internal/build"

	internal "github.com/maistra/istio-workspace/pkg/cmd/internal/session"

	"github.com/spf13/pflag"

	"github.com/maistra/istio-workspace/pkg/shell"

	"github.com/maistra/istio-workspace/pkg/cmd/config"

	gocmd "github.com/go-cmd/cmd"
	"github.com/spf13/cobra"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("develop")

const telepresenceBin = "telepresence"

var DefaultExclusions = []string{"*.log", ".git/"}

// NewCmd creates instance of "develop" Cobra Command with flags and execution logic defined
func NewCmd() *cobra.Command {
	developCmd := &cobra.Command{
		Use:          "develop",
		Short:        "Starts the development flow",
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			if !shell.BinaryExists(telepresenceBin, "Head over to https://www.telepresence.io/reference/install for installation instructions.\n") {
				return fmt.Errorf("unable to find %s on your $PATH", telepresenceBin)
			}
			return config.SyncFullyQualifiedFlags(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			sessionState, sessionClose, err := internal.Sessions(cmd)
			if err != nil {
				return err
			}
			defer sessionClose()

			// HACK: need contract with TP cmd?
			if err := cmd.Flags().Set("deployment", sessionState.DeploymentName); err != nil {
				return err
			}

			if err := build.Build(cmd); err != nil {
				return err
			}

			done := make(chan gocmd.Status, 1)
			defer close(done)

			arguments := parseArguments(cmd)

			go func() {
				tp := gocmd.NewCmdOptions(shell.StreamOutput, telepresenceBin, arguments...)
				shell.RedirectStreams(tp, cmd.OutOrStdout(), cmd.OutOrStderr(), done)
				shell.ShutdownHook(tp, done)
				shell.Start(tp, done)
			}()

			finalStatus := <-done
			return finalStatus.Error
		},
	}

	developCmd.Flags().StringP("deployment", "d", "", "name of the deployment or deployment config")
	developCmd.Flags().StringP("port", "p", "8000", "port to be exposed in format local[:remote]")
	developCmd.Flags().StringP(build.RunFlagName, "r", "", "command to run your application")
	developCmd.Flags().StringP(build.BuildFlagName, "b", "", "command to build your application before run")
	developCmd.Flags().Bool(build.NoBuildFlagName, false, "always skips build")
	developCmd.Flags().Bool("watch", false, "enables watch")
	developCmd.Flags().StringSliceP("watch-include", "w", []string{"."}, "list of directories to watch (relative to the one from which ike has been started)")
	developCmd.Flags().StringSlice("watch-exclude", DefaultExclusions, fmt.Sprintf("list of patterns to exclude (always excludes %v)", DefaultExclusions))
	developCmd.Flags().Int64("watch-interval", 500, "watch interval (in ms)")
	if err := developCmd.Flags().MarkHidden("watch-interval"); err != nil {
		log.Error(err, "failed while trying to hide a flag")
	}
	developCmd.Flags().Bool("offline", false, "avoid calling external sources")
	if err := developCmd.Flags().MarkHidden("offline"); err != nil {
		log.Error(err, "failed while trying to hide a flag")
	}
	developCmd.Flags().StringP("method", "m", "inject-tcp", "telepresence proxying mode - see https://www.telepresence.io/reference/methods")
	developCmd.Flags().StringP("session", "s", "", "create or join an existing session")
	developCmd.Flags().StringP("route", "", "", "specifies traffic route options in the format of type:name=value. "+
		"Defaults to X-Workspace-Route header with current session name value")
	developCmd.Flags().StringP("namespace", "n", "", "target namespace to develop against "+
		"(defaults to default for the current context)")

	developCmd.Flags().VisitAll(config.BindFullyQualifiedFlag(developCmd))

	_ = developCmd.MarkFlagRequired("deployment")
	_ = developCmd.MarkFlagRequired(build.RunFlagName)

	return developCmd
}

func parseArguments(cmd *cobra.Command) []string {
	run := cmd.Flag(build.RunFlagName).Value.String()
	watch, _ := cmd.Flags().GetBool("watch")
	runArgs := strings.Split(run, " ") // default value

	if watch {
		runArgs = []string{
			"ike", "watch",
			"--dir", stringSliceToCSV(cmd.Flags(), "watch-include"),
			"--exclude", stringSliceToCSV(cmd.Flags(), "watch-exclude"),
			"--interval", cmd.Flag("watch-interval").Value.String(),
			"--" + build.RunFlagName, run,
		}
		if cmd.Flag(build.BuildFlagName).Changed {
			runArgs = append(runArgs, "--"+build.BuildFlagName, cmd.Flag(build.BuildFlagName).Value.String())
		}
	}

	tpCmd := append([]string{
		"--deployment", cmd.Flag("deployment").Value.String(),
		"--expose", cmd.Flag("port").Value.String(),
		"--method", cmd.Flag("method").Value.String(),
		"--run"}, runArgs...)

	namespaceFlag := cmd.Flag("namespace")
	if namespaceFlag.Changed {
		tpCmd = append([]string{"--" + namespaceFlag.Name, namespaceFlag.Value.String()}, tpCmd...)
	}

	return tpCmd
}

func stringSliceToCSV(flags *pflag.FlagSet, name string) string {
	slice, _ := flags.GetStringSlice(name)
	return fmt.Sprintf(`"%s"`, strings.Join(slice, ","))
}
