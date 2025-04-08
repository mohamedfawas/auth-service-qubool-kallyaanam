package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/dto"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/interfaces"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/domain/errors"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/domain/models"
	"golang.org/x/crypto/bcrypt"
)

type RegistrationUseCase struct {
	userRepo interfaces.UserRepository
}

func NewRegistrationUseCase(userRepo interfaces.UserRepository) *RegistrationUseCase {
	return &RegistrationUseCase{
		userRepo: userRepo,
	}
}

func (uc *RegistrationUseCase) Register(ctx context.Context, req *dto.RegistrationRequest) (*dto.RegistrationResponse, error) {
	// Check if email exists in users or pending_registrations
	existsInUsers, err := uc.userRepo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existsInUsers {
		return nil, errors.ErrEmailAlreadyExists
	}

	existsInPending, err := uc.userRepo.FindPendingRegistrationByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existsInPending {
		return nil, errors.ErrEmailAlreadyExists
	}

	// Check if phone exists in users or pending_registrations
	existsInUsers, err = uc.userRepo.FindUserByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}
	if existsInUsers {
		return nil, errors.ErrPhoneAlreadyExists
	}

	existsInPending, err = uc.userRepo.FindPendingRegistrationByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}
	if existsInPending {
		return nil, errors.ErrPhoneAlreadyExists
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Generate OTP (will be implemented in Phase 2)
	otp := "123456" // Placeholder for Phase 1

	// Create pending registration
	now := time.Now().UTC()
	id := uuid.New()

	registration := &models.PendingRegistration{
		ID:        id,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  string(hashedPassword),
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
		OTP:       otp,
		OTPExpiry: now.Add(15 * time.Minute),
	}

	// Save to database
	if err := uc.userRepo.CreatePendingRegistration(ctx, registration); err != nil {
		return nil, err
	}

	return &dto.RegistrationResponse{
		ID:      id.String(),
		Message: "Registration initiated successfully. Please check your email for OTP verification.",
	}, nil
}
