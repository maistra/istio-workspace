package e2e_test

import (
	"strings"
	"time"

	"github.com/go-cmd/cmd"
	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/test"
	"github.com/maistra/istio-workspace/test/shell"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Operator installation", func() {

	/*

		Test scenario overview:

		Given N namespaces

		When Operator is installed in X install-mode targeting the namespaces

		Then
			For each target namespace
				Create a dummy Session object(the content is of no importance)
				Verify that the Operator reacted to the Session object creation by checking the log
				Delete the dummy Session

	*/
	Context("on bare k8s", func() {

		var (
			operatorNamespace string
			namespaces        []string

			cleanEnvVariable func()
		)
		BeforeEach(func() {
			if RunsOnOpenshift {
				Skip("Only run on microk8s cluster for complete isolation")
			}

			namespaces = []string{}
			operatorNamespace = generateNamespaceName()
			cleanEnvVariable = test.TemporaryEnvVars("OPERATOR_NAMESPACE", operatorNamespace)
		})

		AfterEach(func() {
			cleanEnvVariable()
			for _, namespace := range namespaces {
				CleanupNamespace(namespace, true)
			}
			CleanupNamespace(operatorNamespace, true)
		})

		CreateNamespace := func() {
			for _, namespace := range namespaces {
				<-shell.Execute(CreateNamespaceCmd(namespace)).Done()
			}
		}

		WatchListExpression := func() string {
			return strings.Join(namespaces, ",")
		}

		VerifyWatchList := func(validateNamespaces ...string) {
			var ikeCmds []*cmd.Cmd
			defer func() {
				for _, ikeCmd := range ikeCmds {
					ikeCmd.Stop()
				}
			}()
			// Start deployments first
			for _, watchNs := range validateNamespaces {
				ikeCreate := RunIke(shell.GetProjectDir(), "create",
					"--deployment", watchNs+"-v1",
					"-n", watchNs,
					"--route", "header:x-test-suite=smoke",
					"--image", "x:x:x", // never used
					"--session", watchNs,
				)
				ikeCmds = append(ikeCmds, ikeCreate)
			}
			// Then wait for completion
			for _, watchNs := range validateNamespaces {
				Eventually(func(contain string) func() bool {
					return func() bool {
						operatorLog := shell.ExecuteInDir(".",
							"kubectl", "logs",
							"deployment/istio-workspace-operator-controller-manager",
							"-n", operatorNamespace,
						)
						<-operatorLog.Done()

						log := strings.Join(operatorLog.Status().Stdout, "")

						return strings.Contains(log, contain)
					}
				}(watchNs), 2*time.Minute, 10*time.Second).Should(BeTrue())
				<-shell.ExecuteInDir(".", "kubectl", "delete", "session", watchNs, "-n", watchNs).Done()
			}

		}
		It("should install to its own namespace", func() {
			bundle := shell.ExecuteInDir(shell.GetProjectDir(), "make", "bundle-run")
			<-bundle.Done()
			Expect(bundle.Status().Exit).To(BeZero())

			Eventually(AllDeploymentsAndPodsReady(operatorNamespace), 10*time.Minute, 5*time.Second).Should(BeTrue())

			VerifyWatchList(operatorNamespace)
		})

		It("should install to the single namespace", func() {
			namespaces = append(namespaces, generateNamespaceName())
			CreateNamespace()

			defer test.TemporaryEnvVars("OPERATOR_WATCH_NAMESPACE", WatchListExpression())()

			bundle := shell.ExecuteInDir(shell.GetProjectDir(), "make", "bundle-run-single")
			<-bundle.Done()
			Expect(bundle.Status().Exit).To(BeZero())

			Eventually(AllDeploymentsAndPodsReady(operatorNamespace), 10*time.Minute, 5*time.Second).Should(BeTrue())

			VerifyWatchList(namespaces...)
		})

		It("should install to multiple namespaces", func() {
			namespaces = append(namespaces, generateNamespaceName(), generateNamespaceName())
			CreateNamespace()

			defer test.TemporaryEnvVars("OPERATOR_WATCH_NAMESPACE", WatchListExpression())()

			bundle := shell.ExecuteInDir(shell.GetProjectDir(), "make", "bundle-run-multi")
			<-bundle.Done()
			Expect(bundle.Status().Exit).To(BeZero())

			Eventually(AllDeploymentsAndPodsReady(operatorNamespace), 10*time.Minute, 5*time.Second).Should(BeTrue())

			VerifyWatchList(namespaces...)
		})

		It("AllNamespace", func() {
			namespaces = append(namespaces, generateNamespaceName(), generateNamespaceName())
			CreateNamespace()

			bundle := shell.ExecuteInDir(shell.GetProjectDir(), "make", "bundle-run-all")
			<-bundle.Done()
			Expect(bundle.Status().Exit).To(BeZero())

			Eventually(AllDeploymentsAndPodsReady(operatorNamespace), 10*time.Minute, 5*time.Second).Should(BeTrue())

			VerifyWatchList(namespaces...)
		})

	})
})
