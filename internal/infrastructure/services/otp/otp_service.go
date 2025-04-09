package otp

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/interfaces"
)

const (
	defaultOTPLength     = 6
	defaultOTPExpiryMins = 15
)

type OTPConfig struct {
	Length     int
	ExpiryMins int
}

type OTPService struct {
	config OTPConfig
}

// NewOTPService creates a new OTP service
func NewOTPService(config OTPConfig) interfaces.OTPService {
	// Use default values if config values are invalid
	if config.Length <= 0 {
		config.Length = defaultOTPLength
	}
	if config.ExpiryMins <= 0 {
		config.ExpiryMins = defaultOTPExpiryMins
	}

	return &OTPService{
		config: config,
	}
}

// GenerateOTP generates a secure random OTP of the specified length
func (s *OTPService) GenerateOTP(ctx context.Context, length int) (string, error) {
	if length <= 0 {
		length = s.config.Length
	}

	// Generate crypto-secure random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to numeric OTP
	const otpChars = "0123456789"
	otp := make([]byte, length)
	for i, b := range bytes {
		otp[i] = otpChars[int(b)%len(otpChars)]
	}

	return string(otp), nil
}

// GetOTPExpiryTime returns the expiry time for OTPs
func (s *OTPService) GetOTPExpiryTime(ctx context.Context) time.Time {
	return time.Now().Add(time.Duration(s.config.ExpiryMins) * time.Minute)
}
