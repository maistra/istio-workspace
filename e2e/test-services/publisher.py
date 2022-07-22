import sys
from http.server import HTTPStatus, BaseHTTPRequestHandler
from socketserver import TCPServer


if len(sys.argv) < 2:
  print("usage: #{$PROGRAM_NAME} port")
  exit(-1)

PORT = int(sys.argv[1])

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(HTTPStatus.OK)
        self.send_header("Content-type", "text/plain")

        self.end_headers()
        self.wfile.write("{\"caller\": \"PublisherA\"}".encode("ascii"))

TCPServer.allow_reuse_address = True
httpd = TCPServer(("", PORT), Handler)
httpd.serve_forever()
