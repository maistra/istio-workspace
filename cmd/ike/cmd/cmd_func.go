package cmd

import (
	"fmt"
	"github.com/spf13/pflag"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	gocmd "github.com/go-cmd/cmd"
)

// StreamOutput sets streaming of output instead of buffering it when running gocmd.Cmd
var StreamOutput = gocmd.Options{
	Buffered:  false,
	Streaming: true,
}

// BufferAndStreamOutput sets buffering and streaming of output when running gocmd.Cmd
// Buffering lets easy inspection of outputs in tests through inspecting gocmd.Cmd.Status.Stdout/err slices
var BufferAndStreamOutput = gocmd.Options{
	Buffered:  true,
	Streaming: true,
}

// Start starts new process (gocmd) and wait until it's done. Status struct is then propagated back to
// done channel passed as argument
func Start(cmd *gocmd.Cmd, done chan gocmd.Status) {
	cmd.Env = os.Environ()
	log.Info("starting command",
		"cmd", cmd.Name,
		"args", fmt.Sprint(cmd.Args),
	)
	status := <-cmd.Start()
	<-cmd.Done()
	done <- status
}

// Execute executes given command in the current directory
// Adds shutdown hook and redirects streams to stdout/err
func Execute(command string) *gocmd.Cmd {
	cmd := strings.Split(command, " ")
	return ExecuteInDir("", cmd[0], cmd[1:]...)
}

// ExecuteInDir executes given command in the defined directory
// Adds shutdown hook and redirects streams to stdout/err
func ExecuteInDir(dir, name string, args ...string) *gocmd.Cmd {
	command := gocmd.NewCmdOptions(BufferAndStreamOutput, name, args...)
	command.Dir = dir
	done := command.Start()
	ShutdownHook(command, done)
	RedirectStreams(command, os.Stdout, os.Stderr, done)
	commandString := command.Name + " " + strings.Join(command.Args, " ")
	fmt.Printf("executing: [%s]\n", commandString)
	return command
}

// ShutdownHook will wait for SIGTERM signal and stop the cmd
// unless done receiving channel passed to it receives status or is closed
func ShutdownHook(cmd *gocmd.Cmd, done <-chan gocmd.Status) {
	go func() {
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
	}()
}

// RedirectStreams redirects Stdout and Stderr of the gocmd.Cmd process to passed io.Writers
func RedirectStreams(src *gocmd.Cmd, stdoutDest, stderrDest io.Writer, done <-chan gocmd.Status) {
	go func() {
	OutOfLoop:
		for {
			select {
			case line, ok := <-src.Stdout:
				if !ok {
					break OutOfLoop
				}
				if _, err := fmt.Fprintln(stdoutDest, line); err != nil {
					log.Error(err, fmt.Sprintf("%s failed executing", src.Name))
				}
			case line, ok := <-src.Stderr:
				if !ok {
					break OutOfLoop
				}
				if _, err := fmt.Fprintln(stderrDest, line); err != nil {
					log.Error(err, fmt.Sprintf("%s failed executing", src.Name))
				}
			case <-done:
				break OutOfLoop
			}
		}
	}()
}

// CurrentDir returns current directory from where binary is executed
func CurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

// BinaryExists ensures that binary with given name (binName) is available on the PATH
// hint lets you customize the error message
func BinaryExists(binName, hint string) bool {
	path, err := exec.LookPath(binName)
	if err != nil {
		log.Error(err, fmt.Sprintf("Couldn't find '%s' installed in your system.\n%s", binName, hint))
		return false
	}

	log.Info(fmt.Sprintf("Found '%s' executable in '%s'.", binName, path))
	log.Info(fmt.Sprintf("See '%s.log' for more details about its execution.", binName))

	return true
}

func stringSliceToCSV(flags *pflag.FlagSet, name string) string {
	slice, _ := flags.GetStringSlice(name)
	return fmt.Sprintf(`"%s"`, strings.Join(slice, ","))
}