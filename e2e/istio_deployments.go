package e2e

import (
	"fmt"
	"os"
	"time"

	"github.com/aslakknutsen/istio-workspace/cmd/ike/cmd"

	"github.com/onsi/gomega"
)

func LoadIstioResources(namespace, dir string) {
	gateway := DownloadInto(dir, "https://raw.githubusercontent.com/Maistra/bookinfo/master/bookinfo-gateway.yaml")
	destinationRules := DownloadInto(dir, "https://raw.githubusercontent.com/istio/istio/release-1.0/samples/bookinfo/networking/destination-rule-all.yaml")
	virtualServices := DownloadInto(dir, "https://raw.githubusercontent.com/istio/istio/release-1.0/samples/bookinfo/networking/virtual-service-all-v1.yaml")

	<-cmd.Execute("oc", "login", "-u", "system:admin").Done()

	CreateFile(dir+"/cr.yaml", MinimalIstioCR)
	<-cmd.ExecuteInDir(dir, "oc", "create", "-n", "istio-operator", "-f", dir+"/cr.yaml").Done()

	gomega.Eventually(func() (string, error) {
		ocGetPods := cmd.Execute("oc", "get", "pods",
			"-n", "istio-system",
			"-l", "job-name=openshift-ansible-istio-installer-job",
			"-o", "go-template='{{range .items}}{{range .status.containerStatuses}}{{.state.terminated.reason}}{{end}}{{end}}'",
		)
		<-ocGetPods.Done()
		return fmt.Sprintf("%v", ocGetPods.Status().Stdout), nil
	}, 5*time.Minute, 5*time.Second).Should(gomega.ContainSubstring("Completed"))

	<-cmd.ExecuteInDir(dir, "oc", "-n", namespace, "apply", "-f", gateway).Done()
	<-cmd.ExecuteInDir(dir, "oc", "-n", namespace, "apply", "-f", destinationRules).Done()
	<-cmd.ExecuteInDir(dir, "oc", "-n", namespace, "apply", "-f", virtualServices).Done()
}

func DeployBookinfoInto(namespace, dir string) {
	bookinfo := DownloadInto(dir, "https://raw.githubusercontent.com/Maistra/bookinfo/master/bookinfo.yaml")
	<-cmd.ExecuteInDir(dir, "oc", "apply", "-n", namespace, "-f", bookinfo).Done()
}

func DeployOperator() {
	projectDir := os.Getenv("CUR_DIR")
	gomega.Expect(projectDir).To(gomega.Not(gomega.BeEmpty()))
	<-cmd.ExecuteInDir(projectDir, "make", "deploy-operator").Done()
}
