// Package validator provides validation utilities.
package validator

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	apperror "github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/errors"
)

var (
	// Common validation patterns
	emailRegex        = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegex        = regexp.MustCompile(`^\+[1-9]\d{1,14}$`) // E.164 format
	passwordMinLength = 8
)

// Validator represents a validator instance.
type Validator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator instance with custom validation rules.
func NewValidator() *Validator {
	v := validator.New()

	// Register custom validation functions
	_ = v.RegisterValidation("e164", validateE164PhoneNumber)
	_ = v.RegisterValidation("strongPassword", validateStrongPassword)

	return &Validator{
		validator: v,
	}
}

// ValidateStruct validates a struct and returns validation errors.
func (v *Validator) ValidateStruct(obj interface{}) map[string]string {
	err := v.validator.Struct(obj)
	if err == nil {
		return nil
	}

	// Create details map to store field-error pairs
	details := make(map[string]string)

	for _, err := range err.(validator.ValidationErrors) {
		field := strings.ToLower(err.Field())
		details[field] = getValidationErrorMessage(err)
	}

	return details
}

// ValidateEmail validates an email address.
func ValidateEmail(email string) *apperror.Error {
	if !emailRegex.MatchString(email) {
		return apperror.ValidationError("Invalid email format")
	}
	return nil
}

// ValidatePhone validates a phone number in E.164 format.
func ValidatePhone(phone string) *apperror.Error {
	if !phoneRegex.MatchString(phone) {
		return apperror.ValidationError("Phone number must be in E.164 format (e.g., +918123456789)")
	}
	return nil
}

// ValidatePassword validates password strength.
func ValidatePassword(password string) *apperror.Error {
	if len(password) < passwordMinLength {
		return apperror.ValidationError("Password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Create validation error with specific details
	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		err := apperror.ValidationError("Password does not meet strength requirements")

		if !hasUpper {
			err.WithDetail("uppercase", "Password must contain at least one uppercase letter")
		}
		if !hasLower {
			err.WithDetail("lowercase", "Password must contain at least one lowercase letter")
		}
		if !hasNumber {
			err.WithDetail("number", "Password must contain at least one number")
		}
		if !hasSpecial {
			err.WithDetail("special", "Password must contain at least one special character")
		}

		return err
	}

	return nil
}

// Custom validator functions

func validateE164PhoneNumber(fl validator.FieldLevel) bool {
	return phoneRegex.MatchString(fl.Field().String())
}

func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return len(password) >= passwordMinLength &&
		strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") &&
		strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") &&
		strings.ContainsAny(password, "0123456789") &&
		strings.ContainsAny(password, "!@#$%^&*()-_=+{}[]|:;<>,.?/~`")
}

// Helper function to get validation error message
func getValidationErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "e164":
		return "Phone number must be in E.164 format (e.g., +918123456789)"
	case "strongPassword":
		return "Password must contain at least 8 characters, including uppercase, lowercase, number, and special character"
	case "min":
		return "Value must be at least " + err.Param()
	case "max":
		return "Value must be at most " + err.Param()
	default:
		return "Validation failed on '" + err.Tag() + "' tag"
	}
}

// ValidateStructToError converts validation details to an apperror.Error
func ValidateStructToError(details map[string]string) *apperror.Error {
	if details == nil || len(details) == 0 {
		return nil
	}

	err := apperror.ValidationError("Validation failed")

	for field, message := range details {
		err.WithDetail(field, message)
	}

	return err
}
