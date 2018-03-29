package controller

import (
	"../../go-zopsmart/appError"
	"../../go-zopsmart/server"
	"../model"
	appUtil "../utility"
	"encoding/json"
	"net/http"
)

const (
	MAX_PER_PAGE int = 20
)

// Request level Caching. Initialize these values Globally for each request.
var (
	storeMap                            map[int]model.Store
	currency                            model.Currency
	extensions                          map[string]bool
	isMultiStoreExtensionEnabled        = false
	isInStoreProcessingExtensionEnabled = false
	stockStrategy                       string
)

func CallFunctionByName(name string, r *http.Request) (interface{}, appError.AppError) {
	organizationId := getOrganizationId(r)
	if organizationId != 0 {
		initializeOrganizationData(organizationId)
	}
	var res interface{}
	var err appError.AppError
	switch name {
	case "ItemGet":
		res, err = ItemGet(r)
	case "ItemPost":
		res, err = ItemPost(r)
	default:
		err = appError.NewValidationError("Method not supported")
	}
	return res, err
}

// Request Level Caching. This needs to be updated for each request.
func initializeOrganizationData(organizationId int) {
	// Fetching organization level data at once
	orgData := appUtil.GetOrganizationData(organizationId)
	currency = model.Currency(orgData.Currency)
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
	// For other methods, we need to read from json Body
	var request = struct{ OrganizationId int }{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		panic(appError.NewValidationError("Incorrect Request"))
	}
	organizationId = request.OrganizationId
	return
}
