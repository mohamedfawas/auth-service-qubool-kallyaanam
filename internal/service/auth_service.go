// internal/service/auth_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/model"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/model/dto"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/repository"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/repository/postgres"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/repository/redis"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/util/logger"
)

// Implementation of the AuthService interface
type authService struct {
	userRepo        repository.UserRepository
	otpService      OTPService
	emailService    EmailService
	securityService SecurityService
	metricsService  MetricsService
	logger          *logger.Logger
}

// NewAuthService creates a new auth service instance
func NewAuthService(
	userRepo repository.UserRepository,
	otpService OTPService,
	emailService EmailService,
	securityService SecurityService,
	metricsService MetricsService,
	logger *logger.Logger,
) AuthService {
	return &authService{
		userRepo:        userRepo,
		otpService:      otpService,
		emailService:    emailService,
		securityService: securityService,
		metricsService:  metricsService,
		logger:          logger,
	}
}

// Register handles user registration including email verification
func (s *authService) Register(ctx context.Context, req *dto.RegistrationRequest) (*dto.RegistrationResponse, error) {
	// Check if user already exists with the same email
	exists, err := s.userRepo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	// Check if user already exists with the same phone
	exists, err = s.userRepo.FindUserByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("phone already exists")
	}

	// Check if there's already a pending registration
	exists, err = s.userRepo.FindPendingRegistrationByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email verification already in progress")
	}

	exists, err = s.userRepo.FindPendingRegistrationByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("phone verification already in progress")
	}

	// Hash password
	hashedPassword, err := s.securityService.HashPassword(ctx, req.Password)
	if err != nil {
		return nil, err
	}

	// Generate and store OTP for verification using Redis
	// Use email as the key for the OTP
	otp, err := s.otpService.GenerateAndStoreOTP(ctx, req.Email)
	if err != nil {
		// If it's a Redis connectivity issue, log it but continue
		// In a production system, you might want a more robust fallback mechanism
		// or you might decide to fail the registration
		return nil, err
	}

	// Create registration record (without OTP fields)
	var registrationID uuid.UUID
	err = s.userRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Create pending registration
		registration := &model.PendingRegistration{
			ID:        uuid.New(),
			Email:     req.Email,
			Phone:     req.Phone,
			Password:  hashedPassword,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour), // Registration expires in 24 hours
		}

		// Save registration in database
		if err := s.userRepo.CreatePendingRegistration(txCtx, registration); err != nil {
			return err
		}

		registrationID = registration.ID
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Send verification email
	if err := s.emailService.SendVerificationEmail(ctx, req.Email, otp); err != nil {
		// Log this error but don't fail the registration
		// The user can request a new OTP later
		// TODO: implement retry mechanism or queue
	}

	return &dto.RegistrationResponse{
		ID:      registrationID.String(),
		Message: "Registration successful. Please verify your email with the OTP sent.",
	}, nil
}

// internal/service/auth_service.go - Improve VerifyEmail method

// VerifyEmail verifies email using OTP and transfers pending registration to users table
func (s *authService) VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) (*dto.VerifyEmailResponse, error) {
	// Add context with timeout to prevent long-running operations
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Track verification activity
	s.metricsService.IncVerificationAttempt(ctx)

	// 1. Verify the OTP
	isValid, err := s.otpService.VerifyOTP(ctx, req.Email, req.OTP)
	if err != nil {
		// Log the error with appropriate classification
		s.logger.VerificationFailure(req.Email, getClientIP(ctx), "otp_verification_error",
			s.logger.Field("error", err))
		s.metricsService.IncVerificationFailure(ctx, "otp_verification_error")

		// Return appropriate error based on the specific error
		if errors.Is(err, redis.ErrRedisUnavailable) {
			return nil, errors.New("verification service temporarily unavailable")
		}
		return nil, fmt.Errorf("error verifying OTP: %w", err)
	}

	if !isValid {
		s.logger.VerificationFailure(req.Email, getClientIP(ctx), "invalid_otp")
		s.metricsService.IncVerificationFailure(ctx, "invalid_otp")
		return nil, errors.New("invalid OTP")
	}

	// 2. Retrieve the pending registration
	pendingReg, err := s.userRepo.GetPendingRegistrationByEmail(ctx, req.Email)
	if err != nil {
		s.logger.VerificationFailure(req.Email, getClientIP(ctx), "database_error",
			s.logger.Field("error", err))
		s.metricsService.IncVerificationFailure(ctx, "database_error")
		return nil, fmt.Errorf("error retrieving pending registration: %w", err)
	}

	if pendingReg == nil {
		s.logger.VerificationFailure(req.Email, getClientIP(ctx), "no_pending_registration")
		s.metricsService.IncVerificationFailure(ctx, "no_pending_registration")
		return nil, errors.New("no pending registration found")
	}

	// 3. Check if the registration has expired
	if time.Now().After(pendingReg.ExpiresAt) {
		s.logger.VerificationFailure(req.Email, getClientIP(ctx), "registration_expired",
			s.logger.Field("expires_at", pendingReg.ExpiresAt))
		s.metricsService.IncVerificationFailure(ctx, "registration_expired")

		// Clean up expired registration
		go func(id uuid.UUID) {
			// Use background context with timeout
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.userRepo.DeletePendingRegistration(cleanupCtx, id); err != nil {
				s.logger.Warn("Failed to clean up expired registration",
					s.logger.Field("id", id.String()),
					s.logger.Field("error", err))
			}
		}(pendingReg.ID)

		return nil, errors.New("registration has expired")
	}

	// 4. Create a new user and remove pending registration within a transaction
	var userID uuid.UUID
	err = s.userRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Create user from pending registration
		user := &model.User{
			ID:        uuid.New(),
			Email:     pendingReg.Email,
			Phone:     pendingReg.Phone,
			Password:  pendingReg.Password, // Password is already hashed
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			IsActive:  true,
		}

		// Save user in database
		if err := s.userRepo.CreateUser(txCtx, user); err != nil {
			// Check for unique constraint violations
			if postgres.IsUniqueConstraintViolation(err, "users_email_key") {
				return errors.New("email already exists in users table")
			}
			if postgres.IsUniqueConstraintViolation(err, "users_phone_key") {
				return errors.New("phone already exists in users table")
			}
			return fmt.Errorf("failed to create user: %w", err)
		}

		// Delete pending registration
		if err := s.userRepo.DeletePendingRegistration(txCtx, pendingReg.ID); err != nil {
			return fmt.Errorf("failed to delete pending registration: %w", err)
		}

		userID = user.ID
		return nil
	})

	if err != nil {
		// Log transaction errors
		s.logger.VerificationFailure(req.Email, getClientIP(ctx), "transaction_error",
			s.logger.Field("error", err))
		s.metricsService.IncVerificationFailure(ctx, "transaction_error")

		// Return appropriate error but don't expose internal details
		if strings.Contains(err.Error(), "email already exists") ||
			strings.Contains(err.Error(), "phone already exists") {
			return nil, errors.New("user already exists with this email or phone")
		}

		return nil, errors.New("verification failed due to database error")
	}

	// Log successful verification
	s.logger.VerificationSuccess(req.Email, getClientIP(ctx))
	s.metricsService.IncVerificationSuccess(ctx)

	// Record verification duration
	duration := time.Since(getRequestStartTime(ctx)).Seconds()
	s.metricsService.VerificationDuration(ctx, duration)

	return &dto.VerifyEmailResponse{
		ID:      userID.String(),
		Email:   req.Email,
		Message: "Email verification successful",
	}, nil
}

// Helper function to get client IP from context (if available)
func getClientIP(ctx context.Context) string {
	if ip, ok := ctx.Value("client_ip").(string); ok {
		return ip
	}
	return "unknown"
}

// Helper function to get request start time from context
func getRequestStartTime(ctx context.Context) time.Time {
	if startTime, ok := ctx.Value("request_start_time").(time.Time); ok {
		return startTime
	}
	return time.Now() // Fallback
}
