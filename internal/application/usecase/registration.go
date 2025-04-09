package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/dto"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/interfaces"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/domain/errors"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/domain/models"
)

type RegistrationUseCase struct {
	userRepo        interfaces.UserRepository
	otpService      interfaces.OTPService
	emailService    interfaces.EmailService
	securityService interfaces.SecurityService
}

func NewRegistrationUseCase(
	userRepo interfaces.UserRepository,
	otpService interfaces.OTPService,
	emailService interfaces.EmailService,
	securityService interfaces.SecurityService,
) *RegistrationUseCase {
	return &RegistrationUseCase{
		userRepo:        userRepo,
		otpService:      otpService,
		emailService:    emailService,
		securityService: securityService,
	}
}

func (uc *RegistrationUseCase) Register(ctx context.Context, req *dto.RegistrationRequest) (*dto.RegistrationResponse, error) {
	// Check if email exists in users or pending_registrations
	existsInUsers, err := uc.userRepo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking email existence in users: %w", err)
	}
	if existsInUsers {
		return nil, errors.ErrEmailAlreadyExists
	}

	existsInPending, err := uc.userRepo.FindPendingRegistrationByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking email existence in pending registrations: %w", err)
	}
	if existsInPending {
		return nil, errors.ErrEmailAlreadyExists
	}

	// Check if phone exists in users or pending_registrations
	existsInUsers, err = uc.userRepo.FindUserByPhone(ctx, req.Phone)
	if err != nil {
		return nil, fmt.Errorf("error checking phone existence in users: %w", err)
	}
	if existsInUsers {
		return nil, errors.ErrPhoneAlreadyExists
	}

	existsInPending, err = uc.userRepo.FindPendingRegistrationByPhone(ctx, req.Phone)
	if err != nil {
		return nil, fmt.Errorf("error checking phone existence in pending registrations: %w", err)
	}
	if existsInPending {
		return nil, errors.ErrPhoneAlreadyExists
	}

	// Use a transaction for the remaining operations
	var registrationID uuid.UUID

	err = uc.userRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Hash the password using the security service
		hashedPassword, err := uc.securityService.HashPassword(txCtx, req.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// Generate OTP
		otp, err := uc.otpService.GenerateOTP(txCtx, 6)
		if err != nil {
			return fmt.Errorf("failed to generate OTP: %w", err)
		}

		// Get current time and calculate expiry times
		now := time.Now()
		otpExpiry := uc.otpService.GetOTPExpiryTime(txCtx)
		registrationExpiry := now.Add(24 * time.Hour)

		// Create pending registration
		registrationID = uuid.New()
		registration := &models.PendingRegistration{
			ID:        registrationID,
			Email:     req.Email,
			Phone:     req.Phone,
			Password:  hashedPassword,
			CreatedAt: now,
			UpdatedAt: now,
			ExpiresAt: registrationExpiry,
			OTP:       otp,
			OTPExpiry: otpExpiry,
		}

		// Save to database
		if err := uc.userRepo.CreatePendingRegistration(txCtx, registration); err != nil {
			return fmt.Errorf("failed to create pending registration: %w", err)
		}

		// Send verification email
		if err := uc.emailService.SendVerificationEmail(txCtx, req.Email, otp); err != nil {
			return fmt.Errorf("failed to send verification email: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	return &dto.RegistrationResponse{
		ID:      registrationID.String(),
		Message: "Registration initiated successfully. Please check your email for OTP verification.",
	}, nil
}
