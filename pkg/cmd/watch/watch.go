package watch

import (
	"strings"

	"github.com/maistra/istio-workspace/pkg/cmd/config"
	"github.com/maistra/istio-workspace/pkg/cmd/develop"
	"github.com/maistra/istio-workspace/pkg/cmd/internal/build"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/shell"
	"github.com/maistra/istio-workspace/pkg/watch"

	"github.com/fsnotify/fsnotify"
	gocmd "github.com/go-cmd/cmd"
	"github.com/spf13/cobra"
)

var logger = log.CreateOperatorAwareLogger("cmd").WithValues("type", "watch")

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
		RunE: watchForChanges,
	}

	watchCmd.Flags().StringP(build.BuildFlagName, "b", "", "command to build your application before run")
	watchCmd.Flags().Bool(build.NoBuildFlagName, false, "always skips build")
	watchCmd.Flags().StringP(build.RunFlagName, "r", "", "command to run your application")
	watchCmd.Flags().StringSliceP("dir", "w", []string{"."}, "list of directories to watch")
	watchCmd.Flags().StringSlice("exclude", develop.DefaultExclusions, "list of patterns to exclude (defaults to telepresence.log which is always excluded)")
	watchCmd.Flags().Int64("interval", 500, "watch interval (in ms)")
	if err := watchCmd.Flags().MarkHidden("interval"); err != nil {
		logger.Error(err, "failed while trying to hide a flag")
	}

	return watchCmd
}

func watchForChanges(command *cobra.Command, args []string) error {
	if err := build.Build(command); err != nil {
		return err
	}

	run := strings.Split(command.Flag(build.RunFlagName).Value.String(), " ")
	done := make(chan gocmd.Status)
	restart := make(chan struct{})

	slice, _ := command.Flags().GetStringSlice("dir")
	excluded, e := command.Flags().GetStringSlice("exclude")
	if e != nil {
		return e
	}
	excluded = append(excluded, develop.DefaultExclusions...)

	ms, _ := command.Flags().GetInt64("interval")
	w, err := watch.CreateWatch(ms).
		WithHandlers(func(events []fsnotify.Event) error {
			for _, event := range events {
				_, _ = command.OutOrStdout().Write([]byte(event.Name + " changed. Restarting process.\n"))
			}

			if err := build.Build(command); err != nil {
				done <- gocmd.Status{
					Error:    err,
					Complete: false,
				}
				return err
			}
			restart <- struct{}{}
			return nil
		}).
		Excluding(excluded...).
		OnPaths(slice...)

	if err != nil {
		return err
	}

	w.Start()
	defer w.Close()

	runDone := make(chan gocmd.Status)
	defer close(runDone)

	go func() {
		var runCmd *gocmd.Cmd
	OutOfLoop:
		for {
			select {
			case <-restart:
				if runCmd != nil {
					err = runCmd.Stop()
					<-runCmd.Done()
					runDone <- gocmd.Status{
						Error:    err,
						Complete: true,
					}
				}
				runCmd = gocmd.NewCmdOptions(shell.StreamOutput, run[0], run[1:]...)
				shell.RedirectStreams(runCmd, command.OutOrStdout(), command.OutOrStderr(), runDone)
				shell.ShutdownHook(runCmd, runDone)
				go shell.Start(runCmd, runDone)
			case status := <-runDone:
				done <- status
				break OutOfLoop
			}
		}
	}()

	restart <- struct{}{}
	status := <-done
	return status.Error
}
