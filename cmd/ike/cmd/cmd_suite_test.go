package cmd_test

import (
	"github.com/onsi/gomega/gexec"
	"testing"

	. "github.com/aslakknutsen/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var mvnBin string
var tpBin string
var tpSleepBin string

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "CLI Suite")
}

var _ = BeforeSuite(func() {
	mvnBin = buildBinary("github.com/aslakknutsen/istio-workspace/test/echo", "mvn")
	tpBin = buildBinary("github.com/aslakknutsen/istio-workspace/test/echo", "telepresence")
	tpSleepBin = buildBinary("github.com/aslakknutsen/istio-workspace/test/echo",
		"telepresence", "-ldflags", "-w -X main.SleepMs=50")
})

var _ = AfterSuite(func() {
	CleanUp(GinkgoT())
	gexec.CleanupBuildArtifacts()
})

