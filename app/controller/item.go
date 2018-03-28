package controller

import (
	"../../go-zopsmart/appError"
	"../../go-zopsmart/server"
	"../model"
	appUtil "../utility"
	"log"
	"net/http"
)

func init() {
	log.Println("init called of item package")
}

const (
	MAX_PER_PAGE int = 20
)

// Request level Caching. Initialize this global for each request.
var (
	storeMap                            map[int]model.Store
	currency                            model.Currency
	extensions                          map[string]bool
	isMultiStoreExtensionEnabled        = false
	isInStoreProcessingExtensionEnabled = false
	stockStrategy                       string
)

type ItemGetListResponse struct {
	Code   int             `json:"code"`
	Status string          `json:"status"`
	Data   ItemGetListData `json:"data"`
}

type ItemGetListData struct {
	Item   []model.ItemStruct `json:"item"`
	Offset int                `json:"offset"`
	Limit  int                `json:"limit"`
	Count  int                `json:"count"`
}

type ItemGetDetailsResponse struct {
	// code status data
	Code   int                         `json:"code"`
	Status string                      `json:"status"`
	Data   map[string]model.ItemStruct `json:"data"`
}

type ItemPostRequest struct {
	ClientItemId   int `json:"clientItemId"`
	OrganizationId int `json:"organizationId"`
}

type storeDataRequest struct {
	StoreId      int     //optional
	SellingPrice float64 // optional
	Mrp          float64
	Discount     float64
	Tax          string // We will convert it later, as ot can be ["CGST":12, "SGST" :1]  or 12
	Barcodes     []string
	Rack         string
	Shelf        string
	Aisle        string
}

func ItemGet(r *http.Request) (interface{}, appError.AppError) {
	mandatoryFields := map[string][]string{
		"organizationId": []string{server.Int},
	}
	optionalFields := map[string][]string{
		"id":           []string{server.Int},
		"page":         []string{server.Int},
		"clientItemId": []string{server.Int, server.IntArray},
		"storeId":      []string{server.Int},
	}
	data := server.GetRequestParams(r, mandatoryFields, optionalFields)

	var response interface{}
	var organizationId, id, page, storeId int
	var clientItemIds []int
	organizationId = data.IntegerParams["organizationId"]
	initializeOrganizationData(organizationId)
	id = data.IntegerParams["id"]
	page = data.IntegerParams["page"]
	storeId = data.IntegerParams["storeId"]
	clientIdArray, ok := data.IntegerArrayParams["clientItemId"]
	if ok {
		clientItemIds = clientIdArray
	} else if clientIdInt, ok := data.IntegerParams["clientItemId"]; ok {
		clientItemIds = []int{clientIdInt}
	}
	if id != 0 || len(clientItemIds) == 1 {
		var item model.ItemStruct
		if id != 0 {
			item = model.GetItemById(id, organizationId)
		} else {
			item = model.GetItemFromClientId(clientItemIds[0], organizationId)
		}
		itemDetails := model.GetItemDetails(item.Id, storeId)
		if itemDetails.StoreSpecificProperty == nil {
			panic(appError.NewModelError("Item details not found in this store"))
		}
		itemDetails = formatItemStruct(organizationId, itemDetails)
		response = ItemGetDetailsResponse{200, "SUCCESS", map[string]model.ItemStruct{"item": itemDetails}}
	} else {
		// Get List. If we decalare it outside,then it will be available evn for the next request
		var (
			offset    = 0
			paginated = true
		)
		if page > 0 {
			offset = (page - 1) * MAX_PER_PAGE
		}
		allItems := model.GetAllItems(clientItemIds, organizationId, storeId, MAX_PER_PAGE, offset, paginated)
		for i, item := range allItems {
			allItems[i] = formatItemStruct(organizationId, item)
		}
		responseData := ItemGetListData{allItems, offset, MAX_PER_PAGE, len(allItems)}
		response = ItemGetListResponse{200, "SUCCESS", responseData}
	}
	return response, appError.AppError{}
}

func formatItemStruct(organizationId int, item model.ItemStruct) model.ItemStruct {
	//storeMap := make(map[int]model.Store)

	if currency.Id == 0 {
		orgData := appUtil.GetOrganizationData(organizationId)
		currency = model.Currency(orgData.Currency)
	}
	storeSpecificProperty := item.StoreSpecificProperty
	for i, storeData := range storeSpecificProperty {
		storeId := storeData.StoreId
		log.Println("StoreId", storeId)
		k, ok := storeMap[storeId]
		if !ok {
			storeInfo := appUtil.GetStoreData(organizationId, storeId)
			k = model.Store(storeInfo)
			storeMap[storeId] = k
		}
		storeData.Store = &k
		storeData.Currency = &currency
		// Not showing aisle, rack and shelf info
		if !isInStoreProcessingExtensionEnabled {
			storeData.Aisle = nil
			storeData.Rack = nil
			storeData.Shelf = nil
		}
		storeSpecificProperty[i] = storeData
	}
	item.StoreSpecificProperty = storeSpecificProperty
	return item
}

// This needs to be called for each request
func initializeOrganizationData(organizationId int) {
	// Fetching organization level data at once
	storeMap = make(map[int]model.Store)
	extensions = make(map[string]bool)
	currency = model.Currency{}
	isMultiStoreExtensionEnabled = appUtil.IsExtensionEnabled(organizationId, appUtil.MULTI_STORE_EXTENSION)
	isInStoreProcessingExtensionEnabled = appUtil.IsExtensionEnabled(organizationId, appUtil.IN_STORE_PROCESSING_EXTENSION)
	config := appUtil.GetOrganizationConfig(organizationId)
	stockStrategy = config["stockStrategy"]
	//taxExclusivePrice := config["taxExclusivePrice"]
}

func ItemPost(r *http.Request) (interface{}, appError.AppError) {
	return nil, appError.NewModelError("Item with id 1 not found")
}
