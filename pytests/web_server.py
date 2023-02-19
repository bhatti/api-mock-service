from http.server import BaseHTTPRequestHandler, HTTPServer
import time

hostName = "localhost"
serverPort = 9090

class MyServer(BaseHTTPRequestHandler):
    def do_GET(self):
        reply = '''
{"account":"21212423423","regions":["us-east-2", "us-west-2"],"name":"sample-id5","id":"us-west2_test1", "taxes": [123, 14], "items": [1.1, 2.0], "boo": [true, false]}
        '''
        self.send_response(200)
        self.send_header("Content-type", "application/json")
        self.end_headers()
        self.wfile.write(bytes(reply, "utf-8"))

    def do_POST(self):
        self.send_response(200)
        for key in self.headers:
            self.send_header(key, self.headers[key])
        self.end_headers()
        length = int(self.headers['content-length'])
        data = self.rfile.read(length)
        self.wfile.write(data) #bytes(data, "utf-8"))

if __name__ == "__main__":
    webServer = HTTPServer((hostName, serverPort), MyServer)
    print("Server started http://%s:%s" % (hostName, serverPort))

    try:
        webServer.serve_forever()
    except KeyboardInterrupt:
        pass

    webServer.server_close()
    print("Server stopped.")
