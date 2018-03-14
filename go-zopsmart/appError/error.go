package appError

import(
	"net/http"
	"fmt"
) 

type AppError struct {
	Code int `json:"code"`
	Status string `json:"status"`
	Message  string `json:"error"`
}

func (e *AppError) Error() string{
	return fmt.Sprintf("Error encountered")
}

func NewValidationError(message string) AppError {
	return AppError{http.StatusBadRequest, "Error", message}
}

func NewModelError(message string) AppError {
	return AppError{http.StatusNotFound, "Error", message}
}

func NewServerError(message string) AppError {
	return AppError{http.StatusInternalServerError, "Error", message}	
}
