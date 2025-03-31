// Package repository contains interfaces and implementations for data access.
package repository

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/domain/entity"
)

// PendingRegistrationRepository defines operations needed for registration.
type PendingRegistrationRepository interface {
	// Create creates a new pending registration.
	Create(ctx context.Context, registration *entity.PendingRegistration) error

	// EmailExists checks if an email already exists in pending registrations.
	EmailExists(ctx context.Context, email string) (bool, error)

	// PhoneExists checks if a phone already exists in pending registrations.
	PhoneExists(ctx context.Context, phone string) (bool, error)
}

// UserRepository defines user-related operations needed for registration.
type UserRepository interface {
	// EmailExists checks if an email already exists.
	EmailExists(ctx context.Context, email string) (bool, error)

	// PhoneExists checks if a phone already exists.
	PhoneExists(ctx context.Context, phone string) (bool, error)
}

// SessionRepository defines session operations for registration.
type SessionRepository interface {
	// StoreRegistrationSession stores a registration session.
	StoreRegistrationSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error

	// SetSessionCookie sets a registration session cookie.
	SetSessionCookie(c *gin.Context, sessionID string, maxAge int)
}

// VerificationRepository defines verification operations for registration.
type VerificationRepository interface {
	// StoreEmailCode stores an email verification code.
	StoreEmailCode(ctx context.Context, email, code string, expiration time.Duration) error

	// StorePhoneCode stores a phone verification code.
	StorePhoneCode(ctx context.Context, phone, code string, expiration time.Duration) error
}
