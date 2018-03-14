package controller

import (
	"net/http"
	"../../go-zopsmart/appError"
)

func CallFunctionByName(name string, r *http.Request) (interface{}, appError.AppError) {
	var res interface{}
	var err appError.AppError
	switch name {
		case "ItemGet": res,err = ItemGet(r)
		case "ItemPost": res, err = ItemPost(r)
		default: err = appError.NewValidationError("Method not supported")
	}
	return res,err
}