package e2e_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-cmd/cmd"

	. "github.com/maistra/istio-workspace/e2e"
	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/pkg/naming"
	"github.com/maistra/istio-workspace/test"
	testshell "github.com/maistra/istio-workspace/test/shell"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("Smoke End To End Tests - against OpenShift Cluster with Istio (maistra)", func() {

	Context("using ike with scenarios", func() {

		var (
			namespace,
			registry,
			scenario,
			sessionName,
			tmpDir string
		)

		JustBeforeEach(func() {
			namespace = generateNamespaceName()
			tmpDir = test.TmpDir(GinkgoT(), "namespace-"+namespace)

			<-testshell.Execute(NewProjectCmd(namespace)).Done()

			PrepareEnv(namespace)

			InstallLocalOperator(namespace)
			Eventually(AllDeploymentsAndPodsReady(namespace), 10*time.Minute, 5*time.Second).Should(BeTrue())
			DeployTestScenario(scenario, namespace)
			sessionName = GenerateSessionName()
		})

		AfterEach(func() {
			if CurrentGinkgoTestDescription().Failed {
				DumpEnvironmentDebugInfo(namespace, tmpDir)
			}
			cleanupNamespace(namespace)
		})

		Context("k8s deployment", func() {

			Context("http protocol", func() {

				BeforeEach(func() {
					scenario = "scenario-1" //nolint:goconst //reason no need for constant (yet)
					registry = GetDockerRegistryInternal()
				})

				Context("basic deployment modifications", func() {

					It("should watch for changes in ratings service and serve it", func() {
						EnsureAllDeploymentPodsAreReady(namespace)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))

						// given we have details code locally
						CreateFile(tmpDir+"/ratings.py", PublisherService)

						ike := RunIke(tmpDir, "develop",
							"--deployment", "ratings-v1",
							"--port", "9080",
							"--method", "inject-tcp",
							"--watch",
							"--run", "python ratings.py 9080",
							"--route", "header:x-test-suite=smoke",
							"--session", sessionName,
							"--namespace", namespace,
						)
						EnsureAllDeploymentPodsAreReady(namespace)
						EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring("PublisherA"))

						// then modify the service
						modifiedDetails := strings.Replace(PublisherService, "PublisherA", "Publisher Ike", 1)
						CreateFile(tmpDir+"/ratings.py", modifiedDetails)

						EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring("Publisher Ike"))

						Stop(ike)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
					})
				})

				Context("deployment create/delete operations", func() {

					It("should watch for changes in ratings service and serve it", func() {
						EnsureAllDeploymentPodsAreReady(namespace)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))

						ChangeNamespace("default")

						// when we start ike to create
						ike1 := RunIke(tmpDir, "create",
							"--deployment", "ratings-v1",
							"-n", namespace,
							"--route", "header:x-test-suite=smoke",
							"--image", registry+"/"+GetDevRepositoryName()+"/istio-workspace-test-prepared-"+PreparedImageV1+":"+GetImageTag(),
							"--session", sessionName,
						)
						Eventually(ike1.Done(), 1*time.Minute).Should(BeClosed())

						// ensure the new service is running
						EnsureAllDeploymentPodsAreReady(namespace)

						// check original response
						EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring(PreparedImageV1), Not(ContainSubstring("ratings-v1")))

						// but also check if prod is intact
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))

						// when we start ike to create with a updated v
						ike2 := RunIke(tmpDir, "create",
							"--deployment", "ratings-v1",
							"-n", namespace,
							"--route", "header:x-test-suite=smoke",
							"--image", registry+"/"+GetDevRepositoryName()+"/istio-workspace-test-prepared-"+PreparedImageV2+":"+GetImageTag(),
							"--session", sessionName,
						)
						Eventually(ike2.Done(), 1*time.Minute).Should(BeClosed())

						// check original response
						EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring(PreparedImageV2), Not(ContainSubstring("ratings-v1")))

						// but also check if prod is intact
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV2)))

						// when we start ike to delete
						ikeDel := RunIke(tmpDir, "delete",
							"--deployment", "ratings-v1",
							"-n", namespace,
							"--session", sessionName,
						)
						Eventually(ikeDel.Done(), 1*time.Minute).Should(BeClosed())

						// check original response
						EnsureSessionRouteIsNotReachable(namespace, sessionName, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV2)))

						// but also check if prod is intact
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
					})

				})
			})

			Context("grpc protocol", func() {
				BeforeEach(func() {
					scenario = "scenario-1.1"
				})

				Context("basic deployment modifications", func() {
					It("should take over ratings service and serve it", func() {
						EnsureAllDeploymentPodsAreReady(namespace)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))

						ike := RunIke(testshell.GetProjectDir(), "develop",
							"--deployment", "ratings-v1",
							"--port", "9081",
							"--method", "inject-tcp",
							"--run", "go run ./test/cmd/test-service -serviceName=PublisherA",
							"--route", "header:x-test-suite=smoke",
							"--session", sessionName,
							"--namespace", namespace,
						)
						EnsureAllDeploymentPodsAreReady(namespace)

						EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring("PublisherA"), ContainSubstring("grpc"))

						Stop(ike)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
					})
				})
			})
		})

		Context("openshift deploymentconfig", func() {

			BeforeEach(func() {
				if !RunsAgainstOpenshift {
					Skip("DeploymentConfig is Openshift-specific resource and it won't work against plain k8s. " +
						"Tests for regular k8s deployment can be found in the same test suite.")
				}
				scenario = "scenario-2"
			})

			It("should watch for changes in ratings service in specified namespace and serve it", func() {
				ChangeNamespace(namespace)
				EnsureAllDeploymentConfigPodsAreReady(namespace)
				EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))

				// given we have details code locally
				CreateFile(tmpDir+"/ratings.py", PublisherService)

				ike := RunIke(tmpDir, "develop",
					"--deployment", "ratings-v1",
					"--port", "9080",
					"--method", "inject-tcp",
					"--watch",
					"--run", "python ratings.py 9080",
					"--route", "header:x-test-suite=smoke",
					"--session", sessionName,
				)
				EnsureAllDeploymentConfigPodsAreReady(namespace)
				EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring("PublisherA"))

				// then modify the service
				modifiedDetails := strings.Replace(PublisherService, "PublisherA", "Publisher Ike", 1)
				CreateFile(tmpDir+"/ratings.py", modifiedDetails)

				EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring("Publisher Ike"))

				Stop(ike)
				EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
			})
		})

		Context("reconcile on change to related resources", func() {

			BeforeEach(func() {
				scenario = "scenario-1"
			})

			It("should watch for changes in ratings service and serve it", func() {
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

				// check original response
				EnsureSessionRouteIsNotReachable(namespace, sessionName, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))

				// but also check if prod is intact
				EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
			})

		})

		Context("verify external integrations", func() {

			Context("Tekton", func() {

				BeforeEach(func() {
					scenario = "scenario-1"
				})

				It("should create, get, and delete", func() {
					defer test.TemporaryEnvVars("TEST_NAMESPACE", namespace, "TEST_SESSION_NAME", sessionName)()

					host := sessionName + "." + GetGatewayHost(namespace)

					<-testshell.ExecuteInProjectRoot("make tekton-deploy").Done()

					EnsureAllDeploymentPodsAreReady(namespace)
					EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))

					<-testshell.ExecuteInProjectRoot("make tekton-test-ike-create").Done()
					Eventually(TaskIsDone(namespace, "ike-create-run"), 5*time.Minute, 5*time.Second).Should(BeTrue())
					Expect(TaskResult(namespace, "ike-create-run", "url")).To(Equal(host))

					// verify session url
					<-testshell.ExecuteInProjectRoot("make tekton-test-ike-session-url").Done()
					Eventually(TaskIsDone(namespace, "ike-session-url-run"), 5*time.Minute, 5*time.Second).Should(BeTrue())
					Expect(TaskResult(namespace, "ike-session-url-run", "url")).To(Equal(host))

					// ensure the new service is running
					EnsureAllDeploymentPodsAreReady(namespace)

					// check original response
					EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring(PreparedImageV1), Not(ContainSubstring("ratings-v1")))

					// but also check if prod is intact
					EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))

					<-testshell.ExecuteInProjectRoot("make tekton-test-ike-delete").Done()
					Eventually(TaskIsDone(namespace, "ike-delete-run"), 5*time.Minute, 5*time.Second).Should(BeTrue())

					// check original response
					EnsureSessionRouteIsNotReachable(namespace, sessionName, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))

					// but also check if prod is intact
					EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))
				})
			})
		})
	})
})

// EnsureAllDeploymentPodsAreReady make sure all Pods are in Ready state in given namespace.
func EnsureAllDeploymentPodsAreReady(namespace string) {
	Eventually(AllDeploymentsAndPodsReady(namespace), 5*time.Minute, 5*time.Second).Should(BeTrue())
}

// EnsureAllDeploymentConfigPodsAreReady make sure all Pods are in Ready state in given namespace.
func EnsureAllDeploymentConfigPodsAreReady(namespace string) {
	Eventually(AllDeploymentConfigsAndPodsReady(namespace), 10*time.Minute, 5*time.Second).Should(BeTrue())
}

// EnsureProdRouteIsReachable can be reached with no special arguments.
func EnsureProdRouteIsReachable(namespace string, matchers ...types.GomegaMatcher) {
	productPageURL := GetIstioIngressHostname() + "/test-service/productpage"

	Eventually(call(productPageURL, map[string]string{
		"Host": GetGatewayHost(namespace)}),
		5*time.Minute, 1*time.Second).Should(And(matchers...))
}

// EnsureSessionRouteIsReachable the manipulated route is reachable.
func EnsureSessionRouteIsReachable(namespace, sessionName string, matchers ...types.GomegaMatcher) {
	productPageURL := GetIstioIngressHostname() + "/test-service/productpage"

	// check original response using headers
	Eventually(call(productPageURL, map[string]string{
		"Host":         GetGatewayHost(namespace),
		"x-test-suite": "smoke"}),
		5*time.Minute, 1*time.Second).Should(And(matchers...))

	// check original response using host route
	Eventually(call(productPageURL, map[string]string{
		"Host": sessionName + "." + GetGatewayHost(namespace)}),
		5*time.Minute, 1*time.Second).Should(And(matchers...))
}

// EnsureSessionRouteIsNotReachable the manipulated route is reachable.
func EnsureSessionRouteIsNotReachable(namespace, sessionName string, matchers ...types.GomegaMatcher) {
	productPageURL := GetIstioIngressHostname() + "/test-service/productpage"

	// check original response using headers
	Eventually(call(productPageURL, map[string]string{
		"Host":         GetGatewayHost(namespace),
		"x-test-suite": "smoke"}),
		5*time.Minute, 1*time.Second).Should(And(matchers...))
}

// ChangeNamespace switch to different namespace - so we also test -n parameter of $ ike.
// That only works for oc cli, as kubectl by default uses `default` namespace.
func ChangeNamespace(namespace string) {
	if RunsAgainstOpenshift {
		<-testshell.Execute("oc project " + namespace).Done()
	}
}

// RunIke runs the ike cli in the given dir.
func RunIke(dir string, arguments ...string) *cmd.Cmd {
	return testshell.ExecuteInDir(dir, "ike", arguments...)
}

// Stop shuts down the process.
func Stop(ike *cmd.Cmd) {
	stopFailed := ike.Stop()
	Expect(stopFailed).ToNot(HaveOccurred())

	Eventually(ike.Done(), 1*time.Minute).Should(BeClosed())
}

// DumpEnvironmentDebugInfo prints tons of noise about the cluster state when test fails.
func DumpEnvironmentDebugInfo(namespace, dir string) {
	pods := GetAllPods(namespace)
	for _, pod := range pods {
		printBanner()
		fmt.Println("Logs of " + pod)
		LogsOf(namespace, pod)
		printBanner()
		StateOf(namespace, pod)
		printBanner()
	}
	GetEvents(namespace)
	DumpTelepresenceLog(dir)
}

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
	<-testshell.Execute("kubectl delete namespace " + namespace + " --wait=false").Done()
}

func call(routeURL string, headers map[string]string) func() (string, error) {
	return func() (string, error) {
		fmt.Printf("[%s] Checking [%s] with headers [%s]...\n", time.Now().Format("2006-01-02 15:04:05.001"), routeURL, headers)
		return GetBodyWithHeaders(routeURL, headers)
	}
}
