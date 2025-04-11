// internal/repository/redis/otp_repository.go
package redis

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/repository"
	redisClient "github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/pkg/redis"
)

var (
	ErrOTPNotFound      = errors.New("OTP not found or expired")
	ErrRedisUnavailable = errors.New("Redis is unavailable")
)

type OTPRepository struct {
	client *redisClient.Client
	prefix string
}

// NewOTPRepository creates a new Redis-based OTP repository
func NewOTPRepository(client *redisClient.Client) repository.OTPRepository {
	return &OTPRepository{
		client: client,
		prefix: "otp:",
	}
}

// StoreOTP stores an OTP with the given key and expiry
func (r *OTPRepository) StoreOTP(ctx context.Context, key string, otp string, expiry time.Duration) error {
	if r.client == nil || r.client.Client == nil {
		return ErrRedisUnavailable
	}
	return r.client.Set(ctx, r.prefix+key, otp, expiry).Err()
}

// GetOTP retrieves an OTP for the given key
func (r *OTPRepository) GetOTP(ctx context.Context, key string) (string, error) {
	if r.client == nil || r.client.Client == nil {
		return "", ErrRedisUnavailable
	}

	val, err := r.client.Get(ctx, r.prefix+key).Result()
	if err == redis.Nil {
		return "", ErrOTPNotFound
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

// DeleteOTP deletes an OTP for the given key
func (r *OTPRepository) DeleteOTP(ctx context.Context, key string) error {
	if r.client == nil || r.client.Client == nil {
		return ErrRedisUnavailable
	}
	return r.client.Del(ctx, r.prefix+key).Err()
}

// VerifyOTP checks if the provided OTP matches the stored OTP
func (r *OTPRepository) VerifyOTP(ctx context.Context, key string, otp string) (bool, error) {
	if r.client == nil || r.client.Client == nil {
		return false, ErrRedisUnavailable
	}

	storedOTP, err := r.GetOTP(ctx, key)
	if err != nil {
		if err == ErrOTPNotFound {
			return false, nil
		}
		return false, err
	}

	if storedOTP == otp {
		// Delete the OTP after successful verification (one-time use)
		if err := r.DeleteOTP(ctx, key); err != nil {
			// Log the error but don't fail the verification
			// Consider implementing a retry mechanism
		}
		return true, nil
	}

	return false, nil
}
