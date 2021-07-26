package e2e_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/go-cmd/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/schollz/progressbar/v3"

	. "github.com/maistra/istio-workspace/e2e"
	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/pkg/naming"
	"github.com/maistra/istio-workspace/test"
	testshell "github.com/maistra/istio-workspace/test/shell"
)

var _ = Describe("Smoke End To End Tests", func() {

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
			} else {
				CleanupNamespace(namespace, false)
			}
		})

		Context("k8s deployment", func() {

			Context("http protocol", func() {

				BeforeEach(func() {
					scenario = "scenario-1" //nolint:goconst //reason no need for constant (yet)
					registry = GetDockerRegistryInternal()
				})

				Context("basic deployment modifications", func() {

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

				Context("deployment create/delete operations", func() {

					It("should watch for changes in ratings service and serve it", func() {
						EnsureAllDeploymentPodsAreReady(namespace)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))
						deploymentCount := GetResourceCount("deployment", namespace)

						ChangeNamespace("default")

						// when we start ike to create
						ikeCreate1 := RunIke(tmpDir, "create",
							"--deployment", "ratings-v1",
							"-n", namespace,
							"--route", "header:x-test-suite=smoke",
							"--image", registry+"/"+GetDevRepositoryName()+"/istio-workspace-test-prepared-"+PreparedImageV1+":"+GetImageTag(),
							"--session", sessionName,
						)
						Eventually(ikeCreate1.Done(), 1*time.Minute).Should(BeClosed())
						testshell.WaitForSuccess(ikeCreate1)

						// ensure the new service is running
						EnsureCorrectNumberOfResources(deploymentCount+1, "deployment", namespace)
						EnsureAllDeploymentPodsAreReady(namespace)

						// check original response
						EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring(PreparedImageV1), Not(ContainSubstring("ratings-v1")))

						// but also check if prod is intact
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))

						// when we start ike to create with a updated v
						ikeCreate2 := RunIke(tmpDir, "create",
							"--deployment", "ratings-v1",
							"-n", namespace,
							"--route", "header:x-test-suite=smoke",
							"--image", registry+"/"+GetDevRepositoryName()+"/istio-workspace-test-prepared-"+PreparedImageV2+":"+GetImageTag(),
							"--session", sessionName,
						)
						Eventually(ikeCreate2.Done(), 1*time.Minute).Should(BeClosed())
						testshell.WaitForSuccess(ikeCreate2)

						// ensure the new service is running
						EnsureCorrectNumberOfResources(deploymentCount+1, "deployment", namespace)
						EnsureAllDeploymentPodsAreReady(namespace)

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
						testshell.WaitForSuccess(ikeDel)

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
						deploymentCount := GetResourceCount("deployment", namespace)

						ike := RunIke(testshell.GetProjectDir(), "develop",
							"--deployment", "ratings-v1",
							"--port", "9081",
							"--method", "inject-tcp",
							"--run", "go run ./test/cmd/test-service -serviceName=PublisherA",
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
						EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring("PublisherA"), ContainSubstring("grpc"))

						Stop(ike)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
					})
				})
			})
		})

		Context("openshift deploymentconfig", func() {

			BeforeEach(func() {
				if !RunsOnOpenshift {
					Skip("DeploymentConfig is Openshift-specific resource and it won't work against plain k8s. " +
						"Tests for regular k8s deployment can be found in the same test suite.")
				}
				scenario = "scenario-2"
			})

			It("should watch for changes in ratings service in specified namespace and serve it", func() {
				ChangeNamespace(namespace)
				EnsureAllDeploymentConfigPodsAreReady(namespace)
				EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
				deploymentCount := GetResourceCount("deploymentconfig", namespace)

				// given we have details code locally
				CreateFile(tmpDir+"/ratings.py", PublisherService)

				ike := RunIke(tmpDir, "develop",
					"--deployment", "dc/ratings-v1",
					"--port", "9080",
					"--method", "inject-tcp",
					"--watch",
					"--run", "python ratings.py 9080",
					"--route", "header:x-test-suite=smoke",
					"--session", sessionName,
				)
				defer func() {
					Stop(ike)
				}()
				go FailOnCmdError(ike, GinkgoT())

				EnsureCorrectNumberOfResources(deploymentCount+1, "deploymentconfig", namespace)
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
				registry = GetDockerRegistryInternal()
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

		Context("verify external integrations", func() {

			Context("Tekton", func() {

				BeforeEach(func() {
					scenario = "scenario-1"
				})

				It("should create, get, and delete", func() {
					defer test.TemporaryEnvVars("TEST_NAMESPACE", namespace, "TEST_SESSION_NAME", sessionName)()

					host := sessionName + "." + GetGatewayHost(namespace)

					testshell.WaitForSuccess(
						testshell.ExecuteInProjectRoot("make tekton-deploy"),
					)

					EnsureAllDeploymentPodsAreReady(namespace)
					EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))

					testshell.WaitForSuccess(
						testshell.ExecuteInProjectRoot("make tekton-test-ike-create"),
					)

					Eventually(TaskIsDone(namespace, "ike-create-run"), 5*time.Minute, 5*time.Second).Should(BeTrue())
					Expect(TaskResult(namespace, "ike-create-run", "url")).To(Equal(host))

					// verify session url
					testshell.WaitForSuccess(
						testshell.ExecuteInProjectRoot("make tekton-test-ike-session-url"),
					)

					Eventually(TaskIsDone(namespace, "ike-session-url-run"), 5*time.Minute, 5*time.Second).Should(BeTrue())
					Expect(TaskResult(namespace, "ike-session-url-run", "url")).To(Equal(host))

					// ensure the new service is running
					EnsureAllDeploymentPodsAreReady(namespace)

					// check original response
					EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring(PreparedImageV1), Not(ContainSubstring("ratings-v1")))

					// but also check if prod is intact
					EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))

					testshell.WaitForSuccess(
						testshell.ExecuteInProjectRoot("make tekton-test-ike-delete"),
					)
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

// EnsureCorrectNumberOfResources make sure the correct number of given resource are in namespace.
func EnsureCorrectNumberOfResources(count int, resource, namespace string) {
	Eventually(MatchResourceCount(count, GetResourceCountFunc(resource, namespace)), 5*time.Minute, 5*time.Second).Should(BeTrue())
}

// EnsureProdRouteIsReachable can be reached with no special arguments.
func EnsureProdRouteIsReachable(namespace string, matchers ...types.GomegaMatcher) {
	productPageURL := GetIstioIngressHostname() + "/productpage"

	Eventually(call(productPageURL, map[string]string{
		"Host": GetGatewayHost(namespace)}),
		10*time.Minute, 1*time.Second).Should(And(matchers...))
}

type stableCountMatcher struct {
	delegate              types.GomegaMatcher
	matchCount            int32
	subsequentOccurrences int32
	flipping              bool
}

func (s *stableCountMatcher) Match(actual interface{}) (success bool, err error) {
	match, err := s.delegate.Match(actual)
	if !match {
		if s.matchCount > 0 {
			s.flipping = true
		}
		s.matchCount = 0

		return false, err
	}

	s.matchCount++

	if s.matchCount < s.subsequentOccurrences {
		return false, errors.Errorf("not enough matches in sequence yet [%d/%d]", s.matchCount, s.subsequentOccurrences)
	}

	return match, err
}

func (s *stableCountMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf(
		"failed to receive stable response after %d times. Response is flipping:%v. latest cause: %s",
		s.subsequentOccurrences, s.flipping, s.delegate.FailureMessage(actual))
}

func (s *stableCountMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf(
		"failed to receive stable response after %d times. Response is flipping:%v. latest cause: %s",
		s.subsequentOccurrences, s.flipping, s.delegate.NegatedFailureMessage(actual))
}

func beStableInSeries(occurrences int32, matcher types.GomegaMatcher) types.GomegaMatcher {
	return &stableCountMatcher{subsequentOccurrences: occurrences, delegate: matcher}
}

// EnsureSessionRouteIsReachable the manipulated route is reachable.
func EnsureSessionRouteIsReachable(namespace, sessionName string, matchers ...types.GomegaMatcher) {
	productPageURL := GetIstioIngressHostname() + "/productpage"

	// check original response using headers
	Eventually(call(productPageURL, map[string]string{
		"Host":         GetGatewayHost(namespace),
		"x-test-suite": "smoke"}),
		10*time.Minute, 1*time.Second).Should(beStableInSeries(8, And(matchers...)))

	// check original response using host route
	Eventually(call(productPageURL, map[string]string{
		"Host": sessionName + "." + GetGatewayHost(namespace)}),
		10*time.Minute, 1*time.Second).Should(beStableInSeries(8, And(matchers...)))
}

// EnsureSessionRouteIsNotReachable the manipulated route is reachable.
func EnsureSessionRouteIsNotReachable(namespace, sessionName string, matchers ...types.GomegaMatcher) {
	productPageURL := GetIstioIngressHostname() + "/productpage"

	// check original response using headers
	Eventually(call(productPageURL, map[string]string{
		"Host":         GetGatewayHost(namespace),
		"x-test-suite": "smoke"}),
		10*time.Minute, 1*time.Second).Should(And(matchers...))
}

// ChangeNamespace switch to different namespace - so we also test -n parameter of $ ike.
// That only works for oc cli, as kubectl by default uses `default` namespace.
func ChangeNamespace(namespace string) {
	if RunsOnOpenshift {
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

func FailOnCmdError(command *cmd.Cmd, t test.TestReporter) {
	<-command.Done()
	if command.Status().Exit != 0 {
		t.Errorf("failed executing %s with code %d", command.Name, command.Status().Exit)
	}
}

// DumpEnvironmentDebugInfo prints tons of noise about the cluster state when test fails.
func DumpEnvironmentDebugInfo(namespace, dir string) {
	GetEvents(namespace)
	DumpTelepresenceLog(dir)
}

func generateNamespaceName() string {
	return "ike-tests-" + naming.RandName(16)
}

func CleanupNamespace(namespace string, wait bool) {
	if keepStr, found := os.LookupEnv("IKE_E2E_KEEP_NS"); found {
		keep, _ := strconv.ParseBool(keepStr)
		if keep {
			return
		}
	}
	CleanupTestScenario(namespace)
	<-testshell.Execute("kubectl delete namespace " + namespace + " --wait=" + strconv.FormatBool(wait)).Done()
}

func call(routeURL string, headers map[string]string) func() (string, error) {
	fmt.Printf("Checking [%s] with headers [%s]\n", routeURL, headers)
	bar := progressbar.Default(-1)

	return func() (string, error) {
		bar.Add(1)

		return GetBodyWithHeaders(routeURL, headers)
	}
}
