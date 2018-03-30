package controller

import (
	"../../go-zopsmart/appError"
	"../../go-zopsmart/server"
	"../model"
	appUtil "../utility"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const (
	MAX_PER_PAGE int = 20
)

// Request level Caching. Initialize these values Globally for each request.
var (
	storeMap                            map[int]model.Store
	currency                            model.Currency
	defaultStoreId                      int
	extensions                          map[string]bool
	isMultiStoreExtensionEnabled        = false
	isInStoreProcessingExtensionEnabled = false
	stockStrategy                       string
)

func CallFunctionByName(name string, r *http.Request) (interface{}, appError.AppError) {
	initializeOrganizationData(r)
	var res interface{}
	var err appError.AppError
	switch name {
	case "ItemGet":
		res, err = ItemGet(r)
	case "ItemPost":
		res, err = ItemPost(r)
	default:
		panic(appError.NewValidationError("Method not supported"))
	}
	return res, err
}

// Request Level Caching. This needs to be updated for each request.
func initializeOrganizationData(r *http.Request) {
	organizationId := getOrganizationId(r)
	if organizationId == 0 {
		return
	}
	// Fetching organization level data at once
	orgData := appUtil.GetOrganizationData(organizationId)
	currency = model.Currency(orgData.Currency)
	defaultStoreId = orgData.DefaultStore.Id
	storeMap = make(map[int]model.Store)
	extensions = make(map[string]bool)
	isMultiStoreExtensionEnabled = appUtil.IsExtensionEnabled(organizationId, appUtil.MULTI_STORE_EXTENSION)
	isInStoreProcessingExtensionEnabled = appUtil.IsExtensionEnabled(organizationId, appUtil.IN_STORE_PROCESSING_EXTENSION)
	config := appUtil.GetOrganizationConfig(organizationId)
	stockStrategy = config["stockStrategy"]
	//taxExclusivePrice := config["taxExclusivePrice"]
	appError.Debug("Cache Update for organization Id :", organizationId)
}

func getOrganizationId(r *http.Request) (organizationId int) {
	if r.Method == "GET" {
		data := server.GetRequestParams(r, map[string][]string{}, map[string][]string{"organizationId": []string{server.Int}})
		organizationId, _ = data.IntegerParams["organizationId"]
		return
	}
	//For other methods, we need to read from json Body
	var request = struct{ OrganizationId int }{}
	requestBody, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(requestBody, &request)
	if err != nil {
		appError.Debug(err)
		panic(appError.NewValidationError("Organization Id must be integer"))
	}
	// Setting the value back, so that controller read it
	r.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
	organizationId = request.OrganizationId
	return
}
