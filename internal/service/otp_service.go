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
	ErrOTPGeneration  = errors.New("failed to generate OTP")
	ErrOTPNotFound    = errors.New("OTP not found or expired")
	ErrOTPMismatch    = errors.New("OTP does not match")
	ErrOTPMaxAttempts = errors.New("maximum verification attempts exceeded")
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

// VerifyOTP verifies an OTP against the stored value
func (s *OTPService) VerifyOTP(ctx context.Context, pendingID uuid.UUID, otpType, otpValue string) error {
	// Get the stored OTP
	otp, err := s.otpRepo.GetOTPByPendingIDAndType(ctx, pendingID, otpType)
	if err != nil {
		return err
	}

	if otp == nil {
		return ErrOTPNotFound
	}

	// Check if max attempts exceeded
	if otp.Attempts >= 3 { // You could make this configurable
		return ErrOTPMaxAttempts
	}

	// Increment attempts
	if err := s.otpRepo.IncrementOTPAttempts(ctx, otp.OTPID); err != nil {
		return err
	}

	// Check if OTP matches
	if otp.OTPValue != otpValue {
		return ErrOTPMismatch
	}

	return nil
}

// MarkVerified marks a verification as complete in the pending registration record
func (s *OTPService) MarkVerified(ctx context.Context, pendingID uuid.UUID, otpType string, regRepo repository.RegistrationRepository) error {
	reg, err := regRepo.GetPendingRegistrationByID(ctx, pendingID)
	if err != nil {
		return err
	}

	if reg == nil {
		return errors.New("pending registration not found")
	}

	field := ""
	if otpType == "email" {
		field = "email_verified"
	} else if otpType == "phone" {
		field = "phone_verified"
	} else {
		return errors.New("invalid OTP type")
	}

	return regRepo.UpdateVerificationStatus(ctx, pendingID, field, true)
}
