package version_test

import (
	"testing"

	"go.uber.org/goleak"

	. "github.com/maistra/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "Version Suite")
}

var _ = SynchronizedAfterSuite(func() {}, func() {
	goleak.VerifyNone(GinkgoT(),
		goleak.IgnoreTopFunction("k8s.io/klog/v2.(*loggingT).flushDaemon"),
		goleak.IgnoreTopFunction("github.com/onsi/ginkgo/internal/specrunner.(*SpecRunner).registerForInterrupts"),
	)
})
