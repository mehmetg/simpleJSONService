package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/go-kit/kit/log"
	"gopkg.in/kothar/brotli-go.v0/dec"
)

var globalData map[string]interface{}
var debug bool

func main() {
	debug = (os.Getenv("SJS_DEBUG") != "")
	fmt.Println(debug)
	var (
		httpAddr     = flag.String("http.addr", ":8000", "Address for HTTP (JSON) server")
		dataFilename = flag.String("datafile", "data.json.br", "JSON file for the test data")
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
	fmt.Printf("Starting service on \"%s\" with datafile \"%s\"\n", *httpAddr, *dataFilename)
	errs := http.ListenAndServe(*httpAddr, nil)

	logger.Log(errs)

}

func readJSONData(logger log.Logger, filename string) {
	fi, fileErr := os.Open(filename)
	if fileErr != nil {
		logger.Log("filereaderror", fileErr.Error())
		globalData = make(map[string]interface{})
		return
	}
	defer fi.Close()
	br := dec.NewBrotliReader(fi)
	defer fi.Close()

	byteArrData, brotliReadErr := ioutil.ReadAll(br)
	if brotliReadErr != nil {
		logger.Log("filebrotlireaderror", fileErr.Error())
		globalData = make(map[string]interface{})
		return
	}

	var data map[string]interface{}
	jsonErr := json.Unmarshal(byteArrData, &data)
	if jsonErr != nil {
		logger.Log("fileformaterror", jsonErr.Error())
		os.Exit(1)
	}
	globalData = data
	if debug {
		for key := range globalData {
			fmt.Println("ccapi/v1/cardData?" + key)
		}
	}
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
