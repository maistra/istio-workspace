package shell

import (
	"os"
	"path"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/spf13/afero"
)

var (
	MvnBin             string
	Tp1WithSleepBin    string
	Tp1FixedVersionBin string
	Tp1VersionFlagBin  string
	Tp2VersionFlagBin  string
	JavaBin            string
)

func StubShellCommands() {
	MvnBin = BuildBinary("github.com/maistra/istio-workspace/test/echo", "mvn")
	JavaBin = BuildBinary("github.com/maistra/istio-workspace/test/echo", "java",
		"-ldflags", "-w -X main.SleepMs=1024")
	_ = BuildBinary("github.com/maistra/istio-workspace/test/echo", "ike",
		"-ldflags", "-w -X main.SleepMs=50")
	Tp1FixedVersionBin = BuildBinary("github.com/maistra/istio-workspace/test/tp_stub", "telepresence", "-ldflags", "-w -X main.Version=v1 -X main.Return=0.234")
	Tp1VersionFlagBin = BuildBinary("github.com/maistra/istio-workspace/test/tp_stub", "telepresence", "-ldflags", "-w -X main.Version=v1")
	Tp2VersionFlagBin = BuildBinary("github.com/maistra/istio-workspace/test/tp_stub", "telepresence", "-ldflags", "-w -X main.Version=v2")
	Tp1WithSleepBin = BuildBinary("github.com/maistra/istio-workspace/test/tp_stub",
		"telepresence", "-ldflags", "-w -X main.SleepMs=256 -X main.Version=v1")
}

func ExecuteCommand(outputChan chan string, execute func() (string, error)) func() {
	return func() {
		defer ginkgo.GinkgoRecover()
		output, err := execute()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		outputChan <- output
	}
}

var appFs = afero.NewOsFs()

func BuildBinary(packagePath, name string, flags ...string) string {
	binPath, err := gexec.Build(packagePath, flags...)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

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
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	err = appFs.Chmod(binPath, os.ModePerm)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	content, err := afero.ReadFile(appFs, src)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = bin.Write(content)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	defer func() {
		err = bin.Close()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	}()

	return bin.Name()
}
