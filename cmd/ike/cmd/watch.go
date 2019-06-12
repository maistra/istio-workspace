package cmd

import (
	"strings"

	"github.com/maistra/istio-workspace/cmd/ike/config"
	"github.com/maistra/istio-workspace/cmd/ike/watch"

	"github.com/fsnotify/fsnotify"
	gocmd "github.com/go-cmd/cmd"
	"github.com/spf13/cobra"
)

// NewWatchCmd creates watch command which observes file system changes in the defined set of directories
// and re-runs build and run command when they occur.
// It is hidden (not user facing) as it's integral part of develop command
func NewWatchCmd() *cobra.Command {
	watchCmd := &cobra.Command{
		Use:          "watch",
		Hidden:       true,
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return config.SyncFlags(cmd)
		},
		RunE: watchForChanges,
	}

	watchCmd.Flags().StringP(buildFlagName, "b", "", "command to build your application before run")
	watchCmd.Flags().Bool(noBuildFlagName, false, "always skips build")
	watchCmd.Flags().StringP(runFlagName, "r", "", "command to run your application")
	watchCmd.Flags().StringSliceP("dir", "w", []string{CurrentDir()}, "list of directories to watch")
	watchCmd.Flags().StringSlice("exclude", defaultExclusions, "list of patterns to exclude (defaults to telepresence.log which is always excluded)")
	watchCmd.Flags().Int64("interval", 500, "watch interval (in ms)")
	if err := watchCmd.Flags().MarkHidden("interval"); err != nil {
		log.Error(err, "failed while trying to hide a flag")
	}

	return watchCmd
}

func watchForChanges(cmd *cobra.Command, args []string) error {
	if err := build(cmd); err != nil {
		return err
	}

	run := strings.Split(cmd.Flag(runFlagName).Value.String(), " ")
	done := make(chan gocmd.Status)
	restart := make(chan struct{})

	slice, _ := cmd.Flags().GetStringSlice("dir")
	excluded, e := cmd.Flags().GetStringSlice("exclude")
	if e != nil {
		return e
	}
	excluded = append(excluded, defaultExclusions...)

	ms, _ := cmd.Flags().GetInt64("interval")
	w, err := watch.CreateWatch(ms).
		WithHandlers(func(events []fsnotify.Event) error {
			for _, event := range events {
				_, _ = cmd.OutOrStdout().Write([]byte(event.Name + " changed. Restarting process.\n"))
			}

			if err := build(cmd); err != nil {
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
				runCmd = gocmd.NewCmdOptions(StreamOutput, run[0], run[1:]...)
				RedirectStreams(runCmd, cmd.OutOrStdout(), cmd.OutOrStderr(), runDone)
				ShutdownHook(runCmd, runDone)
				go Start(runCmd, runDone)
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
