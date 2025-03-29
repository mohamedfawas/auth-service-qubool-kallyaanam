package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/domain/models"
)

// GormUserRepository implements the UserRepository interface using GORM
type GormUserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &GormUserRepository{db: db}
}

// CheckUserExists checks if a user with the given email or phone exists
func (r *GormUserRepository) CheckUserExists(ctx context.Context, email, phone string) (bool, string, error) {
	var user models.User

	// Check email
	result := r.db.WithContext(ctx).Where("LOWER(email) = LOWER(?) AND deleted_at IS NULL", email).First(&user)
	if result.Error == nil {
		return true, "email", nil
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, "", result.Error
	}

	// Check phone
	result = r.db.WithContext(ctx).Where("phone = ? AND deleted_at IS NULL", phone).First(&user)
	if result.Error == nil {
		return true, "phone", nil
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, "", result.Error
	}

	return false, "", nil
}

// CreateUser creates a new user
func (r *GormUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Create(user)
	return result.Error
}
