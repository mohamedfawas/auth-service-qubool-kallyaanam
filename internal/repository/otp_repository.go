package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/domain/models"
)

// GormOTPRepository implements OTPRepository using GORM
type GormOTPRepository struct {
	db *gorm.DB
}

// NewOTPRepository creates a new OTP repository
func NewOTPRepository(db *gorm.DB) *GormOTPRepository {
	return &GormOTPRepository{db: db}
}

// StoreOTP stores a new OTP for verification
func (r *GormOTPRepository) StoreOTP(ctx context.Context, otp *models.VerificationOTP) error {
	// Delete any existing OTPs of the same type for this pending registration
	err := r.db.WithContext(ctx).
		Where("pending_id = ? AND otp_type = ?", otp.PendingID, otp.OTPType).
		Delete(&models.VerificationOTP{}).Error
	if err != nil {
		return err
	}

	// Store the new OTP
	result := r.db.WithContext(ctx).Create(otp)
	return result.Error
}

// GetOTPByPendingIDAndType retrieves an OTP by pending ID and type
func (r *GormOTPRepository) GetOTPByPendingIDAndType(ctx context.Context, pendingID uuid.UUID, otpType string) (*models.VerificationOTP, error) {
	var otp models.VerificationOTP

	// Get the OTP if it's not expired
	result := r.db.WithContext(ctx).
		Where("pending_id = ? AND otp_type = ? AND expires_at > ?", pendingID, otpType, time.Now().UTC()).
		First(&otp)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &otp, nil
}

// IncrementOTPAttempts increments the attempt counter for an OTP
func (r *GormOTPRepository) IncrementOTPAttempts(ctx context.Context, otpID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.VerificationOTP{}).
		Where("otp_id = ?", otpID).
		UpdateColumn("attempts", gorm.Expr("attempts + 1"))

	return result.Error
}

// CleanExpiredOTPs removes expired OTPs
func (r *GormOTPRepository) CleanExpiredOTPs(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&models.VerificationOTP{})

	return result.RowsAffected, result.Error
}
