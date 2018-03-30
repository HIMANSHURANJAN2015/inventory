package utility

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	//"../model" Dont fetch struct from model, or it will cause cyclic import. controller->model->utility->model
)

type StoreResponse struct {
	Code   int
	Status string
	Data   struct {
		Store Store
	}
}

type Store struct {
	Id            int
	Name          string
	ClientStoreId int
	Latitude      string
	Longitude     string
	Address       string
}

type StoreListResponse struct {
	Code   int
	Status string
	Data   struct {
		Store  []Store
		Offset int
		Limit  int
		Count  int
	}
}

type OrganizationResponse struct {
	Code   int
	Status string
	Data   map[string]Organization
}

type Organization struct {
	Id           int
	Name         string
	Currency     Currency
	DefaultStore struct {
		Id int
	}
}

type Currency struct {
	Id     int
	Name   string
	Symbol string
}

type ConfigResponse struct {
	Code   int
	Status string
	Data   map[string]map[string]map[string]string // data->config->inventory are 3 maps
}

// In cases of error, it will return empty struct
func GetStoreData(organizationId, storeId int) Store {
	var storeData Store
	var url = Urls["account-service"]
	url = url + fmt.Sprintf("/store/%d?organizationId=%d", storeId, organizationId)
	data := getExternalData(url)
	var response StoreResponse
	err := json.Unmarshal(data, &response)
	if err != nil {
		log.Println(err)
	} else {
		storeData = response.Data.Store
	}
	return storeData
}

func GetOrganizationData(organizationId int) Organization {
	var organization Organization
	var url = Urls["account-service"]
	url = url + fmt.Sprintf("/organization/%d", organizationId)
	data := getExternalData(url)
	var response OrganizationResponse
	err := json.Unmarshal(data, &response)
	if err != nil {
		log.Println(err)
	} else {
		organization = response.Data["organization"]
	}
	return organization
}

func GetOrganizationConfig(organizationId int) map[string]string {
	var configs = make(map[string]string)
	var url = Urls["account-service"]
	url = url + fmt.Sprintf("/config/inventory?organizationId=%d", organizationId)
	data := getExternalData(url)
	var response ConfigResponse
	err := json.Unmarshal(data, &response)
	if err != nil {
		log.Println(err)
	} else {
		configs = response.Data["config"]["inventory"]
	}
	return configs
}

func GetAllStores(organizationId int) (allStores []Store) {
	var response StoreListResponse
	var url = Urls["account-service"]
	url = url + fmt.Sprintf("/store?organizationId=%d&paginated=false", organizationId)
	data := getExternalData(url)
	err := json.Unmarshal(data, &response)
	if err != nil {
		log.Println(err)
		return
	}
	allStores = response.Data.Store
	return
}

// private
func getExternalData(url string) []byte {
	res, err := http.Get(url)
	if err != nil {
		log.Println(url, " : ", err)
		return nil
	}
	if res.StatusCode != 200 {
		return nil
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(url, " : ", err)
		return nil
	}
	return data
}
