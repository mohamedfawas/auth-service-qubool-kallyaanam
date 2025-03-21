package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/models"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository/postgres"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository/redis"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/notification"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/utils"
	"gorm.io/gorm"
)

// AuthService handles business logic for user authentication
type AuthService struct {
	userRepo        postgres.UserRepository
	otpRepo         redis.OTPRepository
	notificationSvc *notification.Service
	otpConfig       utils.OTPConfig
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(
	userRepo postgres.UserRepository,
	otpRepo redis.OTPRepository,
	notificationSvc *notification.Service,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		otpRepo:         otpRepo,
		notificationSvc: notificationSvc,
		otpConfig:       utils.DefaultOTPConfig(),
	}
}

// RegisterRequest represents the data needed for user registration
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// RegisterResponse represents the response after a successful registration request
type RegisterResponse struct {
	PendingID string    `json:"pending_id"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	ExpiresAt time.Time `json:"expires_at"`
	Message   string    `json:"message"`
}

// Register handles the user registration process
func (s *AuthService) Register(req *RegisterRequest) (*RegisterResponse, error) {
	// Check if email already exists in users table
	emailExists, err := s.userRepo.CheckEmailExists(req.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking email: %w", err)
	}
	if emailExists {
		return nil, errors.New("email already registered")
	}

	// Check if phone already exists in users table
	phoneExists, err := s.userRepo.CheckPhoneExists(req.Phone)
	if err != nil {
		return nil, fmt.Errorf("error checking phone: %w", err)
	}
	if phoneExists {
		return nil, errors.New("phone already registered")
	}

	// Check if email already exists in pending_registrations table
	existingByEmail, err := s.userRepo.FindPendingRegistrationByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error checking pending email: %w", err)
	}
	if existingByEmail != nil {
		return nil, errors.New("email verification already pending")
	}

	// Check if phone already exists in pending_registrations table
	existingByPhone, err := s.userRepo.FindPendingRegistrationByPhone(req.Phone)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error checking pending phone: %w", err)
	}
	if existingByPhone != nil {
		return nil, errors.New("phone verification already pending")
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Create pending registration
	expiryTime := utils.CalculateExpiryTime(s.otpConfig.ExpiryMinutes)
	pendingReg := &models.PendingRegistration{
		Email:         req.Email,
		Phone:         req.Phone,
		PasswordHash:  passwordHash,
		EmailVerified: false,
		PhoneVerified: false,
		ExpiresAt:     expiryTime,
	}

	if err := s.userRepo.CreatePendingRegistration(pendingReg); err != nil {
		return nil, fmt.Errorf("error creating pending registration: %w", err)
	}

	// Generate and store email OTP
	emailOTP, err := utils.GenerateOTP(s.otpConfig.Length)
	if err != nil {
		return nil, fmt.Errorf("error generating email OTP: %w", err)
	}

	if err := s.otpRepo.StoreEmailOTP(
		pendingReg.PendingID.String(),
		req.Email,
		emailOTP,
		time.Duration(s.otpConfig.ExpiryMinutes)*time.Hour,
	); err != nil {
		return nil, fmt.Errorf("error storing email OTP: %w", err)
	}

	fmt.Printf("Email OTP for %s: %s\n", req.Email, emailOTP) // remove this later, INcluded this for development and testing
	// Send email OTP
	if err := s.notificationSvc.SendEmailOTP(req.Email, emailOTP); err != nil {
		// Log the error but don't fail the registration
		fmt.Printf("Failed to send email OTP: %v\n", err)
	}

	// Generate and store phone OTP
	phoneOTP, err := utils.GenerateOTP(s.otpConfig.Length)
	if err != nil {
		return nil, fmt.Errorf("error generating phone OTP: %w", err)
	}

	if err := s.otpRepo.StorePhoneOTP(
		pendingReg.PendingID.String(),
		req.Phone,
		phoneOTP,
		time.Duration(s.otpConfig.ExpiryMinutes)*time.Hour,
	); err != nil {
		return nil, fmt.Errorf("error storing phone OTP: %w", err)
	}

	// Send phone OTP
	if err := s.notificationSvc.SendSMSOTP(req.Phone, phoneOTP); err != nil {
		// Log the error but don't fail the registration
		fmt.Printf("Failed to send SMS OTP: %v\n", err)
	}

	fmt.Printf("Phone OTP for %s: %s\n", req.Phone, phoneOTP) // Remove this later, needed for dev and testing only

	// Return response
	return &RegisterResponse{
		PendingID: pendingReg.PendingID.String(),
		Email:     pendingReg.Email,
		Phone:     pendingReg.Phone,
		ExpiresAt: pendingReg.ExpiresAt,
		Message:   "Registration initiated. Please verify your email and phone using the OTPs sent.",
	}, nil
}
