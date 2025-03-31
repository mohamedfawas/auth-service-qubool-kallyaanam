// internal/repository/postgres/pending_registration_repository.go
package postgres

import (
	"context"
	"errors"

	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/domain/entity"
	apperrors "github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/errors"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/logging"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PendingRegistrationRepo handles database operations for pending registrations.
type PendingRegistrationRepo struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPendingRegistrationRepo creates a new pending registration repository.
func NewPendingRegistrationRepo(db *gorm.DB) *PendingRegistrationRepo {
	return &PendingRegistrationRepo{
		db:     db,
		logger: logging.Logger(),
	}
}

// Create creates a new pending registration.
func (r *PendingRegistrationRepo) Create(ctx context.Context, registration *entity.PendingRegistration) error {
	r.logger.Info("Creating pending registration",
		zap.String("email", registration.Email),
		zap.String("phone", registration.Phone),
	)

	result := r.db.WithContext(ctx).Create(registration)
	if result.Error != nil {
		// Check for duplicate key error
		if isDuplicateKeyError(result.Error) {
			r.logger.Warn("Duplicate registration attempt",
				zap.String("email", registration.Email),
				zap.String("phone", registration.Phone),
			)
			return apperrors.DuplicateError("Email or phone number already exists in pending registrations")
		}

		r.logger.Error("Database error creating pending registration",
			zap.Error(result.Error),
		)
		return apperrors.InternalError(result.Error)
	}

	return nil
}

// EmailExists checks if an email already exists in pending registrations.
func (r *PendingRegistrationRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.PendingRegistration{}).
		Where("email = ?", email).
		Count(&count).Error; err != nil {

		r.logger.Error("Error checking email existence in pending registrations",
			zap.String("email", email),
			zap.Error(err),
		)
		return false, apperrors.InternalError(err)
	}

	return count > 0, nil
}

// PhoneExists checks if a phone already exists in pending registrations.
func (r *PendingRegistrationRepo) PhoneExists(ctx context.Context, phone string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.PendingRegistration{}).
		Where("phone = ?", phone).
		Count(&count).Error; err != nil {

		r.logger.Error("Error checking phone existence in pending registrations",
			zap.String("phone", phone),
			zap.Error(err),
		)
		return false, apperrors.InternalError(err)
	}

	return count > 0, nil
}

// Helper function to check for duplicate key errors
func isDuplicateKeyError(err error) bool {
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) ||
		errors.Is(err, gorm.ErrForeignKeyViolated) ||
		errors.Is(err, gorm.ErrRegistered) ||
		// For PostgreSQL native error messages
		err.Error() == "ERROR: duplicate key value violates unique constraint" ||
		err.Error() == "pq: duplicate key value violates unique constraint")
}
