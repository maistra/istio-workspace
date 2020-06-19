package infra

import (
	"fmt"
	"os"
	"time"

	"github.com/onsi/gomega"

	"github.com/maistra/istio-workspace/test/shell"
)

// BuildTestService builds istio-workspace-test service and pushes it to specified registry.
func BuildTestService() (registry string) {
	projectDir := shell.GetProjectDir()
	setTestNamespace(ImageRepo)
	registry = setDockerRegistryExternal()

	<-shell.ExecuteInDir(".", "bash", "-c", "docker login -u "+user+" -p $(oc whoami -t) "+registry).Done()
	<-shell.ExecuteInDir(projectDir, "make", "docker-build-test", "docker-push-test").Done()
	return
}

// BuildTestServicePreparedImage builds istio-workspace-test-prepared service and pushes it to specified registry.
func BuildTestServicePreparedImage(callerName string) (registry string) {
	projectDir := shell.GetProjectDir()
	setTestNamespace(ImageRepo)
	registry = setDockerRegistryExternal()

	os.Setenv("IKE_TEST_PREPARED_NAME", callerName)

	<-shell.ExecuteInDir(".", "bash", "-c", "docker login -u "+user+" -p $(oc whoami -t) "+registry).Done()
	<-shell.ExecuteInDir(projectDir, "make", "docker-build-test-prepared", "docker-push-test-prepared").Done()
	return
}

// DeployTestScenario deploys a test scenario into the specified namespace.
func DeployTestScenario(scenario, namespace string) {
	projectDir := shell.GetProjectDir()
	setDockerRegistryInternal()
	setDockerEnvForTestServiceDeploy(namespace)

	<-shell.ExecuteInDir(".", "bash", "-c",
		`oc -n `+GetIstioNamespace()+` patch --type='json' smmr default -p '[{"op": "add", "path": "/spec/members/-", "value":"`+namespace+`"}]'`).Done()
	gomega.Eventually(func() string {
		return GetProjectLabels(namespace)
	}, 1*time.Minute).Should(gomega.ContainSubstring("maistra.io/member-of"))
	<-shell.ExecuteInDir(projectDir, "make", "deploy-test-"+scenario).Done()
}

func CleanupTestScenario(namespace string) {
	removeNsSubCmd := `oc get ServiceMeshMemberRoll default -n ` + GetIstioNamespace() + ` -o json | jq -c '.spec.members | map(select(. != "` + namespace + `"))'`
	patchCmd := `oc -n ` + GetIstioNamespace() + ` patch --type='json' smmr default -p "[{\"op\": \"replace\", \"path\": \"/spec/members\", \"value\": $(` + removeNsSubCmd + `) }]"`
	<-shell.ExecuteInDir(".", "bash", "-c", patchCmd).Done()
}

// GetProjectLabels returns labels for a given namespace as a string.
func GetProjectLabels(namespace string) string {
	cmd := shell.ExecuteInDir(".", "bash", "-c", "oc get project "+namespace+" -o jsonpath={.metadata.labels}")
	<-cmd.Done()
	return fmt.Sprintf("%s", cmd.Status().Stdout)
}

func setDockerEnvForTestServiceDeploy(namespace string) {
	setTestNamespace(namespace)
	err := os.Setenv("IKE_SCENARIO_GATEWAY", GetGatewayHost(namespace))
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

// PublisherRuby contains fixed response to be changed by tests.
const PublisherRuby = `
require 'webrick'
require 'json'
require 'net/http'

if ARGV.length < 1 then
    puts "usage: #{$PROGRAM_NAME} port"
    exit(-1)
end

port = Integer(ARGV[0])

server = WEBrick::HTTPServer.new :BindAddress => '0.0.0.0', :Port => port

trap 'INT' do server.shutdown end

server.mount_proc '/' do |req, res|
    res.status = 200
    res.body = {'caller' => 'PublisherA'}.to_json
    res['Content-Type'] = 'application/json'
end

server.start
`
