package e2e_test

import (
	_ "embed"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/maistra/istio-workspace/e2e/infra"
	. "github.com/maistra/istio-workspace/e2e/verify"
	"github.com/maistra/istio-workspace/test"
	"github.com/maistra/istio-workspace/test/scenarios"
	testshell "github.com/maistra/istio-workspace/test/shell"
)

//go:embed test-services/pass-through/main.go
var golangService string

//go:embed test-services/pod_name/main.go
var podNameService string

//go:embed test-services/bookinfo/publisher.py
var publisherService string

var _ = Describe("Fundamental use cases", func() {

	Context("Using ike with existing services", func() {

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

		When("Using Kubernetes cluster and Deployment resource", func() {

			Context("services communicating over HTTP", func() {

				BeforeEach(func() {
					scenario = scenarios.HTTPSeq
					registry = GetInternalContainerRegistry()
				})

				When("changing service locally", func() {

					It("should apply changes and expose modified service through special route", func() {
						EnsureAllDeploymentPodsAreReady(namespace)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("productpage-v1"))
						deploymentCount := GetResourceCount("deployment", namespace)

						By("connecting local product page service to cluster services")
						CreateFile(tmpDir+"/productpage.py", publisherService)

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

						By("modifying local service")
						modifiedDetails := strings.Replace(publisherService, "PublisherA", "Publisher Ike", 1)
						CreateFile(tmpDir+"/productpage.py", modifiedDetails)

						EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring("Publisher Ike"))

						By("disconnecting local service")
						Stop(ike)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("productpage-v1"))
					})
				})

				When("deploying new version of the service to the cluster", func() {

					It("should deploy new instance of the service and make it reachable through special route", func() {
						EnsureAllDeploymentPodsAreReady(namespace)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV1)))
						deploymentCount := GetResourceCount("deployment", namespace)

						ChangeNamespace("default")

						By("creating new version of the service")
						ikeCreate1 := RunIke(tmpDir, "create",
							"--deployment", "ratings-v1",
							"-n", namespace,
							"--route", "header:x-test-suite=smoke",
							"--image", registry+"/"+GetDevRepositoryName()+"/istio-workspace-test-prepared-"+PreparedImageV1+":"+GetImageTag(),
							"--session", sessionName,
						)
						Eventually(ikeCreate1.Done(), 1*time.Minute).Should(BeClosed())
						testshell.WaitForSuccess(ikeCreate1)

						By("ensuring it's running")
						EnsureCorrectNumberOfResources(deploymentCount+1, "deployment", namespace)
						EnsureAllDeploymentPodsAreReady(namespace)

						By("ensuring it responds with new payload")
						EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring(PreparedImageV1), Not(ContainSubstring("ratings-v1")))

						By("ensuring prod route is intact")
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))

						By("creating new version with the same route")
						ikeCreate2 := RunIke(tmpDir, "create",
							"--deployment", "ratings-v1",
							"-n", namespace,
							"--route", "header:x-test-suite=smoke",
							"--image", registry+"/"+GetDevRepositoryName()+"/istio-workspace-test-prepared-"+PreparedImageV2+":"+GetImageTag(),
							"--session", sessionName,
						)
						Eventually(ikeCreate2.Done(), 1*time.Minute).Should(BeClosed())
						testshell.WaitForSuccess(ikeCreate2)

						By("ensuring it was replaced correctly")
						EnsureCorrectNumberOfResources(deploymentCount+1, "deployment", namespace)
						EnsureAllDeploymentPodsAreReady(namespace)

						By("ensuring new version is available")
						EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring(PreparedImageV2), Not(ContainSubstring("ratings-v1")))

						By("ensuring prod route is intact")
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV2)))

						By("removing new version")
						ikeDel := RunIke(tmpDir, "delete",
							"--deployment", "ratings-v1",
							"-n", namespace,
							"--session", sessionName,
						)

						Eventually(ikeDel.Done(), 1*time.Minute).Should(BeClosed())
						testshell.WaitForSuccess(ikeDel)

						By("ensuring session route responds the same as prod")
						EnsureSessionRouteIsNotReachable(namespace, sessionName, ContainSubstring("ratings-v1"), Not(ContainSubstring(PreparedImageV2)))
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
					})

				})
			})

			Context("services communicating over gRPC", func() {
				BeforeEach(func() {
					scenario = scenarios.GRPCSeq
				})

				When("changing service locally", func() {

					It("should apply changes and expose modified service through special route", func() {
						EnsureAllDeploymentPodsAreReady(namespace)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
						deploymentCount := GetResourceCount("deployment", namespace)

						By("locally running modified service")
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

						By("ensuring traffic reaches local service")
						EnsureCorrectNumberOfResources(deploymentCount+1, "deployment", namespace)
						EnsureAllDeploymentPodsAreReady(namespace)
						EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring("PublisherA"), ContainSubstring("grpc"))

						Stop(ike)
						EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
					})
				})
			})
		})

		When("Using Openshift cluster and DeploymentConfig resource", func() {

			BeforeEach(func() {
				if !RunsOnOpenshift {
					Skip("DeploymentConfig is Openshift-specific resource and it won't work against plain k8s. " +
						"Tests for regular k8s deployment can be found in the same test suite.")
				}
				scenario = scenarios.HTTPSeqDC
			})

			When("changing service locally", func() {

				It("should apply changes and expose modified service through special route", func() {
					ChangeNamespace(namespace)
					EnsureAllDeploymentConfigPodsAreReady(namespace)
					EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
					deploymentCount := GetResourceCount("deploymentconfig", namespace)

					By("running the service locally first")
					CreateFile(tmpDir+"/ratings.py", publisherService)

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

					By("modifying service")
					modifiedDetails := strings.Replace(publisherService, "PublisherA", "Publisher Ike", 1)
					CreateFile(tmpDir+"/ratings.py", modifiedDetails)

					EnsureSessionRouteIsReachable(namespace, sessionName, ContainSubstring("Publisher Ike"))

					Stop(ike)
					EnsureProdRouteIsReachable(namespace, ContainSubstring("ratings-v1"))
				})
			})
		})

	})

	Context("Using ike for newly created service", func() {

		var (
			namespace,
			sessionName,
			tmpDir string
		)

		tmpFs := test.NewTmpFileSystem(GinkgoT())

		BeforeEach(func() {
			namespace = GenerateNamespaceName()
			sessionName = GenerateSessionName()

			tmpDir = tmpFs.Dir("namespace-" + namespace)

			<-testshell.Execute(NewProjectCmd(namespace)).Done()

			PrepareEnv(namespace)

			InstallLocalOperator(namespace)
			Eventually(AllDeploymentsAndPodsReady(namespace), 10*time.Minute, 5*time.Second).Should(BeTrue())

		})

		AfterEach(func() {
			if CurrentSpecReport().Failed() {
				DumpEnvironmentDebugInfo(namespace, tmpDir)
			} else {
				CleanupNamespace(namespace, false)
				tmpFs.Cleanup()
			}
		})

		When("creating new service from scratch", func() {

			var restoreEnvVars func()

			JustBeforeEach(func() {
				restoreEnvVars = test.TemporaryEnvVars("IKE_GATEWAY_HOST", GetGatewayHost(namespace))
			})

			JustAfterEach(func() {
				restoreEnvVars()
			})

			It("should act like running in the cluster", func() {

				By("verifying initial resources in the namespace")
				Expect(GetResourceCount("deployment", namespace)).
					To(Equal(1), "istio-workspace operator should already be deployed")
				Expect(GetResourceCount("virtualservice", namespace)).To(BeZero())
				Expect(GetResourceCount("destinationrule", namespace)).To(BeZero())
				Expect(GetResourceCount("gateway", namespace)).To(BeZero())

				newService := tmpFs.Dir(tmpDir+"/pod-name-service") + "/main.go"
				CreateFile(newService, podNameService)
				localPort := "8383"

				ike := RunIke(tmpDir, "develop", "new",
					"--namespace", namespace,
					"--port", "9080",
					"--run", "go run "+newService+" -port "+localPort,
					"--watch",
					"--route", "header:x-test-suite=smoke",
					"--session", sessionName,
					"--method", "inject-tcp",
				)
				defer func() {
					Stop(ike)
				}()
				go FailOnCmdError(ike, GinkgoT())

				EnsureAllDeploymentPodsAreReady(namespace)

				By("adding virtual service")
				Expect(GetResourceCount("virtualservice", namespace)).To(Equal(2))
				By("adding destination rule")
				Expect(GetResourceCount("destinationrule", namespace)).To(Equal(2))
				By("adding gateway")
				Expect(GetResourceCount("gateway", namespace)).To(Equal(1))

				By("verifying service has cluster access")
				Expect(callingLocalService(localPort)()).To(ContainSubstring("running in k8s"))
			})
		})

		When("connecting new service running locally to the existing services in the cluster", func() {

			JustBeforeEach(func() {
				DeployTestScenario(scenarios.HTTPSeq, namespace)
			})

			It("should be able to reach other services within the same namespace", func() {
				EnsureAllDeploymentPodsAreReady(namespace)

				deploymentCount := GetResourceCount("deployment", namespace)
				newService := tmpFs.Dir(tmpDir+"/new-service") + "/main.go"
				CreateFile(newService, golangService)
				localPort := "8282"

				By("connecting local service to cluster services")
				ike := RunIke(tmpDir, "develop", "new",
					"--name", "new-service",
					"--namespace", namespace,
					"--port", "9080",
					"--run", "go run "+newService+" -port "+localPort,
					"--watch",
					"--route", "header:x-test-suite=smoke",
					"--session", sessionName,
					"--method", "inject-tcp",
				)
				defer func() {
					Stop(ike)
				}()
				go FailOnCmdError(ike, GinkgoT())
				// FIX this shouldn't necessarily be a new route
				// it might be swap-deployment actually
				// with 2  extra deployments we have separation of dummy
				// service and newly spawned one running locally
				// Optimize later.
				EnsureCorrectNumberOfResources(deploymentCount+2, "deployment", namespace)
				EnsureAllDeploymentPodsAreReady(namespace)

				modifiedResponse := "proxy response"

				By("ensuring service is accessible locally and can reach already deployed reviews-v1 service")
				Expect(callingLocalService(localPort)()).To(And(ContainSubstring("reviews-v1"), Not(ContainSubstring(modifiedResponse))))

				By("modifying response")
				modifiedService := strings.Replace(golangService, "writer.Write(content)", `writer.Write([]byte("`+modifiedResponse+`: [" + string(content) + "]"))`, 1)
				CreateFile(newService, modifiedService)
				Eventually(callingLocalService(localPort)).Should(ContainSubstring("reviews-v1"), ContainSubstring(modifiedResponse))
			})

		})

	})
})

func callingLocalService(localPort string) func() string {
	curlLocalService := "curl -i http://localhost:" + localPort

	return func() string {
		curl := testshell.Execute(curlLocalService)
		<-curl.Start()

		return joinStdOut(curl)
	}
}

func joinStdOut(command *cmd.Cmd) string {
	return strings.Join(command.Status().Stdout, " ")
}
