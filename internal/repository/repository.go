// internal/repository/repository.go
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/model"
)

type UserRepository interface {
	CreatePendingRegistration(ctx context.Context, registration *model.PendingRegistration) error
	FindUserByEmail(ctx context.Context, email string) (bool, error)
	FindUserByPhone(ctx context.Context, phone string) (bool, error)
	FindPendingRegistrationByEmail(ctx context.Context, email string) (bool, error)
	FindPendingRegistrationByPhone(ctx context.Context, phone string) (bool, error)

	GetPendingRegistrationByEmail(ctx context.Context, email string) (*model.PendingRegistration, error)
	DeletePendingRegistration(ctx context.Context, id uuid.UUID) error
	CreateUser(ctx context.Context, user *model.User) error
	GetPendingRegistrationByEmailWithLock(ctx context.Context, email string) (*model.PendingRegistration, error)

	// WithTransaction executes operations within a transaction
	WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}

// OTPRepository interface for OTP storage
type OTPRepository interface {
	// StoreOTP stores an OTP with the given key and expiry
	StoreOTP(ctx context.Context, key string, otp string, expiry time.Duration) error

	// GetOTP retrieves an OTP for the given key
	GetOTP(ctx context.Context, key string) (string, error)

	// DeleteOTP deletes an OTP for the given key
	DeleteOTP(ctx context.Context, key string) error

	// VerifyOTP checks if the provided OTP matches the stored OTP
	VerifyOTP(ctx context.Context, key string, otp string) (bool, error)
}
