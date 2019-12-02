package infra

import (
	"fmt"
	"os"
	"time"

	"github.com/onsi/gomega"

	"github.com/maistra/istio-workspace/test/shell"
)

// ModifyServerCodeIn changes the code base of a simple python-based web server and puts it in the defined directory
func ModifyServerCodeIn(tmpDir string) {
	CreateFile(tmpDir+"/"+"server.py", modifiedServerPy)
}

// OriginalServerCodeIn puts the original code base of a simple python-based web server in the defined directory
func OriginalServerCodeIn(tmpDir string) {
	CreateFile(tmpDir+"/"+"server.py", origServerPy)
}

// BuildTestService builds istio-workspace-test service and pushes it to specified registry
func BuildTestService(namespace string) (registry string) {
	projectDir := shell.GetProjectDir()
	registry = setDockerEnvForTestServiceBuild(namespace)

	LoginAsTestPowerUser()
	<-shell.ExecuteInDir(".", "bash", "-c", "docker login -u $(oc whoami) -p $(oc whoami -t) "+registry).Done()
	<-shell.ExecuteInDir(projectDir, "make", "docker-build-test", "docker-push-test").Done()
	return
}

// BuildTestServicePreparedImage builds istio-workspace-test-prepared service and pushes it to specified registry
func BuildTestServicePreparedImage(callerName, namespace string) (registry string) {
	projectDir := shell.GetProjectDir()
	registry = setDockerEnvForTestServiceBuild(namespace)

	os.Setenv("IKE_TEST_PREPARED_NAME", callerName)

	LoginAsTestPowerUser()
	<-shell.ExecuteInDir(".", "bash", "-c", "docker login -u $(oc whoami) -p $(oc whoami -t) "+registry).Done()
	<-shell.ExecuteInDir(projectDir, "make", "docker-build-test-prepared", "docker-push-test-prepared").Done()
	return
}

// DeployTestScenario deploys a test scenario into the specified namespace
func DeployTestScenario(scenario, namespace string) {
	projectDir := shell.GetProjectDir()
	setDockerEnvForTestServiceDeploy(namespace)

	LoginAsTestPowerUser()
	if ClientVersion() == 4 {
		<-shell.ExecuteInDir(".", "bash", "-c",
			`oc -n `+GetIstioNamespace()+` patch --type='json' smmr default -p '[{"op": "add", "path": "/spec/members", "value":["'"`+namespace+`"'"]}]'`).Done()
		gomega.Eventually(func() string {
			return GetProjectLabels(namespace)
		}, 1*time.Minute).Should(gomega.ContainSubstring("maistra.io/member-of"))
	}
	<-shell.ExecuteInDir(projectDir, "make", "deploy-test-"+scenario).Done()
}

func CleanupTestScenario(namespace string) {
	if ClientVersion() == 4 {
		LoginAsTestPowerUser()
		removeNsSubCmd := `oc get ServiceMeshMemberRoll default -n ` + GetIstioNamespace() + ` -o json | jq -c '.spec.members | map(select(. != "` + namespace + `"))'`
		patchCmd := `oc -n ` + GetIstioNamespace() + ` patch --type='json' smmr default -p "[{\"op\": \"replace\", \"path\": \"/spec/members\", \"value\": $(` + removeNsSubCmd + `) }]"`
		<-shell.ExecuteInDir(".", "bash", "-c", patchCmd).Done()
	}
}

// GetProjectLabels returns labels for a given namespace as a string
func GetProjectLabels(namespace string) string {
	cmd := shell.ExecuteInDir(".", "bash", "-c", "oc get project "+namespace+" -o jsonpath={.metadata.labels}")
	<-cmd.Done()
	return fmt.Sprintf("%s", cmd.Status().Stdout)
}

func setDockerEnvForTestServiceBuild(namespace string) (registry string) {
	setTestNamespace(namespace)
	return setDockerRegistryExternal()
}

func setDockerEnvForTestServiceDeploy(namespace string) (registry string) {
	setTestNamespace(namespace)
	err := os.Setenv("IKE_SCENARIO_GATEWAY", GetGatewayHost(namespace))
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	return setDockerRegistryInternal()
}

func setTestNamespace(namespace string) {
	err := os.Setenv("TEST_NAMESPACE", namespace)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	setDockerRepository(namespace)
}

// GetGatewayHost returns the host the Gateway in the scenario is bound to (http header Host)
func GetGatewayHost(namespace string) string {
	return namespace + "-test.com"
}

// origServerPy contains original server code used in `ike develop` tests
// Based on https://github.com/datawire/hello-world
const origServerPy = `
from http.server import HTTPStatus, BaseHTTPRequestHandler
from socketserver import TCPServer

PORT = 8000
MESSAGE = "Hello, world!\n".encode("ascii")

class Handler(BaseHTTPRequestHandler):
    """Respond to requests with hello."""

    def do_GET(self):
        """Handle GET"""
        self.send_response(HTTPStatus.OK)
        self.send_header("Content-type", "text/plain")
        self.send_header("Content-length", len(MESSAGE))
        self.end_headers()
        self.wfile.write(MESSAGE)


print("Serving at port", PORT)
TCPServer.allow_reuse_address = True
httpd = TCPServer(("", PORT), Handler)
httpd.serve_forever()

`

// modifiedServerPy contains modified server response used in `ike develop` tests
const modifiedServerPy = `
from http.server import HTTPStatus, BaseHTTPRequestHandler
from socketserver import TCPServer

PORT = 8000
MESSAGE = "Hello, telepresence! Ike Here!\n".encode("ascii")

class Handler(BaseHTTPRequestHandler):
    """Respond to requests with hello."""

    def do_GET(self):
        """Handle GET"""
        self.send_response(HTTPStatus.OK)
        self.send_header("Content-type", "text/plain")
        self.send_header("Content-length", len(MESSAGE))
        self.end_headers()
        self.wfile.write(MESSAGE)


print("Serving at port", PORT)
TCPServer.allow_reuse_address = True
httpd = TCPServer(("", PORT), Handler)
httpd.serve_forever()
`

// PublisherRuby contains fixed response to be changed by tests
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
