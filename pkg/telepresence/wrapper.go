package telepresence

import (
	"fmt"
	"os"
	"strings"

	"emperror.dev/errors"
	gocmd "github.com/go-cmd/cmd"
	"github.com/maistra/istio-workspace/pkg/cmd/execute"
	"github.com/maistra/istio-workspace/pkg/shell"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	// BinaryName is a name of telepresence binary we assume be available on the $PATH.
	BinaryName  = "telepresence"
	installHint = "Head over to https://www.telepresence.io/docs/v1/reference/install/ for installation instructions.\n"
)

var (
	errBinaryNotFound = fmt.Errorf("couldn't find '%s' installed in your system.\n%s\n"+
		"you can specify the version using TELEPRESENCE_VERSION environment variable", BinaryName, installHint)
	errUnsupportedBinary = fmt.Errorf("you are using unsupported version of %s, please install v1.\n%s", BinaryName, installHint)
)

// GetVersion checks which version of telepresence should be used or is available on the path
// TELEPRESENCE_VERSION env variable is checked and if is defined takes precedence.
// If the binary is present on the $PATH then its version is used.
// If all above fails we return error, as there's no telepresence in use nor env var is defined.
func GetVersion() (string, error) {
	if !BinaryAvailable() {
		return "", errBinaryNotFound
	}

	tpVersion, versionSpecified := os.LookupEnv("TELEPRESENCE_VERSION")
	if !versionSpecified {
		// Check if we are dealing with Telepresence v2
		versionCmd := executeCmd("telepresence", "version")
		if versionCmd.Exit == 0 {
			return "", errUnsupportedBinary
		}
		versionCmd = executeCmd("telepresence", "--version")
		tpVersion = strings.Join(versionCmd.Stdout, " ")
	}

	if tpVersion == "" {
		return "", errUnsupportedBinary
	}

	return tpVersion, nil
}

// CreateTpCommand translates `ike develop` command to underlying Telepresence invocation.
func CreateTpCommand(cmd *cobra.Command) ([]string, error) {
	if _, found := cmd.Annotations["telepresence"]; !found {
		return nil, errors.New("command cannot be translated to telepresence invocation")
	}
	tpArgs := []string{
		"--deployment", cmd.Flag("deployment").Value.String(),
		"--method", cmd.Flag("method").Value.String(),
	}
	if cmd.Flags().Changed("port") {
		ports, _ := cmd.Flags().GetStringSlice("port") // ignore error, should only occur if flag does not exist. If it doesn't, it won't be Changed()
		for _, port := range ports {
			tpArgs = append(tpArgs, "--expose", port)
		}
	}

	tpArgs = append(tpArgs, "--run")
	var tpCmd []string
	tpCmd = tpArgs
	subCmd, err := createWrapperExecutionCmd(cmd)
	if err != nil {
		return nil, err
	}
	tpCmd = append(tpCmd, subCmd...)

	namespaceFlag := cmd.Flag("namespace")
	if namespaceFlag.Changed {
		tpCmd = append([]string{"--" + namespaceFlag.Name, namespaceFlag.Value.String()}, tpCmd...)
	}

	return tpCmd, nil
}

func createWrapperExecutionCmd(cmd *cobra.Command) ([]string, error) {
	run := cmd.Flag(execute.RunFlagName).Value.String()
	executable, err := os.Executable()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create executable")
	}
	executeArgs := []string{
		executable, "execute",
		"--" + execute.RunFlagName, run,
	}
	if cmd.Flag(execute.NoBuildFlagName).Changed {
		executeArgs = append(executeArgs, "--"+execute.NoBuildFlagName, cmd.Flag(execute.NoBuildFlagName).Value.String())
	}
	if cmd.Flag(execute.BuildFlagName).Changed {
		executeArgs = append(executeArgs, "--"+execute.BuildFlagName, cmd.Flag(execute.BuildFlagName).Value.String())
	}

	watch, _ := cmd.Flags().GetBool("watch")
	if watch {
		executeArgs = append(executeArgs,
			"--watch",
			"--dir", stringSliceToCSV(cmd.Flags(), "watch-include"),
			"--exclude", stringSliceToCSV(cmd.Flags(), "watch-exclude"),
			"--interval", cmd.Flag("watch-interval").Value.String(),
		)
	}

	return executeArgs, nil
}

func stringSliceToCSV(flags *pflag.FlagSet, name string) string {
	slice, _ := flags.GetStringSlice(name)

	return strings.Join(slice, ",")
}

func executeCmd(cmd string, args ...string) gocmd.Status {
	done := make(chan gocmd.Status, 1)
	defer close(done)

	command := gocmd.NewCmdOptions(shell.BufferAndStreamOutput, cmd, args...)
	shell.Start(command, done)

	return <-done
}

// BinaryAvailable checks if telepresence binary is available on the path.
func BinaryAvailable() bool {
	return shell.BinaryExists(BinaryName, installHint)
}
