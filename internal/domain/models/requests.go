package models

import (
	"github.com/google/uuid"
)

// RegisterRequest represents the request to register a new user
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Phone    string `json:"phone" binding:"required,max=15"`
	Password string `json:"password" binding:"required,min=8,strongpassword"`
}

// VerifyOTPRequest represents the request to verify an OTP
type VerifyOTPRequest struct {
	PendingID uuid.UUID `json:"pending_id"` // Optional, can be extracted from JWT
	OTPType   string    `json:"otp_type" binding:"required,oneof=email phone"`
	OTPValue  string    `json:"otp_value" binding:"required,len=6"`
}

// CompleteRegistrationRequest represents the request to complete registration
type CompleteRegistrationRequest struct {
	PendingID uuid.UUID `json:"pending_id"` // Optional, can be extracted from JWT
}
