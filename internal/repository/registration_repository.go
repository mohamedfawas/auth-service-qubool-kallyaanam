package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/domain/models"
)

// GormRegistrationRepository implements the RegistrationRepository interface using GORM
type GormRegistrationRepository struct {
	db *gorm.DB
}

// NewRegistrationRepository creates a new registration repository
func NewRegistrationRepository(db *gorm.DB) RegistrationRepository {
	return &GormRegistrationRepository{db: db}
}

// CreatePendingRegistration creates a new pending registration
func (r *GormRegistrationRepository) CreatePendingRegistration(ctx context.Context, registration *models.PendingRegistration) error {
	// Normalize email to lowercase
	registration.Email = strings.ToLower(registration.Email)

	result := r.db.WithContext(ctx).Create(registration)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// GetPendingRegistrationByEmail retrieves a pending registration by email
func (r *GormRegistrationRepository) GetPendingRegistrationByEmail(ctx context.Context, email string) (*models.PendingRegistration, error) {
	var registration models.PendingRegistration

	// Only get non-expired registrations
	result := r.db.WithContext(ctx).
		Where("LOWER(email) = LOWER(?) AND expires_at > NOW()", email).
		First(&registration)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &registration, nil
}

// GetPendingRegistrationByPhone retrieves a pending registration by phone
func (r *GormRegistrationRepository) GetPendingRegistrationByPhone(ctx context.Context, phone string) (*models.PendingRegistration, error) {
	var registration models.PendingRegistration

	// Only get non-expired registrations
	result := r.db.WithContext(ctx).
		Where("phone = ? AND expires_at > NOW()", phone).
		First(&registration)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &registration, nil
}

// And add this implementation to your GormRegistrationRepository:
func (r *GormRegistrationRepository) CleanExpiredRegistrations(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&models.PendingRegistration{})

	return result.RowsAffected, result.Error
}

// Add this method to the GormRegistrationRepository
// GetPendingRegistrationByID retrieves a pending registration by ID
func (r *GormRegistrationRepository) GetPendingRegistrationByID(ctx context.Context, id uuid.UUID) (*models.PendingRegistration, error) {
	var registration models.PendingRegistration

	// Only get non-expired registrations
	result := r.db.WithContext(ctx).
		Where("pending_id = ? AND expires_at > NOW()", id).
		First(&registration)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &registration, nil
}

// UpdateVerificationStatus updates the verification status of a pending registration
func (r *GormRegistrationRepository) UpdateVerificationStatus(ctx context.Context, id uuid.UUID, field string, value bool) error {
	// Create a map for the update to allow dynamic field name
	updates := map[string]interface{}{
		field: value,
	}

	result := r.db.WithContext(ctx).
		Model(&models.PendingRegistration{}).
		Where("pending_id = ?", id).
		Updates(updates)

	return result.Error
}
