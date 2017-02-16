package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"

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
type pathResponse struct {
	Cards        []interface{} `json:"cards"`
	TotalCount   int           `json:"total_count"`
	PerPageCount int           `json:"per_page_count"`
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

type rawDataRequest struct {
	Offset int
	Limit  int
}

type pathRequest dataRequest

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
func makeGetPathEndpoint(svc dataTestService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(dataRequest)
		data, err := svc.GetPath(req.Path)
		if err != nil {
			return pathResponse{make([]interface{}, 0), 0, 10}, nil
		}
		mapData := data.(map[string]interface{})
		totalCount, tcOK := mapData["total_count"].(float64)
		perPageCount, ppcOK := mapData["per_page_count"].(float64)
		cardData, idsOK := mapData["cards"].([]interface{})
		if !ppcOK || !tcOK || !idsOK {
			fmt.Println("makeGetPathEndpoint error!!")
			fmt.Println(tcOK)
			fmt.Println(ppcOK)
			fmt.Println(idsOK)
			fmt.Println(mapData["cards"])
			fmt.Println(mapData["total_count"])
			fmt.Println(mapData["per_page_count"])
			return pathResponse{make([]interface{}, 0), 0, 10}, nil
		}
		return pathResponse{cardData, int(totalCount), int(perPageCount)}, nil
	}
}

func makeAllDataEndpoint(svc dataTestService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(rawDataRequest)
		data, err := svc.GetAllData(req.Offset, req.Limit)
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
		decodeRawDataRequest,
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

	getPathHandler := kithttp.NewServer(
		ctx,
		makeGetPathEndpoint(svc),
		decodePathRequest,
		encodeResponse,
		opts...,
	)

	getAllDataHandler := kithttp.NewServer(
		ctx,
		makeAllDataEndpoint(svc),
		decodeRawDataRequest,
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

	r.Handle("/all/{offset}/{limit}", getAllDataHandler).Methods("GET")
	r.Handle("/ccapi/v1/cardData", getPathHandler).Methods("GET")
	r.Handle("/data/{path}", getDataHandler).Methods("GET")
	r.Handle("/data/{path}", postDataHandler).Methods("POST")
	r.Handle("/data/{path}", deleteDataHandler).Methods("DELETE")
	r.Handle("/status", statusHandler).Methods("GET")
	return r
}

func decodeRawDataRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var requestPayload rawDataRequest
	vars := mux.Vars(r)
	offset, _ := strconv.Atoi(vars["offset"])
	limit, limitErr := strconv.Atoi(vars["limit"])
	if limitErr != nil {
		return nil, ErrMalformedData
	}
	requestPayload.Offset = offset
	requestPayload.Limit = limit
	return requestPayload, nil
}

func decodePathRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var requestPayload dataRequest
	query := r.URL.RawQuery
	requestPayload.Path = query
	return requestPayload, nil
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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	jsonEncoder := json.NewEncoder(w)
	return jsonEncoder.Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	//kithttpError := err.(kithttp.Error)

	switch err {
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
