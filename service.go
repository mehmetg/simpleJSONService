package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrPathNotFound is returned when path we're looking for is not defined in
// cardData map
var ErrPathNotFound = errors.New("path not found")

// ErrOperationFailed is returned when path we tried to delete still exists
var ErrOperationFailed = errors.New("operation failed")

// ErrMalformedData is returned when data is not proper
var ErrMalformedData = errors.New("malformed data")

var rwMutex sync.RWMutex

// DataTestService provides card test data
type DataTestService interface {
	Status() (string, error)
	GetData(string) (interface{}, error)
	GetPath(string) (interface{}, error)
	PostData(string, interface{}) (interface{}, error)
	DeleteData(string) (interface{}, error)
	GetAllData(int, int) (interface{}, error)
}

type dataTestService struct{}

func (dataTestService) Status() (string, error) {
	return time.Now().Format(time.RFC850), nil
}

// GetData returns card data from the map
func (dataTestService) GetData(id string) (interface{}, error) {
	return getCardData(id)
}

// GetPath returns card data array from the map by path
func (dataTestService) GetPath(path string) (interface{}, error) {
	fmt.Println(path)
	rwMutex.Lock()
	defer rwMutex.Unlock()
	data, mapOk := globalData[path].(map[string]interface{})
	if !mapOk {
		fmt.Println("Path error")
		return nil, ErrPathNotFound
	}
	cardIDs, idsOK := data["cards"].([]interface{})
	if !idsOK {
		fmt.Println("ID error")
		return nil, ErrPathNotFound
	}
	allCards, allCardsOK := globalData["all_data"].(map[string]interface{})
	if !allCardsOK {
		fmt.Println(globalData["all_cards"])
		fmt.Println("All cards error")
		return nil, ErrPathNotFound
	}
	var cardData []interface{}
	for _, cardID := range cardIDs {
		strID := cardID.(string)
		cardData = append(cardData, allCards[strID])
	}
	processedData := make(map[string]interface{})
	processedData["per_page_count"] = data["per_page_count"]
	processedData["total_count"] = data["total_count"]
	processedData["cards"] = cardData
	return processedData, nil
}

// GetAllData returns all data with offset and limit
func (dataTestService) GetAllData(offset int, limit int) ([]interface{}, error) {
	rwMutex.Lock()
	defer rwMutex.Unlock()
	var data []interface{}
	rawData := globalData["all_data"].(map[string]interface{})
	count := 0
	for _, value := range rawData {
		if count >= offset {
			data = append(data, value)
		}
		if count >= offset+limit {
			break
		}
		count++
	}
	return data, nil
}

// PostData sets card data from to the map
func (dataTestService) PostData(id string, data interface{}) (interface{}, error) {
	byteData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	rwMutex.Lock()
	globalData["all_data"].(map[string]interface{})[id] = string(byteData)
	rwMutex.Unlock()
	return getCardData(id)
}

// DeleteData deletes data from to the map
func (dataTestService) DeleteData(id string) (interface{}, error) {
	rwMutex.Lock()
	delete(globalData["all_data"].(map[string]interface{}), id)
	rwMutex.Unlock()
	_, err := getCardData(id)
	if err != nil {
		return nil, nil
	}
	return nil, ErrOperationFailed
}

func getCardData(id string) (interface{}, error) {
	rwMutex.Lock()
	defer rwMutex.Unlock()
	data, mapOk := globalData["all_data"].(map[string]interface{})[id]
	if !mapOk {
		return nil, ErrPathNotFound
	}
	return data, nil
}
