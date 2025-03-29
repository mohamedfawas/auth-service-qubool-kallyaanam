package models

import (
	"time"

	"github.com/google/uuid"
)

// PendingRegistration represents a user registration awaiting verification
type PendingRegistration struct {
	PendingID        uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"pending_id"`
	Email            string            `gorm:"uniqueIndex:idx_email_not_expired;size:255;not null" json:"email"`
	Phone            string            `gorm:"uniqueIndex:idx_phone_not_expired;size:15;not null" json:"phone"`
	PasswordHash     []byte            `gorm:"not null" json:"-"`
	EmailVerified    bool              `gorm:"default:false;index" json:"email_verified"`                            // Added index
	PhoneVerified    bool              `gorm:"default:false;index" json:"phone_verified"`                            // Added index
	CreatedAt        time.Time         `gorm:"default:now();index" json:"created_at"`                                // Added index
	ExpiresAt        time.Time         `gorm:"not null;default:now() + interval '24 hours';index" json:"expires_at"` // Index already exists
	VerificationOTPs []VerificationOTP `gorm:"foreignKey:PendingID;references:PendingID" json:"-"`
}
