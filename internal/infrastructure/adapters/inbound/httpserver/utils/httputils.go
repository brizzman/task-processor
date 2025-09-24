package utils

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type HTTPResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    any 		`json:"data,omitempty"`
	Errors  []string    `json:"errors,omitempty"`
}

func SendSuccess(w http.ResponseWriter, r *http.Request, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	json.NewEncoder(w).Encode(HTTPResponse{
		Success: true,
		Data:    data,
	})
}

func SendError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	json.NewEncoder(w).Encode(HTTPResponse{
		Success: false,
		Message: message,
	})
}

func SendValidationError(w http.ResponseWriter, r *http.Request, err error) {
	var errors []string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, e.Error())
		}
	} else {
		errors = append(errors, err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	
	json.NewEncoder(w).Encode(HTTPResponse{
		Success: false,
		Message: "Validation failed",
		Errors:  errors,
	})
}

func ValidateStruct(s any) error {
	return validate.Struct(s)
}