package repository

import (
	"context"

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
}

// RegistrationRepository defines the interface for pending registration operations
type RegistrationRepository interface {
	// CreatePendingRegistration creates a new pending registration
	CreatePendingRegistration(ctx context.Context, registration *models.PendingRegistration) error

	// GetPendingRegistrationByEmail retrieves a pending registration by email
	GetPendingRegistrationByEmail(ctx context.Context, email string) (*models.PendingRegistration, error)

	// GetPendingRegistrationByPhone retrieves a pending registration by phone
	GetPendingRegistrationByPhone(ctx context.Context, phone string) (*models.PendingRegistration, error)
}
