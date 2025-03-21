package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered user
type User struct {
	UserID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"user_id"`
	Email        string    `gorm:"type:varchar(255);not null;unique" json:"email"`
	Phone        string    `gorm:"type:varchar(15);not null;unique" json:"phone"`
	PasswordHash []byte    `gorm:"type:bytea;not null" json:"-"`
	CreatedAt    time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"default:now()" json:"updated_at"`
}

// PendingRegistration represents a user registration that is awaiting verification
type PendingRegistration struct {
	PendingID     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"pending_id"`
	Email         string    `gorm:"type:varchar(255);not null" json:"email"`
	Phone         string    `gorm:"type:varchar(15);not null" json:"phone"`
	PasswordHash  []byte    `gorm:"type:bytea;not null" json:"-"`
	EmailVerified bool      `gorm:"default:false" json:"email_verified"`
	PhoneVerified bool      `gorm:"default:false" json:"phone_verified"`
	CreatedAt     time.Time `gorm:"default:now()" json:"created_at"`
	ExpiresAt     time.Time `gorm:"not null" json:"expires_at"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// TableName specifies the table name for the PendingRegistration model
func (PendingRegistration) TableName() string {
	return "pending_registrations"
}
