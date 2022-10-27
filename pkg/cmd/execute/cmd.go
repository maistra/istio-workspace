package execute

import (
	"fmt"
	"os"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/fsnotify/fsnotify"
	gocmd "github.com/go-cmd/cmd"
	"github.com/go-logr/logr"
	"github.com/maistra/istio-workspace/pkg/cmd/config"
	"github.com/maistra/istio-workspace/pkg/hook"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/shell"
	"github.com/maistra/istio-workspace/pkg/watch"
	"github.com/spf13/cobra"
)

const (
	// BuildFlagName is a name of the flag defining build process.
	BuildFlagName = "build"
	// NoBuildFlagName is a nme of the flag which disables build execution.
	NoBuildFlagName = "no-build"
	// RunFlagName is a name of the flag which defines process to be executed.
	RunFlagName = "run"
)

type cmdCtrl int

const (
	start   cmdCtrl = iota // 0
	restart                // 1
	stop                   // 2

)

// DefaultExclusions is a slices with glob patterns excluded by default.
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
			return errors.Wrap(config.SyncFullyQualifiedFlags(cmd), "failed syncing flags")
		},
		RunE: execute,
	}

	executeCmd.Flags().StringP(BuildFlagName, "b", "", "command to build your application before run")
	executeCmd.Flags().Bool(NoBuildFlagName, false, "always skips build")
	executeCmd.Flags().StringP(RunFlagName, "r", "", "command to run your application")
	// Watch config
	executeCmd.Flags().Bool("watch", false, "enables watch")
	executeCmd.Flags().StringSlice("dir", []string{"."}, "list of directories to watch (defaults to current directory)")
	// Empty slice as we are always adding DefaultExclusions while constructing the watch
	executeCmd.Flags().StringSlice("exclude", []string{}, fmt.Sprintf("list of patterns to exclude (always excludes %v)", DefaultExclusions))
	executeCmd.Flags().Int64("interval", 500, "watch interval (in ms)")
	if err := executeCmd.Flags().MarkHidden("interval"); err != nil {
		logger().Error(err, "failed while trying to hide a flag")
	}

	return executeCmd
}

func execute(command *cobra.Command, args []string) error {
	watcher := func(cmdChan chan cmdCtrl) (func(), error) {
		dirs, _ := command.Flags().GetStringSlice("dir")
		excluded, e := command.Flags().GetStringSlice("exclude")
		if e != nil {
			return nil, errors.Wrap(e, "failed obtaining exclude flag")
		}
		excluded = append(excluded, DefaultExclusions...)

		ms, _ := command.Flags().GetInt64("interval")
		restartHandler := func(events []fsnotify.Event) error {
			for _, event := range events {
				_, _ = command.OutOrStdout().Write([]byte(event.Name + " changed. Restarting process.\n"))
			}
			cmdChan <- restart

			return nil
		}

		w, err := watch.CreateWatch(ms).
			WithHandlers(restartHandler).
			Excluding(excluded...).
			OnPaths(dirs...)

		if err != nil {
			return nil, errors.WrapIf(err, "failed handling watch event")
		}

		w.Start()

		return w.Close, nil
	}

	stopPrevious := make(chan struct{})
	defer close(stopPrevious)

	cmdChan := make(chan cmdCtrl)
	defer close(cmdChan)

	if w, e := command.Flags().GetBool("watch"); w && e == nil {
		closeWatch, err := watcher(cmdChan)
		if err != nil {
			return errors.WrapIf(err, "failed watching")
		}
		defer closeWatch()
	} else if e != nil {
		return errors.Wrap(e, "failed obtaining watch flag")
	}

	end := make(chan struct{}, 1)
	defer close(end)

	go func() {
		for i := range cmdChan {
			switch i {
			case start:
				go buildAndRun(buildExecutor(command), runExecutor(command), stopPrevious, cmdChan)
			case restart:
				stopPrevious <- struct{}{}
			case stop:
				end <- struct{}{}
			}
		}
	}()

	cmdChan <- start

	<-end

	return nil
}

type stopper func() error
type executor func(cmdChan chan cmdCtrl) stopper

func buildExecutor(command *cobra.Command) executor {
	buildFlag := command.Flag(BuildFlagName)
	skipBuild, _ := command.Flags().GetBool(NoBuildFlagName)

	shouldRunBuild := buildFlag.Changed && !skipBuild
	if !shouldRunBuild {
		return func(chan cmdCtrl) stopper { return func() error { return nil } } // NOOP
	}

	buildCmd := command.Flag(BuildFlagName).Value.String()
	buildArgs := strings.Split(buildCmd, " ")

	return func(chan cmdCtrl) stopper {
		b := gocmd.NewCmdOptions(shell.StreamOutput, buildArgs[0], buildArgs[1:]...)
		b.Env = os.Environ()
		shell.RedirectStreams(b, command.OutOrStdout(), command.OutOrStderr())
		logger().V(1).Info("starting build command",
			"cmd", b.Name,
			"args", fmt.Sprint(b.Args),
		)

		hook.Register(func() error {
			return b.Stop() //nolint:wrapcheck //reason shutdownhook, it's where all ends anyway
		})

		<-b.Start()

		status := b.Status()
		if status.Error != nil {
			logger().Error(status.Error, "failed to run build command")
		}

		return b.Stop
	}
}

func runExecutor(command *cobra.Command) executor {
	runCmd := command.Flag("run").Value.String()
	runArgs := strings.Split(runCmd, " ")

	return func(cmdChan chan cmdCtrl) stopper {
		r := gocmd.NewCmdOptions(shell.StreamOutput, runArgs[0], runArgs[1:]...)
		r.Env = os.Environ()
		shell.RedirectStreams(r, command.OutOrStdout(), command.OutOrStderr())

		logger().V(1).Info("starting run command",
			"cmd", r.Name,
			"args", fmt.Sprint(r.Args),
		)

		go func(statusChan <-chan gocmd.Status) {
			hook.Register(func() error {
				cmdChan <- stop

				return r.Stop() //nolint:wrapcheck //reason shutdownhook, it's where all ends anyway
			})

			status := <-statusChan
			if status.Exit > 0 {
				logger().Error(status.Error, fmt.Sprintf("failed to run [%s] command", command.Name()))
				time.Sleep(4 * time.Second) // to avoid too frequent restarts of instantly failing process so that user can actually notice
				cmdChan <- restart

				return
			}

			if status.Complete {
				cmdChan <- stop
			}
		}(r.Start())

		return r.Stop
	}
}

func buildAndRun(builder, runner executor, stopPrevious chan struct{}, cmdChan chan cmdCtrl) {
	_ = builder(cmdChan)
	stopRun := runner(cmdChan)

	if _, ok := <-stopPrevious; ok {
		if e := stopRun(); e != nil {
			logger().Error(e, "failed while trying to stop RUNNING process")
		}
		cmdChan <- start
	}
}
