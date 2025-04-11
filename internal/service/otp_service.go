// internal/service/otp_service.go
package service

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/repository"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/repository/redis"
)

// OTPConfig holds OTP service configuration
type OTPConfig struct {
	Length     int
	ExpiryMins int
}

// Implementation of the OTPService interface
type otpService struct {
	config  OTPConfig
	otpRepo repository.OTPRepository
}

// NewOTPService creates a new OTP service instance
func NewOTPService(config OTPConfig, otpRepo repository.OTPRepository) OTPService {
	return &otpService{
		config:  config,
		otpRepo: otpRepo,
	}
}

// GenerateOTP generates a new OTP of specified length
func (s *otpService) GenerateOTP(ctx context.Context, length int) (string, error) {
	if length <= 0 {
		length = s.config.Length
	}

	const otpChars = "0123456789"
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	otpCharsLength := len(otpChars)
	for i := 0; i < length; i++ {
		buffer[i] = otpChars[int(buffer[i])%otpCharsLength]
	}

	return string(buffer), nil
}

// GetOTPExpiryTime returns the configured expiry time for OTPs
func (s *otpService) GetOTPExpiryTime(ctx context.Context) time.Time {
	return time.Now().Add(time.Duration(s.config.ExpiryMins) * time.Minute)
}

// GenerateAndStoreOTP generates a new OTP and stores it in Redis
func (s *otpService) GenerateAndStoreOTP(ctx context.Context, key string) (string, error) {
	otp, err := s.GenerateOTP(ctx, s.config.Length)
	if err != nil {
		return "", err
	}

	expiry := time.Duration(s.config.ExpiryMins) * time.Minute
	err = s.otpRepo.StoreOTP(ctx, key, otp, expiry)
	if err != nil {
		// If Redis is unavailable, we still return the OTP
		// This allows the system to continue working but OTP validation won't work
		if err == redis.ErrRedisUnavailable {
			// Log the error but proceed without storing the OTP
			// In a real system, you'd want to handle this better, perhaps with a fallback mechanism
			return otp, nil
		}
		return "", err
	}

	return otp, nil
}

// VerifyOTP verifies an OTP for a given key
func (s *otpService) VerifyOTP(ctx context.Context, key string, otp string) (bool, error) {
	return s.otpRepo.VerifyOTP(ctx, key, otp)
}
