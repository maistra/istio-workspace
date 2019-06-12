package e2e_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/maistra/istio-workspace/cmd/ike/cmd"
	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/pkg/naming"
	"github.com/maistra/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Smoke End To End Tests - against OpenShift Cluster with Istio (maistra)", func() {

	// Can't be ran without a session (not using --swap-deployment)
	XContext("using ike develop in offline mode", func() {

		tmpPath := test.NewTmpPath()

		var (
			appName,
			tmpDir string
		)

		BeforeEach(func() {
			tmpPath.SetPath(path.Dir(cmd.CurrentDir())+"/dist", os.Getenv("PATH"))
			appName = naming.RandName(16)
			tmpDir = test.TmpDir(GinkgoT(), "app-"+appName)
			Expect(cmd.BinaryExists("ike", "make sure you have binary in the ./dist folder. Try make compile at least")).To(BeTrue())
		})

		AfterEach(func() {
			tmpPath.Restore()
			<-cmd.Execute("oc delete project " + appName).Done()
		})

		It("should watch python code changes and replace service when they occur", func() {

			CreateNewApp(appName)
			Eventually(callGetOn(appName), 1*time.Minute).Should(Equal("Hello, world!\n"))

			OriginalServerCodeIn(tmpDir)
			ikeWithWatch := cmd.ExecuteInDir(tmpDir, "ike", "develop",
				"--deployment", appName,
				"--port", "8000",
				"--method", "inject-tcp",
				"--watch",
				"--run", "python3 server.py",
				"--offline",
			)
			Eventually(callGetOn(appName), 3*time.Minute, 200*time.Millisecond).Should(Equal("Hello, world!\n"))

			ModifyServerCodeIn(tmpDir)

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
			scenario string
		)

		JustBeforeEach(func() {
			tmpPath.SetPath(path.Dir(cmd.CurrentDir())+"/dist", os.Getenv("PATH"))
			namespace = naming.RandName(16)
			tmpDir = test.TmpDir(GinkgoT(), "namespace-"+namespace)
			Expect(cmd.BinaryExists("ike", "make sure you have binary in the ./dist folder. Try make compile at least")).To(BeTrue())

			<-cmd.Execute("oc login -u developer").Done()
			<-cmd.Execute("oc new-project " + namespace).Done()
			UpdateSecurityConstraintsFor(namespace)
			BuildTestService(namespace)
			DeployTestScenario(scenario, namespace)
		})

		AfterEach(func() {
			tmpPath.Restore()
			<-cmd.Execute("oc delete project " + namespace).Done()
		})

		Context("scenario-1-basic-deployment", func() {
			BeforeEach(func() {
				scenario = "scenario-1"
			})

			It("should watch for changes in ratings service and serve it", func() {
				Eventually(AllPodsNotInState(namespace, "Running"), 3*time.Minute, 2*time.Second).
					Should(ContainSubstring("No resources found"))

				Eventually(func() (string, error) {
					return GetBody("http://istio-ingressgateway-istio-system.127.0.0.1.nip.io/productpage")
				}, 3*time.Minute, 1*time.Second).Should(ContainSubstring("ratings-v1"))

				// given we have details code locally
				CreateFile(tmpDir+"/ratings.rb", DetailsRuby)

				// when we start ike with watch
				ikeWithWatch := cmd.ExecuteInDir(tmpDir, "ike", "develop",
					"--deployment", "ratings-v1",
					"--port", "9080",
					"--watch",
					"--run", "ruby ratings.rb 9080",
					"--route", "header:x-test-suite=smoke",
				)

				// ensure the new service is running
				Eventually(AllPodsNotInState(namespace, "Running"), 3*time.Minute, 2*time.Second).
					Should(ContainSubstring("No resources found"))

				// and modify the service
				modifiedDetails := strings.Replace(DetailsRuby, "PublisherA", "Publisher Ike", 1)
				CreateFile(tmpDir+"/ratings.rb", modifiedDetails)

				// then
				Eventually(func() (string, error) {
					fmt.Printf("[%s] checking...\n", time.Now().Format("2006-01-02 15:04:05.001"))
					return GetBodyWithHeaders("http://istio-ingressgateway-istio-system.127.0.0.1.nip.io/productpage", map[string]string{"x-test-suite": "smoke"})
				}, 3*time.Minute, 1*time.Second).Should(ContainSubstring("Publisher Ike"))

				stopFailed := ikeWithWatch.Stop()
				Expect(stopFailed).ToNot(HaveOccurred())

				Eventually(ikeWithWatch.Done(), 1*time.Minute).Should(BeClosed())
			})

			It("should watch for changes in ratings service in specified namespace and serve it", func() {
				verifyThatResponseMatchesModifiedService(tmpDir, namespace)
			})
		})

		Context("scenario-2-basic-deploymentconfig", func() {
			BeforeEach(func() {
				scenario = "scenario-2"
			})

			It("should watch for changes in ratings service in specified namespace and serve it", func() {
				verifyThatResponseMatchesModifiedService(tmpDir, namespace)
			})
		})
	})
})

func verifyThatResponseMatchesModifiedService(tmpDir, namespace string) {
	Eventually(AllPodsNotInState(namespace, "Running"), 3*time.Minute, 2*time.Second).
		Should(ContainSubstring("No resources found"))

	Eventually(func() (string, error) {
		return GetBody("http://istio-ingressgateway-istio-system.127.0.0.1.nip.io/productpage")
	}, 3*time.Minute, 1*time.Second).Should(ContainSubstring("ratings-v1"))

	// switch to different namespace
	<-cmd.Execute("oc project myproject").Done()

	// given we have details code locally
	CreateFile(tmpDir+"/ratings.rb", DetailsRuby)

	// when we start ike with watch
	ikeWithWatch := cmd.ExecuteInDir(tmpDir, "ike", "develop",
		"--deployment", "ratings-v1",
		"-n", namespace,
		"--port", "9080",
		"--method", "inject-tcp",
		"--watch",
		"--run", "ruby ratings.rb 9080",
		"--route", "header:x-test-suite=smoke",
	)

	// ensure the new service is running
	Eventually(AllPodsNotInState(namespace, "Running"), 3*time.Minute, 2*time.Second).
		Should(ContainSubstring("No resources found"))

	// and modify the service
	modifiedDetails := strings.Replace(DetailsRuby, "PublisherA", "Publisher Ike", 1)
	CreateFile(tmpDir+"/ratings.rb", modifiedDetails)

	// then
	Eventually(func() (string, error) {
		fmt.Printf("[%s] checking...\n", time.Now().Format("2006-01-02 15:04:05.001"))
		return GetBodyWithHeaders("http://istio-ingressgateway-istio-system.127.0.0.1.nip.io/productpage", map[string]string{"x-test-suite": "smoke"})
	}, 3*time.Minute, 1*time.Second).Should(ContainSubstring("Publisher Ike"))

	stopFailed := ikeWithWatch.Stop()
	Expect(stopFailed).ToNot(HaveOccurred())

	Eventually(ikeWithWatch.Done(), 1*time.Minute).Should(BeClosed())
}

func callGetOn(name string) func() (string, error) {
	return func() (string, error) {
		return GetBody(fmt.Sprintf("http://%[1]s-%[1]s.127.0.0.1.nip.io", name))
	}
}
