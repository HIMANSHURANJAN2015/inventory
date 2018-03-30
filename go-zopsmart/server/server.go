package server

import (
	"../appError"
	"../db"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type ControllerHandler interface {
	GetResponse(string, *http.Request) (interface{}, appError.AppError)
}

// Can accept configuraion if required
func StartServer(controllerHandler ControllerHandler) {
	fmt.Println(time.Now().Format("2006-01-02 03:04:05 PM"), " HTTP Server started")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		router(controllerHandler, w, r)
	})
	// Start the HTTP listener
	err := http.ListenAndServe(":8083", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func router(controllerHandler ControllerHandler, w http.ResponseWriter, r *http.Request) {
	tx, dbErr := db.StartTransaction()
	if dbErr != nil {
		fmt.Println("Unable to start transaction", dbErr)
	}
	defer func() {
		if r := recover(); r != nil {
			dbErr = tx.Rollback()
			log.Println("Error during rollback:", dbErr)
			fmt.Println("Recovered in Base controller router", r)
			err, ok := r.(appError.AppError)
			if ok {
				WriteError(w, err)
			} else {
				fmt.Println("", err, ok)
			}
		}
		tx.Commit()
	}()
	httpVerb := r.Method
	urlParts := GetURIParts(r)
	// splitting "/" on  basis of "/" gives an array of 2 elements
	controllerName := strings.Title(urlParts[1])
	if controllerName == "" {
		err := appError.NewModelError("Controller '' does not exists")
		WriteError(w, err)
		return
	}
	functionName := controllerName + strings.Title(strings.ToLower(httpVerb))
	log.Println("Calling function", functionName)
	// Error/response from controllers/handlers will be retireved here
	response, appErr := controllerHandler.GetResponse(functionName, r)
	if appErr.Code != 0 {
		WriteError(w, appErr)
		return
	}
	bytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	WriteJsonResponse(w, bytes)
}
