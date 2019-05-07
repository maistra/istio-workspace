package cmd_test

import (
	"os"
	"path"
	"testing"

	"github.com/onsi/gomega/gexec"
	"github.com/spf13/afero"
	"go.uber.org/goleak"

	. "github.com/maistra/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	mvnBin     string
	tpBin      string
	tpSleepBin string
	ikeBin     string
	javaBin    string
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "CLI Suite")
}

var _ = BeforeSuite(func() {
	mvnBin = buildBinary("github.com/maistra/istio-workspace/test/echo", "mvn")
	javaBin = buildBinary("github.com/maistra/istio-workspace/test/echo", "java",
		"-ldflags", "-w -X main.SleepMs=50")
	ikeBin = buildBinary("github.com/maistra/istio-workspace/test/echo", "ike",
		"-ldflags", "-w -X main.SleepMs=50")
	tpBin = buildBinary("github.com/maistra/istio-workspace/test/echo", "telepresence")
	tpSleepBin = buildBinary("github.com/maistra/istio-workspace/test/echo",
		"telepresence", "-ldflags", "-w -X main.SleepMs=50")
})

var _ = AfterSuite(func() {
	CleanUpTmpFiles(GinkgoT())
	gexec.CleanupBuildArtifacts()
	goleak.VerifyNone(GinkgoT(),
		goleak.IgnoreTopFunction("github.com/maistra/istio-workspace/vendor/k8s.io/klog.(*loggingT).flushDaemon"),
		goleak.IgnoreTopFunction("github.com/maistra/istio-workspace/vendor/github.com/onsi/ginkgo/internal/specrunner.(*SpecRunner).registerForInterrupts"),
	)
})

// -- Test helpers

func executeCommand(outputChan chan string, execute func() (string, error)) func() {
	return func() {
		defer GinkgoRecover()
		output, err := execute()
		Expect(err).NotTo(HaveOccurred())
		outputChan <- output
	}
}

var appFs = afero.NewOsFs()

func buildBinary(packagePath, name string, flags ...string) string { //nolint[:unparam]

	binPath, err := gexec.Build(packagePath, flags...)
	Expect(err).ToNot(HaveOccurred())

	// gexec.Build from Ginkgo does not allow to specify `-o` flag for the final binary name
	// thus we rename the binary instead. TODO: pull request to ginkgo
	if name != path.Base(packagePath) {
		finalName := copyBinary(appFs, binPath, name)
		_ = os.Remove(binPath)
		return finalName
	}

	return binPath
}

func copyBinary(appFs afero.Fs, src, dest string) string {
	binPath := path.Dir(src) + "/" + dest
	bin, err := appFs.Create(binPath)
	Expect(err).ToNot(HaveOccurred())

	err = appFs.Chmod(binPath, os.ModePerm)
	Expect(err).ToNot(HaveOccurred())

	content, err := afero.ReadFile(appFs, src)
	Expect(err).ToNot(HaveOccurred())
	_, err = bin.Write(content)
	Expect(err).ToNot(HaveOccurred())

	defer func() {
		err = bin.Close()
		Expect(err).ToNot(HaveOccurred())
	}()

	return bin.Name()
}
