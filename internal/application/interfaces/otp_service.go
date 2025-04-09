package interfaces

import (
	"context"
	"time"
)

// OTPService defines the interface for OTP operations
type OTPService interface {
	// GenerateOTP generates a new OTP of specified length
	GenerateOTP(ctx context.Context, length int) (string, error)

	// GetOTPExpiryTime returns the configured expiry time for OTPs
	GetOTPExpiryTime(ctx context.Context) time.Time
}
