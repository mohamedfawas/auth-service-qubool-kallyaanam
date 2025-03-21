package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// OTPConfig holds configuration for OTP generation
type OTPConfig struct {
	Length        int
	ExpiryMinutes int
}

// DefaultOTPConfig returns a default OTP configuration
func DefaultOTPConfig() OTPConfig {
	return OTPConfig{
		Length:        6,
		ExpiryMinutes: 15,
	}
}

// GenerateOTP creates a random OTP of the specified length
func GenerateOTP(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("OTP length must be greater than zero")
	}

	const digits = "0123456789"
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		result[i] = digits[num.Int64()]
	}

	return string(result), nil
}

// CalculateExpiryTime calculates the expiry time based on minutes
func CalculateExpiryTime(minutes int) time.Time {
	return time.Now().UTC().Add(time.Duration(minutes) * time.Minute)
}
