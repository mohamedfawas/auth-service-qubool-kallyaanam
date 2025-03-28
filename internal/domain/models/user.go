package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PendingRegistration represents a user registration awaiting verification
type PendingRegistration struct {
	PendingID        uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"pending_id"`
	Email            string            `gorm:"uniqueIndex:idx_email_not_expired;size:255;not null" json:"email"`
	Phone            string            `gorm:"uniqueIndex:idx_phone_not_expired;size:15;not null" json:"phone"`
	PasswordHash     []byte            `gorm:"not null" json:"-"`
	EmailVerified    bool              `gorm:"default:false" json:"email_verified"`
	PhoneVerified    bool              `gorm:"default:false" json:"phone_verified"`
	CreatedAt        time.Time         `gorm:"default:now()" json:"created_at"`
	ExpiresAt        time.Time         `gorm:"not null;default:now() + interval '24 hours';index" json:"expires_at"`
	VerificationOTPs []VerificationOTP `gorm:"foreignKey:PendingID;references:PendingID" json:"-"`
}

// VerificationOTP represents an OTP for email or phone verification
type VerificationOTP struct {
	OTPID     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"otp_id"`
	PendingID uuid.UUID `gorm:"type:uuid;not null" json:"pending_id"`
	OTPType   string    `gorm:"size:10;not null" json:"otp_type"` // "email" or "phone"
	OTPValue  string    `gorm:"size:10;not null" json:"-"`        // Actual OTP
	Attempts  int       `gorm:"default:0" json:"attempts"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
}

// User represents a fully registered and verified user
type User struct {
	UserID       uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"user_id"`
	Email        string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Phone        string         `gorm:"uniqueIndex;size:15;not null" json:"phone"`
	PasswordHash []byte         `gorm:"not null" json:"-"`
	Blocked      bool           `gorm:"default:false" json:"blocked"`
	LastLogin    *time.Time     `json:"last_login,omitempty"`
	CreatedAt    time.Time      `gorm:"default:now()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// RegisterRequest represents the data sent by the client to register
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Phone    string `json:"phone" binding:"required,max=15"`
	Password string `json:"password" binding:"required,min=8"`
}

// RegisterResponse represents the response sent back to the client
type RegisterResponse struct {
	PendingID uuid.UUID `json:"pending_id,omitempty"`
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}
