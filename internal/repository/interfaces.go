package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/domain/models"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// CheckUserExists checks if a user with the given email or phone exists
	CheckUserExists(ctx context.Context, email, phone string) (bool, string, error)

	// GetUserByEmail retrieves a user by their email
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)

	// GetUserByPhone retrieves a user by their phone number
	GetUserByPhone(ctx context.Context, phone string) (*models.User, error)

	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *models.User) error
}

// RegistrationRepository defines the interface for pending registration operations
type RegistrationRepository interface {
	// CreatePendingRegistration creates a new pending registration
	CreatePendingRegistration(ctx context.Context, registration *models.PendingRegistration) error

	// GetPendingRegistrationByEmail retrieves a pending registration by email
	GetPendingRegistrationByEmail(ctx context.Context, email string) (*models.PendingRegistration, error)

	// GetPendingRegistrationByPhone retrieves a pending registration by phone
	GetPendingRegistrationByPhone(ctx context.Context, phone string) (*models.PendingRegistration, error)

	// GetPendingRegistrationsByEmail retrieves pending registrations by email// Add this method to the RegistrationRepository interface
	// GetPendingRegistrationByID retrieves a pending registration by ID
	GetPendingRegistrationByID(ctx context.Context, id uuid.UUID) (*models.PendingRegistration, error)

	// Add this to your RegistrationRepository interface in interfaces.go:
	CleanExpiredRegistrations(ctx context.Context) (int64, error)

	UpdateVerificationStatus(ctx context.Context, id uuid.UUID, field string, value bool) error
}

// OTPRepository defines the interface for OTP operations
type OTPRepository interface {
	// StoreOTP stores a new OTP for verification
	StoreOTP(ctx context.Context, otp *models.VerificationOTP) error

	// GetOTPByPendingIDAndType retrieves an OTP by pending ID and type
	GetOTPByPendingIDAndType(ctx context.Context, pendingID uuid.UUID, otpType string) (*models.VerificationOTP, error)

	// IncrementOTPAttempts increments the attempt counter for an OTP
	IncrementOTPAttempts(ctx context.Context, otpID uuid.UUID) error

	// CleanExpiredOTPs removes expired OTPs
	CleanExpiredOTPs(ctx context.Context) (int64, error)
}

// RateLimitRepository defines the interface for rate limiting operations
type RateLimitRepository interface {
	// IncrementCounter increments a counter for rate limiting
	IncrementCounter(ctx context.Context, key string, expiry time.Duration) (int, error)

	// GetCounter gets the current count for a rate limit key
	GetCounter(ctx context.Context, key string) (int, error)

	// ResetCounter resets a rate limit counter
	ResetCounter(ctx context.Context, key string) error
}

// SessionRepository defines the interface for session management
type SessionRepository interface {
	// StoreSession stores a mapping between a session ID and a pending registration ID
	StoreSession(ctx context.Context, sessionID string, pendingID string, expiry time.Duration) error

	// GetSession retrieves a pending registration ID by session ID
	GetSession(ctx context.Context, sessionID string) (string, error)

	// DeleteSession removes a session
	DeleteSession(ctx context.Context, sessionID string) error
}
