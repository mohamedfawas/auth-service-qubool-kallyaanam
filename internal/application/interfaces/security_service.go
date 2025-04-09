package interfaces

import "context"

// SecurityService defines the interface for security operations
type SecurityService interface {
	// SanitizeInput sanitizes user input to prevent injection attacks
	SanitizeInput(ctx context.Context, input string) string

	// ValidatePassword checks if a password meets security requirements
	ValidatePassword(ctx context.Context, password string) (bool, string)

	// HashPassword hashes a password securely
	HashPassword(ctx context.Context, password string) (string, error)

	// VerifyPassword verifies if a password matches its hash
	VerifyPassword(ctx context.Context, hashedPassword, password string) bool
}
