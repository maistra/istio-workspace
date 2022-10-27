package telepresence_test

import (
	"testing"

	"github.com/maistra/istio-workspace/test/shell"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/goleak"
)

func TestTelepresenceWrapper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Telepresence Wrapper Suite")
}

var current goleak.Option

var _ = SynchronizedBeforeSuite(func() []byte {
	shell.StubShellCommands()
	current = goleak.IgnoreCurrent()

	return []byte{}
}, func([]byte) {})

var _ = SynchronizedAfterSuite(func() {}, func() {
	goleak.VerifyNone(GinkgoT(), current)
})
