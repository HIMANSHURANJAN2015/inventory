package server

import (
	"encoding/json"
	"net/http"
	"io"
	"log"
	"../appError"	
)

type ErrorString struct {
	Status string `json:"status"`
	Message  string `json:"error"`
}

func WriteJsonResponse(w http.ResponseWriter, bytes []byte) {
	// Set headers here
	log.Println("reached here")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(bytes)
}

func WriteError(w http.ResponseWriter, err appError.AppError) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(err.Code)
	response, err1 := json.Marshal(err)
	if err1 != nil {
		http.Error(w, err1.Error(), 500)
	}
	io.WriteString(w, string(response))
}

