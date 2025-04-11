// auth-service-qubool-kallyaanam/internal/service/interfaces.go
package service

import (
	"context"
	"time"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/model/dto"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Register registers a new user and sends verification OTP
	Register(ctx context.Context, req *dto.RegistrationRequest) (*dto.RegistrationResponse, error)
	// VerifyEmail verifies email using OTP sent during registration
	VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) (*dto.VerifyEmailResponse, error)
	// Additional methods would be added here (login, verify, etc.)
}

// internal/service/interfaces.go (update the OTPService interface)
// OTPService defines the interface for OTP operations
type OTPService interface {
	// GenerateOTP generates a new OTP of specified length
	GenerateOTP(ctx context.Context, length int) (string, error)
	// GetOTPExpiryTime returns the configured expiry time for OTPs
	GetOTPExpiryTime(ctx context.Context) time.Time
	// GenerateAndStoreOTP generates a new OTP and stores it in Redis
	GenerateAndStoreOTP(ctx context.Context, key string) (string, error)
	// VerifyOTP verifies an OTP for a given key
	VerifyOTP(ctx context.Context, key string, otp string) (bool, error)
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
	// Registration metrics
	IncRegistrationAttempt(ctx context.Context)
	IncRegistrationSuccess(ctx context.Context)
	IncRegistrationFailure(ctx context.Context, reason string)
	RegistrationDuration(ctx context.Context, seconds float64)

	// Verification metrics
	IncVerificationAttempt(ctx context.Context)
	IncVerificationSuccess(ctx context.Context)
	IncVerificationFailure(ctx context.Context, reason string)
	VerificationDuration(ctx context.Context, seconds float64)
}
