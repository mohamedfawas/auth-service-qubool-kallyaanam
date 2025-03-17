package models

import (
	"time"

	"github.com/google/uuid"
)

type PendingRegistration struct {
	PendingID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email         string    `gorm:"type:varchar(255);not null"`
	Phone         string    `gorm:"type:varchar(15);not null"`
	PasswordHash  []byte    `gorm:"type:bytea;not null"`
	EmailVerified bool      `gorm:"default:false"`
	PhoneVerified bool      `gorm:"default:false"`
	CreatedAt     time.Time `gorm:"default:now()"`
	ExpiresAt     time.Time `gorm:"default:now() + interval '24 hours'"`
}
