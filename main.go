package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"context"

	"github.com/go-kit/kit/log"
)

var globalData map[string]interface{}

func main() {
	var (
		httpAddr     = flag.String("http.addr", ":8000", "Address for HTTP (JSON) server")
		dataFilename = flag.String("datafile", "data.json", "JSON file for the test data")
	)
	flag.Parse()

	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = &serializedLogger{Logger: logger}
	logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)

	httpLogger := log.NewContext(logger).With("component", "http")
	fileIOLogger := log.NewContext(logger).With("component", "fileio")

	readJSONData(fileIOLogger, *dataFilename)

	ctx := context.Background()
	svc := dataTestService{}

	http.Handle("/", MakeHandler(ctx, svc, httpLogger))
	errs := http.ListenAndServe(*httpAddr, nil)

	logger.Log(errs)

}

func readJSONData(logger log.Logger, filename string) {
	file, fileErr := ioutil.ReadFile(filename)
	if fileErr != nil {
		logger.Log("filereaderror", fileErr.Error())
		globalData = make(map[string]interface{})
		return
	}
	var data map[string]interface{}
	jsonErr := json.Unmarshal(file, &data)
	if jsonErr != nil {
		logger.Log("fileformaterror", jsonErr.Error())
		os.Exit(1)
	}
	globalData = data

}

type serializedLogger struct {
	mtx sync.Mutex
	log.Logger
}

func (l *serializedLogger) Log(keyvals ...interface{}) error {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	return l.Logger.Log(keyvals...)
}
