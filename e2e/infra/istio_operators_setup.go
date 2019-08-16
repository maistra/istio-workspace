package infra

import (
	"os"

	"github.com/maistra/istio-workspace/test/shell"

	"github.com/onsi/gomega"
)

// LoadIstio calls make load-istio target and waits until operator sets up the mesh
func LoadIstio() {
	projectDir := os.Getenv("PROJECT_DIR")
	<-shell.ExecuteInDir(projectDir, "make", "load-istio").Done()
}

// BuildOperator builds istio-workspace operator and pushes it to specified registry
func BuildOperator() (registry string) {
	projectDir := os.Getenv("PROJECT_DIR")
	_, registry = setDockerEnvForOperatorBuild()
	LoginAsTestPowerUser()
	<-shell.ExecuteInDir(".", "bash", "-c", "docker login -u $(oc whoami) -p $(oc whoami -t) " + registry).Done()
	<-shell.ExecuteInDir(projectDir, "make", "docker-build", "docker-push").Done()
	return
}

func CreateOperatorNamespace() (namespace string) {
	namespace, _ = setDockerEnvForOperatorDeploy()
	LoginAsTestPowerUser()
	<-shell.Execute(NewProjectCmd(namespace)).Done()
	return
}

// DeployLocalOperator deploys istio-workspace operator into specified namespace
func DeployLocalOperator(namespace string) {
	projectDir := os.Getenv("PROJECT_DIR")
	gomega.Expect(projectDir).To(gomega.Not(gomega.BeEmpty()))
	LoginAsTestPowerUser()

	setDockerEnvForLocalOperatorBuild(namespace)
	os.Setenv("IKE_IMAGE_NAME", "istio-workspace")
	<-shell.Execute("docker tag $IKE_DOCKER_REGISTRY/istio-workspace-operator/$IKE_IMAGE_NAME:$IKE_IMAGE_TAG $IKE_DOCKER_REGISTRY/" + namespace + "/$IKE_IMAGE_NAME:$IKE_IMAGE_TAG").Done() //nolint[:lll]
	<-shell.ExecuteInDir(".", "bash", "-c", "docker push $IKE_DOCKER_REGISTRY/" + namespace + "/$IKE_IMAGE_NAME:$IKE_IMAGE_TAG").Done()

	setDockerEnvForLocalOperatorDeploy(namespace)
	<-shell.Execute("ike install-operator -l -n " + namespace).Done()
}

func setOperatorNamespace() (namespace string) {
	operatorNS := "istio-workspace-operator"

	err := os.Setenv("OPERATOR_NAMESPACE", operatorNS)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	setDockerRepository(operatorNS)
	return operatorNS
}

func setLocalOperatorNamespace(namespace string) {
	err := os.Setenv("OPERATOR_NAMESPACE", namespace)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	setDockerRepository(namespace)
}

func GetClusterHost() string {
	if host, found := os.LookupEnv("IKE_CLUSTER_HOST"); found {
		return host
	}
	return "127.0.0.1.nip.io"
}
