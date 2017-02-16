# simpleJSONService
Simple service with /status (GET), /all (GET), /data (GET, POST, DELETE) endpoints to host JSON data for testing.
The service can be preloaded using a data.json file.

# Usage
* ```go get github.com/mehmetg/simpleJSONService```
From project root:
* ```./build_all.sh```
* ```./simpleJsonService -http.addr <host:port> -datafile <brotli compressed jsonfile>```

