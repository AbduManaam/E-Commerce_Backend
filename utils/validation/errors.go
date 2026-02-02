package validator

import "github.com/go-playground/validator/v10"

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func FormatErrors(err error) []ValidationError {
	var errors []ValidationError

	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, ValidationError{
			Field:   err.Field(),
			Message: messageForTag(err),
		})
	}

	return errors
}

func messageForTag(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "must be at least " + err.Param()
	case "max":
		return "must be at most " + err.Param()
	case "gt":
		return "must be greater than " + err.Param()
	case "oneof":
		return "must be one of: " + err.Param()
	default:
		return "is invalid"
	}
}
