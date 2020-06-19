package build

import (
	"fmt"
	"strings"

	"github.com/maistra/istio-workspace/pkg/shell"

	gocmd "github.com/go-cmd/cmd"
	"github.com/spf13/cobra"
)

const (
	BuildFlagName   = "build"
	NoBuildFlagName = "no-build"
	RunFlagName     = "run"
)

// Build expects that cmd has Build and no-Build flags defined.
// otherwise it fails.
func Build(cmd *cobra.Command) error {
	buildFlag := cmd.Flag(BuildFlagName)
	if buildFlag == nil {
		return fmt.Errorf("expecting '%s' command to have '%s' flag defined", cmd.Name(), BuildFlagName)
	}

	skipBuild, err := cmd.Flags().GetBool(NoBuildFlagName)
	if err != nil {
		return fmt.Errorf("expecting '%s' command to have '%s' flag defined", cmd.Name(), NoBuildFlagName)
	}

	if buildFlag.Changed && !skipBuild {
		buildCmd := cmd.Flag(BuildFlagName).Value.String()
		buildArgs := strings.Split(buildCmd, " ")
		build := gocmd.NewCmdOptions(shell.StreamOutput, buildArgs[0], buildArgs[1:]...)

		done := make(chan gocmd.Status, 1)
		defer close(done)

		shell.RedirectStreams(build, cmd.OutOrStdout(), cmd.OutOrStderr(), done)

		buildStatusChan := build.Start()
		buildStatus := <-buildStatusChan

		done <- buildStatus

		if buildStatus.Error != nil {
			return buildStatus.Error
		}
	}

	return nil
}
