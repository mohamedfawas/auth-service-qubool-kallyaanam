package models

import (
	"time"

	"github.com/google/uuid"
)

// RegisterResponse represents the response sent back to the client after registration
type RegisterResponse struct {
	PendingID uuid.UUID `json:"-"` // Not exposed to client
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// VerifyOTPResponse represents the response after OTP verification
type VerifyOTPResponse struct {
	OTPType  string `json:"otp_type"`
	Verified bool   `json:"verified"`
	Message  string `json:"message"`
}

// CompleteRegistrationResponse represents the response after completing registration
type CompleteRegistrationResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
