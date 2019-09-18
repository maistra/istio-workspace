package delete_test

import (
	"testing"

	"github.com/maistra/istio-workspace/test/shell"

	"github.com/onsi/gomega/gexec"
	"go.uber.org/goleak"

	. "github.com/maistra/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDeleteCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "Delete Command Suite")
}

var _ = BeforeSuite(shell.StubShellCommands)

var _ = AfterSuite(func() {
	CleanUpTmpFiles(GinkgoT())
	gexec.CleanupBuildArtifacts()
	goleak.VerifyNone(GinkgoT(),
		goleak.IgnoreTopFunction("github.com/maistra/istio-workspace/vendor/k8s.io/klog.(*loggingT).flushDaemon"),
		goleak.IgnoreTopFunction("github.com/maistra/istio-workspace/vendor/github.com/onsi/ginkgo/internal/specrunner.(*SpecRunner).registerForInterrupts"),
	)
})
