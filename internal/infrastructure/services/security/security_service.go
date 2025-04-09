package security

import (
	"context"
	"html"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/interfaces"
)

// SecurityConfig contains the configuration for the security service
type SecurityConfig struct {
	BcryptCost       int
	MinPasswordChars int
}

// SecurityService implements the SecurityService interface
type SecurityService struct {
	config SecurityConfig
}

// NewSecurityService creates a new SecurityService instance
func NewSecurityService(config SecurityConfig) interfaces.SecurityService {
	// Use default values if not provided
	if config.BcryptCost <= 0 {
		config.BcryptCost = 12 // Higher is more secure but slower
	}
	if config.MinPasswordChars <= 0 {
		config.MinPasswordChars = 8
	}

	return &SecurityService{
		config: config,
	}
}

// SanitizeInput sanitizes user input to prevent injection attacks
func (s *SecurityService) SanitizeInput(ctx context.Context, input string) string {
	// Convert HTML entities
	sanitized := html.EscapeString(input)

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	// Remove any control characters
	sanitized = regexp.MustCompile(`[\x00-\x1F]`).ReplaceAllString(sanitized, "")

	return sanitized
}

// ValidatePassword checks if a password meets security requirements
func (s *SecurityService) ValidatePassword(ctx context.Context, password string) (bool, string) {
	if len(password) < s.config.MinPasswordChars {
		return false, "Password must be at least 8 characters long"
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
		case unicode.IsNumber(char):
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
func (s *SecurityService) HashPassword(ctx context.Context, password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.config.BcryptCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// VerifyPassword verifies if a password matches its hash
func (s *SecurityService) VerifyPassword(ctx context.Context, hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
