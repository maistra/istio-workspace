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

		var projectName string

		BeforeEach(func() {
			projectName = strings.ReplaceAll(CurrentGinkgoTestDescription().TestText, " ", "-") + "-" + naming.RandName(16)
			projectName = strings.ReplaceAll(projectName, "should", "ike")
			<-testshell.Execute(NewProjectCmd(projectName)).Done()
			PushOperatorImage(projectName)
		})

		AfterEach(func() {
			if CurrentGinkgoTestDescription().Failed {
				pods := GetAllPods(projectName)
				for _, pod := range pods {
					printBanner()
					fmt.Println("Logs of " + pod)
					fmt.Println(LogsOf(projectName, pod))
					printBanner()
					StateOf(projectName, pod)
					printBanner()
				}
				GetEvents(projectName)
			}
			<-testshell.Execute(DeleteProjectCmd(projectName)).Done()
		})

		It("should install into specified namespace", func() {
			// when
			<-testshell.Execute("ike install-operator -l -n " + projectName).Done()

			// then
			Eventually(AllPodsReady(projectName), 2*time.Minute, 5*time.Second).Should(BeTrue())
			operatorPodName := GetAllPods(projectName)[0]
			Expect(operatorPodName).To(ContainSubstring("istio-workspace-"))
			ensureOperatorPodIsRunning(operatorPodName, projectName)
		})

		It("should install into current namespace", func() {
			// when
			<-testshell.Execute("ike install-operator --local").Done()

			// then
			Eventually(AllPodsReady(projectName), 2*time.Minute, 5*time.Second).Should(BeTrue())
			operatorPodName := GetAllPods(projectName)[0]
			Expect(operatorPodName).To(ContainSubstring("istio-workspace-"))
			ensureOperatorPodIsRunning(operatorPodName, projectName)
		})

	})
})

func ensureOperatorPodIsRunning(operatorPodName, projectName string) {
	podDetails := testshell.Execute("oc get pod " + operatorPodName + " -o yaml")
	<-podDetails.Done()

	detailsYaml := strings.Join(podDetails.Status().Stdout, "\n")
	Expect(detailsYaml).To(ContainSubstring(`    command:
    - ike
    env:
    - name: WATCH_NAMESPACE
      value: ` + projectName))
}
