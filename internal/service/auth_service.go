package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/domain/models"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository"
	appErrors "github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/errors"
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
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrPhoneNotVerified   = errors.New("phone not verified")
	ErrCreateUser         = errors.New("failed to create user")
	ErrSessionNotFound    = errors.New("session not found or expired")
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo         repository.UserRepository
	registrationRepo repository.RegistrationRepository
	otpService       *OTPService
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repository.UserRepository, registrationRepo repository.RegistrationRepository, otpService *OTPService) *AuthService {
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
		return nil, appErrors.Wrap(err, "failed to check user existence")
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
		return nil, appErrors.Wrap(err, "failed to check pending registration by email")
	}
	if pendingByEmail != nil {
		return nil, ErrPendingEmailExists
	}

	// Check for existing pending registration with same phone
	pendingByPhone, err := s.registrationRepo.GetPendingRegistrationByPhone(ctx, req.Phone)
	if err != nil {
		return nil, appErrors.Wrap(err, "failed to check pending registration by phone")
	}
	if pendingByPhone != nil {
		return nil, ErrPendingPhoneExists
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, appErrors.Wrap(err, "failed to hash password")
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

	if err := s.registrationRepo.CreatePendingRegistration(ctx, pendingReg); err != nil {
		return nil, appErrors.Wrap(err, "failed to create pending registration")
	}

	// Return response
	resp := &models.RegisterResponse{
		PendingID: pendingID,
		Email:     req.Email,
		Phone:     req.Phone,
		ExpiresAt: expiresAt,
	}

	return resp, nil
}

// Add this method to AuthService to expose the registrationRepo
func (s *AuthService) GetRegistrationRepo() repository.RegistrationRepository {
	return s.registrationRepo
}

// CompleteRegistration creates a user from a verified pending registration
func (s *AuthService) CompleteRegistration(ctx context.Context, pendingID uuid.UUID) (*models.CompleteRegistrationResponse, error) {
	// Get the pending registration
	reg, err := s.registrationRepo.GetPendingRegistrationByID(ctx, pendingID)
	if err != nil {
		return nil, ErrDatabaseOperation
	}

	if reg == nil {
		return nil, errors.New("pending registration not found")
	}

	// Check if both email and phone are verified
	if !reg.EmailVerified {
		return nil, ErrEmailNotVerified
	}

	if !reg.PhoneVerified {
		return nil, ErrPhoneNotVerified
	}

	// Create user in database
	user := &models.User{
		UserID:       uuid.New(),
		Email:        reg.Email,
		Phone:        reg.Phone,
		PasswordHash: reg.PasswordHash,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	// Save user to database
	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, ErrCreateUser
	}

	// Return success response
	return &models.CompleteRegistrationResponse{
		UserID:    user.UserID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}
