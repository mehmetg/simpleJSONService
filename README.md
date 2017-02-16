# simpleJSONService
Simple service with /status (GET), /all (GET), /data (GET, POST, DELETE) endpoints to host JSON data for testing.
The service can be preloaded using a data.json.br file.

# Usage
* ```go get github.com/mehmetg/simpleJSONService```
* From project root: ```./build_all.sh```
* ```./sjs-<linux|darwin>-amd64 -http.addr <host:port> -datafile <brotli compressed jsonfile>```

