package shell

import (
	"fmt"
	"os"
	"strings"

	"github.com/maistra/istio-workspace/pkg/shell"

	gocmd "github.com/go-cmd/cmd"
	"github.com/google/shlex"
)

// Execute executes given command in the current directory.
func Execute(command string) *gocmd.Cmd {
	cmd, _ := shlex.Split(command)
	return ExecuteInDir("", cmd[0], cmd[1:]...)
}

// ExecuteAll executes all passed commands in sequence, waiting for every single one to finish.
// before starting next one.
func ExecuteAll(commands ...string) {
	for _, command := range commands {
		<-Execute(command).Done()
	}
}

// ExecuteInDir executes given command in the defined directory
// Redirects streams to stdout/err.
func ExecuteInDir(dir, name string, args ...string) *gocmd.Cmd {
	command := gocmd.NewCmdOptions(shell.BufferAndStreamOutput, name, args...)
	command.Dir = dir
	done := command.Start()
	shell.RedirectStreams(command, os.Stdout, os.Stderr, done)
	commandString := command.Name + " " + strings.Join(command.Args, " ")
	if !strings.Contains(commandString, "oc login") {
		fmt.Printf("executing: [%s]\n", commandString)
	} else {
		fmt.Println("executing [oc login .....]")
	}
	return command
}

func GetProjectDir() string {
	projectDir, found := os.LookupEnv("PROJECT_DIR")
	if !found {
		return "."
	}
	return projectDir
}
