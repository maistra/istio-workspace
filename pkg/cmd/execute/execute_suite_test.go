package execute_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"go.uber.org/goleak"

	. "github.com/maistra/istio-workspace/test"
	"github.com/maistra/istio-workspace/test/shell"
)

func TestWatchCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "Execute Command Suite")
}

var current goleak.Option

var _ = SynchronizedBeforeSuite(func() []byte {
	current = goleak.IgnoreCurrent()
	shell.StubShellCommands()

	return []byte{}
}, func([]byte) {})

var _ = SynchronizedAfterSuite(func() {}, func() {
	CleanUpTmpFiles(GinkgoT())
	gexec.CleanupBuildArtifacts()
	goleak.VerifyNone(GinkgoT(), current)
})
