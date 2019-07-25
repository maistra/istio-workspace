package infra

// OrigServerPy contains original server code used in `ike develop` tests
// Based on https://github.com/datawire/hello-world
const OrigServerPy = `
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

// ModifiedServerPy contains modified server response used in `ike develop` tests
const ModifiedServerPy = `
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
