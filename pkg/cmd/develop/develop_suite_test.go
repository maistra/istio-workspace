package develop_test

import (
	"testing"

	"github.com/maistra/istio-workspace/test/shell"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"go.uber.org/goleak"
)

func TestDevelopCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Develop Command Suite")
}

var current goleak.Option

var _ = SynchronizedBeforeSuite(func() []byte {
	shell.StubShellCommands()
	current = goleak.IgnoreCurrent()

	return []byte{}
}, func([]byte) {})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
	goleak.VerifyNone(GinkgoT(), current)
})
