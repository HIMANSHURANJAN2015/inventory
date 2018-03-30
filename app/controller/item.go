package controller

import (
	zError "../../go-zopsmart/appError"
	"../../go-zopsmart/server"
	"../model"
	appUtil "../utility"
	"encoding/json"
	"fmt"
	"net/http"
)

func init() {
	zError.Debug("init called of item package")
}

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

func ItemGet(r *http.Request) (response interface{}, appError zError.AppError) {
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

	// clientItemId can be and integer or array of integers
	var organizationId, id, page, storeId int
	var clientItemIds []int
	organizationId = data.IntegerParams["organizationId"]
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
		// response will contain only 1 item, uniquely identified by id or clientItemId
		var item model.ItemStruct
		var itemPtr *model.ItemStruct
		if id != 0 {
			itemPtr = model.GetItemById(id, organizationId)
			if itemPtr == nil {
				panic(zError.NewModelError(fmt.Sprintf("Item with id %d not found", id)))
			}
		} else {
			itemPtr = model.GetItemFromClientId(clientItemIds[0], organizationId)
			if itemPtr == nil {
				panic(zError.NewModelError(fmt.Sprintf("Item with given client id %d not found", clientItemIds[0])))
			}
		}
		item = *itemPtr
		itemDetails := model.GetItemDetails(item.Id, storeId)
		if itemDetails.StoreSpecificProperty == nil {
			panic(zError.NewModelError("Item details not found in this store"))
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
	return
}

func formatItemStruct(organizationId int, item model.ItemStruct) model.ItemStruct {
	storeSpecificProperty := item.StoreSpecificProperty
	for i, storeData := range storeSpecificProperty {
		storeId := storeData.StoreId
		k, ok := storeMap[storeId]
		if !ok {
			storeInfo := appUtil.GetStoreData(organizationId, storeId)
			k = model.Store(storeInfo)
			storeMap[storeId] = k
		}
		if k.Id != 0 {
			storeData.Store = &k
		}
		if currency.Id != 0 {
			storeData.Currency = &currency
		}
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

// StoreData is mandatory
func ItemPost(r *http.Request) (response interface{}, appError zError.AppError) {
	// Unable to use ItemStuct of model because, we are reading StoreSpecificData(Request) and returning StoreSpecificProperty(Response)
	var requestData = struct {
		ClientItemId      int
		OrganizationId    int
		StoreSpecificData []model.StoreData
	}{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestData)
	if err != nil {
		zError.Debug("In Item Post controller", err)
		panic(zError.NewValidationError("Incorrect request format"))
	}
	organizationId := requestData.OrganizationId
	zError.Debug(requestData.StoreSpecificData[0])
	// Validation of Request
	switch {
	case requestData.ClientItemId == 0:
		panic(zError.NewValidationError("Missing clientItemId"))
	case requestData.OrganizationId == 0:
		panic(zError.NewValidationError("Missing organizationId"))
	case len(requestData.StoreSpecificData) == 0:
		panic(zError.NewValidationError("Invalid store data"))
	case (model.GetItemFromClientId(requestData.ClientItemId, organizationId)) != nil:
		panic(zError.NewValidationError("Item Already Exists with given clientItemId"))
	}
	// Fetching all storeIds and caching it
	allStores := appUtil.GetAllStores(organizationId)
	for _, storeInfo := range allStores {
		storeMap[storeInfo.Id] = model.Store(storeInfo)
	}
	var storeDataToAdd = make(map[int]model.StoreData)
	for _, storeInfo := range requestData.StoreSpecificData {
		if !isMultiStoreExtensionEnabled {
			// Taking defaultStoreId
			storeInfo.StoreId = defaultStoreId
		}
		_, validStore := storeMap[storeInfo.StoreId]
		switch {
		case storeInfo.StoreId == 0:
			panic(zError.NewValidationError("Store Id cannot be empty"))
		case !validStore:
			panic(zError.NewValidationError("Invalid storeId"))
		case storeInfo.Mrp == 0:
			panic(zError.NewValidationError("Please pass mrp"))
		case storeInfo.Mrp < storeInfo.Discount:
			panic(zError.NewValidationError("Mrp cannot be less than discount"))
		case storeInfo.Stock < 0:
			panic(zError.NewValidationError("Stock cannot be negative"))
		}
		storeDataToAdd[storeInfo.StoreId] = storeInfo
	}
	itemId := model.AddItem(requestData.ClientItemId, organizationId)
	model.AddStoreData(itemId, storeDataToAdd, nil)

	itemDetails := model.GetItemDetails(itemId, 0)
	itemDetails = formatItemStruct(organizationId, itemDetails)
	response = ItemGetDetailsResponse{200, "SUCCESS", map[string]model.ItemStruct{"item": itemDetails}}
	return
}
