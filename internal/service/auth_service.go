// internal/service/auth_service.go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/model"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/repository"
)

type AuthService struct {
	userRepo        repository.UserRepository
	otpService      OTPService
	emailService    EmailService
	securityService SecurityService
}

func NewAuthService(
	userRepo repository.UserRepository,
	otpService OTPService,
	emailService EmailService,
	securityService SecurityService,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		otpService:      otpService,
		emailService:    emailService,
		securityService: securityService,
	}
}

func (s *AuthService) Register(ctx context.Context, req *model.RegistrationRequest) (*model.RegistrationResponse, error) {
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

	// Generate OTP for verification
	otp, err := s.otpService.GenerateOTP(ctx, 6)
	if err != nil {
		return nil, err
	}

	// Create registration record
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
			OTP:       otp,
			OTPExpiry: s.otpService.GetOTPExpiryTime(ctx),
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

	return &model.RegistrationResponse{
		ID:      registrationID.String(),
		Message: "Registration successful. Please verify your email with the OTP sent.",
	}, nil
}
