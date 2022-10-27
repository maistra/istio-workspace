package execute_test

import (
	"testing"

	"github.com/maistra/istio-workspace/test/shell"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"go.uber.org/goleak"
)

func TestExecuteCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Execute Command Suite")
}

var current goleak.Option

var sleepBin string

var _ = SynchronizedBeforeSuite(func() []byte {
	sleepBin = shell.BuildBinary("github.com/maistra/istio-workspace/test/echo",
		"sleepy", "-ldflags", "-w -X main.SleepMs=40000")
	current = goleak.IgnoreCurrent()
	shell.StubShellCommands()

	return []byte{}
}, func([]byte) {})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
	goleak.VerifyNone(GinkgoT(), current)
})
