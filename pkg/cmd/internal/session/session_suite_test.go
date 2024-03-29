package internal_test

import (
	"path"
	"testing"

	. "github.com/maistra/istio-workspace/test"
	"github.com/maistra/istio-workspace/test/shell"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"go.uber.org/goleak"
)

func TestDevelopCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Session Internal Logic Suite")
}

var current goleak.Option

var tmpPath = NewTmpPath()
var _ = SynchronizedBeforeSuite(func() []byte {
	current = goleak.IgnoreCurrent()
	tpPath := shell.BuildBinary("github.com/maistra/istio-workspace/test/tp_stub", "telepresence", "-ldflags", "-w -X main.Version=v1")

	tmpPath = NewTmpPath()
	tmpPath.SetPath(path.Dir(tpPath))

	return []byte{}
}, func([]byte) {})

var _ = SynchronizedAfterSuite(func() {}, func() {
	tmpPath.Restore()
	gexec.CleanupBuildArtifacts()
	goleak.VerifyNone(GinkgoT(), current)
})
