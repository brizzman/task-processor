package validator

import (
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

func (v *Validator) ValidateStruct(s any) error {
	return v.validate.Struct(s)
}

func (v *Validator) ValidationErrorsToStrings(err error) []string {
	var errors []string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, e.Error())
		}
	} else {
		errors = append(errors, err.Error())
	}
	return errors
}