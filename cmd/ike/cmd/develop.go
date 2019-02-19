package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/aslakknutsen/istio-workspace/cmd/ike/config"
	"github.com/aslakknutsen/istio-workspace/cmd/ike/watch"

	"github.com/fsnotify/fsnotify"
	gocmd "github.com/go-cmd/cmd"
	"github.com/spf13/cobra"
)

const telepresenceBin = "telepresence"

var excludeTpLog = []string{"telepresence.log"}

var streamOutput = gocmd.Options{
	Buffered:  false,
	Streaming: true,
}

func NewDevelopCmd() *cobra.Command {

	developCmd := &cobra.Command{
		Use:   "develop",
		Short: "Starts the development flow",

		PreRunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			if !telepresenceExists() {
				return fmt.Errorf("unable to find %s on your $PATH", telepresenceBin)
			}

			return config.SyncFlags(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]

			if err := build(cmd); err != nil {
				return err
			}

			done := make(chan gocmd.Status)
			runTp := make(chan struct{})

			if cmd.Flag("watch").Changed {
				slice, _ := cmd.Flags().GetStringSlice("watch")

				excluded, e := cmd.Flags().GetStringSlice("watch-exclude")
				if e != nil {
					return e
				}
				excluded = append(excluded, excludeTpLog...)

				w, err := watch.CreateWatch().
					WithHandler(func(event fsnotify.Event) error {
						_, _ = cmd.OutOrStdout().Write([]byte(event.Name + " changed. Restarting process.\n"))
						if err := build(cmd); err != nil {
							return err
						}
						runTp <- struct{}{}
						return nil
					}).
					Excluding(excluded...).
					OnPaths(slice...)

				if err != nil {
					return err
				}

				defer w.Close()
				w.Watch()
			}

			go func() {
				var tp *gocmd.Cmd
				for {
					<-runTp
					if tp != nil {
						_ = tp.Stop()
					}

					tp = gocmd.NewCmdOptions(streamOutput, telepresenceBin, parseArguments(cmd)...)
					go redirectStreamsToCmd(tp, cmd)
					go notifyTelepresenceOnClose(tp, done)
					go start(tp, done)
				}
			}()

			runTp <- struct{}{}

			finalStatus := <-done
			return finalStatus.Error
		},
	}

	developCmd.Flags().StringP("deployment", "d", "", "name of the deployment or deployment config")
	developCmd.Flags().IntP("port", "p", 8000, "port to be exposed")
	developCmd.Flags().StringP("run", "r", "", "command to run your application")
	developCmd.Flags().StringP("build", "b", "", "command to build your application before run")
	developCmd.Flags().Bool("no-build", false, "always skips build")
	developCmd.Flags().StringSliceP("watch", "w", []string{currentDir()}, "list of directories to watch")
	developCmd.Flags().StringSlice("watch-exclude", excludeTpLog, "list of patterns to exclude (defaults to telepresence.log which is always excluded)")
	developCmd.Flags().StringP("method", "m", "inject-tcp", "telepresence proxying mode - see https://www.telepresence.io/reference/methods")

	developCmd.Flags().VisitAll(config.BindFullyQualifiedFlag(developCmd))

	_ = developCmd.MarkFlagRequired("deployment")
	_ = developCmd.MarkFlagRequired("run")

	return developCmd
}

func start(tp *gocmd.Cmd, done chan gocmd.Status) {
	tp.Env = os.Environ()
	status := <-tp.Start()
	if status.Complete {
		done <- status
	}
}

func notifyTelepresenceOnClose(tp *gocmd.Cmd, done chan gocmd.Status) {
	hookChan := make(chan os.Signal, 1)
	signal.Notify(hookChan, os.Interrupt, syscall.SIGTERM)
	<-hookChan
	var err error
	if tp != nil {
		err = tp.Stop()
	}
	done <- gocmd.Status{
		Error: err,
	}
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

func build(cmd *cobra.Command) error {

	buildFlag := cmd.Flag("build")
	skipBuild, _ := cmd.Flags().GetBool("no-build")
	if buildFlag.Changed && !skipBuild {
		buildCmd := cmd.Flag("build").Value.String()
		buildArgs := strings.Split(buildCmd, " ")
		log.Info("Starting build", "build-cmd", buildCmd)
		build := gocmd.NewCmdOptions(streamOutput, buildArgs[0], buildArgs[1:]...)

		go redirectStreamsToCmd(build, cmd)

		buildStatusChan := build.Start()
		buildStatus := <-buildStatusChan

		if buildStatus.Error != nil {
			return buildStatus.Error
		}
	}

	return nil
}

func currentDir() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	return exPath
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
