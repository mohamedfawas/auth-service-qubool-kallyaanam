// Package service contains business logic for the application.
package service

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/domain/dto"
)

// RegistrationService handles user registration operations.
type RegistrationService interface {
	// Register initiates the registration process for a new user.
	// It creates a pending registration, generates verification codes,
	// and sends them to the user's email and phone.
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, string, error)
}

// NotificationService handles sending notifications.
type NotificationService interface {
	// SendEmailVerification sends an email verification code.
	SendEmailVerification(ctx context.Context, email, code string) error

	// SendSMSVerification sends an SMS verification code.
	SendSMSVerification(ctx context.Context, phone, code string) error
}

// SessionService handles session management.
type SessionService interface {
	// CreateRegistrationSession creates a new registration session.
	CreateRegistrationSession(ctx context.Context, data interface{}) (string, error)

	// SetRegistrationSessionCookie sets the registration session cookie.
	SetRegistrationSessionCookie(c *gin.Context, sessionID string)
}
