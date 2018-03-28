package main

import (
	"log"
	"net/http"

	"../go-zopsmart/appError"
	"../go-zopsmart/db"
	"../go-zopsmart/server"
	"../go-zopsmart/utility"
	"./controller"
	appUtility "./utility"
)

func init() {
	log.Println("init called of main")
}

var config = &configuration{}

type handler string

var controllerHandler handler

type configuration struct {
	Database    db.MysqlConfig `json:"Database"`
	CallService appUtility.Config
}

func main() {
	// Loading the configuration file
	utility.LoadJsonFromFile("../config/config.json", config)
	// Loading call service extension
	appUtility.Configure(config.CallService)
	// Connect to database
	db.Configure(config.Database, "inventory_service")
	//Starting the server
	server.StartServer(controllerHandler)
}

func (h handler) GetResponse(function string, r *http.Request) (response interface{}, err appError.AppError) {
	response, err = controller.CallFunctionByName(function, r)
	return
}
