package service

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/domain/models"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// Error definitions
var (
	ErrEmailExists        = errors.New("email already registered")
	ErrPhoneExists        = errors.New("phone already registered")
	ErrPendingEmailExists = errors.New("email already has a pending registration")
	ErrPendingPhoneExists = errors.New("phone already has a pending registration")
	ErrPasswordEncryption = errors.New("failed to encrypt password")
	ErrDatabaseOperation  = errors.New("database operation failed")
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo         repository.UserRepository
	registrationRepo repository.RegistrationRepository
	otpService       *OTPService
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo repository.UserRepository,
	registrationRepo repository.RegistrationRepository,
	otpService *OTPService,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		registrationRepo: registrationRepo,
		otpService:       otpService,
	}
}

// Register handles the user registration process
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error) {
	// Normalize inputs
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)

	// Check if user already exists in users table
	exists, field, err := s.userRepo.CheckUserExists(ctx, req.Email, req.Phone)
	if err != nil {
		return nil, ErrDatabaseOperation
	}
	if exists {
		if field == "email" {
			return nil, ErrEmailExists
		}
		return nil, ErrPhoneExists
	}

	// Check for existing pending registration with same email
	pendingByEmail, err := s.registrationRepo.GetPendingRegistrationByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrDatabaseOperation
	}
	if pendingByEmail != nil {
		return nil, ErrPendingEmailExists
	}

	// Check for existing pending registration with same phone
	pendingByPhone, err := s.registrationRepo.GetPendingRegistrationByPhone(ctx, req.Phone)
	if err != nil {
		return nil, ErrDatabaseOperation
	}
	if pendingByPhone != nil {
		return nil, ErrPendingPhoneExists
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrPasswordEncryption
	}

	// Create pending registration
	pendingID := uuid.New()
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	pendingReg := &models.PendingRegistration{
		PendingID:     pendingID,
		Email:         req.Email,
		Phone:         req.Phone,
		PasswordHash:  passwordHash,
		EmailVerified: false,
		PhoneVerified: false,
		CreatedAt:     time.Now().UTC(),
		ExpiresAt:     expiresAt,
	}

	// Create the registration
	if err := s.registrationRepo.CreatePendingRegistration(ctx, pendingReg); err != nil {
		return nil, ErrDatabaseOperation
	}

	// Generate OTPs
	emailOTP, phoneOTP, err := s.otpService.GenerateOTPsForRegistration(ctx, pendingID)
	if err != nil {
		return nil, err
	}

	// In development mode, log the OTPs instead of sending them
	log.Printf("OTPs for %s: Email OTP: %s, Phone OTP: %s",
		req.Email, emailOTP, phoneOTP)
	// In production, you would send these via email and SMS
	// This will be handled by the handler

	// Return response
	resp := &models.RegisterResponse{
		PendingID: pendingID,
		Email:     req.Email,
		Phone:     req.Phone,
		ExpiresAt: expiresAt,
	}

	return resp, nil
}
