// auth-service-qubool-kallyaanam/internal/handler/auth_handler.go
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/model"
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
	var request model.RegistrationRequest
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
