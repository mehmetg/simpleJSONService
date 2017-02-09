package main

import (
	"encoding/json"
	"errors"
	"time"
)

// ErrPathNotFound is returned when path we're looking for is not defined in
// cardData map
var ErrPathNotFound = errors.New("path not found")

// ErrOperationFailed is returned when path we tried to delete still exists
var ErrOperationFailed = errors.New("operation failed")

// ErrMalformedData is returned when data is not proper
var ErrMalformedData = errors.New("malformed data")

// DataTestService provides card test data
type DataTestService interface {
	Status() (string, error)
	GetData(string) (map[string]interface{}, error)
	PostData(string, map[string]interface{}) (map[string]interface{}, error)
	DeleteData(string) (map[string]interface{}, error)
}

type dataTestService struct{}

func (dataTestService) Status() (string, error) {
	return time.Now().Format(time.RFC850), nil
}

// GetData returns card data from the map
func (dataTestService) GetData(path string) (map[string]interface{}, error) {
	return getCardData(path)
}

// PostData sets card data from to the map
func (dataTestService) PostData(path string, data map[string]interface{}) (map[string]interface{}, error) {
	byteData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	cardData[path] = string(byteData)
	return getCardData(path)
}

// DeleteData deletes data from to the map
func (dataTestService) DeleteData(path string) (map[string]interface{}, error) {
	delete(cardData, path)
	_, err := getCardData(path)
	if err != nil {
		return nil, nil
	}
	return nil, ErrOperationFailed
}

func getCardData(path string) (map[string]interface{}, error) {
	data, mapOk := cardData[path]
	arrData, jsonOk := data.(string)
	if mapOk && jsonOk {
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(arrData), &jsonData); err != nil {
			return nil, err
		}
		return jsonData, nil
	}
	return nil, ErrPathNotFound
}
