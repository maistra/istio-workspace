package infra

import (
	"os"
	"time"

	"github.com/aslakknutsen/istio-workspace/cmd/ike/cmd"

	"github.com/onsi/gomega"
)

func LoadIstio(dir string) {
	<-cmd.Execute("oc", "login", "-u", "system:admin").Done()

	CreateFile(dir+"/cr.yaml", minimalIstioCR)
	<-cmd.ExecuteInDir(dir, "oc", "create", "-n", "istio-operator", "-f", dir+"/cr.yaml").Done()

	gomega.Eventually(PodCompletedStatus("istio-system", "job-name=openshift-ansible-istio-installer-job"),
		10*time.Minute, 5*time.Second).Should(gomega.ContainSubstring("Completed"))
}

func DeployBookinfoInto(namespace, dir string) {
	<-cmd.Execute("oc", "login", "-u", "system:admin").Done()

	CreateFile(dir+"/session_role.yaml", sessionRole)
	CreateFile(dir+"/session_rolebinding.yaml", developerSessionRoleBinding)
	<-cmd.ExecuteInDir(dir, "oc", "-n", namespace, "apply", "-f", dir+"/session_role.yaml").Done()
	<-cmd.ExecuteInDir(dir, "oc", "-n", namespace, "apply", "-f", dir+"/session_rolebinding.yaml").Done()

	gateway := DownloadInto(dir, "https://raw.githubusercontent.com/Maistra/bookinfo/master/bookinfo-gateway.yaml")
	destinationRules := DownloadInto(dir, "https://raw.githubusercontent.com/istio/istio/release-1.0/samples/bookinfo/networking/destination-rule-all.yaml")
	virtualServices := DownloadInto(dir, "https://raw.githubusercontent.com/istio/istio/release-1.0/samples/bookinfo/networking/virtual-service-all-v1.yaml")

	<-cmd.ExecuteInDir(dir, "oc", "-n", namespace, "apply", "-f", gateway).Done()
	<-cmd.ExecuteInDir(dir, "oc", "-n", namespace, "apply", "-f", destinationRules).Done()
	<-cmd.ExecuteInDir(dir, "oc", "-n", namespace, "apply", "-f", virtualServices).Done()

	bookinfo := DownloadInto(dir, "https://raw.githubusercontent.com/Maistra/bookinfo/master/bookinfo.yaml")
	<-cmd.ExecuteInDir(dir, "oc", "apply", "-n", namespace, "-f", bookinfo).Done()
	<-cmd.ExecuteInDir(dir, "oc", "delete", "deployment", "reviews-v2", "-n", namespace).Done()
	<-cmd.ExecuteInDir(dir, "oc", "delete", "deployment", "reviews-v3", "-n", namespace).Done()
}

func BuildOperator() {
	projectDir := os.Getenv("CUR_DIR")

	operatorNS := "istio-workspace-operator"
	dockerRegistry := "docker-registry-default.127.0.0.1.nip.io:80"

	err := os.Setenv("OPERATOR_NAMESPACE", operatorNS)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	err = os.Setenv("DOCKER_REPOSITORY", operatorNS)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	err = os.Setenv("DOCKER_REGISTRY", dockerRegistry)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	<-cmd.Execute("oc", "login", "-u", "admin", "-p", "admin").Done()

	<-cmd.Execute("bash", "-c", "echo $(oc whoami)").Done()
	<-cmd.Execute("bash", "-c", "echo $(oc whoami -t)").Done()

	<-cmd.ExecuteInDir(projectDir, "make", "docker-build").Done()
	<-cmd.Execute("docker", "login", "-u $(oc whoami)", "-p $(oc whoami -t)", dockerRegistry).Done()
	<-cmd.ExecuteInDir(projectDir, "make", "docker-push").Done()
}

func DeployOperator() {
	projectDir := os.Getenv("CUR_DIR")
	gomega.Expect(projectDir).To(gomega.Not(gomega.BeEmpty()))
	<-cmd.Execute("oc", "login", "-u", "system:admin").Done()

	err := os.Setenv("OPERATOR_NAMESPACE", "istio-workspace-operator")
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	err = os.Setenv("DOCKER_IMAGE_TAG", "latest")
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

const sessionRole = `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: sessions
rules:
- apiGroups:
  - istio.openshift.com
  resources:
  - sessions
  verbs:
  - '*'
`

const developerSessionRoleBinding = `kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: session_developer
subjects:
- kind: User
  name: developer
roleRef:
  kind: Role
  name: sessions
  apiGroup: rbac.authorization.k8s.io
`
