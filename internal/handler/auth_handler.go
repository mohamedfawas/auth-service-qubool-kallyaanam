// auth-service-qubool-kallyaanam/internal/handler/auth_handler.go
package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/middleware"
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
	// Create verification-specific rate limiter
	verificationRateLimiter := middleware.VerificationRateLimiter(middleware.VerificationRateLimiterConfig{
		MaxAttemptsPerPeriod: 5,
		PeriodMinutes:        15,
		BlockDurationMinutes: 60,
	}, h.logger)

	router.POST("/register", h.Register)
	router.POST("/verify-email", verificationRateLimiter, h.VerifyEmail)
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

// VerifyEmail handles the email verification endpoint
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	start := time.Now().UTC()

	// Extract client info for logging and context
	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Create context with client info
	ctx := context.WithValue(c.Request.Context(), "client_ip", clientIP)
	ctx = context.WithValue(ctx, "request_start_time", start)
	ctx = context.WithValue(ctx, "user_agent", userAgent)

	// Log verification attempt
	h.logger.VerificationAttempt("", clientIP, userAgent)

	// Parse and validate request
	var request model.VerifyEmailRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.metricsService.IncVerificationFailure(c, "invalid_request")
		h.logger.VerificationFailure("", clientIP, "invalid_request",
			h.logger.Field("error", err))
		response.BadRequest(c, "Invalid request", err)
		return
	}

	// Sanitize inputs
	request.Email = h.securityService.SanitizeInput(ctx, request.Email)
	request.OTP = h.securityService.SanitizeInput(ctx, request.OTP)

	// Verify the email
	verifyResp, err := h.authService.VerifyEmail(ctx, &request)

	// Record metrics for the duration
	duration := time.Since(start).Seconds()
	h.metricsService.VerificationDuration(c, duration)

	// Handle response based on result
	if err != nil {
		var statusCode int
		var errorMsg string

		// Determine appropriate status code and error message based on error
		// Use specific error cases but don't expose internal details
		switch {
		case strings.Contains(err.Error(), "no pending registration"):
			statusCode = http.StatusNotFound
			errorMsg = "No verification in progress for this email"
		case strings.Contains(err.Error(), "invalid OTP"):
			statusCode = http.StatusBadRequest
			errorMsg = "Invalid verification code"
		case strings.Contains(err.Error(), "expired"):
			statusCode = http.StatusGone
			errorMsg = "Verification expired, please register again"
		case strings.Contains(err.Error(), "user already exists"):
			statusCode = http.StatusConflict
			errorMsg = "Account already exists"
		case strings.Contains(err.Error(), "unavailable"):
			statusCode = http.StatusServiceUnavailable
			errorMsg = "Service temporarily unavailable, please try again later"
		default:
			statusCode = http.StatusInternalServerError
			errorMsg = "Failed to complete verification"
		}

		// Send secure error response that doesn't expose details
		response.Error(c, statusCode, errorMsg, &ResponseError{Message: errorMsg})
		return
	}

	// Return standardized success response
	response.Success(c, "Email verification successful", gin.H{
		"id":      verifyResp.ID,
		"email":   verifyResp.Email,
		"message": verifyResp.Message,
	})
}

// ResponseError implements error interface for secure error responses
type ResponseError struct {
	Message string
}

func (e *ResponseError) Error() string {
	return e.Message
}
