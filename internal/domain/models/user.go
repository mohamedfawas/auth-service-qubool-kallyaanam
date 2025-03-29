package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a fully registered and verified user
type User struct {
	UserID       uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"user_id"`
	Email        string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Phone        string         `gorm:"uniqueIndex;size:15;not null" json:"phone"`
	PasswordHash []byte         `gorm:"not null" json:"-"`
	Blocked      bool           `gorm:"default:false" json:"blocked"`
	IsDeleted    bool           `gorm:"default:false" json:"is_deleted"`
	LastLogin    *time.Time     `json:"last_login,omitempty"`
	CreatedAt    time.Time      `gorm:"default:now();index" json:"created_at"` // Added index
	UpdatedAt    time.Time      `gorm:"default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
