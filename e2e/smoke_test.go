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
			tmpDir string
			scenario string
		)

		JustBeforeEach(func() {
			namespace = generateNamespaceName()
			tmpDir = test.TmpDir(GinkgoT(), "namespace-"+namespace)

			<-testshell.Execute(NewProjectCmd(namespace)).Done()

			UpdateSecurityConstraintsFor(namespace)
			EnablePullingImages(namespace)
			InstallLocalOperator(namespace)
			DeployTestScenario(scenario, namespace)
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
					scenario = "scenario-1"
				})

				Context("basic deployment modifications", func() {

					It("should watch for changes in ratings service and serve it", func() {
						EnsureAllPodsAreReady(namespace)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))

						// given we have details code locally
						CreateFile(tmpDir+"/ratings.rb", PublisherRuby)

						ike := RunIke(tmpDir, "develop",
							"--deployment", "ratings-v1",
							"--port", "9080",
							"--method", "inject-tcp",
							"--watch",
							"--run", "ruby ratings.rb 9080",
							"--route", "header:x-test-suite=smoke",
						)
						EnsureAllPodsAreReady(namespace)
						EnsureSessionRouteIsReachable(namespace, ContainSubstring("PublisherA"))

						// then modify the service
						modifiedDetails := strings.Replace(PublisherRuby, "PublisherA", "Publisher Ike", 1)
						CreateFile(tmpDir+"/ratings.rb", modifiedDetails)

						EnsureSessionRouteIsReachable(namespace, ContainSubstring("Publisher Ike"))

						Stop(ike)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
					})

					It("should watch for changes in ratings service in specified namespace and serve it", func() {
						EnsureAllPodsAreReady(namespace)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))

						// given oc context is a different namespace
						ChangeNamespace("default")

						// given we have details code locally
						CreateFile(tmpDir+"/ratings.rb", PublisherRuby)

						ike := RunIke(tmpDir, "develop",
							"--namespace", namespace,
							"--deployment", "ratings-v1",
							"--port", "9080",
							"--method", "inject-tcp",
							"--watch",
							"--run", "ruby ratings.rb 9080",
							"--route", "header:x-test-suite=smoke",
						)
						EnsureAllPodsAreReady(namespace)
						EnsureSessionRouteIsReachable(namespace, ContainSubstring("PublisherA"))

						// then modify the service
						modifiedDetails := strings.Replace(PublisherRuby, "PublisherA", "Publisher Ike", 1)
						CreateFile(tmpDir+"/ratings.rb", modifiedDetails)

						EnsureSessionRouteIsReachable(namespace, ContainSubstring("Publisher Ike"))

						Stop(ike)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
					})
				})

				Context("deployment create/delete operations", func() {
					registry := GetDockerRegistryInternal()

					It("should watch for changes in ratings service and serve it", func() {
						EnsureAllPodsAreReady(namespace)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))

						ChangeNamespace("default")

						// when we start ike to create
						ike1 := RunIke(tmpDir, "create",
							"--deployment", "ratings-v1",
							"-n", namespace,
							"--route", "header:x-test-suite=smoke",
							"--image", registry+"/"+ImageRepo+"/istio-workspace-test-prepared-"+PreparedImageV1+":latest",
							"-s", "test-session",
						)
						Eventually(ike1.Done(), 1*time.Minute).Should(BeClosed())

						// ensure the new service is running
						EnsureAllPodsAreReady(namespace)

						// check original response
						EnsureSessionRouteIsReachable(namespace, ContainSubstring(PreparedImageV1), Not(ContainSubstring("ratings-v1")))

						// but also check if prod is intact
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
						//ShouldNot(ContainSubstring(PreparedImageV1))

						// when we start ike to create with a updated v
						ike2 := RunIke(tmpDir, "create",
							"--deployment", "ratings-v1",
							"-n", namespace,
							"--route", "header:x-test-suite=smoke",
							"--image", registry+"/"+ImageRepo+"/istio-workspace-test-prepared-"+PreparedImageV2+":latest",
							"-s", "test-session",
						)
						Eventually(ike2.Done(), 1*time.Minute).Should(BeClosed())

						// check original response
						EnsureSessionRouteIsReachable(namespace, ContainSubstring(PreparedImageV2), Not(ContainSubstring("ratings-v1")))

						// but also check if prod is intact
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV2)))

						// when we start ike to delete
						ikeDel := RunIke(tmpDir, "delete",
							"--deployment", "ratings-v1",
							"-n", namespace,
							"-s", "test-session",
						)
						Eventually(ikeDel.Done(), 1*time.Minute).Should(BeClosed())

						// check original response
						EnsureSessionRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV2)))

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
						EnsureAllPodsAreReady(namespace)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))

						ike := RunIke(testshell.GetProjectDir(), "develop",
							"--deployment", "ratings-v1",
							"--port", "9081",
							"--method", "inject-tcp",
							"--run", "go run ./test/cmd/test-service -serviceName=PublisherA",
							"--route", "header:x-test-suite=smoke",
						)
						EnsureAllPodsAreReady(namespace)

						EnsureSessionRouteIsReachable(namespace, ContainSubstring("PublisherA"), ContainSubstring("grpc"))

						Stop(ike)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
					})
				})
			})
		})

		Context("openshift deploymentconfig", func() {

			BeforeEach(func() {
				scenario = "scenario-2"
			})

			It("should watch for changes in ratings service in specified namespace and serve it", func() {
				EnsureAllPodsAreReady(namespace)
				EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))

				// given we have details code locally
				CreateFile(tmpDir+"/ratings.rb", PublisherRuby)

				ike := RunIke(tmpDir, "develop",
					"--deployment", "ratings-v1",
					"--port", "9080",
					"--method", "inject-tcp",
					"--watch",
					"--run", "ruby ratings.rb 9080",
					"--route", "header:x-test-suite=smoke",
				)
				EnsureAllPodsAreReady(namespace)
				EnsureSessionRouteIsReachable(namespace, ContainSubstring("PublisherA"))

				// then modify the service
				modifiedDetails := strings.Replace(PublisherRuby, "PublisherA", "Publisher Ike", 1)
				CreateFile(tmpDir+"/ratings.rb", modifiedDetails)

				EnsureSessionRouteIsReachable(namespace, ContainSubstring("Publisher Ike"))

				Stop(ike)
				EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
			})
		})
	})
})

// EnsureAllPodsAreReady make sure all Pods are in Ready state in given namespace.
func EnsureAllPodsAreReady(namespace string) {
	Eventually(AllPodsReady(namespace), 5*time.Minute, 5*time.Second).Should(BeTrue())
}

// EnsureProdRouteIsReachable can be reached with no special arguments.
func EnsureProdRouteIsReachable(namespace string, matchers ...types.GomegaMatcher) {
	productPageURL := GetIstioIngressHostname() + "/test-service/productpage"

	Eventually(call(productPageURL, map[string]string{
		"Host": GetGatewayHost(namespace)}),
		3*time.Minute, 1*time.Second).Should(And(matchers...))
}

// EnsureSessionRouteIsReachable the manipulated route is reachable.
func EnsureSessionRouteIsReachable(namespace string, matchers ...types.GomegaMatcher) {
	productPageURL := GetIstioIngressHostname() + "/test-service/productpage"

	// check original response
	Eventually(call(productPageURL, map[string]string{
		"Host":         GetGatewayHost(namespace),
		"x-test-suite": "smoke"}),
		3*time.Minute, 1*time.Second).Should(And(matchers...))
}

// ChangeNamespace switch to different namespace - so we also test -n parameter of $ ike.
func ChangeNamespace(namespace string) {
	<-testshell.Execute("oc project default").Done()
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

func DumpEnvironmentDebugInfo(namespace, dir string) {
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
	<-testshell.Execute("oc delete project " + namespace).Done()
}

func call(routeURL string, headers map[string]string) func() (string, error) {
	return func() (string, error) {
		fmt.Printf("[%s] Checking [%s] with headers [%s]...\n", time.Now().Format("2006-01-02 15:04:05.001"), routeURL, headers)
		return GetBodyWithHeaders(routeURL, headers)
	}
}
