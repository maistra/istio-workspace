package create_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"go.uber.org/goleak"

	"github.com/maistra/istio-workspace/test/shell"
)

func TestCreateCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Create Command Suite")
}

var current goleak.Option

var _ = SynchronizedBeforeSuite(func() []byte {
	current = goleak.IgnoreCurrent()
	shell.StubShellCommands()

	return []byte{}
}, func([]byte) {})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
	goleak.VerifyNone(GinkgoT(), current)
})
