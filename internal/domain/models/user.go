package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PendingRegistration struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Phone     string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"phone"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"` // Hashed password, not exposed in JSON
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	OTP       string    `gorm:"type:varchar(10);not null" json:"-"` // OTP not exposed in JSON
	OTPExpiry time.Time `gorm:"not null" json:"otp_expiry"`
}

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Phone     string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"phone"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"` // Hashed password, not exposed in JSON
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

// BeforeCreate will set UUID before creating
func (p *PendingRegistration) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// BeforeCreate will set UUID before creating
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
