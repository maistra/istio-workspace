package e2e_test

import (
	"os"
	"strings"
	"time"

	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/test"
	testshell "github.com/maistra/istio-workspace/test/shell"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("End To End Tests - non standard scenarios", func() {

	Context("using ike with scenarios", func() {

		var (
			namespace,
			gwNamespace,
			scenario,
			sessionName,
			tmpDir string
		)

		tmpFs := test.NewTmpFileSystem(GinkgoT())

		Context("Gateway in another namespace", func() {

			var restoreEnvVars func()

			BeforeEach(func() {
				scenario = "scenario-1" //nolint:goconst //reason no need for constant (yet)

				namespace = generateNamespaceName()
				gwNamespace = "gw-" + namespace
				tmpDir = tmpFs.Dir("namespace-" + namespace)

				<-testshell.Execute(CreateNamespaceCmd(namespace)).Done()
				<-testshell.Execute(CreateNamespaceCmd(gwNamespace)).Done()
				restoreEnvVars = test.TemporaryEnvVars("TEST_GW_NAMESPACE", gwNamespace)

				PrepareEnvForOpenshift(namespace)

				InstallMultiNamespaceOperator(namespace, gwNamespace)
				Eventually(AllDeploymentsAndPodsReady(namespace), 10*time.Minute, 5*time.Second).Should(BeTrue())
				DeployTestScenario(scenario, namespace)
				sessionName = GenerateSessionName()
			})

			AfterEach(func() {
				if CurrentSpecReport().Failed() {
					PrintFailureDetails(namespace, tmpDir)
				} else {
					CleanupNamespace(namespace, false)
					tmpFs.Cleanup()
				}
				<-testshell.Execute(DeleteNamespaceCmd(gwNamespace)).Done()
				os.Setenv("OPERATOR_WATCH_NAMESPACE", "")
				restoreEnvVars()
			})

			It("should watch for changes in connected service and serve it", func() {
				EnsureAllDeploymentPodsAreReady(namespace)
				EnsureProdRouteIsReachable(namespace, ContainSubstring("productpage-v1"))
				deploymentCount := GetResourceCount("deployment", namespace)

				// given we have details code locally
				CreateFile(tmpDir+"/productpage.py", PublisherService)

				ike := RunIke(tmpDir, "develop",
					"--deployment", "deployment/productpage-v1",
					"--port", "9080",
					"--method", "inject-tcp",
					"--watch",
					"--run", "python productpage.py 9080",
					"--route", "header:x-test-suite=smoke",
					"--session", sessionName,
					"--namespace", namespace,
				)
				defer func() {
					Stop(ike)
				}()
				go FailOnCmdError(ike, GinkgoT())

				EnsureCorrectNumberOfResources(deploymentCount+1, "deployment", namespace)
				EnsureAllDeploymentPodsAreReady(namespace)
				EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring("PublisherA"))

				// then modify the service
				modifiedDetails := strings.Replace(PublisherService, "PublisherA", "Publisher Ike", 1)
				CreateFile(tmpDir+"/productpage.py", modifiedDetails)

				EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring("Publisher Ike"))

				Stop(ike)
				EnsureProdRouteIsReachable(namespace, ContainSubstring("productpage-v1"))
			})
		})

	})

})
