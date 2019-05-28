package {{.Package}}

import (
	"go.uber.org/goleak"
	"testing"

	. "github.com/maistra/istio-workspace/test"

	{{.GinkgoImport}}
	{{.GomegaImport}}
)

func Test{{.FormattedName}}(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "{{.FormattedName}} Suite")
}

var _ = SynchronizedAfterSuite(func() {}, func() {
	goleak.VerifyNone(GinkgoT(),
		goleak.IgnoreTopFunction("github.com/maistra/istio-workspace/vendor/k8s.io/klog.(*loggingT).flushDaemon"),
		goleak.IgnoreTopFunction("github.com/maistra/istio-workspace/vendor/github.com/onsi/ginkgo/internal/specrunner.(*SpecRunner).registerForInterrupts"),
	)
})
