// Package entity contains the domain entities of the application.
package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a fully registered user with verified email and phone.
type User struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email         string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	Phone         string         `gorm:"type:varchar(20);uniqueIndex;not null"`
	PasswordHash  string         `gorm:"type:varchar(255);not null"`
	EmailVerified bool           `gorm:"default:true;not null"`
	PhoneVerified bool           `gorm:"default:true;not null"`
	CreatedAt     time.Time      `gorm:"not null"`
	UpdatedAt     time.Time      `gorm:"not null"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the table name used by GORM to `users`
func (User) TableName() string {
	return "users"
}

// BeforeCreate is a GORM hook that sets default values before creating a record
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	now := time.Now().UTC()
	u.CreatedAt = now
	u.UpdatedAt = now
	return nil
}
