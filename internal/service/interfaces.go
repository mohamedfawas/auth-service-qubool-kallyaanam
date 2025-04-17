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

	// Login authenticates a user and generates JWT tokens
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)

	// RefreshToken refreshes a user's JWT tokens
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error)

	// Logout logs out a user
	Logout(ctx context.Context, req *dto.LogoutRequest) error
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

	// GenerateJWT generates a JWT token for the authenticated user
	GenerateJWT(ctx context.Context, userID, role string, lastLogin time.Time) (string, error)

	// GenerateRefreshToken generates a refresh token for the authenticated user
	GenerateRefreshToken(ctx context.Context, userID string) (string, string, error)

	// ValidateJWT validates the JWT token and returns the claims
	ValidateJWT(ctx context.Context, token string) (map[string]interface{}, error)

	// ExtractTokenID extracts the token ID from the JWT token
	ExtractTokenID(ctx context.Context, token string) (string, error)
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

	// Login metrics
	IncLoginAttempt(ctx context.Context)
	IncLoginSuccess(ctx context.Context)
	IncLoginFailure(ctx context.Context, reason string)
	LoginDuration(ctx context.Context, seconds float64)

	// Token refresh metrics
	IncTokenRefreshAttempt(ctx context.Context)
	IncTokenRefreshSuccess(ctx context.Context)
	IncTokenRefreshFailure(ctx context.Context, reason string)
	TokenRefreshDuration(ctx context.Context, seconds float64)

	// Logout metrics
	IncLogoutAttempt(ctx context.Context)
	IncLogoutSuccess(ctx context.Context)
	IncLogoutFailure(ctx context.Context, reason string)
	LogoutDuration(ctx context.Context, seconds float64)
}

// RedisService defines operations for Redis
type RedisService interface {
	// Token operations
	StoreRefreshToken(ctx context.Context, tokenID, token string, data TokenData) error
	GetRefreshTokenData(ctx context.Context, tokenID string) (*TokenData, error)
	DeleteRefreshToken(ctx context.Context, tokenID string) error

	// Blacklist operations
	BlacklistToken(ctx context.Context, tokenID string, expiry time.Duration) error
	IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error)

	// Throttling operations
	IncrementLoginAttempts(ctx context.Context, ip string) (int64, error)
	IsLoginThrottled(ctx context.Context, ip string) (bool, error)

	// Login history
	StoreLoginHistory(ctx context.Context, userID, userAgent, ip string) error
}
