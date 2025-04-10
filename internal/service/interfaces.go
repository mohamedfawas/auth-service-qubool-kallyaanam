// auth-service-qubool-kallyaanam/internal/service/interfaces.go
package service

import (
	"context"
	"time"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/model"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Register registers a new user and sends verification OTP
	Register(ctx context.Context, req *model.RegistrationRequest) (*model.RegistrationResponse, error)
	// Additional methods would be added here (login, verify, etc.)
}

// OTPService defines the interface for OTP operations
type OTPService interface {
	// GenerateOTP generates a new OTP of specified length
	GenerateOTP(ctx context.Context, length int) (string, error)
	// GetOTPExpiryTime returns the configured expiry time for OTPs
	GetOTPExpiryTime(ctx context.Context) time.Time
	// Additional methods would be added here (validate OTP, etc.)
}

// EmailService defines the interface for email operations
type EmailService interface {
	// SendVerificationEmail sends verification email with OTP
	SendVerificationEmail(ctx context.Context, to string, otp string) error
	// Additional methods would be added here (send reset password email, etc.)
}

// SecurityService defines security-related operations
type SecurityService interface {
	// SanitizeInput cleans input to prevent XSS
	SanitizeInput(ctx context.Context, input string) string
	// ValidatePassword checks if a password meets security requirements
	ValidatePassword(ctx context.Context, password string) (bool, string)
	// HashPassword hashes a password securely
	HashPassword(ctx context.Context, password string) (string, error)
	// VerifyPassword checks if a password matches its hash
	VerifyPassword(ctx context.Context, hashedPassword, password string) bool
}

// MetricsService defines methods for recording metrics
type MetricsService interface {
	// IncRegistrationAttempt increments the registration attempt counter
	IncRegistrationAttempt(ctx context.Context)
	// IncRegistrationSuccess increments the successful registration counter
	IncRegistrationSuccess(ctx context.Context)
	// IncRegistrationFailure increments the failed registration counter with a reason
	IncRegistrationFailure(ctx context.Context, reason string)
	// RegistrationDuration records the duration of a registration operation
	RegistrationDuration(ctx context.Context, duration float64)
}
