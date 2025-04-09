// internal/model/user.go
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PendingRegistration struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Phone     string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"phone"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	OTP       string    `gorm:"type:varchar(10);not null" json:"-"`
	OTPExpiry time.Time `gorm:"not null" json:"otp_expiry"`
}

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Phone     string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"phone"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

func (p *PendingRegistration) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// Add a struct for registration request and response
type RegistrationRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type RegistrationResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}
