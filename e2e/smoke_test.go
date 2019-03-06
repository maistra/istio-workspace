package e2e_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aslakknutsen/istio-workspace/pkg/naming"

	"github.com/aslakknutsen/istio-workspace/cmd/ike/cmd"
	"github.com/aslakknutsen/istio-workspace/e2e"
	"github.com/aslakknutsen/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Smoke End To End Tests - against OpenShift Cluster with Istio (maistra)", func() {

	Context("using ike develop in offline mode", func() {

		tmpPath := test.NewTmpPath()

		var (
			appName,
			tmpDir string
		)

		BeforeEach(func() {
			tmpPath.SetPath(os.Getenv("PATH"), path.Dir(cmd.CurrentDir())+"/dist")
			appName = naming.RandName(16)
			tmpDir = test.TmpDir(GinkgoT(), "app-"+appName)
			Expect(cmd.BinaryExists("ike", "make sure you have binary in the ./dist folder. Try make compile at least")).To(BeTrue())
		})

		AfterEach(func() {
			tmpPath.Restore()
			<-cmd.Execute("oc", "delete", "project", appName).Done()
		})

		It("should watch python code changes and replace service when they occur", func() {

			e2e.CreateNewApp(appName)
			Eventually(callGetOn(appName), 1*time.Minute).Should(Equal("Hello, world!\n"))

			e2e.OriginalServerCodeIn(tmpDir)
			ikeWithWatch := cmd.ExecuteInDir(tmpDir, "ike", "develop",
				"--deployment", appName,
				"--port", "8000",
				"--method", "inject-tcp",
				"--watch",
				"--run", "python3 server.py",
			)
			Eventually(callGetOn(appName), 3*time.Minute, 200*time.Millisecond).Should(Equal("Hello, world!\n"))

			e2e.ModifyServerCodeIn(tmpDir)

			Eventually(callGetOn(appName), 3*time.Minute, 200*time.Millisecond).Should(Equal("Hello, telepresence! Ike Here!\n"))

			Expect(ikeWithWatch.Stop()).ToNot(HaveOccurred())
			Eventually(ikeWithWatch.Done(), 1*time.Minute).Should(BeClosed())
		})

	})

	Context("using ike develop with istio-bookinfo example", func() {

		tmpPath := test.NewTmpPath()

		var (
			namespace,
			tmpDir string
		)

		BeforeEach(func() {
			tmpPath.SetPath(os.Getenv("PATH"), path.Dir(cmd.CurrentDir())+"/dist")
			namespace = naming.RandName(16)
			tmpDir = test.TmpDir(GinkgoT(), "namespace-"+namespace)
			Expect(cmd.BinaryExists("ike", "make sure you have binary in the ./dist folder. Try make compile at least")).To(BeTrue())

			<-cmd.Execute("oc", "new-project", namespace).Done()
			e2e.UpdateSecurityConstraintsFor(namespace)
			e2e.LoadIstioResources(namespace, tmpDir)
			e2e.DeployBookinfoInto(namespace, tmpDir)
		})

		AfterEach(func() {
			tmpPath.Restore()
			<-cmd.Execute("oc", "delete", "project", namespace).Done()
		})

		It("should watch for changes in details service and serve it", func() {
			// before: check is original product page is up and running
			Eventually(func() (string, error) {
				return e2e.GetBody("http://istio-ingressgateway-istio-system.127.0.0.1.nip.io/productpage")
			}, 1*time.Minute).Should(ContainSubstring("PublisherA"))

			// given we have details code locally
			details, err := e2e.GetBody("https://raw.githubusercontent.com/istio/istio/master/samples/bookinfo/src/details/details.rb")
			Expect(err).ToNot(HaveOccurred())
			e2e.CreateFile(tmpDir+"/details.rb", details)

			// when we start ike with watch
			ikeWithWatch := cmd.ExecuteInDir(tmpDir, "ike", "develop",
				"--deployment", "details-v1",
				"--port", "9080",
				"--method", "inject-tcp",
				"--watch",
				"--run", "ruby details.rb 9080",
			)

			// and modify the service
			modifiedDetails := strings.Replace(details, "PublisherA", "Publisher Ike", 1)
			e2e.CreateFile(tmpDir+"/details.rb", modifiedDetails)

			// then
			Eventually(func() (string, error) {
				return e2e.GetBody("http://istio-ingressgateway-istio-system.127.0.0.1.nip.io/productpage")
			}, 1*time.Minute).Should(ContainSubstring("Publisher Ike"))

			Expect(ikeWithWatch.Stop()).ToNot(HaveOccurred())
			Eventually(ikeWithWatch.Done(), 1*time.Minute).Should(BeClosed())
		})

	})
})

func callGetOn(name string) func() (string, error) {
	return func() (string, error) {
		return e2e.GetBody(fmt.Sprintf("http://%[1]s-%[1]s.127.0.0.1.nip.io", name))
	}
}
