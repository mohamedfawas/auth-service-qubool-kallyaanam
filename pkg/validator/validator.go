package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Update the Setup function to register the new validator
func Setup() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {

		// Register custom validation for password strength
		v.RegisterValidation("strongpassword", validatePasswordStrength)
	}
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

// validatePasswordStrength validates password strength
func validatePasswordStrength(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Check length (already handled by min=8 in the binding tag)
	if len(password) < 8 {
		return false
	}

	// Check for uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)

	// Check for lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)

	// Check for number
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	// Check for special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	// Require at least 3 of the 4 character types
	typeCount := 0
	if hasUpper {
		typeCount++
	}
	if hasLower {
		typeCount++
	}
	if hasNumber {
		typeCount++
	}
	if hasSpecial {
		typeCount++
	}

	return typeCount >= 3
}
