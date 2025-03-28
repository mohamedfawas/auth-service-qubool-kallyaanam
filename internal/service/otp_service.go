package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/config"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/domain/models"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository"
)

// OTP error definitions
var (
	ErrOTPGeneration  = errors.New("failed to generate OTP")
	ErrOTPStorage     = errors.New("failed to store OTP")
	ErrOTPExpired     = errors.New("OTP has expired")
	ErrOTPMaxAttempts = errors.New("maximum OTP verification attempts exceeded")
	ErrInvalidOTP     = errors.New("invalid OTP")
	ErrOTPNotFound    = errors.New("OTP not found")
)

// OTPService handles OTP generation and verification
type OTPService struct {
	otpRepo          repository.OTPRepository
	registrationRepo repository.RegistrationRepository
	rateLimitRepo    repository.RateLimitRepository
	config           *config.OTPConfig
}

// NewOTPService creates a new OTP service
func NewOTPService(
	otpRepo repository.OTPRepository,
	registrationRepo repository.RegistrationRepository,
	rateLimitRepo repository.RateLimitRepository,
	config *config.OTPConfig,
) *OTPService {
	return &OTPService{
		otpRepo:          otpRepo,
		registrationRepo: registrationRepo,
		rateLimitRepo:    rateLimitRepo,
		config:           config,
	}
}

// GenerateOTP generates and stores a new OTP for a pending registration
func (s *OTPService) GenerateOTP(ctx context.Context, pendingID uuid.UUID, otpType string) (string, error) {
	// Check if pending registration exists
	registration, err := s.registrationRepo.GetPendingRegistrationByID(ctx, pendingID)
	if err != nil {
		return "", err
	}
	if registration == nil {
		return "", errors.New("pending registration not found or expired")
	}

	// Generate OTP
	otpValue, err := s.generateRandomOTP(s.config.Length)
	if err != nil {
		return "", ErrOTPGeneration
	}

	// Store OTP
	otp := &models.VerificationOTP{
		OTPID:     uuid.New(),
		PendingID: pendingID,
		OTPType:   otpType,
		OTPValue:  otpValue,
		Attempts:  0,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Duration(s.config.ExpiryMinutes) * time.Minute),
	}

	if err := s.otpRepo.StoreOTP(ctx, otp); err != nil {
		return "", ErrOTPStorage
	}

	return otpValue, nil
}

// generateRandomOTP generates a random numeric OTP of specified length
func (s *OTPService) generateRandomOTP(length int) (string, error) {
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

// GenerateOTPsForRegistration generates both email and phone OTPs for a registration
func (s *OTPService) GenerateOTPsForRegistration(ctx context.Context, pendingID uuid.UUID) (string, string, error) {
	// Generate email OTP
	emailOTP, err := s.GenerateOTP(ctx, pendingID, "email")
	if err != nil {
		return "", "", fmt.Errorf("failed to generate email OTP: %w", err)
	}

	// Generate phone OTP
	phoneOTP, err := s.GenerateOTP(ctx, pendingID, "phone")
	if err != nil {
		return "", "", fmt.Errorf("failed to generate phone OTP: %w", err)
	}

	return emailOTP, phoneOTP, nil
}

// VerifyOTP verifies an OTP for a pending registration
func (s *OTPService) VerifyOTP(ctx context.Context, pendingID uuid.UUID, otpType string, otpValue string) (bool, error) {
	// Get the OTP
	otp, err := s.otpRepo.GetOTPByPendingIDAndType(ctx, pendingID, otpType)
	if err != nil {
		return false, err
	}
	if otp == nil {
		return false, ErrOTPNotFound
	}

	// Check if OTP is expired
	if time.Now().After(otp.ExpiresAt) {
		return false, ErrOTPExpired
	}

	// Check attempts
	if otp.Attempts >= s.config.MaxAttempts {
		return false, ErrOTPMaxAttempts
	}

	// Increment attempts
	if err := s.otpRepo.IncrementOTPAttempts(ctx, otp.OTPID); err != nil {
		return false, err
	}

	// Verify OTP
	return otp.OTPValue == otpValue, nil
}

// CleanupExpiredData removes expired registrations and OTPs
func (s *OTPService) CleanupExpiredData(ctx context.Context) error {
	// Clean expired OTPs
	_, err := s.otpRepo.CleanExpiredOTPs(ctx)
	if err != nil {
		return err
	}

	// Clean expired registrations
	_, err = s.registrationRepo.CleanExpiredRegistrations(ctx)
	return err
}

// OTPPair holds both email and phone OTPs
type OTPPair struct {
	EmailOTP string
	PhoneOTP string
}

// GetOTPsForRegistration retrieves both OTPs for a registration (for development only)
func (s *OTPService) GetOTPsForRegistration(ctx context.Context, pendingID uuid.UUID) (*OTPPair, error) {
	emailOTP, err := s.otpRepo.GetOTPByPendingIDAndType(ctx, pendingID, "email")
	if err != nil {
		return nil, err
	}

	phoneOTP, err := s.otpRepo.GetOTPByPendingIDAndType(ctx, pendingID, "phone")
	if err != nil {
		return nil, err
	}

	if emailOTP == nil || phoneOTP == nil {
		return nil, errors.New("OTPs not found")
	}

	return &OTPPair{
		EmailOTP: emailOTP.OTPValue,
		PhoneOTP: phoneOTP.OTPValue,
	}, nil
}
