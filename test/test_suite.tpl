package {{.Package}}

import (
	"testing"

	. "github.com/maistra/istio-workspace/test"

	{{.GinkgoImport}}
	{{.GomegaImport}}
)

func Test{{.FormattedName}}(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "{{.FormattedName}} Suite")
}
