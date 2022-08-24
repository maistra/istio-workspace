package infra

import (
	"fmt"
	"time"

	"github.com/go-cmd/cmd"
	. "github.com/onsi/gomega"

	"github.com/maistra/istio-workspace/test"
	testshell "github.com/maistra/istio-workspace/test/shell"
)

// RunIke runs the ike cli in the given dir.
func RunIke(dir string, arguments ...string) *cmd.Cmd {
	return testshell.ExecuteInDir(dir, "ike", arguments...)
}

// Stop shuts down the process.
func Stop(ike *cmd.Cmd) {
	if ike.Status().Complete {
		return
	}
	stopFailed := ike.Stop()
	Expect(stopFailed).ToNot(HaveOccurred())

	Eventually(ike.Done(), 1*time.Minute).Should(BeClosed())
}

func FailOnCmdError(command *cmd.Cmd, t test.TestReporter) {
	<-command.Done()
	fmt.Println(command.Status().Exit)
	fmt.Println(command.Status().Stdout)
	fmt.Println(command.Status().Stderr)
	if command.Status().Exit != 0 && command.Status().Exit != 130 { // do not panic on SIGINT
		t.Errorf("failed executing %s with code %d", command.Name, command.Status().Exit)
	}
}
