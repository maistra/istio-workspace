package e2e_test

import (
	"fmt"
	"strings"
	"time"

	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/test"
	testshell "github.com/maistra/istio-workspace/test/shell"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Smoke End To End Tests - Faulty scenarios", func() {

	Context("exit codes", func() {

		It("should return non 0 on failed command", func() {
			completion := testshell.ExecuteInDir(".", "bash", "-c", "ike missing-command")
			<-completion.Done()
			Expect(completion.Status().Exit).Should(Equal(23))
		})

	})

	Context("using ike cli", func() {

		var (
			namespace,
			tmpDir string
		)

		JustBeforeEach(func() {
			namespace = generateNamespaceName()
			tmpDir = test.TmpDir(GinkgoT(), "namespace-"+namespace)

			<-testshell.Execute(NewProjectCmd(namespace)).Done()

			PrepareEnv(namespace)
			InstallLocalOperator(namespace)
			Eventually(AllPodsReady(namespace), 2*time.Minute, 5*time.Second).Should(BeTrue())
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

		Describe("session cleanup", func() {

			It("should remove session if non-existing deployment is specified", func() {

				// given
				ikeWithWatch := testshell.ExecuteInDir(tmpDir, "ike", "develop",
					"--deployment", "non-existing-deployment",
					"-n", namespace,
					"--port", "9080",
					"--method", "inject-tcp",
					"--watch",
					"--run", "ruby ratings.rb 9080",
					"--route", "header:x-test-suite=smoke",
				)
				Eventually(ikeWithWatch.Done(), 2*time.Minute).Should(BeClosed())
				Expect(ikeWithWatch.Status().Exit).ToNot(Equal(0))

				// when
				sessions := testshell.ExecuteInDir(tmpDir, "kubectl", "get", "sessions", "-n", namespace)
				Eventually(sessions.Done(), 1*time.Minute).Should(BeClosed())

				// then
				stdErr := strings.Join(sessions.Status().Stderr, " ")
				Expect(stdErr).To(ContainSubstring("No resources found"))
			})

		})
	})
})
