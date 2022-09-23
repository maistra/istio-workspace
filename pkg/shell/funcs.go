package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	gocmd "github.com/go-cmd/cmd"
	"github.com/go-logr/logr"

	"github.com/maistra/istio-workspace/pkg/log"
)

var (
	// StreamOutput sets streaming of output instead of buffering it when running gocmd.Cmd.
	StreamOutput = gocmd.Options{
		Buffered:  false,
		Streaming: true,
	}

	// BufferAndStreamOutput sets buffering and streaming of output when running gocmd.Cmd
	// Buffering lets easy inspection of outputs in tests through inspecting gocmd.Cmd.Status.Stdout/err slices.
	BufferAndStreamOutput = gocmd.Options{
		Buffered:  true,
		Streaming: true,
	}

	logger = func() logr.Logger {
		return log.Log.WithValues("type", "shell")
	}
)

// Start starts new process (gocmd) and wait until it's done. Status struct is then propagated back to
// done channel passed as argument.
func Start(cmd *gocmd.Cmd, done chan gocmd.Status) {
	cmd.Env = os.Environ()
	logger().V(1).Info("starting command",
		"cmd", cmd.Name,
		"args", fmt.Sprint(cmd.Args),
	)
	status := <-cmd.Start()
	<-cmd.Done()
	done <- status
}

// RedirectStreams redirects Stdout and Stderr of the gocmd.Cmd process to passed io.Writers.
func RedirectStreams(src *gocmd.Cmd, stdoutDest, stderrDest io.Writer) {
	go func() {
	OutOfLoop:
		for {
			select {
			case line, ok := <-src.Stdout:
				if !ok {
					break OutOfLoop
				}
				if _, err := fmt.Fprintln(stdoutDest, line); err != nil {
					logger().Error(err, fmt.Sprintf("%s failed executing", src.Name))
				}
			case line, ok := <-src.Stderr:
				if !ok {
					break OutOfLoop
				}
				if _, err := fmt.Fprintln(stderrDest, line); err != nil {
					logger().Error(err, fmt.Sprintf("%s failed executing", src.Name))
				}
			case <-src.Done():
				break OutOfLoop
			}
		}
	}()
}

// CurrentDir returns current directory from where binary is executed.
func CurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return dir
}

// BinaryExists ensures that binary with given name (binName) is available on the PATH.
// hint lets you customize the error message.
func BinaryExists(binName, hint string) bool {
	path, err := exec.LookPath(binName)
	if err != nil {
		if hint != "" {
			logger().Error(err, fmt.Sprintf("Couldn't find '%s' installed in your system.\n%s", binName, hint))
		}

		return false
	}
	logger().V(1).Info(fmt.Sprintf("Found '%s' executable in '%s'.", binName, path))
	logger().V(1).Info(fmt.Sprintf("See '%s.log' for more details about its execution.", binName))

	return true
}
