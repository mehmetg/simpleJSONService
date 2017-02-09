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
	GetData(string) (interface{}, error)
	PostData(string, interface{}) (interface{}, error)
	DeleteData(string) (interface{}, error)
	GetAllData() (interface{}, error)
}

type dataTestService struct{}

func (dataTestService) Status() (string, error) {
	return time.Now().Format(time.RFC850), nil
}

// GetData returns card data from the map
func (dataTestService) GetData(path string) (interface{}, error) {
	return getCardData(path)
}

// GetAllData returns all data in the map
func (dataTestService) GetAllData() (map[string]interface{}, error) {
	return cardData, nil
}

// PostData sets card data from to the map
func (dataTestService) PostData(path string, data interface{}) (interface{}, error) {
	byteData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	cardData[path] = string(byteData)
	return getCardData(path)
}

// DeleteData deletes data from to the map
func (dataTestService) DeleteData(path string) (interface{}, error) {
	delete(cardData, path)
	_, err := getCardData(path)
	if err != nil {
		return nil, nil
	}
	return nil, ErrOperationFailed
}

func getCardData(path string) (interface{}, error) {
	data, mapOk := cardData[path]
	if !mapOk {
		return nil, ErrPathNotFound
	}
	return data, nil
}
