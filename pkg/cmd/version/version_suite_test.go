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
		goleak.IgnoreTopFunction("github.com/maistra/istio-workspace/vendor/k8s.io/klog.(*loggingT).flushDaemon"),
		goleak.IgnoreTopFunction("github.com/maistra/istio-workspace/vendor/github.com/onsi/ginkgo/internal/specrunner.(*SpecRunner).registerForInterrupts"),
	)
})
