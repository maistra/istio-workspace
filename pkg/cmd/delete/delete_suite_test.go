package delete_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"go.uber.org/goleak"

	. "github.com/maistra/istio-workspace/test"
	"github.com/maistra/istio-workspace/test/shell"
)

func TestDeleteCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "Delete Command Suite")
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
