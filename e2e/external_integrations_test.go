package e2e_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/maistra/istio-workspace/e2e/infra"
	. "github.com/maistra/istio-workspace/e2e/verify"
	"github.com/maistra/istio-workspace/test"
	testshell "github.com/maistra/istio-workspace/test/shell"
)

var _ = Describe("External integrations", func() {

	var (
		namespace,
		scenario,
		sessionName,
		tmpDir string
	)

	tmpFs := test.NewTmpFileSystem(GinkgoT())

	JustBeforeEach(func() {
		namespace = GenerateNamespaceName()
		tmpDir = tmpFs.Dir("namespace-" + namespace)

		<-testshell.Execute(NewProjectCmd(namespace)).Done()

		PrepareEnv(namespace)

		InstallLocalOperator(namespace)
		Eventually(AllDeploymentsAndPodsReady(namespace), 10*time.Minute, 5*time.Second).Should(BeTrue())

		// FIX Smelly to rely on global state. Scenario is set in subsequent beforeEach for given context
		DeployTestScenario(scenario, namespace)
		sessionName = GenerateSessionName()
	})

	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			DumpEnvironmentDebugInfo(namespace, tmpDir)
		} else {
			CleanupNamespace(namespace, false)
			tmpFs.Cleanup()
		}
	})

	When("Using ike with Tekton Pipelines", func() {

		BeforeEach(func() {
			scenario = "http-seq"
		})

		It("should build and expose service preview through session url", func() {
			defer test.TemporaryEnvVars("TEST_NAMESPACE", namespace, "TEST_SESSION_NAME", sessionName)()

			host := sessionName + "." + GetGatewayHost(namespace)
			By("deploying Tekton tasks")
			testshell.WaitForSuccess(
				testshell.ExecuteInProjectRoot("make tekton-deploy"),
			)

			EnsureAllDeploymentPodsAreReady(namespace)
			EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))

			By("running tekton ike-create task")
			testshell.WaitForSuccess(
				testshell.ExecuteInProjectRoot("make tekton-test-ike-create"),
			)

			Eventually(TaskIsDone(namespace, "ike-create-run"), 5*time.Minute, 5*time.Second).Should(BeTrue())
			Expect(TaskResult(namespace, "ike-create-run", "url")).To(Equal(host))

			By("creating preview url")
			testshell.WaitForSuccess(
				testshell.ExecuteInProjectRoot("make tekton-test-ike-session-url"),
			)

			Eventually(TaskIsDone(namespace, "ike-session-url-run"), 5*time.Minute, 5*time.Second).Should(BeTrue())
			Expect(TaskResult(namespace, "ike-session-url-run", "url")).To(Equal(host))

			By("ensuring the new service is running")
			EnsureAllDeploymentPodsAreReady(namespace)

			By("checking new service is reachable")
			EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring(PreparedImageV1), Not(ContainSubstring("ratings-v1")))

			By("ensuring production version is intact")
			EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))

			By("deleting service preview (e.g. after PR merge)")
			testshell.WaitForSuccess(
				testshell.ExecuteInProjectRoot("make tekton-test-ike-delete"),
			)
			Eventually(TaskIsDone(namespace, "ike-delete-run"), 5*time.Minute, 5*time.Second).Should(BeTrue())

			By("ensuring new version is not reachable after removal")
			EnsureSessionRouteIsNotReachable(namespace, sessionName, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))

			By("ensuring production version is still available")
			EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))
		})
	})
})
