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

// DetailsRuby contains details service code base
// from https://raw.githubusercontent.com/istio/istio/master/samples/bookinfo/src/details/details.rb
const DetailsRuby = `
#!/usr/bin/ruby
#
# Copyright 2017 Istio Authors
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

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
