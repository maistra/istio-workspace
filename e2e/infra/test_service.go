package infra

import (
	"os"

	"github.com/onsi/gomega"

	"github.com/maistra/istio-workspace/test/shell"
)

// ModifyServerCodeIn changes the code base of a simple python-based web server and puts it in the defined directory
func ModifyServerCodeIn(tmpDir string) {
	CreateFile(tmpDir+"/"+"server.py", ModifiedServerPy)
}

// OriginalServerCodeIn puts the original code base of a simple python-based web server in the defined directory
func OriginalServerCodeIn(tmpDir string) {
	CreateFile(tmpDir+"/"+"server.py", OrigServerPy)
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
	if ClientVersion() == 4 {
		<-shell.ExecuteInDir(".", "bash", "-c",
			"oc get ServiceMeshMemberRoll default -n istio-system -o json | jq '.spec.members[.spec.members | length] |= \""+
				namespace+"\"' | oc apply -f - -n istio-system").Done()
	}
	<-shell.ExecuteInDir(projectDir, "make", "deploy-test-"+scenario).Done()
}

func setDockerEnvForTestServiceBuild(namespace string) (registry string) {
	setTestNamespace(namespace)
	return setDockerRegistryExternal()
}

func setDockerEnvForTestServiceDeploy(namespace string) (registry string) {
	setTestNamespace(namespace)
	return setDockerRegistryInternal()
}

func setTestNamespace(namespace string) {
	err := os.Setenv("TEST_NAMESPACE", namespace)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	setDockerRepository(namespace)
}
