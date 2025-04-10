// internal/service/otp_service.go
package service

import (
	"context"
	"crypto/rand"
	"time"
)

// OTPConfig holds OTP service configuration
type OTPConfig struct {
	Length     int
	ExpiryMins int
}

// Implementation of the OTPService interface
type otpService struct {
	config OTPConfig
}

// NewOTPService creates a new OTP service instance
func NewOTPService(config OTPConfig) OTPService {
	return &otpService{
		config: config,
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
