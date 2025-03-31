package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/config"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/domain/dto"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/domain/entity"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/errors"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/repository"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/auth"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/logging"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/validator"
	"go.uber.org/zap"
)

// registrationService implements the RegistrationService interface.
type registrationService struct {
	config              *config.Config
	pendingRegRepo      repository.PendingRegistrationRepository
	userRepo            repository.UserRepository
	notificationService NotificationService
	sessionService      SessionService
	logger              *zap.Logger
}

// NewRegistrationService creates a new registration service.
func NewRegistrationService(
	config *config.Config,
	pendingRegRepo repository.PendingRegistrationRepository,
	userRepo repository.UserRepository,
	notificationService NotificationService,
	sessionService SessionService,
) RegistrationService {
	return &registrationService{
		config:              config,
		pendingRegRepo:      pendingRegRepo,
		userRepo:            userRepo,
		notificationService: notificationService,
		sessionService:      sessionService,
		logger:              logging.Logger(),
	}
}

// Register implements the registration process.
func (s *registrationService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, string, error) {
	// Validate request data
	if err := s.validateRegistrationData(ctx, req); err != nil {
		return nil, "", err
	}

	// Hash the password
	passwordHash, err := auth.HashPassword(req.Password, s.config.Auth.BcryptCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, "", errors.InternalError(err)
	}

	// Generate verification codes
	emailCode, err := auth.GenerateRandomCode(s.config.Validation.EmailCodeLength)
	if err != nil {
		s.logger.Error("Failed to generate email verification code", zap.Error(err))
		return nil, "", errors.InternalError(err)
	}

	phoneCode, err := auth.GenerateRandomCode(s.config.Validation.PhoneCodeLength)
	if err != nil {
		s.logger.Error("Failed to generate phone verification code", zap.Error(err))
		return nil, "", errors.InternalError(err)
	}

	// Create pending registration
	pendingRegistration := &entity.PendingRegistration{
		ID:                    uuid.New(),
		Email:                 req.Email,
		Phone:                 req.Phone,
		PasswordHash:          passwordHash,
		EmailVerificationCode: emailCode,
		PhoneVerificationCode: phoneCode,
		EmailVerified:         false,
		PhoneVerified:         false,
		ExpiresAt:             time.Now().Add(s.config.Auth.RegistrationExpiry),
	}

	// Save pending registration
	if err := s.pendingRegRepo.Create(ctx, pendingRegistration); err != nil {
		return nil, "", errors.InternalError(err)
	}

	// Send verification codes
	go s.sendVerificationCodes(req.Email, req.Phone, emailCode, phoneCode)

	// Create session for registration
	sessionData := map[string]interface{}{
		"registration_id": pendingRegistration.ID.String(),
		"email":           req.Email,
		"phone":           req.Phone,
		"created_at":      time.Now().Unix(),
	}

	// Store session data
	sessionID, err := s.sessionService.CreateRegistrationSession(ctx, sessionData)
	if err != nil {
		s.logger.Error("Failed to create registration session", zap.Error(err))
		return nil, "", errors.InternalError(err)
	}

	// Return response and sessionID
	return &dto.RegisterResponse{
		Email: req.Email,
		Phone: req.Phone,
	}, sessionID, nil
}

// validateRegistrationData validates the registration request data.
func (s *registrationService) validateRegistrationData(ctx context.Context, req dto.RegisterRequest) error {
	// Validate email format
	if err := validator.ValidateEmail(req.Email); err != nil {
		return err
	}

	// Validate phone format
	if err := validator.ValidatePhone(req.Phone); err != nil {
		return err
	}

	// Validate password strength
	if err := validator.ValidatePassword(req.Password); err != nil {
		return err
	}

	// Check if email already exists in users
	emailExists, err := s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		s.logger.Error("Error checking email existence in users",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return errors.InternalError(err)
	}

	if emailExists {
		s.logger.Info("Registration attempt with existing email", zap.String("email", req.Email))
		return errors.DuplicateError("Email is already registered")
	}

	// Check if phone already exists in users
	phoneExists, err := s.userRepo.PhoneExists(ctx, req.Phone)
	if err != nil {
		s.logger.Error("Error checking phone existence in users",
			zap.String("phone", req.Phone),
			zap.Error(err),
		)
		return errors.InternalError(err)
	}

	if phoneExists {
		s.logger.Info("Registration attempt with existing phone", zap.String("phone", req.Phone))
		return errors.DuplicateError("Phone number is already registered")
	}

	// Check if email already exists in pending registrations
	emailExistsPending, err := s.pendingRegRepo.EmailExists(ctx, req.Email)
	if err != nil {
		s.logger.Error("Error checking email existence in pending registrations",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return errors.InternalError(err)
	}

	if emailExistsPending {
		s.logger.Info("Registration attempt with email in pending state", zap.String("email", req.Email))
		return errors.DuplicateError("Email is already pending verification")
	}

	// Check if phone already exists in pending registrations
	phoneExistsPending, err := s.pendingRegRepo.PhoneExists(ctx, req.Phone)
	if err != nil {
		s.logger.Error("Error checking phone existence in pending registrations",
			zap.String("phone", req.Phone),
			zap.Error(err),
		)
		return errors.InternalError(err)
	}

	if phoneExistsPending {
		s.logger.Info("Registration attempt with phone in pending state", zap.String("phone", req.Phone))
		return errors.DuplicateError("Phone number is already pending verification")
	}

	return nil
}

// sendVerificationCodes sends verification codes to email and phone.
func (s *registrationService) sendVerificationCodes(email, phone, emailCode, phoneCode string) {
	ctx := context.Background()

	// Send email verification code
	if err := s.notificationService.SendEmailVerification(ctx, email, emailCode); err != nil {
		s.logger.Error("Failed to send email verification code",
			zap.String("email", email),
			zap.Error(err),
		)
	}

	// Send SMS verification code
	if err := s.notificationService.SendSMSVerification(ctx, phone, phoneCode); err != nil {
		s.logger.Error("Failed to send SMS verification code",
			zap.String("phone", phone),
			zap.Error(err),
		)
	}
}
