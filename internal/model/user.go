// internal/model/user.go
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Phone        string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"phone"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Role         string         `gorm:"type:varchar(50);not null" json:"role"`
	IsVerified   bool           `gorm:"not null;default:false" json:"is_verified"`
	LastLoginAt  time.Time      `json:"last_login_at"`
	CreatedAt    time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	IsActive     bool           `gorm:"not null;default:true" json:"is_active"`
}
