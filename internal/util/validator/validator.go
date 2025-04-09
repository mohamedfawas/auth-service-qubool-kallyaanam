// internal/util/validator/validator.go
package validator

import (
	"regexp"
)

// EmailValidator validates email format
func ValidateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	return match
}

// PhoneValidator validates phone number format
// This is a simple example - adapt to your requirements
func ValidatePhone(phone string) bool {
	pattern := `^\+?[0-9]{10,15}$`
	match, _ := regexp.MatchString(pattern, phone)
	return match
}

// PasswordValidator checks if password meets requirements
func ValidatePassword(password string, minLength int) bool {
	if len(password) < minLength {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}
