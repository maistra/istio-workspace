package telepresence

import (
	"os"
	"strings"

	"github.com/maistra/istio-workspace/pkg/shell"

	gocmd "github.com/go-cmd/cmd"
)

const (
	BinaryName                 = "telepresence"
	DefaultTelepresenceVersion = "0.101"
)

// GetVersion checks which version of telepresence should be used or is available on the path
// TELEPRESENCE_VERSION env variable is checked and if is defined takes precedence.
// If the binary is present on the $PATH then its version is used.
// If all above fails we return telepresence.DefaultTelepresenceVersion
func GetVersion() string {
	tpVersion, found := os.LookupEnv("TELEPRESENCE_VERSION")
	if !found && BinaryAvailable() {
		done := make(chan gocmd.Status, 1)
		defer close(done)

		go func() {
			tp := gocmd.NewCmdOptions(shell.BufferAndStreamOutput, "telepresence", "--version")
			shell.RedirectStreams(tp, os.Stdout, os.Stderr, done)
			shell.Start(tp, done)
		}()

		finalStatus := <-done
		tpVersion = strings.Join(finalStatus.Stdout, " ")
	}

	if tpVersion == "" {
		tpVersion = DefaultTelepresenceVersion
	}
	return tpVersion
}

// BinaryAvailable checks if telepresence binary is available on the path
func BinaryAvailable() bool {
	return shell.BinaryExists(BinaryName, "Head over to https://www.telepresence.io/reference/install for installation instructions.\n")
}
