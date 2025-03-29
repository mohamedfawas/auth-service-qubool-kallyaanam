package models

import (
	"time"

	"github.com/google/uuid"
)

// VerificationOTP represents an OTP for email or phone verification
type VerificationOTP struct {
	OTPID     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"otp_id"`
	PendingID uuid.UUID `gorm:"type:uuid;not null;index" json:"pending_id"` // Added index
	OTPType   string    `gorm:"size:10;not null;index" json:"otp_type"`     // Added index
	OTPValue  string    `gorm:"size:10;not null" json:"-"`                  // Actual OTP
	Attempts  int       `gorm:"default:0" json:"attempts"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"` // Index already exists
}
