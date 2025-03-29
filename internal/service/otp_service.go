package service

import (
	"context"
	"crypto/rand"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/domain/models"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository"
)

var (
	ErrOTPGeneration = errors.New("failed to generate OTP")
)

// OTPService handles OTP generation and verification
type OTPService struct {
	otpRepo   *repository.GormOTPRepository
	otpLength int
	otpExpiry time.Duration
}

// NewOTPService creates a new OTP service
func NewOTPService(otpRepo *repository.GormOTPRepository) *OTPService {
	return &OTPService{
		otpRepo:   otpRepo,
		otpLength: 6,                // Default to 6 digits
		otpExpiry: 15 * time.Minute, // Default to 15 minutes
	}
}

// GenerateOTP generates a new OTP for the given pending registration and type
func (s *OTPService) GenerateOTP(ctx context.Context, pendingID uuid.UUID, otpType string) (string, error) {
	// Generate a random 6-digit OTP
	otp, err := s.generateRandomOTP()
	if err != nil {
		return "", ErrOTPGeneration
	}

	// Create OTP record
	verificationOTP := &models.VerificationOTP{
		OTPID:     uuid.New(),
		PendingID: pendingID,
		OTPType:   otpType,
		OTPValue:  otp,
		Attempts:  0,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(s.otpExpiry),
	}

	// Store OTP in database
	if err := s.otpRepo.StoreOTP(ctx, verificationOTP); err != nil {
		return "", err
	}

	return otp, nil
}

// generateRandomOTP generates a random 6-digit OTP
func (s *OTPService) generateRandomOTP() (string, error) {
	const digits = "0123456789"
	result := make([]byte, s.otpLength)

	_, err := rand.Read(result)
	if err != nil {
		return "", err
	}

	// Convert random bytes to digits
	for i := range result {
		result[i] = digits[int(result[i])%len(digits)]
	}

	return string(result), nil
}

// GenerateAndLogOTPs generates and logs OTPs for both email and phone
func (s *OTPService) GenerateAndLogOTPs(ctx context.Context, pendingID uuid.UUID) error {
	// Generate email OTP
	emailOTP, err := s.GenerateOTP(ctx, pendingID, "email")
	if err != nil {
		return err
	}

	// Generate phone OTP
	phoneOTP, err := s.GenerateOTP(ctx, pendingID, "phone")
	if err != nil {
		return err
	}

	// Log the OTPs (for development purposes only)
	log.Printf("Generated OTPs for pending registration %s: Email OTP: %s, Phone OTP: %s",
		pendingID, emailOTP, phoneOTP)

	return nil
}
