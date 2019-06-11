package infra

import (
	"os"
	"time"

	"github.com/maistra/istio-workspace/cmd/ike/cmd"

	"github.com/onsi/gomega"
)

// LoadIstio calls make load-istio target and waits until operator sets up mesh
func LoadIstio() {
	projectDir := os.Getenv("CUR_DIR")
	<-cmd.Execute("oc login -u system:admin").Done()
	<-cmd.ExecuteInDir(projectDir, "make", "load-istio").Done()
	gomega.Eventually(PodCompletedStatus("istio-system", "job-name=openshift-ansible-istio-installer-job"),
		10*time.Minute, 5*time.Second).Should(gomega.ContainSubstring("Completed"))
}

// BuildTestService builds istio-workspace-test service and pushes it to specified registry
func BuildTestService(namespace string) (registry string) {
	projectDir := os.Getenv("CUR_DIR")
	registry = setDockerEnvForTestServiceBuild(namespace)

	<-cmd.Execute("oc login -u admin -p admin").Done()
	<-cmd.ExecuteInDir(".", "bash", "-c", "docker login -u $(oc whoami) -p $(oc whoami -t) "+registry).Done()
	<-cmd.ExecuteInDir(projectDir, "make", "docker-build-test", "docker-push-test").Done()
	return
}

// DeployTestScenario deploys a test scenario into the specified namespace
func DeployTestScenario(scenario, namespace string) {
	projectDir := os.Getenv("CUR_DIR")
	setDockerEnvForTestServiceDeploy(namespace)

	<-cmd.Execute("oc login -u system:admin").Done()
	<-cmd.ExecuteInDir(projectDir, "make", "deploy-test-"+scenario).Done()
}

// BuildOperator builds istio-workspace operator and pushes it to specified registry
func BuildOperator() (registry string) {
	projectDir := os.Getenv("CUR_DIR")
	_, registry = setDockerEnvForOperatorBuild()
	<-cmd.Execute("oc login -u admin -p admin").Done()
	<-cmd.ExecuteInDir(".", "bash", "-c", "docker login -u $(oc whoami) -p $(oc whoami -t) "+registry).Done()
	<-cmd.ExecuteInDir(projectDir, "make", "docker-build", "docker-push").Done()
	return
}

// DeployOperator deploys istio-workspace operator into specified namespace
func DeployOperator() (namespace string) {
	projectDir := os.Getenv("CUR_DIR")
	gomega.Expect(projectDir).To(gomega.Not(gomega.BeEmpty()))
	<-cmd.Execute("oc login -u system:admin").Done()

	namespace, _ = setDockerEnvForOperatorDeploy()

	<-cmd.ExecuteInDir(projectDir, "ike", "install").Done()
	return
}

func setDockerEnvForTestServiceBuild(namespace string) (registry string) {
	setTestNamespace(namespace)
	return setDockerRegistryExternal()
}

func setDockerEnvForTestServiceDeploy(namespace string) (registry string) {
	setTestNamespace(namespace)
	return setDockerRegistryInternal()
}

func setDockerEnvForOperatorBuild() (namespace, registry string) {
	ns := setOperatorNamespace()
	repo := setDockerRegistryExternal()
	return ns, repo
}

func setDockerEnvForOperatorDeploy() (namespace, registry string) {
	ns := setOperatorNamespace()
	repo := setDockerRegistryInternal()
	return ns, repo
}

func setTestNamespace(namespace string) {
	err := os.Setenv("TEST_NAMESPACE", namespace)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	setDockerRepository(namespace)
}

func setOperatorNamespace() (namespace string) {
	operatorNS := "istio-workspace-operator"

	err := os.Setenv("OPERATOR_NAMESPACE", operatorNS)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	setDockerRepository(operatorNS)
	return operatorNS
}

func setDockerRepository(namespace string) {
	err := os.Setenv("IKE_DOCKER_REPOSITORY", namespace)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
}

func setDockerRegistryInternal() (registry string) {
	registry = "172.30.1.1:5000"
	setDockerRegistry(registry)
	return registry
}

func setDockerRegistryExternal() (registry string) {
	registry = "docker-registry-default.127.0.0.1.nip.io:80"
	setDockerRegistry(registry)
	return registry
}

func setDockerRegistry(registry string) {
	err := os.Setenv("IKE_DOCKER_REGISTRY", registry)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
}
