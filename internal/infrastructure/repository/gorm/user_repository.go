package gorm

import (
	"context"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/interfaces"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/domain/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) interfaces.UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreatePendingRegistration(ctx context.Context, registration *models.PendingRegistration) error {
	return r.db.WithContext(ctx).Create(registration).Error
}

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) FindUserByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("phone = ?", phone).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) FindPendingRegistrationByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.PendingRegistration{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) FindPendingRegistrationByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.PendingRegistration{}).Where("phone = ?", phone).Count(&count).Error
	return count > 0, err
}
