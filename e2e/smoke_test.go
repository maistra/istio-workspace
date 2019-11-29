package e2e_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	. "github.com/maistra/istio-workspace/e2e"
	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/pkg/naming"
	"github.com/maistra/istio-workspace/pkg/shell"
	"github.com/maistra/istio-workspace/test"
	testshell "github.com/maistra/istio-workspace/test/shell"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Smoke End To End Tests - against OpenShift Cluster with Istio (maistra)", func() {

	// Can't be ran without a session (not using --swap-deployment)
	XContext("using ike develop in offline mode", func() {

		var (
			appName,
			tmpDir string
		)

		BeforeEach(func() {
			appName = generateNamespaceName()
			tmpDir = test.TmpDir(GinkgoT(), "app-"+appName)
			Expect(shell.BinaryExists("ike", "make sure you have binary in the ./dist folder. Try make compile at least")).To(BeTrue())
		})

		AfterEach(func() { cleanupNamespace(appName) })

		It("should watch python code changes and replace service when they occur", func() {

			CreateNewApp(appName)
			Eventually(callGetOn(appName), 1*time.Minute).Should(Equal("Hello, world!\n"))

			OriginalServerCodeIn(tmpDir)
			ikeWithWatch := testshell.ExecuteInDir(tmpDir, "ike", "develop",
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

	Context("using ike with scenarios", func() {

		var (
			namespace,
			tmpDir string
			scenario string
		)

		JustBeforeEach(func() {
			namespace = generateNamespaceName()
			tmpDir = test.TmpDir(GinkgoT(), "namespace-"+namespace)

			LoginAsTestPowerUser()
			<-testshell.Execute(NewProjectCmd(namespace)).Done()

			UpdateSecurityConstraintsFor(namespace)
			PushOperatorImage(namespace)
			InstallLocalOperator(namespace)
			BuildTestService(namespace)
			DeployTestScenario(scenario, namespace)
		})

		AfterEach(func() {
			if CurrentGinkgoTestDescription().Failed {
				pods := GetAllPods(namespace)
				for _, pod := range pods {
					printBanner()
					fmt.Println("Logs of " + pod)
					fmt.Println(LogsOf(namespace, pod))
					printBanner()
					StateOf(namespace, pod)
					printBanner()
				}
				GetEvents(namespace)
				DumpTelepresenceLog(tmpDir)
			}
			cleanupNamespace(namespace)
		})

		Describe("k8s deployment", func() {

			BeforeEach(func() {
				scenario = "scenario-1"
			})

			Context("basic deployment modifications", func() {
				It("should watch for changes in ratings service and serve it", func() {
					verifyThatResponseMatchesModifiedService(tmpDir, namespace)
				})

				It("should watch for changes in ratings service in specified namespace and serve it", func() {
					verifyThatResponseMatchesModifiedService(tmpDir, namespace)
				})
			})

			Context("deployment create/delete operations", func() {
				var registry string
				preparedImageV1 := "prepared-image"
				preparedImageV2 := "image-prepared"

				JustBeforeEach(func() {
					BuildTestServicePreparedImage(preparedImageV1, namespace)
					BuildTestServicePreparedImage(preparedImageV2, namespace)
					registry = GetDockerRegistryInternal()
				})

				It("should watch for changes in ratings service and serve it", func() {
					productPageURL := GetIstioIngressHostname() + "/test-service/productpage"

					Eventually(AllPodsReady(namespace), 5*time.Minute, 5*time.Second).Should(BeTrue())

					Eventually(call(productPageURL, map[string]string{
						"Host": GetGatewayHost(namespace)}),
						3*time.Minute, 1*time.Second).
						Should(And(ContainSubstring("ratings-v1"), Not(ContainSubstring(preparedImageV1))))

					// switch to different namespace - so we also test -n parameter of $ ike
					<-testshell.Execute("oc project default").Done()

					// when we start ike to create
					ikeWithCreateV1 := testshell.ExecuteInDir(tmpDir, "ike", "create",
						"--deployment", "ratings-v1",
						"-n", namespace,
						"--route", "header:x-test-suite=smoke",
						"--image", registry+"/"+namespace+"/istio-workspace-test-prepared-"+preparedImageV1+":latest",
						"-s", "test-session",
					)
					Eventually(ikeWithCreateV1.Done(), 1*time.Minute).Should(BeClosed())

					// ensure the new service is running
					Eventually(AllPodsReady(namespace), 5*time.Minute, 5*time.Second).Should(BeTrue())

					// check original response
					Eventually(call(productPageURL, map[string]string{
						"Host":         GetGatewayHost(namespace),
						"x-test-suite": "smoke"}),
						3*time.Minute, 1*time.Second).
						Should(And(ContainSubstring(preparedImageV1), Not(ContainSubstring("ratings-v1"))))

					// but also check if prod is intact
					Eventually(call(productPageURL, map[string]string{}), 3*time.Minute, 1*time.Second).
						ShouldNot(ContainSubstring(preparedImageV1))

					// when we start ike to create with a updated v
					ikeWithCreateV2 := testshell.ExecuteInDir(tmpDir, "ike", "create",
						"--deployment", "ratings-v1",
						"-n", namespace,
						"--route", "header:x-test-suite=smoke",
						"--image", registry+"/"+namespace+"/istio-workspace-test-prepared-"+preparedImageV2+":latest",
						"-s", "test-session",
					)
					Eventually(ikeWithCreateV2.Done(), 1*time.Minute).Should(BeClosed())

					// check original response
					Eventually(call(productPageURL, map[string]string{
						"Host":         GetGatewayHost(namespace),
						"x-test-suite": "smoke"}),
						3*time.Minute, 1*time.Second).
						Should(And(ContainSubstring(preparedImageV2), Not(ContainSubstring("ratings-v1"))))

					// but also check if prod is intact
					Eventually(call(productPageURL, map[string]string{}), 3*time.Minute, 1*time.Second).
						ShouldNot(ContainSubstring(preparedImageV2))

					// when we start ike to delete
					ikeWithDelete := testshell.ExecuteInDir(tmpDir, "ike", "delete",
						"--deployment", "ratings-v1",
						"-n", namespace,
						"-s", "test-session",
					)
					Eventually(ikeWithDelete.Done(), 1*time.Minute).Should(BeClosed())

					// check original response
					Eventually(call(productPageURL, map[string]string{
						"Host":         GetGatewayHost(namespace),
						"x-test-suite": "smoke"}),
						3*time.Minute, 1*time.Second).
						Should(And(ContainSubstring("ratings-v1"), Not(ContainSubstring(preparedImageV2))))

					// but also check if prod is intact
					Eventually(call(productPageURL, map[string]string{
						"Host": GetGatewayHost(namespace)}),
						3*time.Minute, 1*time.Second).
						Should(And(ContainSubstring("ratings-v1"), Not(ContainSubstring(preparedImageV2))))
				})

			})

		})

		Context("openshift deploymentconfig", func() {

			BeforeEach(func() {
				scenario = "scenario-2"
			})

			It("should watch for changes in ratings service in specified namespace and serve it", func() {
				verifyThatResponseMatchesModifiedService(tmpDir, namespace)
			})
		})
	})
})

func printBanner() {
	fmt.Println("---------------------------------------------------------------------")
}

func generateNamespaceName() string {
	return "ike-tests-" + naming.RandName(16)
}

func cleanupNamespace(namespace string) {
	if keepStr, found := os.LookupEnv("IKE_E2E_KEEP_NS"); found {
		keep, _ := strconv.ParseBool(keepStr)
		if keep {
			return
		}
	}
	CleanupTestScenario(namespace)
	<-testshell.Execute("oc delete project " + namespace).Done()
}

func verifyThatResponseMatchesModifiedService(tmpDir, namespace string) {
	productPageURL := GetIstioIngressHostname() + "/test-service/productpage"

	Eventually(AllPodsReady(namespace), 5*time.Minute, 5*time.Second).Should(BeTrue())

	Eventually(call(productPageURL, map[string]string{
		"Host": GetGatewayHost(namespace)}),
		3*time.Minute, 1*time.Second).Should(ContainSubstring("ratings-v1"))

	// switch to different namespace - so we also test -n parameter of $ ike
	<-testshell.Execute("oc project default").Done()

	// given we have details code locally
	CreateFile(tmpDir+"/ratings.rb", PublisherRuby)

	// when we start ike with watch
	ikeWithWatch := testshell.ExecuteInDir(tmpDir, "ike", "develop",
		"--deployment", "ratings-v1",
		"-n", namespace,
		"--port", "9080",
		"--method", "inject-tcp",
		"--watch",
		"--run", "ruby ratings.rb 9080",
		"--route", "header:x-test-suite=smoke",
	)

	// ensure the new service is running
	Eventually(AllPodsReady(namespace), 5*time.Minute, 5*time.Second).Should(BeTrue())

	// check original response
	Eventually(call(productPageURL, map[string]string{
		"Host":         GetGatewayHost(namespace),
		"x-test-suite": "smoke"}),
		3*time.Minute, 1*time.Second).Should(ContainSubstring("PublisherA"))

	// then modify the service
	modifiedDetails := strings.Replace(PublisherRuby, "PublisherA", "Publisher Ike", 1)
	CreateFile(tmpDir+"/ratings.rb", modifiedDetails)

	// then verify new content being served
	Eventually(call(productPageURL, map[string]string{
		"Host":         GetGatewayHost(namespace),
		"x-test-suite": "smoke"}),
		3*time.Minute, 1*time.Second).Should(ContainSubstring("Publisher Ike"))

	// but also check if prod is intact
	Eventually(call(productPageURL, map[string]string{
		"Host": GetGatewayHost(namespace)}),
		3*time.Minute, 1*time.Second).ShouldNot(And(ContainSubstring("PublisherA"), ContainSubstring("Publisher Ike")))

	stopFailed := ikeWithWatch.Stop()
	Expect(stopFailed).ToNot(HaveOccurred())

	Eventually(ikeWithWatch.Done(), 1*time.Minute).Should(BeClosed())
}

func call(routeURL string, headers map[string]string) func() (string, error) { //nolint[:unparam]
	return func() (string, error) {
		fmt.Printf("[%s] Checking [%s] with headers [%s]...\n", time.Now().Format("2006-01-02 15:04:05.001"), routeURL, headers)
		return GetBodyWithHeaders(routeURL, headers)
	}
}

func callGetOn(name string) func() (string, error) {
	return func() (string, error) {
		return GetBody(fmt.Sprintf("http://%[1]s-%[1]s."+GetClusterHost(), name))
	}
}
