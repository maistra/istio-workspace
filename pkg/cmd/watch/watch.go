package watch

import (
	"fmt"
	"github.com/maistra/istio-workspace/pkg/cmd/config"
	"github.com/maistra/istio-workspace/pkg/cmd/develop"
	"github.com/maistra/istio-workspace/pkg/cmd/internal/build"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/shell"
	"github.com/maistra/istio-workspace/pkg/watch"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	gocmd "github.com/go-cmd/cmd"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
)

var logger = func() logr.Logger {
	return log.Log.WithValues("type", "watch")
}

// NewCmd creates watch command which observes file system changes in the defined set of directories
// and re-runs build and run command when they occur.
// It is hidden (not user facing) as it's integral part of develop command.
func NewCmd() *cobra.Command {
	watchCmd := &cobra.Command{
		Use:          "watch",
		Hidden:       true,
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return config.SyncFullyQualifiedFlags(cmd)
		},
		RunE: watchForRealChanges,
	}

	watchCmd.Flags().StringP(build.BuildFlagName, "b", "", "command to build your application before run")
	watchCmd.Flags().Bool(build.NoBuildFlagName, false, "always skips build")
	watchCmd.Flags().StringP(build.RunFlagName, "r", "", "command to run your application")
	watchCmd.Flags().StringSliceP("dir", "w", []string{"."}, "list of directories to watch")
	watchCmd.Flags().StringSlice("exclude", develop.DefaultExclusions, "list of patterns to exclude (defaults to telepresence.log which is always excluded)")
	watchCmd.Flags().Int64("interval", 500, "watch interval (in ms)")
	if err := watchCmd.Flags().MarkHidden("interval"); err != nil {
		logger().Error(err, "failed while trying to hide a flag")
	}
	watchCmd.Flags().Bool("kill", false, "to kill during testing")
	if err := watchCmd.Flags().MarkHidden("kill"); err != nil {
		logger().Error(err, "failed while trying to hide a flag")
	}

	return watchCmd
}

func watchForRealChanges(command *cobra.Command, args []string) error {

	// build first
	// execute the --run cmd

	// watch for changes
	// --> build
	// -> restart execution of --run

	watcher := func(restart chan int32) (func(), error) {
		dirs, _ := command.Flags().GetStringSlice("dir")
		excluded, e := command.Flags().GetStringSlice("exclude")
		if e != nil {
			return nil, e
		}
		excluded = append(excluded, develop.DefaultExclusions...)

		ms, _ := command.Flags().GetInt64("interval")
		w, err := watch.CreateWatch(ms).
			WithHandlers(func(events []fsnotify.Event) error {
				for _, event := range events {
					_, _ = command.OutOrStdout().Write([]byte(event.Name + " changed. Restarting process.\n"))
				}
				restart <- 1
				return nil
			}).
			Excluding(excluded...).
			OnPaths(dirs...)

		if err != nil {
			return nil, err
		}

		w.Start()
		return w.Close, nil
	}

	kill := make(chan struct{})
	defer close(kill)

	restart := make(chan int32)
	defer close(restart)

	closeWatch, err := watcher(restart)
	if err != nil {
		return err
	}
	defer closeWatch()

	go func() {
		for i := range restart {
			if i > 0 { // not initial restart
				kill <- struct{}{}
			}
			go buildAndRun(buildExecutor(command), runExecutor(command), kill, nil)
		}
	}()

	hookChan := make(chan os.Signal, 1)
	testSigtermGuard := make(chan struct{})
	defer close(testSigtermGuard)

	go simulateSigterm(command, testSigtermGuard, hookChan)

	signal.Notify(hookChan, os.Interrupt, syscall.SIGTERM)
	defer func() {
		signal.Stop(hookChan)
		close(hookChan)
	}()

	restart <- 0

	<-hookChan

	kill <- struct{}{}

	return nil

	// RULES
	// if first --run fails jump out
	// if first build fails - jump out
	// if subsequent --run fails ... hangs with watch still being there
	// if subsequent build fails ... hangs with watch still being there

	// TODO send sth on the kill channel still -> we need to kill build goroutine which was fired first --> restart being a counter
	// TODO clean up Build.command / string parsing (optional flags --no-build etc)
	// TODO RULES from above?
	// TODO clean up the code

}
type stopper func() error
type executor func() stopper

func buildExecutor(command *cobra.Command) executor {
	return func() stopper {
		buildCmd := command.Flag("build").Value.String() // TODO build-no-build review
		buildArgs := strings.Split(buildCmd, " ")

		b := gocmd.NewCmdOptions(shell.StreamOutput, buildArgs[0], buildArgs[1:]...)
		shell.RedirectStreams(b, command.OutOrStdout(), command.OutOrStderr())
		<-b.Start()
		<-b.Done()

		return func() error { return b.Stop() }
	}
}

func runExecutor(command *cobra.Command) executor {
	return func() stopper {
		runCmd := command.Flag("run").Value.String()
		runArgs := strings.Split(runCmd, " ")
		r := gocmd.NewCmdOptions(shell.StreamOutput, runArgs[0], runArgs[1:]...)
		shell.RedirectStreams(r, command.OutOrStdout(), command.OutOrStderr())
		r.Start()
		return func() error {
			return r.Stop()
		}
	}
}

func buildAndRun(builder, runner executor, kill chan struct{}, status chan gocmd.Status) {

	stopBuild := builder()
	stopRun := runner()

	for {
		select {
		case <-kill:
			if e := stopBuild(); e != nil {
				fmt.Println(e.Error())
			}
			if e := stopRun(); e != nil {
				fmt.Println(e.Error())
			}
			return
		}
	}

}

//
func simulateSigterm(command *cobra.Command, testSigtermGuard chan struct{}, hookChan chan os.Signal) {
	for {
		select {
		case <-testSigtermGuard:
			return
		default:
			if command.Flag("kill").Value.String() == "true" {
				hookChan <- syscall.SIGTERM
				return
			}
		}
	}
}
