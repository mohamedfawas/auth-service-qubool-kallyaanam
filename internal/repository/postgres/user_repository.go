// internal/repository/postgres/user_repository.go
package postgres

import (
	"context"

	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/domain/entity"
	apperrors "github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/errors"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/logging"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserRepo handles database operations for users.
type UserRepo struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserRepo creates a new user repository.
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		db:     db,
		logger: logging.Logger(),
	}
}

// EmailExists checks if an email already exists in users table.
func (r *UserRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("email = ?", email).
		Count(&count).Error; err != nil {

		r.logger.Error("Error checking email existence",
			zap.String("email", email),
			zap.Error(err),
		)
		return false, apperrors.InternalError(err)
	}

	return count > 0, nil
}

// PhoneExists checks if a phone already exists in users table.
func (r *UserRepo) PhoneExists(ctx context.Context, phone string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("phone = ?", phone).
		Count(&count).Error; err != nil {

		r.logger.Error("Error checking phone existence",
			zap.String("phone", phone),
			zap.Error(err),
		)
		return false, apperrors.InternalError(err)
	}

	return count > 0, nil
}
