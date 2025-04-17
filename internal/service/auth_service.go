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
	redisService    RedisService
}

// NewAuthService creates a new auth service instance
func NewAuthService(
	userRepo repository.UserRepository,
	otpService OTPService,
	emailService EmailService,
	securityService SecurityService,
	metricsService MetricsService,
	logger *logger.Logger,
	redisService RedisService,
) AuthService {
	return &authService{
		userRepo:        userRepo,
		otpService:      otpService,
		emailService:    emailService,
		securityService: securityService,
		metricsService:  metricsService,
		logger:          logger,
		redisService:    redisService,
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

// internal/service/auth_service.go

// VerifyEmail verifies email using OTP and transfers pending registration to users table
func (s *authService) VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) (*dto.VerifyEmailResponse, error) {
	// Extract client info for logging
	clientIP := getClientIP(ctx)
	startTime := getRequestStartTime(ctx)

	s.logger.Info("Email verification attempt",
		s.logger.Field("email", req.Email),
		s.logger.Field("ip", clientIP),
		s.logger.Field("request_id", ctx.Value("request_id")))

	// 1. Verify the OTP
	isValid, err := s.otpService.VerifyOTP(ctx, req.Email, req.OTP)
	if err != nil {
		if errors.Is(err, redis.ErrOTPNotFound) {
			s.logger.VerificationFailure(req.Email, clientIP, "otp_expired")
			return nil, errors.New("OTP expired or not found")
		}
		s.logger.VerificationFailure(req.Email, clientIP, "otp_verification_error",
			s.logger.Field("error", err.Error()))
		return nil, fmt.Errorf("error verifying OTP")
	}

	if !isValid {
		s.logger.VerificationFailure(req.Email, clientIP, "invalid_otp")
		return nil, errors.New("invalid OTP")
	}

	// 2. Retrieve the pending registration with row locking to prevent race conditions
	pendingReg, err := s.userRepo.GetPendingRegistrationByEmailWithLock(ctx, req.Email)
	if err != nil {
		s.logger.VerificationFailure(req.Email, clientIP, "database_error",
			s.logger.Field("error", err.Error()))
		return nil, fmt.Errorf("error retrieving pending registration")
	}

	if pendingReg == nil {
		s.logger.VerificationFailure(req.Email, clientIP, "no_pending_registration")
		return nil, errors.New("no pending registration found")
	}

	// 3. Check if the registration has expired
	if time.Now().After(pendingReg.ExpiresAt) {
		s.logger.VerificationFailure(req.Email, clientIP, "registration_expired")
		return nil, errors.New("registration has expired")
	}

	// 4. Create a new user and remove pending registration within a transaction
	var userID uuid.UUID

	// Wrap transaction with detailed error handling and logging
	err = s.userRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Create user from pending registration
		user := &model.User{
			ID:           uuid.New(),
			Email:        pendingReg.Email,
			Phone:        pendingReg.Phone,
			PasswordHash: pendingReg.Password, // Password is already hashed
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			IsActive:     true,
			IsVerified:   true,
			Role:         "user",
			LastLoginAt:  time.Now(),
		}

		// Save user in database
		if err := s.userRepo.CreateUser(txCtx, user); err != nil {
			// Log internal details but don't expose them
			s.logger.VerificationFailure(req.Email, clientIP, "user_creation_failed",
				s.logger.Field("error", err.Error()))

			if errors.Is(err, postgres.ErrDuplicateKey) {
				return errors.New("account already exists")
			}

			return fmt.Errorf("failed to create user") // Generalized user-facing error
		}

		// Delete pending registration
		if err := s.userRepo.DeletePendingRegistration(txCtx, pendingReg.ID); err != nil {
			s.logger.VerificationFailure(req.Email, clientIP, "pending_registration_deletion_failed",
				s.logger.Field("error", err.Error()))
			return fmt.Errorf("failed to complete verification process")
		}

		userID = user.ID
		return nil
	})

	if err != nil {
		// Don't expose internal errors to the client
		if strings.Contains(err.Error(), "account already exists") {
			return nil, errors.New("account already exists")
		}
		return nil, errors.New("verification failed")
	}

	// Calculate and log processing time
	processingTime := time.Since(startTime).Seconds()
	s.logger.VerificationSuccess(req.Email, clientIP)
	// Log additional data separately
	s.logger.Info("Verification processing completed",
		s.logger.Field("duration_seconds", processingTime),
		s.logger.Field("user_id", userID.String()))

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

// Add to auth_service.go

// Login authenticates a user and generates JWT tokens
func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// Extract client info
	clientIP, _ := ctx.Value("client_ip").(string)
	userAgent, _ := ctx.Value("user_agent").(string)

	// Check if login attempts are throttled for this IP
	isThrottled, err := s.redisService.IsLoginThrottled(ctx, clientIP)
	if err != nil {
		s.logger.Error("Error checking login throttling",
			s.logger.Field("client_ip", clientIP),
			s.logger.Field("error", err.Error()))
		// Continue processing in case of error
	}

	if isThrottled {
		return nil, errors.New("too many login attempts")
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Log error but return generic message to user
		s.logger.Error("Error finding user during login",
			s.logger.Field("email", req.Email),
			s.logger.Field("error", err.Error()))
		return nil, errors.New("invalid credentials")
	}

	// Check if user exists
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Check if email is verified
	if !user.IsVerified {
		return nil, errors.New("email not verified")
	}

	// Verify password
	if !s.securityService.VerifyPassword(ctx, user.PasswordHash, req.Password) {
		// Add delay to prevent timing attacks
		time.Sleep(300 * time.Millisecond)
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	accessToken, err := s.securityService.GenerateJWT(ctx, user.ID.String(), user.Role, user.LastLoginAt)
	if err != nil {
		s.logger.Error("Error generating JWT token",
			s.logger.Field("user_id", user.ID.String()),
			s.logger.Field("error", err.Error()))
		return nil, errors.New("failed to generate access token")
	}

	// Generate refresh token
	refreshToken, tokenID, err := s.securityService.GenerateRefreshToken(ctx, user.ID.String())
	if err != nil {
		s.logger.Error("Error generating refresh token",
			s.logger.Field("user_id", user.ID.String()),
			s.logger.Field("error", err.Error()))
		return nil, errors.New("failed to generate refresh token")
	}
	// Store refresh token in Redis
	tokenData := TokenData{
		UserID:    user.ID.String(),
		TokenID:   tokenID,
		UserRole:  user.Role,
		IssuedAt:  time.Now(),
		UserAgent: userAgent,
		ClientIP:  clientIP,
	}

	if err := s.redisService.StoreRefreshToken(ctx, tokenID, refreshToken, tokenData); err != nil {
		s.logger.Error("Error storing refresh token",
			s.logger.Field("user_id", user.ID.String()),
			s.logger.Field("error", err.Error()))
		return nil, errors.New("failed to store refresh token")
	}

	// Store login history
	if err := s.redisService.StoreLoginHistory(ctx, user.ID.String(), userAgent, clientIP); err != nil {
		// Log but don't fail the login
		s.logger.Warn("Failed to store login history",
			s.logger.Field("user_id", user.ID.String()),
			s.logger.Field("error", err.Error()))
	}

	// Update last login time
	user.LastLoginAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		// Log error but continue - this shouldn't block login
		s.logger.Warn("Failed to update last login time",
			s.logger.Field("user_id", user.ID.String()),
			s.logger.Field("error", err.Error()))
	}

	// Return response
	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserRole:     user.Role,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
	// Extract token ID
	tokenID, err := s.securityService.ExtractTokenID(ctx, req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Get token data from Redis
	tokenData, err := s.redisService.GetRefreshTokenData(ctx, tokenID)
	if err != nil {
		s.logger.Error("Error retrieving refresh token data",
			s.logger.Field("token_id", tokenID),
			s.logger.Field("error", err.Error()))
		return nil, errors.New("failed to validate refresh token")
	}

	// Check if token exists
	if tokenData == nil {
		return nil, errors.New("refresh token not found or expired")
	}

	// Verify token hasn't been blacklisted
	isBlacklisted, err := s.redisService.IsTokenBlacklisted(ctx, tokenID)
	if err != nil {
		s.logger.Error("Error checking token blacklist",
			s.logger.Field("token_id", tokenID),
			s.logger.Field("error", err.Error()))
		return nil, errors.New("failed to validate refresh token")
	}

	if isBlacklisted {
		return nil, errors.New("refresh token has been revoked")
	}

	// Delete the used refresh token (token rotation)
	if err := s.redisService.DeleteRefreshToken(ctx, tokenID); err != nil {
		s.logger.Error("Error deleting used refresh token",
			s.logger.Field("token_id", tokenID),
			s.logger.Field("error", err.Error()))
		// Continue despite error
	}

	// Get user to ensure they still exist and are active
	userID := tokenData.UserID
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		s.logger.Error("Error finding user during token refresh",
			s.logger.Field("user_id", userID),
			s.logger.Field("error", err.Error()))
		return nil, errors.New("failed to validate user")
	}

	if user == nil || !user.IsActive {
		return nil, errors.New("user not found or inactive")
	}

	// Generate new access token
	accessToken, err := s.securityService.GenerateJWT(ctx, userID, tokenData.UserRole, user.LastLoginAt)
	if err != nil {
		s.logger.Error("Error generating access token",
			s.logger.Field("user_id", userID),
			s.logger.Field("error", err.Error()))
		return nil, errors.New("failed to generate access token")
	}

	// Generate new refresh token (rotation)
	newRefreshToken, newTokenID, err := s.securityService.GenerateRefreshToken(ctx, userID)
	if err != nil {
		s.logger.Error("Error generating new refresh token",
			s.logger.Field("user_id", userID),
			s.logger.Field("error", err.Error()))
		return nil, errors.New("failed to generate refresh token")
	}

	// Store new refresh token
	newTokenData := TokenData{
		UserID:    userID,
		TokenID:   newTokenID,
		UserRole:  tokenData.UserRole,
		IssuedAt:  time.Now(),
		UserAgent: tokenData.UserAgent,
		ClientIP:  tokenData.ClientIP,
	}

	if err := s.redisService.StoreRefreshToken(ctx, newTokenID, newRefreshToken, newTokenData); err != nil {
		s.logger.Error("Error storing new refresh token",
			s.logger.Field("user_id", userID),
			s.logger.Field("error", err.Error()))
		return nil, errors.New("failed to store refresh token")
	}

	return &dto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Add Logout method
func (s *authService) Logout(ctx context.Context, req *dto.LogoutRequest) error {
	// Extract token ID from access token
	accessTokenID, err := s.securityService.ExtractTokenID(ctx, req.AccessToken)
	if err != nil {
		return errors.New("invalid access token")
	}

	// Extract token ID from refresh token
	refreshTokenID, err := s.securityService.ExtractTokenID(ctx, req.RefreshToken)
	if err != nil {
		return errors.New("invalid refresh token")
	}

	// Delete refresh token
	if err := s.redisService.DeleteRefreshToken(ctx, refreshTokenID); err != nil {
		s.logger.Error("Error deleting refresh token during logout",
			s.logger.Field("token_id", refreshTokenID),
			s.logger.Field("error", err.Error()))
		// Continue despite error
	}

	// Blacklist access token until it expires
	accessTokenClaims, err := s.securityService.ValidateJWT(ctx, req.AccessToken)
	if err == nil { // Only blacklist if token is valid
		// Extract expiry time
		expUnix, ok := accessTokenClaims["exp"].(float64)
		if ok {
			exp := time.Unix(int64(expUnix), 0)
			ttl := time.Until(exp)

			// Blacklist the token
			if err := s.redisService.BlacklistToken(ctx, accessTokenID, ttl); err != nil {
				s.logger.Error("Error blacklisting access token",
					s.logger.Field("token_id", accessTokenID),
					s.logger.Field("error", err.Error()))
				// Continue despite error
			}
		}
	}

	return nil
}
