package e2e_test

import (
	"time"

	. "github.com/maistra/istio-workspace/e2e"
	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/test"
	testshell "github.com/maistra/istio-workspace/test/shell"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resources reconciliation", func() {

	var (
		namespace,
		registry,
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

	Context("reconcile on change to related resources", func() {

		BeforeEach(func() {
			scenario = "scenario-1"
			registry = GetInternalContainerRegistry()
		})

		It("should create/delete deployment with prepared image", func() {
			EnsureAllDeploymentPodsAreReady(namespace)
			EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))

			// when we start ike to create
			ikeCreate := RunIke(tmpDir, "create",
				"--deployment", "ratings-v1",
				"-n", namespace,
				"--route", "header:x-test-suite=smoke",
				"--image", registry+"/"+GetDevRepositoryName()+"/istio-workspace-test-prepared-"+PreparedImageV1+":"+GetImageTag(),
				"--session", sessionName,
			)
			Eventually(ikeCreate.Done(), 1*time.Minute).Should(BeClosed())
			testshell.WaitForSuccess(ikeCreate)

			// ensure the new service is running
			EnsureAllDeploymentPodsAreReady(namespace)

			// check original response
			EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring(PreparedImageV1), Not(ContainSubstring("ratings-v1")))

			// but also check if prod is intact
			EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))

			// then reset scenario
			DeployTestScenario(scenario, namespace)

			// check original response is still intact
			EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring(PreparedImageV1), Not(ContainSubstring("ratings-v1")))

			// but also check if prod is intact
			EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))

			// when we start ike to delete
			ikeDel := RunIke(tmpDir, "delete",
				"--deployment", "ratings-v1",
				"-n", namespace,
				"--session", sessionName,
			)
			Eventually(ikeDel.Done(), 1*time.Minute).Should(BeClosed())
			testshell.WaitForSuccess(ikeDel)

			// check original response
			EnsureSessionRouteIsNotReachable(namespace, sessionName, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))

			// but also check if prod is intact
			EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
		})
	})
})
