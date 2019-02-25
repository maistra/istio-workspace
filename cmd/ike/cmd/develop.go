package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"

	"github.com/aslakknutsen/istio-workspace/cmd/ike/config"

	gocmd "github.com/go-cmd/cmd"
	"github.com/spf13/cobra"
)

const telepresenceBin = "telepresence"

var excludeLogs = []string{"*.log"}

// NewDevelopCmd creates instance of "develop" Cobra Command with flags and execution logic defined
func NewDevelopCmd() *cobra.Command {

	developCmd := &cobra.Command{
		Use:   "develop",
		Short: "Starts the development flow",

		PreRunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			if !binaryExists(telepresenceBin, "Head over to https://www.telepresence.io/reference/install for installation instructions.\n") {
				return fmt.Errorf("unable to find %s on your $PATH", telepresenceBin)
			}

			return config.SyncFlags(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]

			if err := build(cmd); err != nil {
				return err
			}

			done := make(chan gocmd.Status, 1)
			defer close(done)

			arguments := parseArguments(cmd)

			go func() {
				tp := gocmd.NewCmdOptions(streamOutput, telepresenceBin, arguments...)
				go redirectStreamsToCmd(tp, cmd, done)
				go shutdownHook(tp, done)
				start(tp, done)
			}()

			finalStatus := <-done

			return finalStatus.Error
		},
	}

	developCmd.Flags().StringP("deployment", "d", "", "name of the deployment or deployment config")
	developCmd.Flags().IntP("port", "p", 8000, "port to be exposed")
	developCmd.Flags().StringP(runFlagName, "r", "", "command to run your application")
	developCmd.Flags().StringP(buildFlagName, "b", "", "command to build your application before run")
	developCmd.Flags().Bool(noBuildFlagName, false, "always skips build")
	developCmd.Flags().Bool("watch", false, "enables watch")
	developCmd.Flags().StringSliceP("watch-include", "w", []string{currentDir()}, "list of directories to watch")
	developCmd.Flags().StringSlice("watch-exclude", excludeLogs, "list of patterns to exclude (defaults to telepresence.log which is always excluded)")
	developCmd.Flags().Int64("watch-interval", 500, "watch interval (in ms)")
	if err := developCmd.Flags().MarkHidden("watch-interval"); err != nil {
		log.Error(err, "failed while trying to hide a flag")
	}
	developCmd.Flags().StringP("method", "m", "inject-tcp", "telepresence proxying mode - see https://www.telepresence.io/reference/methods")

	developCmd.Flags().VisitAll(config.BindFullyQualifiedFlag(developCmd))

	_ = developCmd.MarkFlagRequired("deployment")
	_ = developCmd.MarkFlagRequired(runFlagName)

	return developCmd
}

func parseArguments(cmd *cobra.Command) []string {
	run := cmd.Flag(runFlagName).Value.String()
	watch, _ := cmd.Flags().GetBool("watch")
	runArgs := strings.Split(run, " ") // default value

	if watch {
		runArgs = []string{
			"ike", "watch",
			"--dir", flag(cmd.Flags(), "watch-include"),
			"--exclude", flag(cmd.Flags(), "watch-exclude"),
			"--interval", cmd.Flag("watch-interval").Value.String(),
			"--" + runFlagName, run,
		}
		if cmd.Flag(buildFlagName).Changed {
			runArgs = append(runArgs, "--"+buildFlagName, cmd.Flag(buildFlagName).Value.String())
		}
	}

	return append([]string{
		"--swap-deployment", cmd.Flag("deployment").Value.String(),
		"--expose", cmd.Flag("port").Value.String(),
		"--method", cmd.Flag("method").Value.String(),
		"--run"}, runArgs...)
}

func flag(flags *pflag.FlagSet, name string) string {
	slice, _ := flags.GetStringSlice(name)
	return fmt.Sprintf(`"%s"`, strings.Join(slice, ","))
}
