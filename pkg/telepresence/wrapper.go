package telepresence

import (
	"fmt"
	"os"
	"strings"

	gocmd "github.com/go-cmd/cmd"

	"github.com/maistra/istio-workspace/pkg/shell"
)

const (
	// BinaryName is a name of telepresence binary we assume be available on the $PATH.
	BinaryName  = "telepresence"
	installHint = "Head over to https://www.telepresence.io/docs/v1/reference/install/ for installation instructions.\n"
)

var errorNoTelepresenceHint = fmt.Errorf("couldn't find '%s' installed in your system.\n%s\n"+
	"you can specify the version using TELEPRESENCE_VERSION environment variable", BinaryName, installHint)

// GetVersion checks which version of telepresence should be used or is available on the path
// TELEPRESENCE_VERSION env variable is checked and if is defined takes precedence.
// If the binary is present on the $PATH then its version is used.
// If all above fails we return error, as there's no telepresence in use nor env var is defined.
func GetVersion() (string, error) {
	tpVersion, found := os.LookupEnv("TELEPRESENCE_VERSION")
	if !found && BinaryAvailable() {
		done := make(chan gocmd.Status, 1)
		defer close(done)

		go func() {
			tp := gocmd.NewCmdOptions(shell.BufferAndStreamOutput, "telepresence", "--version")
			shell.Start(tp, done)
		}()

		finalStatus := <-done
		tpVersion = strings.Join(finalStatus.Stdout, " ")
	}

	if tpVersion == "" {
		return "", errorNoTelepresenceHint
	}

	return tpVersion, nil
}

// BinaryAvailable checks if telepresence binary is available on the path.
func BinaryAvailable() bool {
	return shell.BinaryExists(BinaryName, installHint)
}
