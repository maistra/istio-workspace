package e2e_test

import (
	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/pkg/naming"
	testshell "github.com/maistra/istio-workspace/test/shell"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"strings"
	"time"
)

var _ = Describe("Operator Installation Tests", func() {

	Context("local installation", func() {

		It("install ike operator into specified namespace", func() {
			// given
			projectName := "ike-local-installation-defined-namespace-" + naming.RandName(16)
			<-testshell.Execute(NewProjectCmd(projectName)).Done()
			defer func() {
				<-testshell.Execute(DeleteProjectCmd(projectName)).Done()
			}()
			PushOperatorImage(projectName)

			// when
			<-testshell.Execute("ike install-operator -l -n " + projectName).Done()

			// then
			operatorPodName := GetAllPods(projectName)[0]
			Eventually(AllPodsReady(projectName), 5*time.Minute, 5*time.Second).Should(BeTrue())
			Expect(operatorPodName).To(ContainSubstring("istio-workspace-"))
			ensureOperatorPodIsRunning(operatorPodName, projectName)
		})

		It("install ike operator into current namespace", func() {
			// given
			projectName := "ike-local-installation-current-namespace-" + naming.RandName(16)
			<-testshell.Execute(NewProjectCmd(projectName)).Done()
			defer func() {
				<-testshell.Execute(DeleteProjectCmd(projectName)).Done()
			}()
			PushOperatorImage(projectName)

			// when
			<-testshell.Execute("ike install-operator --local").Done()

			// then
			operatorPodName := GetAllPods(projectName)[0]
			Eventually(AllPodsReady(projectName), 5*time.Minute, 5*time.Second).Should(BeTrue())
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
