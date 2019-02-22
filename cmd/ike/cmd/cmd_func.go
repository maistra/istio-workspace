package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	gocmd "github.com/go-cmd/cmd"
	"github.com/spf13/cobra"
)

var streamOutput = gocmd.Options{
	Buffered:  false,
	Streaming: true,
}

func start(cmd *gocmd.Cmd, done chan gocmd.Status) {
	cmd.Env = os.Environ()
	log.Info("starting command",
		"cmd", cmd.Name,
		"args", fmt.Sprint(cmd.Args),
	)
	status := <-cmd.Start()
	<-cmd.Done()
	done <- status
}

func shutdownHook(cmd *gocmd.Cmd, done <-chan gocmd.Status) {
	hookChan := make(chan os.Signal, 1)
	signal.Notify(hookChan, os.Interrupt, syscall.SIGTERM)
	defer func() {
		signal.Stop(hookChan)
		close(hookChan)
	}()
OutOfLoop:
	for {
		select {
		case _, ok := <-hookChan:
			if !ok {
				break OutOfLoop
			}
			_ = cmd.Stop()
			<-cmd.Done()
			break OutOfLoop
		case <-done:
			break OutOfLoop
		}
	}
}

func redirectStreamsToCmd(src *gocmd.Cmd, dest *cobra.Command, done <-chan gocmd.Status) {
OutOfLoop:
	for {
		select {
		case line, ok := <-src.Stdout:
			if !ok {
				break OutOfLoop
			}
			if _, err := fmt.Fprintln(dest.OutOrStdout(), line); err != nil {
				log.Error(err, fmt.Sprintf("%s failed executing", src.Name))
			}
		case line, ok := <-src.Stderr:
			if !ok {
				break OutOfLoop
			}
			if _, err := fmt.Fprintln(dest.OutOrStderr(), line); err != nil {
				log.Error(err, fmt.Sprintf("%s failed executing", src.Name))
			}
		case <-done:
			break OutOfLoop
		}
	}
}

func currentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

func binaryExists(binName, hint string) bool {
	path, err := exec.LookPath(binName)
	if err != nil {
		log.Error(err, fmt.Sprintf("Couldn't find '%s' installed in your system.\n%s", binName, hint))
		return false
	}

	log.Info(fmt.Sprintf("Found '%s' executable in '%s'.", binName, path))
	log.Info(fmt.Sprintf("See '%s.log' for more details about its execution.", binName))

	return true
}
