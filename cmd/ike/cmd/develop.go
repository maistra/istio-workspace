package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/aslakknutsen/istio-workspace/cmd/ike/config"

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
			config.SyncFlag(cmd, "deployment")
			config.SyncFlag(cmd, "run")
			config.SyncFlag(cmd, "port")
			config.SyncFlag(cmd, "method")

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) { //nolint[:unparam]
			var tp = exec.Command(telepresenceBin, parseArguments(cmd)...)
			startWithRedirectedStreams(cmd, tp)

			if err := tp.Wait(); err != nil {
				log.Error(err, fmt.Sprintf("%s failed", telepresenceBin))
				os.Exit(1)
			}
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

func startWithRedirectedStreams(cmd *cobra.Command, exCmd *exec.Cmd) {
	stdoutIn, _ := exCmd.StdoutPipe()
	stderrIn, _ := exCmd.StderrPipe()
	stdout := io.MultiWriter(os.Stdout, cmd.OutOrStdout())
	stderr := io.MultiWriter(os.Stderr, cmd.OutOrStderr())
	var errStdout, errStderr error
	err := exCmd.Start()
	if err != nil {
		log.Error(err, fmt.Sprintf("failed to start '%s'", telepresenceBin))
		os.Exit(1)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()
	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
		wg.Done()
	}()

	if errStderr != nil || errStdout != nil {
		log.V(9).Info("Failed to copy either of stdout or stderr")
	}

	wg.Wait()
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
