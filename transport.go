package main

import (
	"encoding/json"
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
	Path string      `json:"path,omitempty"`
	Data interface{} `json:"data,omitempty"`
}
type dataResponse struct {
	Path string      `json:"path"`
	Data interface{} `json:"data"`
	Err  string      `json:"err"`
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

func makeGetDataEndpoint(svc dataTestService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(dataRequest)
		data, err := svc.GetData(req.Path)
		if err != nil {
			return dataResponse{req.Path, data, err.Error()}, err
		}
		return dataResponse{req.Path, data, ""}, nil
	}
}

func makeAllDataEndpoint(svc dataTestService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		data, err := svc.GetAllData()
		if err != nil {
			return dataResponse{Data: data, Err: err.Error()}, err
		}
		return dataResponse{Data: data, Err: ""}, nil
	}
}

func makePostDataEndpoint(svc dataTestService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(dataRequest)
		data, err := svc.PostData(req.Path, req.Data)
		if err != nil {
			return dataResponse{req.Path, data, err.Error()}, err
		}
		return dataResponse{req.Path, data, ""}, nil
	}
}

func makeDeleteDataEndpoint(svc dataTestService) endpoint.Endpoint {
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
		decodeNoDataRequest,
		encodeResponse,
		opts...,
	)

	getDataHandler := kithttp.NewServer(
		ctx,
		makeGetDataEndpoint(svc),
		decodeDataRequest,
		encodeResponse,
		opts...,
	)

	getAllDataHandler := kithttp.NewServer(
		ctx,
		makeAllDataEndpoint(svc),
		decodeNoDataRequest,
		encodeResponse,
		opts...,
	)

	postDataHandler := kithttp.NewServer(
		ctx,
		makePostDataEndpoint(svc),
		decodeDataRequest,
		encodeResponse,
		opts...,
	)

	deleteDataHandler := kithttp.NewServer(
		ctx,
		makeDeleteDataEndpoint(svc),
		decodeDataRequest,
		encodeResponse,
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/all", getAllDataHandler).Methods("GET")
	r.Handle("/data/{path}", getDataHandler).Methods("GET")
	r.Handle("/data/{path}", postDataHandler).Methods("POST")
	r.Handle("/data/{path}", deleteDataHandler).Methods("DELETE")
	r.Handle("/status", statusHandler).Methods("GET")
	return r
}

func decodeNoDataRequest(_ context.Context, r *http.Request) (interface{}, error) {
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
