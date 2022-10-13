package infra

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/onsi/gomega"

	"github.com/maistra/istio-workspace/test/shell"
)

// BuildTestService builds istio-workspace-test service and pushes it to specified registry.
func BuildTestService() (registry string) {
	projectDir := shell.GetProjectDir()
	registry = SetExternalContainerRegistry()
	if RunsOnOpenshift {
		shell.WaitForSuccess(
			shell.ExecuteInDir(".", "bash", "-c", "docker login -u "+user+" -p $(oc whoami -t) "+registry),
		)
	}
	shell.WaitForSuccess(
		shell.ExecuteInDir(projectDir, "make", "container-image-test", "container-push-test"),
	)

	return
}

// BuildTestServicePreparedImage builds istio-workspace-test-prepared service and pushes it to specified registry.
func BuildTestServicePreparedImage(callerName string) (registry string) {
	projectDir := shell.GetProjectDir()
	registry = SetExternalContainerRegistry()

	os.Setenv("IKE_TEST_PREPARED_NAME", callerName)
	if RunsOnOpenshift {
		<-shell.ExecuteInDir(".", "bash", "-c", "docker login -u "+user+" -p $(oc whoami -t) "+registry).Done()
	}
	<-shell.ExecuteInDir(projectDir, "make", "container-image-test-prepared", "container-push-test-prepared").Done()

	return
}

// DeployTestScenario deploys a test scenario into the specified namespace.
func DeployTestScenario(scenario, namespace string) {
	projectDir := shell.GetProjectDir()
	SetInternalContainerRegistry()
	setContainerEnvForTestServiceDeploy(namespace)
	if RunsOnOpenshift {
		<-shell.ExecuteInDir(".", "bash", "-c",
			`oc -n `+GetIstioNamespace()+` patch --type='json' smmr default -p '[{"op": "add", "path": "/spec/members/-", "value":"`+namespace+`"}]'`).Done()
		gomega.Eventually(func() string {
			return GetProjectLabels(namespace)
		}, 1*time.Minute).Should(gomega.ContainSubstring("maistra.io/member-of"))
	} else {
		shell.WaitForSuccess(
			shell.ExecuteInDir(".", "bash", "-c", "kubectl label namespace "+namespace+" istio-injection=enabled --overwrite=true"),
		)
	}
	shell.WaitForSuccess(
		shell.ExecuteInDir(projectDir, "make", "deploy-test-"+scenario),
	)
}

func CleanupTestScenario(namespace string) {
	if RunsOnOpenshift {
		removeNsSubCmd := `oc get ServiceMeshMemberRoll default -n ` + GetIstioNamespace() + ` -o json | jq -c '.spec.members | map(select(. != "` + namespace + `"))'`
		patchCmd := `oc -n ` + GetIstioNamespace() + ` patch --type='json' smmr default -p "[{\"op\": \"replace\", \"path\": \"/spec/members\", \"value\": $(` + removeNsSubCmd + `) }]"`
		<-shell.ExecuteInDir(".", "bash", "-c", patchCmd).Done()
	}
}

// GetProjectLabels returns labels for a given namespace as a string.
func GetProjectLabels(namespace string) string {
	cmd := shell.ExecuteInDir(".", "bash", "-c", "kubectl get namespace "+namespace+" -o jsonpath={.metadata.labels}")
	<-cmd.Done()

	return fmt.Sprintf("%s", cmd.Status().Stdout)
}

func setContainerEnvForTestServiceDeploy(namespace string) {
	setTestNamespace(namespace)
	err := os.Setenv("IKE_GATEWAY_HOST", GetGatewayHost(namespace))
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
}

func setTestNamespace(namespace string) {
	err := os.Setenv("TEST_NAMESPACE", namespace)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
}

// GetGatewayHost returns the host the Gateway in the scenario is bound to (http header Host).
func GetGatewayHost(namespace string) string {
	return namespace + "-test.com"
}

const charset = "abcdefghijklmnopqrstuvwxyz"

// stringWithCharset returns a random string of length based on charset.
func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		ri, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[ri.Int64()]
	}

	return string(b)
}

// GenerateSessionName returns a random safe string to be used as a session name.
func GenerateSessionName() string {
	return stringWithCharset(8, charset)
}
