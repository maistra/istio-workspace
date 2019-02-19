package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/aslakknutsen/istio-workspace/cmd/ike/config"

	gocmd "github.com/go-cmd/cmd"
	"github.com/spf13/cobra"
)

const telepresenceBin = "telepresence"

func NewDevelopCmd() *cobra.Command {

	developCmd := &cobra.Command{
		Use:   "develop",
		Short: "starts the development flow",

		PreRunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			if !telepresenceExists() {
				return fmt.Errorf("unable to find %s on your $PATH", telepresenceBin)
			}

			return config.SyncFlags(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			options := gocmd.Options{
				Buffered:  false,
				Streaming: true,
			}
			tp := gocmd.NewCmdOptions(options, telepresenceBin, parseArguments(cmd)...)

			go redirectStreamsToCmd(tp, cmd)

			tpStatusChan := tp.Start()
			finalStatus := <-tpStatusChan

			return finalStatus.Error
		},
	}

	developCmd.Flags().StringP("deployment", "d", "", "name of the deployment or deployment config")
	developCmd.Flags().IntP("port", "p", 8000, "port to be exposed")
	developCmd.Flags().StringP("run", "r", "", "command to run your application")
	developCmd.Flags().StringP("method", "m", "inject-tcp", "telepresence proxying mode - see https://www.telepresence.io/reference/methods")

	developCmd.Flags().VisitAll(config.BindFullyQualifiedFlag(developCmd))

	_ = developCmd.MarkFlagRequired("deployment")
	_ = developCmd.MarkFlagRequired("run")

	return developCmd
}

func redirectStreamsToCmd(src *gocmd.Cmd, dest *cobra.Command) {
	for {
		select {
		case line, ok := <-src.Stdout:
			if !ok {
				return
			}
			if _, err := fmt.Fprintln(dest.OutOrStdout(), line); err != nil {
				log.Error(err, fmt.Sprintf("%s failed executing", src.Name))
			}
		case line, ok := <-src.Stderr:
			if !ok {
				return
			}
			if _, err := fmt.Fprintln(dest.OutOrStderr(), line); err != nil {
				log.Error(err, fmt.Sprintf("%s failed executing", src.Name))
			}
		}
	}
}

func parseArguments(cmd *cobra.Command) []string {
	run, _ := cmd.Flags().GetString("run")
	runArgs := strings.Split(run, " ")
	return append([]string{
		"--swap-deployment", cmd.Flag("deployment").Value.String(),
		"--expose", cmd.Flag("port").Value.String(),
		"--method", cmd.Flag("method").Value.String(),
		"--run"}, runArgs...)
}

func telepresenceExists() bool {
	path, err := exec.LookPath(telepresenceBin)
	if err != nil {
		log.Error(err, fmt.Sprintf("Couldn't find '%s' installed in your system.\n"+
			"Head over to https://www.telepresence.io/reference/install for installation instructions.\n", telepresenceBin))
		return false
	}

	log.Info(fmt.Sprintf("Found '%s' executable in '%s'.", telepresenceBin, path))
	log.Info(fmt.Sprintf("See '%s.log' for more details about its execution.", telepresenceBin))

	return true
}
