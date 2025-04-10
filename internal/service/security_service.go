// internal/service/security_service.go
package service

import (
	"context"
	"html"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	BcryptCost       int
	MinPasswordChars int
}

// Implementation of the SecurityService interface
type securityService struct {
	config SecurityConfig
}

// NewSecurityService creates a new security service instance
func NewSecurityService(config SecurityConfig) SecurityService {
	return &securityService{
		config: config,
	}
}

// SanitizeInput cleans input to prevent XSS
func (s *securityService) SanitizeInput(ctx context.Context, input string) string {
	// Escape HTML
	sanitized := html.EscapeString(input)
	// Trim spaces
	sanitized = strings.TrimSpace(sanitized)
	return sanitized
}

// ValidatePassword checks if a password meets security requirements
func (s *securityService) ValidatePassword(ctx context.Context, password string) (bool, string) {
	if len(password) < s.config.MinPasswordChars {
		return false, "Password must be at least " + string(rune(s.config.MinPasswordChars)) + " characters long"
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "Password must contain at least one uppercase letter"
	}
	if !hasLower {
		return false, "Password must contain at least one lowercase letter"
	}
	if !hasNumber {
		return false, "Password must contain at least one number"
	}
	if !hasSpecial {
		return false, "Password must contain at least one special character"
	}

	return true, ""
}

// HashPassword hashes a password securely
func (s *securityService) HashPassword(ctx context.Context, password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.config.BcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword checks if a password matches its hash
func (s *securityService) VerifyPassword(ctx context.Context, hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
