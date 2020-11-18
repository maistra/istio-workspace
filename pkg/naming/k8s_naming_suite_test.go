package naming_test

import (
	"math/rand"
	"testing"
	"time"

	"go.uber.org/goleak"

	. "github.com/maistra/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNamingGenerator(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "Names Generator Suite")
}

var _ = SynchronizedAfterSuite(func() {}, func() {
	goleak.VerifyNone(GinkgoT(),
		goleak.IgnoreTopFunction("k8s.io/klog.(*loggingT).flushDaemon"),
		goleak.IgnoreTopFunction("github.com/onsi/ginkgo/internal/specrunner.(*SpecRunner).registerForInterrupts"),
	)
})
