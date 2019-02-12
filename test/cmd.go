package test

import (
	"bytes"
	"strings"

	"github.com/spf13/cobra"
)

type Cmd cobra.Command

// ValidateArgumentsOf will not run actual command but let the initialization and validation happen
func ValidateArgumentsOf(command *cobra.Command) *Cmd {
	command.Run = emptyRun
	command.RunE = emptyRunE
	return (*Cmd)(command)
}

func (command *Cmd) Passing(args ...string) (output string, err error) {
	cmd := (*cobra.Command)(command)
	output, err = executeCommand(cmd, args...)
	return output, err
}

func executeCommand(cmd *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(cmd, args...)
	return output, err
}

func executeCommandC(cmd *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)
	cmd.Root().SetArgs(append(strings.Split(cmd.CommandPath(), " ")[1:], args...))
	c, err = cmd.ExecuteC()
	return c, buf.String(), err
}

func emptyRun(cmd *cobra.Command, args []string) {}

func emptyRunE(cmd *cobra.Command, args []string) error {
	return nil
}
