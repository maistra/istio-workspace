package infra

import (
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

	CreateFile(dir+"/cr.yaml", minimalIstioCR)
	<-cmd.ExecuteInDir(dir, "oc", "create", "-n", "istio-operator", "-f", dir+"/cr.yaml").Done()

	gomega.Eventually(PodCompletedStatus("istio-system", "job-name=openshift-ansible-istio-installer-job"), 5*time.Minute, 5*time.Second).Should(gomega.ContainSubstring("Completed"))

	<-cmd.ExecuteInDir(dir, "oc", "-n", namespace, "apply", "-f", gateway).Done()
	<-cmd.ExecuteInDir(dir, "oc", "-n", namespace, "apply", "-f", destinationRules).Done()
	<-cmd.ExecuteInDir(dir, "oc", "-n", namespace, "apply", "-f", virtualServices).Done()
}

func DeployBookinfoInto(namespace, dir string) {
	<-cmd.Execute("oc", "login", "-u", "developer").Done()
	bookinfo := DownloadInto(dir, "https://raw.githubusercontent.com/Maistra/bookinfo/master/bookinfo.yaml")
	<-cmd.ExecuteInDir(dir, "oc", "apply", "-n", namespace, "-f", bookinfo).Done()
}

func DeployOperator() {
	projectDir := os.Getenv("CUR_DIR")
	gomega.Expect(projectDir).To(gomega.Not(gomega.BeEmpty()))
	<-cmd.Execute("oc", "login", "-u", "system:admin").Done()

	err := os.Setenv("OPERATOR_NAMESPACE", "istio-workspace-operator")
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	<-cmd.Execute("oc", "new-project", "istio-workspace-operator").Done()

	<-cmd.ExecuteInDir(projectDir, "make", "deploy-operator").Done()
}

// minimalIstioCR is a minimal custom resource required to install an Istio Control Plane.
// This will deploy a control plane using the CentOS-based community Istio images.
const minimalIstioCR = `
apiVersion: "istio.openshift.com/v1alpha1"
kind: "Installation"
metadata:
  name: "istio-installation"
  namespace: istio-operator
`
