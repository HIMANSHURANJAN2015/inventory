package model

import (
	"../../go-zopsmart/appError"
	"../../go-zopsmart/db"
	"../../go-zopsmart/utility"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
)

func init() {
	log.Println("init called of item model")
}

type ItemStruct struct {
	Id                    int         `json:"id"`
	ClientItemId          int         `json:"clientItemId"`
	OrganizationId        int         `json:"-"`
	StoreSpecificProperty []StoreData `json:"storeSpecificProperty,omitempty"`
}

// Saving as pointer, as marshal will make ecode the value it points to. And pointers can save nil
type StoreData struct {
	SellingPrice float64   `json:"sellingPrice"`
	Mrp          float64   `json:"mrp"`
	Discount     float64   `json:"discount"`
	Stock        int       `json:"stock"` // Will be true,false or integer based on strategy set. But in Db it is saved as int
	Tax          *string   `json:"tax"`   // Can be [{"CGST":"2"},{"SGST":"2"}]  or 12. So saving as string and returning string
	Barcodes     []*string `json:"barcodes"`
	Store        *Store    `json:"store"`
	StoreId      int       `json:"storeId"`
	Currency     *Currency `json:"currency"`
	Aisle        *string   `json:"aisle"`
	Rack         *string   `json:"rack"`
	Shelf        *string   `json:"shelf"`
}

type Store struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	ClientStoreId int    `json:"clientStoreId"`
	Latitude      string `json:"latitude"`
	Longitude     string `json:"longitude"`
	Address       string `json:"address"`
}

type Currency struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

func GetItemById(id int, organizationId int) *ItemStruct {
	row := db.Row("select id, client_item_id, organization_id from items where id =? and organization_id = ? and deleted_at is null", id, organizationId)
	var item ItemStruct
	err := row.Scan(&item.Id, &item.ClientItemId, &item.OrganizationId)
	if err != nil {
		return nil
		//panic(appError.NewModelError(fmt.Sprintf("Item with id %d not found", id)))
	}
	return &item
}

func GetItemFromClientId(clientItemId, organizationId int) *ItemStruct {
	row := db.Row("select id, client_item_id, organization_id from items where client_item_id =? and organization_id = ? and deleted_at is null", clientItemId, organizationId)
	var item ItemStruct
	err := row.Scan(&item.Id, &item.ClientItemId, &item.OrganizationId)
	if err != nil {
		return nil
	}
	return &item
}

// storeId is optional
func GetItemDetails(id int, storeId int) ItemStruct {
	var temp = make(map[int]StoreData)
	var item ItemStruct
	// nil is a different type in Go and being a staticly typed language,we cant assign nil to another type
	query := `select i.id,i.client_item_id,i.organization_id,ifnull(s.store_id,0),ifnull(s.mrp,0),ifnull(s.discount,0),
		ifnull(s.stock,0),s.aisle,s.rack,s.shelf,s.tax,b.barcode from items i 
		left join item_store s on i.id=s.item_id left join 
		item_barcodes b on b.item_store_id=s.id 
		where i.id=?`
	var rows *sql.Rows
	var err error
	if storeId != 0 {
		query = query + " and s.store_id = ?"
		rows, err = db.Select(query, id, storeId)
	} else {
		rows, err = db.Select(query, id)
	}
	defer rows.Close()
	//log.Println("Get details query ",query)
	if err != nil {
		log.Println(err)
		panic(appError.NewModelError(fmt.Sprintf("Item with id %d not found", id)))
	}
	for rows.Next() {
		var data StoreData
		var barcode *string
		err := rows.Scan(&item.Id, &item.ClientItemId, &item.OrganizationId, &data.StoreId, &data.Mrp, &data.Discount, &data.Stock, &data.Aisle, &data.Rack, &data.Shelf, &data.Tax, &barcode)
		if err != nil {
			panic(err)
		}
		if data.StoreId == 0 {
			// No Storedata found for this item
			continue
		}
		// Iterating to club Barcodes for each store
		val, ok := temp[data.StoreId]
		if ok {
			barcodes := val.Barcodes
			data.Barcodes = append(barcodes, barcode)
		} else if barcode != nil {
			data.Barcodes = []*string{barcode}
		}
		temp[data.StoreId] = data
	}
	for _, v := range temp {
		// Adding store specific data only when its there
		item.StoreSpecificProperty = append(item.StoreSpecificProperty, v)
	}
	return item
}

func GetAllItems(clientItemIds []int, organizationId, storeId, maxPerPage, offset int, paginated bool) []ItemStruct {
	query := `select i.id from items i`
	if storeId != 0 {
		query = query + ` LEFT JOIN item_store s ON s.item_id=i.id AND s.store_id =` + strconv.Itoa(storeId)
	}
	query = query + " WHERE i.organization_id = " + strconv.Itoa(organizationId)
	if len(clientItemIds) != 0 {
		query = query + " AND i.client_item_id in (" + utility.IntArrayToString(clientItemIds, ",") + ")"
	}
	query = query + " ORDER BY i.id DESC"
	if paginated {
		query = query + " LIMIT " + strconv.Itoa(maxPerPage) + " OFFSET " + strconv.Itoa(offset)
	}
	rows, err := db.Select(query)
	log.Println("getlist query", query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var items []ItemStruct
	for rows.Next() {
		var id int
		rows.Scan(&id)
		item := GetItemDetails(id, storeId)
		if item.Id != 0 {
			items = append(items, item)
		}
	}
	return items
}

func AddItem(clientItemId int, organizationId int) (id int) {
	res := db.Insert("insert into items (client_item_id, organization_id) values (?,?)", clientItemId, organizationId)
	insertId, _ := res.LastInsertId()
	id = int(insertId)
	return
}

func AddStoreData(itemId int, storeDataToAdd map[int]StoreData, storeDataToUpdate map[int]StoreData) {
	for storeId, storeData := range storeDataToAdd {
		res := db.Insert("insert into item_store(item_id, store_id, mrp, discount,stock,aisle,rack,shelf,tax) values (?,?,?,?,?,?,?,?,?)", itemId, storeId, storeData.Mrp, storeData.Discount, storeData.Stock, storeData.Aisle, storeData.Rack, storeData.Shelf, storeData.Tax)
		insertId, _ := res.LastInsertId()
		itemStoreId := int(insertId)
		// Inserting barcodes
		query := "insert into item_barcodes(item_store_id,barcode) values " // (1,"barcode"),(1."barcode")
		var barcodes = storeData.Barcodes
		for _, barcode := range barcodes {
			query = query + fmt.Sprintf("(%d, '%s'),", itemStoreId, *barcode)
		}
		if len(barcodes) > 0 {
			query = strings.TrimRight(query, ",")
			res = db.Insert(query)
		}
	}
}
