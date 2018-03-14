
package controller

import (
	"../model"
	"net/http"
	"log"
	"../../go-zopsmart/server"
	"../../go-zopsmart/appError"
	//appUtil "../utility"
)

func init() {
	log.Println("init called of item package")
}

const (
	MAX_PER_PAGE int = 20
)

type ItemGetListResponse struct {
	Code int 	`json:"code"`
	Status string	`json:"status"`
	Data ItemGetListData	`json:"data"`
}

type ItemGetListData struct {
		Item []model.ItemStruct `json:"item"`
		Offset int `json:"offset"`
		Limit int `json:"limit"`
		Count int `json:"count"`
}

type ItemGetDetailsResponse struct {
	// code status data
	Code int `json:"code"`
	Status string `json:"status"`
	Data map[string]model.ItemStruct `json:"data"`
}

type ItemPostRequest struct {
	ClientItemId int `json:"clientItemId"`
	OrganizationId int `json:"organizationId"`

}

type storeDataRequest struct {
	StoreId int //optional 
	SellingPrice float64 // optional
	Mrp float64
	Discount float64
	Tax string // We will convert it later, as ot can be ["CGST":12, "SGST" :1]  or 12 
	Barcodes []string
	Rack string
	Shelf string
	Aisle string
}

func ItemGet(r *http.Request) (interface{}, appError.AppError) {
	mandatoryFields := map[string][]string{
		"organizationId" : []string{server.Int},
	}
	optionalFields := map[string][]string {
		"id" : []string{server.Int},
		"page": []string{server.Int},
		"clientItemId": []string{server.Int, server.IntArray},
		"storeId": []string{server.Int},
	}
	data := server.GetRequestParams(r, mandatoryFields, optionalFields)

	var response interface{}
	var organizationId, id, page,storeId int
	var clientItemIds []int
	organizationId = data.IntegerParams["organizationId"]
	id = data.IntegerParams["id"]
	page = data.IntegerParams["page"]
	storeId = data.IntegerParams["storeId"]
	clientIdArray, ok := data.IntegerArrayParams["clientItemId"]
	if ok {
		clientItemIds = clientIdArray
	} else if clientIdInt,ok := data.IntegerParams["clientItemId"]; ok {
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
		//formatItemStruct(organizationId, &itemDetails)
		response = ItemGetDetailsResponse{200, "SUCCESS", map[string]model.ItemStruct{"item":itemDetails}}	
	} else {
		// Get List. If we decalare it outside,then it will be available evn for the next request
		var (
			offset = 0
			paginated = true
		)
		if page > 0 {
			offset = (page-1)*MAX_PER_PAGE
		}
		allItems := model.GetAllItems(clientItemIds,organizationId, storeId, MAX_PER_PAGE, offset, paginated)
		responseData := ItemGetListData{allItems, offset, MAX_PER_PAGE, len(allItems)}
		response = ItemGetListResponse{200, "SUCCESS", responseData}
	}
	// If required filter out those values of store
	return response, appError.AppError{}
}

func formatItemStruct(organizationId int, item *model.ItemStruct) {
	//storeId := item.StoreId
	// orgData := appUtil.GetOrganizationData(organizationId)
	// currency := orgData.Currency
	// log.Println(orgData.Currency)
	// storeData := appUtil.GetStoreData(organizationId, storeId)
	// log.Println(storeData)
	// config := appUtil.GetOrganizationConfig(organizationId)
	// stockStrategy := config["stockStrategy"]
	// taxExclusivePrice := config["taxExclusivePrice"]
	// log.Println(config)
	// isMultiStoreExtensionEnabled := appUtil.IsExtensionEnabled(organizationId, appUtil.MULTI_STORE_EXTENSION)
	// isInStoreProcessingExtensionEnabled := appUtil.IsExtensionEnabled(organizationId, appUtil.IN_STORE_PROCESSING_EXTENSION)
	// log.Println(isMultiStoreExtensionEnabled)
	// log.Println(isInStoreProcessingExtensionEnabled)
	// //item.Store = storeData
	// //item.Currency = currency
	// log.Println(item)
}


func ItemPost(r *http.Request) (interface{}, appError.AppError) {
	return nil, appError.NewModelError("Item with id 1 not found")
}

// func ItemPost(w http.ResponseWriter, r *http.Request) {
// 	log.Println("Reached Post")
// 	// mandatoryFields = []string{"clientItemId", "organizationId"}
// 	// optionalFields = []string{"storeSpecificData"}
// 	log.Println("Request body",r.Body)
// 	var data ItemPostRequest
// 	err := json.NewDecoder(r.Body).Decode(&data)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer r.Body.Close()
// 	log.Println("clientItemId",  data.ClientItemId, data.OrganizationId)
// 	isMultiStoreExtensionEnabled := utility.IsExtensionEnabled(utility.MULTI_STORE_EXTENSION, data.OrganizationId)
// 	log.Println(isMultiStoreExtensionEnabled)
// }

//localhost:8083/item?clientItemId%5B0%5D=123&clientItemId%5B1%5D=987&clientItemId%5B2%5D=767&clientItemId%5B3%5D=345&organizationId=1&paginated=0
//localhost:8083/item?clientItemId=1&clientItemId=2&clientItemId=3&organizationId=45
