package utils

import (
	"encoding/json"
	"net/http"

	"task-processor/internal/infrastructure/shared/validator"
)

type HTTPResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    any 		`json:"data,omitempty"`
	Errors  []string    `json:"errors,omitempty"`
}

func SendSuccess(w http.ResponseWriter, r *http.Request, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	_  = json.NewEncoder(w).Encode(HTTPResponse{
		Success: true,
		Data:    data,
	})
}

func SendError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	_ = json.NewEncoder(w).Encode(HTTPResponse{
		Success: false,
		Message: message,
	})
}

func SendValidationError(w http.ResponseWriter, r *http.Request, v *validator.Validator, err error) {
	errors := v.ValidationErrorsToStrings(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	
	_ = json.NewEncoder(w).Encode(HTTPResponse{
		Success: false,
		Message: "Validation failed",
		Errors:  errors,
	})
}