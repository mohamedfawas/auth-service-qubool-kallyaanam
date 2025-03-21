package postgres

import (
	"time"

	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/models"
	"gorm.io/gorm"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreatePendingRegistration creates a new pending registration
func (r *UserRepository) CreatePendingRegistration(registration *models.PendingRegistration) error {
	return r.db.Create(registration).Error
}

// FindPendingRegistrationByEmail finds a pending registration by email
func (r *UserRepository) FindPendingRegistrationByEmail(email string) (*models.PendingRegistration, error) {
	var registration models.PendingRegistration
	err := r.db.Where("email = ?", email).First(&registration).Error
	if err != nil {
		return nil, err
	}
	return &registration, nil
}

// FindPendingRegistrationByPhone finds a pending registration by phone
func (r *UserRepository) FindPendingRegistrationByPhone(phone string) (*models.PendingRegistration, error) {
	var registration models.PendingRegistration
	err := r.db.Where("phone = ?", phone).First(&registration).Error
	if err != nil {
		return nil, err
	}
	return &registration, nil
}

// CheckEmailExists checks if an email exists in the users table
func (r *UserRepository) CheckEmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckPhoneExists checks if a phone number exists in the users table
func (r *UserRepository) CheckPhoneExists(phone string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("phone = ?", phone).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteExpiredRegistrations removes all expired pending registrations
func (r *UserRepository) DeleteExpiredRegistrations() error {
	return r.db.Where("expires_at < ?", time.Now().UTC()).Delete(&models.PendingRegistration{}).Error
}
