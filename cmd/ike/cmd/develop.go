package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var (
	deploymentName string
	port           int
	runnable       string
	method         string
)

func init() {
	developCmd.PersistentFlags().StringVarP(&deploymentName, "deployment", "d", "", "name of the deployment or deployment config")
	developCmd.PersistentFlags().IntVarP(&port, "port", "p", 8000, "port to be exposed")
	developCmd.PersistentFlags().StringVarP(&runnable, "run", "r", "", "command to run your application")
	developCmd.PersistentFlags().StringVarP(&method, "method", "m", "inject-tcp", "telepresence proxying mode - see https://www.telepresence.io/reference/methods")

	_ = developCmd.MarkPersistentFlagRequired("deployment")
	_ = developCmd.MarkPersistentFlagRequired("run")
	rootCmd.AddCommand(developCmd)
}

const telepresenceBin = "telepresence"

var developCmd = &cobra.Command{
	Use:   "develop",
	Short: "starts the development flow",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {

		checkIfTelepresenceExists()

		var tp = exec.Command(telepresenceBin, parseArguments()...)

		redirectStreams(tp)

		err := tp.Wait()

		if err != nil {
			log.Error(err, fmt.Sprintf("%s failed", telepresenceBin))
			os.Exit(1)
		}

	},
}

func redirectStreams(command *exec.Cmd) {
	stdoutIn, _ := command.StdoutPipe()
	stderrIn, _ := command.StderrPipe()
	var stdoutBuf, stderrBuf bytes.Buffer
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	var errStdout, errStderr error
	err := command.Start()
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

func parseArguments() []string {
	runArgs := strings.Split(runnable, " ")
	return append([]string{
		"--swap-deployment", deploymentName,
		"--expose", strconv.Itoa(port),
		"--method", method,
		"--run"}, runArgs...)
}

func checkIfTelepresenceExists() {
	path, err := exec.LookPath(telepresenceBin)
	if err != nil {
		log.Error(err, fmt.Sprintf("Couldn't find '%s' installed in your system.\n"+
			"Head over to https://www.telepresence.io/reference/install for installation instructions.\n", telepresenceBin))
		os.Exit(1)
	} else {
		log.Info(fmt.Sprintf("Found '%s' executable in '%s'.", telepresenceBin, path))
		log.Info(fmt.Sprintf("See '%s.log' for more details about its execution.", telepresenceBin))
	}
}
