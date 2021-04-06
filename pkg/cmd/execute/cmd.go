package execute

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	gocmd "github.com/go-cmd/cmd"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/maistra/istio-workspace/pkg/cmd/config"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/shell"
	"github.com/maistra/istio-workspace/pkg/watch"
)

const (
	BuildFlagName   = "build"
	NoBuildFlagName = "no-build"
	RunFlagName     = "run"
)

var DefaultExclusions = []string{"*.log", ".git/"}

var logger = func() logr.Logger {
	return log.Log.WithValues("type", "execute")
}

// NewCmd creates execute command which triggers defined build and/or run script
// When --watch is defined it will continuously observe file system changes in the defined set of directories
// and re-runs build and run command when they occur.
// It is hidden (not user facing) as it's integral part of develop command.
func NewCmd() *cobra.Command {
	executeCmd := &cobra.Command{
		Use:          "execute",
		Hidden:       true,
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return config.SyncFullyQualifiedFlags(cmd)
		},
		RunE: execute,
	}

	executeCmd.Flags().StringP(BuildFlagName, "b", "", "command to build your application before run")
	executeCmd.Flags().Bool(NoBuildFlagName, false, "always skips build")
	executeCmd.Flags().StringP(RunFlagName, "r", "", "command to run your application")
	// Watch config
	executeCmd.Flags().Bool("watch", false, "enables watch")
	executeCmd.Flags().StringSliceP("dir", "w", []string{"."}, "list of directories to watch")
	executeCmd.Flags().StringSlice("exclude", DefaultExclusions, "list of patterns to exclude (defaults to telepresence.log which is always excluded)")
	executeCmd.Flags().Int64("interval", 500, "watch interval (in ms)")
	if err := executeCmd.Flags().MarkHidden("interval"); err != nil {
		logger().Error(err, "failed while trying to hide a flag")
	}

	// To enable SIGTERM emulation for testing purposes
	executeCmd.Flags().Bool("kill", false, "to kill during testing")
	if err := executeCmd.Flags().MarkHidden("kill"); err != nil {
		logger().Error(err, "failed while trying to hide a flag")
	}

	return executeCmd
}

func execute(command *cobra.Command, args []string) error {
	watcher := func(restart chan int32) (func(), error) {
		dirs, _ := command.Flags().GetStringSlice("dir")
		excluded, e := command.Flags().GetStringSlice("exclude")
		if e != nil {
			return nil, e
		}
		excluded = append(excluded, DefaultExclusions...)

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

	if w, e := command.Flags().GetBool("watch"); w && e == nil {
		closeWatch, err := watcher(restart)
		if err != nil {
			return err
		}
		defer closeWatch()
	} else if e != nil {
		return e
	}

	go func() {
		for i := range restart {
			if i > 0 { // not initial restart
				kill <- struct{}{}
			}
			go buildAndRun(buildExecutor(command), runExecutor(command), kill)
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
}

type stopper func() error
type executor func() stopper

func buildExecutor(command *cobra.Command) executor {
	return func() stopper {
		buildFlag := command.Flag(BuildFlagName)

		skipBuild, _ := command.Flags().GetBool(NoBuildFlagName)

		shouldRunBuild := buildFlag.Changed && !skipBuild
		if !shouldRunBuild {
			return func() error { return nil }
		}

		buildCmd := command.Flag("build").Value.String()
		buildArgs := strings.Split(buildCmd, " ")

		b := gocmd.NewCmdOptions(shell.StreamOutput, buildArgs[0], buildArgs[1:]...)
		b.Env = os.Environ()
		shell.RedirectStreams(b, command.OutOrStdout(), command.OutOrStderr())
		logger().V(1).Info("starting build command",
			"cmd", b.Name,
			"args", fmt.Sprint(b.Args),
		)

		<-b.Start()
		<-b.Done()

		status := b.Status()
		if status.Error != nil {
			logger().Error(status.Error, "failed to run build command")
		}

		return b.Stop
	}
}

func runExecutor(command *cobra.Command) executor {
	return func() stopper {
		runCmd := command.Flag("run").Value.String()
		runArgs := strings.Split(runCmd, " ")
		r := gocmd.NewCmdOptions(shell.StreamOutput, runArgs[0], runArgs[1:]...)
		r.Env = os.Environ()
		shell.RedirectStreams(r, command.OutOrStdout(), command.OutOrStderr())
		logger().V(1).Info("starting run command",
			"cmd", r.Name,
			"args", fmt.Sprint(r.Args),
		)
		go func(statusChan <-chan gocmd.Status) {
			status := <-statusChan
			if status.Error != nil {
				logger().Error(status.Error, "failed to run run command")
			}
		}(r.Start())

		return r.Stop
	}
}

func buildAndRun(builder, runner executor, kill chan struct{}) {
	_ = builder()
	stopRun := runner()

	<-kill

	if e := stopRun(); e != nil {
		logger().Error(e, "failed while trying to stop RUN process")
	}
}

// simulateSigterm allow us to simulate a SIGTERM when running cobra command inside a test.
// Triggered by setting the command flag "--kill" to true when you want SIGTERM to occur.
func simulateSigterm(command *cobra.Command, testSigtermGuard chan struct{}, hookChan chan os.Signal) {
	const enabled = "true"
	if command.Annotations["test"] != enabled {
		return
	}

	for {
		select {
		case <-testSigtermGuard:
			return
		default:
			if command.Flag("kill").Value.String() == enabled {
				hookChan <- syscall.SIGTERM

				return
			}
		}
	}
}
