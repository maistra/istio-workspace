package openshift_test

import (
	"testing"

	. "github.com/maistra/istio-workspace/test"

	"go.uber.org/goleak"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOpenshift(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "Openshift Suite")
}

var _ = SynchronizedAfterSuite(func() {}, func() {
	goleak.VerifyNone(GinkgoT(),
		goleak.IgnoreTopFunction("k8s.io/klog/v2.(*loggingT).flushDaemon"),
		goleak.IgnoreTopFunction("github.com/onsi/ginkgo/internal/specrunner.(*SpecRunner).registerForInterrupts"),
	)
})
