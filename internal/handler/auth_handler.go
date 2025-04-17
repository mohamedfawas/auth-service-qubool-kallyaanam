// auth-service-qubool-kallyaanam/internal/handler/auth_handler.go
package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/model/dto"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/service"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/util/logger"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/util/response"
)

type AuthHandler struct {
	authService     service.AuthService // Changed from *service.AuthService to service.AuthService
	securityService service.SecurityService
	metricsService  service.MetricsService
	logger          *logger.Logger
}

func NewAuthHandler(
	authService service.AuthService, // Changed from *service.AuthService to service.AuthService
	securityService service.SecurityService,
	metricsService service.MetricsService,
	logger *logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		securityService: securityService,
		metricsService:  metricsService,
		logger:          logger,
	}
}

func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {

	router.POST("/register", h.Register)
	router.POST("/verify-email", h.VerifyEmail)
	router.POST("/login", h.Login)
	router.POST("/refresh-token", h.RefreshToken)
	router.POST("/logout", h.Logout)
}

func (h *AuthHandler) Register(c *gin.Context) {
	start := time.Now().UTC()

	// Extract client info for logging
	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Start metrics tracking
	h.metricsService.IncRegistrationAttempt(c)

	// Log registration attempt
	h.logger.RegistrationAttempt("", clientIP, userAgent)

	// Parse and validate request
	var request dto.RegistrationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.metricsService.IncRegistrationFailure(c, "invalid_request")
		h.logger.RegistrationFailure("", clientIP, "invalid_request", zap.Error(err))
		response.BadRequest(c, "Invalid request", err)
		return
	}

	// Sanitize inputs
	request.Email = h.securityService.SanitizeInput(c, request.Email)
	request.Phone = h.securityService.SanitizeInput(c, request.Phone)

	// Validate password strength
	if isValid, reason := h.securityService.ValidatePassword(c, request.Password); !isValid {
		h.metricsService.IncRegistrationFailure(c, "weak_password")
		h.logger.RegistrationFailure(request.Email, clientIP, reason)
		response.BadRequest(c, "Password validation failed", &ValidationError{Reason: reason})
		return
	}

	// Register the user
	resp, err := h.authService.Register(c, &request)

	// Record metrics for the duration
	duration := time.Since(start).Seconds()
	h.metricsService.RegistrationDuration(c, duration)

	// Handle response based on result
	if err != nil {
		var statusCode int
		var errorType string
		var errMsg string

		// Determine appropriate status code and error type based on error
		switch err.Error() {
		case "email already exists", "phone already exists":
			statusCode = http.StatusConflict
			errorType = "duplicate_user"
			errMsg = "User with this email or phone already exists"
		default:
			statusCode = http.StatusInternalServerError
			errorType = "server_error"
			errMsg = "Failed to process registration"
		}

		h.metricsService.IncRegistrationFailure(c, errorType)
		h.logger.RegistrationFailure(request.Email, clientIP, err.Error())

		response.Error(c, statusCode, errMsg, err)
		return
	}

	// Log successful registration
	h.metricsService.IncRegistrationSuccess(c)

	// Return standardized response
	response.Created(c, "Registration successful. Please verify your email with the OTP sent.", gin.H{
		"id": resp.ID,
	})
}

// ValidationError represents a password validation error
type ValidationError struct {
	Reason string
}

func (e *ValidationError) Error() string {
	return e.Reason
}

// internal/handler/auth_handler.go

// internal/handler/auth_handler.go - Update VerifyEmail handler

// internal/handler/auth_handler.go

// VerifyEmail handles the email verification endpoint
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	// Add request ID for tracing
	requestID := uuid.New().String()
	ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
	c.Request = c.Request.WithContext(ctx)

	start := time.Now().UTC()

	// Extract client info for logging
	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Start metrics tracking
	h.metricsService.IncVerificationAttempt(c)

	// Log verification attempt
	h.logger.VerificationAttempt("", clientIP, userAgent)
	h.logger.Info("Request ID for verification attempt",
		h.logger.Field("request_id", requestID))

	// Parse and validate request
	var request dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.metricsService.IncVerificationFailure(c, "invalid_request")
		h.logger.VerificationFailure("", clientIP, "invalid_request",
			h.logger.Field("error", err.Error()),
			h.logger.Field("request_id", requestID))
		response.BadRequest(c, "Invalid request", nil) // Don't expose validation errors
		return
	}

	// Sanitize inputs to prevent attacks
	request.Email = h.securityService.SanitizeInput(c, request.Email)
	request.OTP = h.securityService.SanitizeInput(c, request.OTP)

	// Verify the email
	verifyResp, err := h.authService.VerifyEmail(c, &request)

	// Record metrics for the duration
	duration := time.Since(start).Seconds()
	h.metricsService.VerificationDuration(c, duration)

	// Handle response based on result
	if err != nil {
		var statusCode int
		var errorType string
		var errorMsg string

		// Map internal errors to user-friendly messages
		// without exposing sensitive details
		switch {
		case strings.Contains(err.Error(), "no pending registration") ||
			strings.Contains(err.Error(), "OTP expired") ||
			strings.Contains(err.Error(), "not found"):
			statusCode = http.StatusNotFound
			errorType = "verification_not_found"
			errorMsg = "Verification request not found or expired"
		case strings.Contains(err.Error(), "invalid OTP"):
			statusCode = http.StatusBadRequest
			errorType = "invalid_otp"
			errorMsg = "Invalid verification code"
		case strings.Contains(err.Error(), "registration has expired"):
			statusCode = http.StatusGone
			errorType = "registration_expired"
			errorMsg = "Registration has expired, please register again"
		case strings.Contains(err.Error(), "account already exists"):
			statusCode = http.StatusConflict
			errorType = "already_verified"
			errorMsg = "This email is already verified"
		default:
			statusCode = http.StatusInternalServerError
			errorType = "server_error"
			errorMsg = "Failed to verify email"
		}

		h.metricsService.IncVerificationFailure(c, errorType)
		h.logger.VerificationFailure(request.Email, clientIP, errorType,
			h.logger.Field("status_code", statusCode),
			h.logger.Field("request_id", requestID),
			h.logger.Field("duration_seconds", duration))

		// Send generalized error response - never expose internal error details
		response.Error(c, statusCode, errorMsg, nil)
		return
	}

	// Log successful verification
	h.metricsService.IncVerificationSuccess(c)
	h.logger.VerificationSuccess(request.Email, clientIP)
	// Log additional details separately
	h.logger.Info("Verification succeeded with details",
		h.logger.Field("user_id", verifyResp.ID),
		h.logger.Field("request_id", requestID),
		h.logger.Field("duration_seconds", duration))

	// Return standardized success response
	response.Success(c, "Email verification successful", gin.H{
		"id":    verifyResp.ID,
		"email": verifyResp.Email,
	})
}

// ResponseError implements error interface for secure error responses
type ResponseError struct {
	Message string
}

func (e *ResponseError) Error() string {
	return e.Message
}

// Add to auth_handler.go

// Login handles the user login endpoint
func (h *AuthHandler) Login(c *gin.Context) {
	start := time.Now().UTC()

	// Extract client info for logging
	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()
	requestID := uuid.New().String()

	// Add request ID to context for tracing
	ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
	c.Request = c.Request.WithContext(ctx)

	// Start metrics tracking
	h.metricsService.IncLoginAttempt(ctx)

	// Log login attempt (without credentials)
	h.logger.Info("Login attempt",
		h.logger.Field("client_ip", clientIP),
		h.logger.Field("user_agent", userAgent),
		h.logger.Field("request_id", requestID))

	// Parse and validate request
	var request dto.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.metricsService.IncLoginFailure(ctx, "invalid_request")
		h.logger.Warn("Login failure: invalid request",
			h.logger.Field("error", err.Error()),
			h.logger.Field("client_ip", clientIP),
			h.logger.Field("request_id", requestID))
		response.BadRequest(c, "Invalid request format", nil)
		return
	}

	// Sanitize inputs
	request.Email = h.securityService.SanitizeInput(ctx, request.Email)

	// Perform login
	loginResp, err := h.authService.Login(ctx, &request)

	// Record metrics for the duration
	duration := time.Since(start).Seconds()
	h.metricsService.LoginDuration(ctx, duration)

	// Handle response based on result
	if err != nil {
		var statusCode int
		var errorType string
		var errorMsg string

		// Map internal errors to user-friendly messages
		switch {
		case strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "invalid credentials"):
			statusCode = http.StatusUnauthorized
			errorType = "invalid_credentials"
			errorMsg = "Invalid email or password"
		case strings.Contains(err.Error(), "not verified"):
			statusCode = http.StatusForbidden
			errorType = "unverified_account"
			errorMsg = "Email not verified. Please verify your email first"
		default:
			statusCode = http.StatusInternalServerError
			errorType = "server_error"
			errorMsg = "Authentication failed"
		}

		h.metricsService.IncLoginFailure(ctx, errorType)
		h.logger.Warn("Login failure",
			h.logger.Field("error_type", errorType),
			h.logger.Field("email", request.Email),
			h.logger.Field("client_ip", clientIP),
			h.logger.Field("request_id", requestID))

		response.Error(c, statusCode, errorMsg, nil)
		return
	}

	// Log successful login
	h.metricsService.IncLoginSuccess(ctx)
	h.logger.Info("Login successful",
		h.logger.Field("email", request.Email),
		h.logger.Field("client_ip", clientIP),
		h.logger.Field("request_id", requestID))

	// Return standardized response
	response.Success(c, "Login successful", loginResp)
}

// Add these methods to the AuthHandler

// RefreshToken handles the token refresh endpoint
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	start := time.Now().UTC()

	// Extract client info for logging
	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()
	requestID := uuid.New().String()

	// Add request ID to context for tracing
	ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
	ctx = context.WithValue(ctx, "client_ip", clientIP)
	ctx = context.WithValue(ctx, "user_agent", userAgent)
	c.Request = c.Request.WithContext(ctx)

	// Start metrics tracking
	h.metricsService.IncTokenRefreshAttempt(ctx)

	// Log token refresh attempt
	h.logger.Info("Token refresh attempt",
		h.logger.Field("client_ip", clientIP),
		h.logger.Field("user_agent", userAgent),
		h.logger.Field("request_id", requestID))

	// Parse and validate request
	var request dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.metricsService.IncTokenRefreshFailure(ctx, "invalid_request")
		h.logger.Warn("Token refresh failure: invalid request",
			h.logger.Field("error", err.Error()),
			h.logger.Field("client_ip", clientIP),
			h.logger.Field("request_id", requestID))
		response.BadRequest(c, "Invalid request format", nil)
		return
	}

	// Perform token refresh
	refreshResp, err := h.authService.RefreshToken(ctx, &request)

	// Record metrics for the duration
	duration := time.Since(start).Seconds()
	h.metricsService.TokenRefreshDuration(ctx, duration)

	// Handle response based on result
	if err != nil {
		var statusCode int
		var errorType string
		var errorMsg string

		// Map internal errors to user-friendly messages
		switch {
		case strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "expired"):
			statusCode = http.StatusUnauthorized
			errorType = "invalid_token"
			errorMsg = "Refresh token is invalid or expired"
		case strings.Contains(err.Error(), "revoked"):
			statusCode = http.StatusUnauthorized
			errorType = "token_revoked"
			errorMsg = "Token has been revoked"
		default:
			statusCode = http.StatusInternalServerError
			errorType = "server_error"
			errorMsg = "Failed to refresh token"
		}

		h.metricsService.IncTokenRefreshFailure(ctx, errorType)
		h.logger.Warn("Token refresh failure",
			h.logger.Field("error_type", errorType),
			h.logger.Field("client_ip", clientIP),
			h.logger.Field("request_id", requestID))

		response.Error(c, statusCode, errorMsg, nil)
		return
	}

	// Log successful token refresh
	h.metricsService.IncTokenRefreshSuccess(ctx)
	h.logger.Info("Token refresh successful",
		h.logger.Field("client_ip", clientIP),
		h.logger.Field("request_id", requestID))

	// Return standardized response
	response.Success(c, "Token refreshed successfully", refreshResp)
}

// Logout handles the logout endpoint
func (h *AuthHandler) Logout(c *gin.Context) {
	start := time.Now().UTC()

	// Extract client info for logging
	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()
	requestID := uuid.New().String()

	// Add request ID to context for tracing
	ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
	c.Request = c.Request.WithContext(ctx)

	// Start metrics tracking
	h.metricsService.IncLogoutAttempt(ctx)

	// Log logout attempt
	h.logger.Info("Logout attempt",
		h.logger.Field("client_ip", clientIP),
		h.logger.Field("user_agent", userAgent),
		h.logger.Field("request_id", requestID))

	// Parse and validate request
	var request dto.LogoutRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.metricsService.IncLogoutFailure(ctx, "invalid_request")
		h.logger.Warn("Logout failure: invalid request",
			h.logger.Field("error", err.Error()),
			h.logger.Field("client_ip", clientIP),
			h.logger.Field("request_id", requestID))
		response.BadRequest(c, "Invalid request format", nil)
		return
	}

	// Perform logout
	err := h.authService.Logout(ctx, &request)

	// Record metrics for the duration
	duration := time.Since(start).Seconds()
	h.metricsService.LogoutDuration(ctx, duration)

	// Handle response based on result
	if err != nil {
		h.metricsService.IncLogoutFailure(ctx, "invalid_token")
		h.logger.Warn("Logout failure",
			h.logger.Field("error", err.Error()),
			h.logger.Field("client_ip", clientIP),
			h.logger.Field("request_id", requestID))

		response.Error(c, http.StatusUnauthorized, "Invalid tokens", nil)
		return
	}

	// Log successful logout
	h.metricsService.IncLogoutSuccess(ctx)
	h.logger.Info("Logout successful",
		h.logger.Field("client_ip", clientIP),
		h.logger.Field("request_id", requestID))

	// Return standardized response
	response.Success(c, "Logged out successfully", nil)
}
