package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// OTPRepository handles OTP storage and verification
type OTPRepository struct {
	client *redis.Client
}

// NewOTPRepository creates a new instance of OTPRepository
func NewOTPRepository(client *redis.Client) *OTPRepository {
	return &OTPRepository{client: client}
}

// StoreEmailOTP stores an email OTP with expiration
func (r *OTPRepository) StoreEmailOTP(pendingID, email, otp string, expiry time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("email_otp:%s", pendingID)
	return r.client.Set(ctx, key, otp, expiry).Err()
}

// StorePhoneOTP stores a phone OTP with expiration
func (r *OTPRepository) StorePhoneOTP(pendingID, phone, otp string, expiry time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("phone_otp:%s", pendingID)
	return r.client.Set(ctx, key, otp, expiry).Err()
}

// VerifyEmailOTP verifies an email OTP
func (r *OTPRepository) VerifyEmailOTP(pendingID, otp string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("email_otp:%s", pendingID)

	storedOTP, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil // OTP not found or expired
		}
		return false, err
	}

	return storedOTP == otp, nil
}

// VerifyPhoneOTP verifies a phone OTP
func (r *OTPRepository) VerifyPhoneOTP(pendingID, otp string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("phone_otp:%s", pendingID)

	storedOTP, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil // OTP not found or expired
		}
		return false, err
	}

	return storedOTP == otp, nil
}
