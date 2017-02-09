package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
)

type statusResponse struct {
	Time string `json:"time"`
	Err  string `json:"err"`
}
type dataRequest struct {
	Path string                 `json:"path,omitempty"`
	Data map[string]interface{} `json:"data,omitempty"`
}
type dataResponse struct {
	Path string                 `json:"path"`
	Data map[string]interface{} `json:"data"`
	Err  string                 `json:"err"`
}
type setDataRequest struct {
	Path string `json:"path"`
	Data string `json:"data"`
}

func makeStatusEndpoint(svc dataTestService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		time, err := svc.Status()
		if err != nil {
			return statusResponse{time, err.Error()}, nil
		}
		return statusResponse{time, ""}, nil
	}
}

func makeGetCardDataEndpoint(svc dataTestService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(dataRequest)
		data, err := svc.GetData(req.Path)
		if err != nil {
			return dataResponse{req.Path, data, err.Error()}, err
		}
		return dataResponse{req.Path, data, ""}, nil
	}
}

func makePostCardDataEndpoint(svc dataTestService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(dataRequest)
		data, err := svc.PostData(req.Path, req.Data)
		if err != nil {
			return dataResponse{req.Path, data, err.Error()}, err
		}
		return dataResponse{req.Path, data, ""}, nil
	}
}

func makeDeleteCardDataEndpoint(svc dataTestService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(dataRequest)
		data, err := svc.DeleteData(req.Path)
		if err != nil {
			return dataResponse{req.Path, data, err.Error()}, err
		}
		return dataResponse{req.Path, data, ""}, nil
	}
}

// MakeHandler returns a handler for the card data test service.
func MakeHandler(ctx context.Context, svc dataTestService, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	statusHandler := kithttp.NewServer(
		ctx,
		makeStatusEndpoint(svc),
		decodeStatusRequest,
		encodeResponse,
		opts...,
	)

	getCardDataHandler := kithttp.NewServer(
		ctx,
		makeGetCardDataEndpoint(svc),
		decodeDataRequest,
		encodeResponse,
		opts...,
	)

	postCardDataHandler := kithttp.NewServer(
		ctx,
		makePostCardDataEndpoint(svc),
		decodeDataRequest,
		encodeResponse,
		opts...,
	)

	deleteCardDataHandler := kithttp.NewServer(
		ctx,
		makeDeleteCardDataEndpoint(svc),
		decodeDataRequest,
		encodeResponse,
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/data/{path}", getCardDataHandler).Methods("GET")
	r.Handle("/data/{path}", postCardDataHandler).Methods("POST")
	r.Handle("/data/{path}", deleteCardDataHandler).Methods("DELETE")
	r.Handle("/status", statusHandler).Methods("GET")
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		fmt.Println(t)
		return nil
	})

	return r
}

func decodeStatusRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return "", nil
}

func decodeDataRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var requestPayload dataRequest
	vars := mux.Vars(r)
	method := r.Method
	path, ok := vars["path"]
	if !ok {
		return nil, ErrMalformedData
	}
	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		if method != "GET" {
			return nil, ErrMalformedData
		}
	}
	requestPayload.Path = path

	if method == "POST" && requestPayload.Data == nil {
		return nil, ErrMalformedData
	}
	return requestPayload, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	jsonEncoder := json.NewEncoder(w)
	return jsonEncoder.Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	kithttpError := err.(kithttp.Error)

	switch kithttpError.Err {
	case ErrPathNotFound:
		w.WriteHeader(http.StatusNotFound)
		break
	case ErrMalformedData:
		w.WriteHeader(http.StatusUnprocessableEntity)
		break
	case ErrOperationFailed:
		w.WriteHeader(http.StatusExpectationFailed)
		break
	default:
		w.WriteHeader(http.StatusInternalServerError)
		break
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
