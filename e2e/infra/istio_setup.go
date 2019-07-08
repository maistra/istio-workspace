package infra

import (
	"os"
	"time"

	"github.com/maistra/istio-workspace/pkg/shell"

	"github.com/maistra/istio-workspace/pkg/cmd/config"

	"github.com/onsi/gomega"
)

// LoadIstio calls make load-istio target and waits until operator sets up mesh
func LoadIstio() {
	projectDir := os.Getenv("CUR_DIR")
	<-shell.Execute("oc login -u system:admin").Done()
	<-shell.ExecuteInDir(projectDir, "make", "load-istio").Done()
	gomega.Eventually(PodCompletedStatus("istio-system", "job-name=openshift-ansible-istio-installer-job"),
		10*time.Minute, 5*time.Second).Should(gomega.ContainSubstring("Completed"))
}

// BuildTestService builds istio-workspace-test service and pushes it to specified registry
func BuildTestService(namespace string) (registry string) {
	projectDir := os.Getenv("CUR_DIR")
	registry = setDockerEnvForTestServiceBuild(namespace)

	<-shell.Execute("oc login -u admin -p admin").Done()
	<-shell.ExecuteInDir(".", "bash", "-c", "docker login -u $(oc whoami) -p $(oc whoami -t) "+registry).Done()
	<-shell.ExecuteInDir(projectDir, "make", "docker-build-test", "docker-push-test").Done()
	return
}

// DeployTestScenario deploys a test scenario into the specified namespace
func DeployTestScenario(scenario, namespace string) {
	projectDir := os.Getenv("CUR_DIR")
	setDockerEnvForTestServiceDeploy(namespace)

	<-shell.Execute("oc login -u system:admin").Done()
	<-shell.ExecuteInDir(projectDir, "make", "deploy-test-"+scenario).Done()
}

// BuildOperator builds istio-workspace operator and pushes it to specified registry
func BuildOperator() (registry string) {
	projectDir := os.Getenv("CUR_DIR")
	_, registry = setDockerEnvForOperatorBuild()
	<-shell.Execute("oc login -u admin -p admin").Done()
	<-shell.ExecuteInDir(".", "bash", "-c", "docker login -u $(oc whoami) -p $(oc whoami -t) "+registry).Done()
	<-shell.ExecuteInDir(projectDir, "make", "docker-build", "docker-push").Done()
	return
}

func CreateOperatorNamespace() (namespace string) {
	namespace, _ = setDockerEnvForOperatorDeploy()
	<-shell.Execute("oc login -u admin -p admin").Done()
	<-shell.Execute("oc new-project " + namespace).Done()
	return
}

// DeployOperator deploys istio-workspace operator into specified namespace
func DeployOperator() (namespace string) {
	projectDir := os.Getenv("CUR_DIR")
	gomega.Expect(projectDir).To(gomega.Not(gomega.BeEmpty()))
	<-shell.Execute("oc login -u admin -p admin").Done()

	namespace, _ = setDockerEnvForOperatorDeploy()

	<-shell.ExecuteInDir(projectDir, "ike", "install-operator").Done()
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
	err := os.Setenv(config.EnvDockerRepository, namespace)
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
	err := os.Setenv(config.EnvDockerRegistry, registry)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
}
