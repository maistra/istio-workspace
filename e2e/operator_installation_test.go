package e2e_test

import (
	"fmt"

	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/pkg/naming"
	testshell "github.com/maistra/istio-workspace/test/shell"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"strings"
	"time"
)

var _ = Describe("Operator Installation Tests", func() {

	Context("local ike operator installation", func() {

		var namespace string

		BeforeEach(func() {
			namespace = strings.ReplaceAll(CurrentGinkgoTestDescription().TestText, " ", "-") + "-" + naming.RandName(16)
			namespace = strings.ReplaceAll(namespace, "should", "ike")
			<-testshell.Execute(NewProjectCmd(namespace)).Done()
			PrepareEnv(namespace)
			SetDockerRegistryInternal()
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
			}
			<-testshell.Execute(DeleteProjectCmd(namespace)).Done()
		})

		It("should install into specified namespace", func() {
			// when
			<-testshell.Execute("ike install -l -n " + namespace).Done()

			// then
			Eventually(AllDeploymentsAndPodsReady(namespace), 5*time.Minute, 5*time.Second).Should(BeTrue())
			operatorPodName := GetAllPods(namespace)[0]
			Expect(operatorPodName).To(ContainSubstring("istio-workspace-"))
			ensureOperatorPodIsRunning(operatorPodName, namespace)
		})

		It("should install into current namespace", func() {
			if !RunsAgainstOpenshift {
				Skip("This is OpenShift specific test which assumes current namespace/project is set and oc available." +
					"We also cover installation to specific namespace with -n flag (see test above).")
			}

			// given
			ChangeNamespace(namespace)

			// when
			<-testshell.Execute("ike install --local").Done()

			// then
			Eventually(AllDeploymentsAndPodsReady(namespace), 5*time.Minute, 5*time.Second).Should(BeTrue())
			operatorPodName := GetAllPods(namespace)[0]
			Expect(operatorPodName).To(ContainSubstring("istio-workspace-"))
			ensureOperatorPodIsRunning(operatorPodName, namespace)
		})

	})
})

func ensureOperatorPodIsRunning(operatorPodName, namespace string) {
	podDetails := testshell.Execute("kubectl get pod " + operatorPodName + " -n " + namespace + " -o yaml")
	<-podDetails.Done()

	detailsYaml := strings.Join(podDetails.Status().Stdout, "\n")
	Expect(detailsYaml).To(ContainSubstring(`    command:
    - ike
    env:
    - name: WATCH_NAMESPACE
      value: ` + namespace))
}
