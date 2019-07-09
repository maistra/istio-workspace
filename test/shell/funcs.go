package shell

import (
	"fmt"
	"os"
	"strings"

	"github.com/maistra/istio-workspace/pkg/shell"

	gocmd "github.com/go-cmd/cmd"
)

// Execute executes given command in the current directory
// Adds shutdown hook and redirects streams to stdout/err
func Execute(command string) *gocmd.Cmd {
	cmd := strings.Split(command, " ")
	return ExecuteInDir("", cmd[0], cmd[1:]...)
}

// ExecuteInDir executes given command in the defined directory
// Adds shutdown hook and redirects streams to stdout/err
func ExecuteInDir(dir, name string, args ...string) *gocmd.Cmd {
	command := gocmd.NewCmdOptions(shell.BufferAndStreamOutput, name, args...)
	command.Dir = dir
	done := command.Start()
	shell.RedirectStreams(command, os.Stdout, os.Stderr, done)
	commandString := command.Name + " " + strings.Join(command.Args, " ")
	fmt.Printf("executing: [%s]\n", commandString)
	return command
}
