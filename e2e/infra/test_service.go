package infra

import (
	"os"

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
			"oc get ServiceMeshMemberRoll default -n "+GetIstioNamespace()+" -o json | jq '.spec.members[.spec.members | length] |= \""+
				namespace+"\"' | oc apply -f - -n "+GetIstioNamespace()).Done()
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
