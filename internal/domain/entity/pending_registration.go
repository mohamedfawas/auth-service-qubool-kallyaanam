package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PendingRegistration represents a user registration that is pending email/phone verification.
// Once verification is complete, the data will be moved to the permanent User entity.
type PendingRegistration struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email                 string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	Phone                 string         `gorm:"type:varchar(20);uniqueIndex;not null"`
	PasswordHash          string         `gorm:"type:varchar(255);not null"`
	EmailVerificationCode string         `gorm:"type:varchar(8)"`
	PhoneVerificationCode string         `gorm:"type:varchar(8)"`
	EmailVerified         bool           `gorm:"default:false"`
	PhoneVerified         bool           `gorm:"default:false"`
	CreatedAt             time.Time      `gorm:"not null"`
	UpdatedAt             time.Time      `gorm:"not null"`
	ExpiresAt             time.Time      `gorm:"not null"`
	DeletedAt             gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the table name used by GORM to `pending_registrations`
func (PendingRegistration) TableName() string {
	return "pending_registrations"
}

// BeforeCreate is a GORM hook that sets default values before creating a record
func (p *PendingRegistration) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now
	// Default expiration is 24 hours if not set
	if p.ExpiresAt.IsZero() {
		p.ExpiresAt = now.Add(24 * time.Hour)
	}
	return nil
}
