// internal/repository/redis/verification_repository.go
package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/logging"
	"go.uber.org/zap"
)

// VerificationRepo handles verification codes in Redis.
type VerificationRepo struct {
	client *redis.Client
	logger *zap.Logger
}

// NewVerificationRepo creates a new verification repository.
func NewVerificationRepo(client *redis.Client) *VerificationRepo {
	return &VerificationRepo{
		client: client,
		logger: logging.Logger(),
	}
}

const (
	// emailVerificationPrefix is used for email verification code keys
	emailVerificationPrefix = "email_verification:"
	// phoneVerificationPrefix is used for phone verification code keys
	phoneVerificationPrefix = "phone_verification:"
)

// StoreEmailCode stores an email verification code.
func (r *VerificationRepo) StoreEmailCode(ctx context.Context, email, code string, expiration time.Duration) error {
	key := emailVerificationPrefix + email
	return r.storeCode(ctx, key, code, expiration)
}

// StorePhoneCode stores a phone verification code.
func (r *VerificationRepo) StorePhoneCode(ctx context.Context, phone, code string, expiration time.Duration) error {
	key := phoneVerificationPrefix + phone
	return r.storeCode(ctx, key, code, expiration)
}

// storeCode is a helper method to store a verification code.
func (r *VerificationRepo) storeCode(ctx context.Context, key, code string, expiration time.Duration) error {
	err := r.client.Set(ctx, key, code, expiration).Err()
	if err != nil {
		r.logger.Error("Failed to store verification code",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	r.logger.Debug("Verification code stored",
		zap.String("key", key),
		zap.Duration("expiration", expiration),
	)
	return nil
}
