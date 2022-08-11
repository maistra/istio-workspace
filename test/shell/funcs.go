package shell

import (
	"fmt"
	"os"
	"strings"

	gocmd "github.com/go-cmd/cmd"
	"github.com/google/shlex"
	"github.com/onsi/gomega"

	"github.com/maistra/istio-workspace/pkg/shell"
)

// WaitForSuccess checks if cmds ended just fine.
func WaitForSuccess(cmds ...*gocmd.Cmd) {
	for _, cmd := range cmds {
		<-cmd.Done()
		gomega.Expect(cmd.Status().Exit).To(gomega.BeZero())
	}
}

// Execute executes given command in the current directory.
func Execute(command string) *gocmd.Cmd {
	cmd, _ := shlex.Split(command)

	return ExecuteInDir("", cmd[0], cmd[1:]...)
}

// ExecuteInProjectRoot runs given command in project root folder (e.g. handy for make).
func ExecuteInProjectRoot(command string) *gocmd.Cmd {
	cmd, _ := shlex.Split(command)

	return ExecuteInDir(GetProjectDir(), cmd[0], cmd[1:]...)
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
	_ = command.Start()
	shell.RedirectStreams(command, os.Stdout, os.Stderr)
	commandString := command.Name + " " + strings.Join(command.Args, " ")
	if !strings.Contains(commandString, "oc login") {
		fmt.Printf("executing: [%s]\n", commandString)
	} else {
		fmt.Println("executing [oc login .....]")
	}

	return command
}

func GetProjectDir() string {
	if projectDir, found := os.LookupEnv("PROJECT_DIR"); found {
		return projectDir
	}

	gitRoot := Execute("git rev-parse --show-toplevel")
	<-gitRoot.Done()
	if gitRoot.Status().Error == nil {
		return strings.Join(gitRoot.Status().Stdout, "")
	}

	return "."
}
