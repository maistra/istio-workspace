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

var current goleak.Option

var _ = SynchronizedBeforeSuite(func() []byte {
	current = goleak.IgnoreCurrent()

	return []byte{}
}, func([]byte) {})

var _ = SynchronizedAfterSuite(func() {}, func() {
	goleak.VerifyNone(GinkgoT(), current)
})
