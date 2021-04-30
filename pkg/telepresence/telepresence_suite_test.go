package telepresence_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/goleak"

	. "github.com/maistra/istio-workspace/test"
	"github.com/maistra/istio-workspace/test/shell"
)

func TestTelepresenceWrapper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "Telepresence Wrapper Suite")
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
