package infra

import (
	"os"

	"github.com/maistra/istio-workspace/test/shell"

	"github.com/maistra/istio-workspace/pkg/cmd/config"

	"github.com/onsi/gomega"
)

// LoadIstio calls make load-istio target and waits until operator sets up mesh
func LoadIstio() {
	projectDir := os.Getenv("PROJECT_DIR")
	<-shell.ExecuteInDir(projectDir, "make", "load-istio").Done()
}

// BuildTestService builds istio-workspace-test service and pushes it to specified registry
func BuildTestService(namespace string) (registry string) {
	projectDir := os.Getenv("PROJECT_DIR")
	registry = setDockerEnvForTestServiceBuild(namespace)

	LoginAsTestPowerUser()
	<-shell.ExecuteInDir(".", "bash", "-c", "docker login -u $(oc whoami) -p $(oc whoami -t) "+registry).Done()
	<-shell.ExecuteInDir(projectDir, "make", "docker-build-test", "docker-push-test").Done()
	return
}

// DeployTestScenario deploys a test scenario into the specified namespace
func DeployTestScenario(scenario, namespace string) {
	projectDir := os.Getenv("PROJECT_DIR")
	setDockerEnvForTestServiceDeploy(namespace)

	LoginAsTestPowerUser()
	//<-shell.ExecuteInDir(".", "bash", "-c", "oc get ServiceMeshMemberRoll default -n istio-system -o json | jq '.spec.members[.spec.members | length] |= \""+namespace+"\"' | oc apply -f - -n istio-system").Done()
	<-shell.ExecuteInDir(projectDir, "make", "deploy-test-"+scenario).Done()
}

// BuildOperator builds istio-workspace operator and pushes it to specified registry
func BuildOperator() (registry string) {
	projectDir := os.Getenv("PROJECT_DIR")
	_, registry = setDockerEnvForOperatorBuild()
	LoginAsTestPowerUser()
	<-shell.ExecuteInDir(".", "bash", "-c", "docker login -u $(oc whoami) -p $(oc whoami -t) "+registry).Done()
	<-shell.ExecuteInDir(projectDir, "make", "docker-build", "docker-push").Done()
	return
}

func CreateOperatorNamespace() (namespace string) {
	namespace, _ = setDockerEnvForOperatorDeploy()
	LoginAsTestPowerUser()
	<-shell.Execute("oc new-project " + namespace).Done()
	return
}

// DeployLocalOperator deploys istio-workspace operator into specified namespace
func DeployLocalOperator(namespace string) {
	projectDir := os.Getenv("PROJECT_DIR")
	gomega.Expect(projectDir).To(gomega.Not(gomega.BeEmpty()))
	LoginAsTestPowerUser()

	setDockerEnvForLocalOperatorBuild(namespace)
	os.Setenv("IKE_IMAGE_NAME", "istio-workspace")
	<-shell.ExecuteInDir(".", "bash", "-c", "docker tag $IKE_DOCKER_REGISTRY/istio-workspace-operator/$IKE_IMAGE_NAME:$IKE_IMAGE_TAG $IKE_DOCKER_REGISTRY/"+namespace+"/$IKE_IMAGE_NAME:$IKE_IMAGE_TAG").Done() //nolint[:lll]
	<-shell.ExecuteInDir(".", "bash", "-c", "docker push $IKE_DOCKER_REGISTRY/"+namespace+"/$IKE_IMAGE_NAME:$IKE_IMAGE_TAG").Done()

	setDockerEnvForLocalOperatorDeploy(namespace)
	<-shell.ExecuteInDir(".", "bash", "-c", "ike install-operator -l -n "+namespace).Done()
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

func setDockerEnvForLocalOperatorBuild(namespace string) string {
	setLocalOperatorNamespace(namespace)
	repo := setDockerRegistryExternal()
	return repo
}

func setDockerEnvForOperatorDeploy() (namespace, registry string) {
	ns := setOperatorNamespace()
	repo := setDockerRegistryInternal()
	return ns, repo
}

func setDockerEnvForLocalOperatorDeploy(namespace string) string {
	setLocalOperatorNamespace(namespace)
	repo := setDockerRegistryInternal()
	return repo
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

func setLocalOperatorNamespace(namespace string) {
	err := os.Setenv("OPERATOR_NAMESPACE", namespace)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	setDockerRepository(namespace)
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
