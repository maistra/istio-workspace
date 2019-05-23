package infra

import (
	"os"
	"time"

	"github.com/maistra/istio-workspace/cmd/ike/cmd"

	"github.com/onsi/gomega"
)

func LoadIstio() {
	projectDir := os.Getenv("CUR_DIR")
	<-cmd.Execute("oc login -u system:admin").Done()
	<-cmd.ExecuteInDir(projectDir, "make", "load-istio").Done()
	gomega.Eventually(PodCompletedStatus("istio-system", "job-name=openshift-ansible-istio-installer-job"),
		10*time.Minute, 5*time.Second).Should(gomega.ContainSubstring("Completed"))
}

func DeployBookinfoInto(namespace string) {
	projectDir := os.Getenv("CUR_DIR")
	<-cmd.Execute("oc login -u system:admin").Done()
	<-cmd.ExecuteInDir(projectDir, "make", "deploy-bookinfo", "EXAMPLE_NAMESPACE="+namespace).Done()
}

func BuildOperator() (registry string) {
	projectDir := os.Getenv("CUR_DIR")
	_, registry = setDockerEnv()
	<-cmd.Execute("oc login -u admin -p admin").Done()
	<-cmd.ExecuteInDir(".", "bash", "-c", "docker login -u $(oc whoami) -p $(oc whoami -t) "+registry).Done()
	<-cmd.ExecuteInDir(projectDir, "make", "docker-build", "docker-push").Done()
	return
}

func DeployOperator() (namespace string) {
	projectDir := os.Getenv("CUR_DIR")
	gomega.Expect(projectDir).To(gomega.Not(gomega.BeEmpty()))
	<-cmd.Execute("oc login -u system:admin").Done()

	namespace, _ = setDockerEnv()
	// override and use internal address on Deployment
	err := os.Setenv("DOCKER_REGISTRY", "172.30.1.1:5000")
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	<-cmd.ExecuteInDir(projectDir, "make", "deploy-operator").Done()
	return
}

func setDockerEnv() (operatorNS, dockerRegistry string) {
	operatorNS = "istio-workspace-operator"
	dockerRegistry = "docker-registry-default.127.0.0.1.nip.io:80"

	err := os.Setenv("OPERATOR_NAMESPACE", operatorNS)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	err = os.Setenv("DOCKER_REPOSITORY", operatorNS)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	err = os.Setenv("DOCKER_REGISTRY", dockerRegistry)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	return
}
