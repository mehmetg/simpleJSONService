package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"golang.org/x/net/context"
)

var cardData = map[string]interface{}{}

func main() {
	str := json.RawMessage(`{"test": "hello123"}`)
	j, _ := json.Marshal(&str)
	cardData["test3"] = string(j)
	var (
		httpAddr     = flag.String("http.addr", ":8000", "Address for HTTP (JSON) server")
		retryMax     = flag.Int("retry.max", 3, "per-request retries to different instances")
		retryTimeout = flag.Duration("retry.timeout", 500*time.Millisecond, "per-request timeout, including retries")
	)
	flag.Parse()

	fmt.Println("Hello, World!")
	fmt.Println(reflect.TypeOf(httpAddr))
	fmt.Println(*httpAddr)
	fmt.Println(reflect.TypeOf(retryMax))
	fmt.Println(*retryMax)
	fmt.Println(reflect.TypeOf(retryTimeout))
	fmt.Println(*retryTimeout)

	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = &serializedLogger{Logger: logger}
	logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)

	httpLogger := log.NewContext(logger).With("component", "http")

	ctx := context.Background()
	svc := dataTestService{}

	http.Handle("/", MakeHandler(ctx, svc, httpLogger))
	errs := http.ListenAndServe(*httpAddr, nil)

	logger.Log(errs)

}

type serializedLogger struct {
	mtx sync.Mutex
	log.Logger
}
