package cmd

import (
	"fmt"
	"strings"

	gocmd "github.com/go-cmd/cmd"
	"github.com/spf13/cobra"
)

const (
	buildFlagName   = "build"
	noBuildFlagName = "no-build"
	runFlagName     = "run"
)

// build expects that cmd has build and no-build flags defined.
// otherwise it fails
func build(cmd *cobra.Command) error {
	buildFlag := cmd.Flag(buildFlagName)
	if buildFlag == nil {
		return fmt.Errorf("expecting '%s' command to have '%s' flag defined", cmd.Name(), buildFlagName)
	}

	skipBuild, err := cmd.Flags().GetBool(noBuildFlagName)
	if err != nil {
		return fmt.Errorf("expecting '%s' command to have '%s' flag defined", cmd.Name(), noBuildFlagName)
	}

	if buildFlag.Changed && !skipBuild {
		buildCmd := cmd.Flag(buildFlagName).Value.String()
		buildArgs := strings.Split(buildCmd, " ")
		log.Info("Starting build", "build-cmd", buildCmd)
		build := gocmd.NewCmdOptions(StreamOutput, buildArgs[0], buildArgs[1:]...)

		done := make(chan gocmd.Status, 1)
		defer close(done)

		go RedirectStreamsToCmd(build, cmd, done)

		buildStatusChan := build.Start()
		buildStatus := <-buildStatusChan

		done <- buildStatus

		if buildStatus.Error != nil {
			return buildStatus.Error
		}
	}

	return nil
}
