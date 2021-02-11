package e2e_test

import (
	"fmt"
	"strings"
	"time"

	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/test"
	"github.com/maistra/istio-workspace/test/shell"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Operator End to End Tests", func() {

	Context("using install mode", func() {

		var (
			operatorNamepace string
			namespaces       []string

			cleanEnvVariable func()
		)
		BeforeEach(func() {
			namespaces = []string{}
			operatorNamepace = generateNamespaceName()
			cleanEnvVariable = test.TemporaryEnvVars("OPERATOR_NAMESPACE", operatorNamepace)
		})

		AfterEach(func() {
			cleanEnvVariable()
			for _, namespace := range namespaces {
				cleanupNamespace(namespace)
			}
			cleanupNamespace(operatorNamepace)
		})
		CreateNamespace := func() {
			for _, namespace := range namespaces {
				<-shell.Execute(NewProjectCmd(namespace)).Done()
			}
		}
		WatchListExpression := func() string {
			return strings.Join(namespaces, ",")
		}
		VerifyWatchList := func(validateNamespaces ...string) {
			for _, watchNs := range validateNamespaces {
				ikeCreate := RunIke(shell.GetProjectDir(), "create",
					"--deployment", watchNs+"-v1",
					"-n", watchNs,
					"--route", "header:x-test-suite=smoke",
					"--image", "x:x:x", // never used
					"--session", watchNs,
				)

				Eventually(func(contain string) func() bool {
					return func() bool {
						operatorLog := shell.ExecuteInDir(".",
							"kubectl", "logs",
							"deployment/istio-workspace-operator-controller-manager",
							"-n", operatorNamepace,
						)
						<-operatorLog.Done()

						log := strings.Join(operatorLog.Status().Stdout, "")
						fmt.Println(">>>>>>CONTAINS ", contain)
						fmt.Println(">>>>>>LOG  ", log)
						return strings.Contains(log, contain)
					}
				}(watchNs), 1*time.Minute, 5*time.Second).Should(BeTrue())

				ikeCreate.Stop()
			}

		}
		XIt("OwnNamespace", func() {
			if RunsAgainstOpenshift {
				Skip("Only run on microk8s cluster for complete isolation")
			}

			bundle := shell.ExecuteInDir(shell.GetProjectDir(), "make", "bundle-run")
			<-bundle.Done()
			Expect(bundle.Status().Exit).To(Equal((0)))

			Eventually(AllDeploymentsAndPodsReady(operatorNamepace), 10*time.Minute, 5*time.Second).Should(BeTrue())

			VerifyWatchList(operatorNamepace)
		})

		It("SingleNamespace", func() {
			if RunsAgainstOpenshift {
				Skip("Only run on microk8s cluster for complete isolation")
			}

			namespaces = append(namespaces, generateNamespaceName())
			CreateNamespace()

			defer test.TemporaryEnvVars("OPERATOR_WATCH_NAMESPACE", WatchListExpression())()

			bundle := shell.ExecuteInDir(shell.GetProjectDir(), "make", "bundle-run-single")
			<-bundle.Done()
			Expect(bundle.Status().Exit).To(Equal((0)))

			Eventually(AllDeploymentsAndPodsReady(operatorNamepace), 10*time.Minute, 5*time.Second).Should(BeTrue())

			VerifyWatchList(namespaces...)
		})

		It("MultiNamespace", func() {
			if RunsAgainstOpenshift {
				Skip("Only run on microk8s cluster for complete isolation")
			}

			namespaces = append(namespaces, generateNamespaceName(), generateNamespaceName())
			CreateNamespace()

			defer test.TemporaryEnvVars("OPERATOR_WATCH_NAMESPACE", WatchListExpression())()

			bundle := shell.ExecuteInDir(shell.GetProjectDir(), "make", "bundle-run-multi")
			<-bundle.Done()
			Expect(bundle.Status().Exit).To(Equal((0)))

			Eventually(AllDeploymentsAndPodsReady(operatorNamepace), 10*time.Minute, 5*time.Second).Should(BeTrue())

			VerifyWatchList(namespaces...)
		})
		It("AllNamespace", func() {
			if RunsAgainstOpenshift {
				Skip("Only run on microk8s cluster for complete isolation")
			}

			namespaces = append(namespaces, generateNamespaceName(), generateNamespaceName())
			CreateNamespace()

			bundle := shell.ExecuteInDir(shell.GetProjectDir(), "make", "bundle-run-all")
			<-bundle.Done()
			Expect(bundle.Status().Exit).To(Equal((0)))

			Eventually(AllDeploymentsAndPodsReady(operatorNamepace), 10*time.Minute, 5*time.Second).Should(BeTrue())

			VerifyWatchList(namespaces...)
		})

	})
})
