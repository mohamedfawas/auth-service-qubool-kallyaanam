package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Setup initializes custom validators
func Setup() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Register custom validation for phone number in E.164 format
		v.RegisterValidation("e164", validateE164)
	}
}

// validateE164 validates if a string is a valid E.164 phone number
func validateE164(fl validator.FieldLevel) bool {
	phoneNumber := fl.Field().String()
	// E.164 format: + followed by 1-15 digits
	re := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	return re.MatchString(phoneNumber)
}

// FormatValidationErrors converts validator errors into a user-friendly format
func FormatValidationErrors(err error) map[string]string {
	if err == nil {
		return nil
	}

	errors := make(map[string]string)

	// Type switch for different error types
	switch err.(type) {
	case validator.ValidationErrors:
		for _, err := range err.(validator.ValidationErrors) {
			field := strings.ToLower(err.Field())
			switch err.Tag() {
			case "required":
				errors[field] = "This field is required"
			case "email":
				errors[field] = "Must be a valid email address"
			case "e164":
				errors[field] = "Must be a valid phone number in E.164 format (e.g. +12345678901)"
			case "min":
				errors[field] = fmt.Sprintf("Must be at least %s characters long", err.Param())
			case "max":
				errors[field] = fmt.Sprintf("Must be at most %s characters long", err.Param())
			default:
				errors[field] = fmt.Sprintf("Failed validation on %s", err.Tag())
			}
		}
	default:
		errors["general"] = err.Error()
	}

	return errors
}
