package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/dto"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/interfaces"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/usecase"
	domainErrors "github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/domain/errors"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/services/logger"
)

type RegistrationHandler struct {
	registrationUseCase *usecase.RegistrationUseCase
	securityService     interfaces.SecurityService
	metricsService      interfaces.MetricsService
	logger              *logger.Logger
}

func NewRegistrationHandler(
	registrationUseCase *usecase.RegistrationUseCase,
	securityService interfaces.SecurityService,
	metricsService interfaces.MetricsService,
	logger *logger.Logger,
) *RegistrationHandler {
	return &RegistrationHandler{
		registrationUseCase: registrationUseCase,
		securityService:     securityService,
		metricsService:      metricsService,
		logger:              logger,
	}
}

func (h *RegistrationHandler) Register(c *gin.Context) {
	// Start timer for metrics
	startTime := time.Now()

	// Get client information
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Increment metrics for attempt
	h.metricsService.IncRegistrationAttempt(c)

	var request dto.RegistrationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.RegistrationFailure("", ipAddress, "invalid_request_format",
			zap.Error(err))
		h.metricsService.IncRegistrationFailure(c, "invalid_request_format")

		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Invalid request format",
			"error":   err.Error(),
		})
		return
	}

	// Sanitize inputs
	request.Email = h.securityService.SanitizeInput(c, request.Email)
	request.Phone = h.securityService.SanitizeInput(c, request.Phone)

	// Log the registration attempt with sanitized data
	h.logger.RegistrationAttempt(request.Email, ipAddress, userAgent)

	// Input validation
	if request.Email == "" || request.Phone == "" || request.Password == "" {
		h.logger.RegistrationFailure(request.Email, ipAddress, "missing_required_fields")
		h.metricsService.IncRegistrationFailure(c, "missing_required_fields")

		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Missing required fields",
			"error":   domainErrors.ErrRequiredFieldMissing.Error(),
		})
		return
	}

	// Validate password strength
	valid, reason := h.securityService.ValidatePassword(c, request.Password)
	if !valid {
		h.logger.RegistrationFailure(request.Email, ipAddress, "weak_password",
			zap.String("reason", reason))
		h.metricsService.IncRegistrationFailure(c, "weak_password")

		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Password is too weak",
			"error":   reason,
		})
		return
	}

	// Call use case
	response, err := h.registrationUseCase.Register(c.Request.Context(), &request)
	if err != nil {
		var statusCode int
		var errorType string

		// Handle specific errors
		switch err {
		case domainErrors.ErrEmailAlreadyExists:
			statusCode = http.StatusConflict
			errorType = "email_exists"
		case domainErrors.ErrPhoneAlreadyExists:
			statusCode = http.StatusConflict
			errorType = "phone_exists"
		case domainErrors.ErrInvalidEmail:
			statusCode = http.StatusBadRequest
			errorType = "invalid_email"
		case domainErrors.ErrInvalidPhone:
			statusCode = http.StatusBadRequest
			errorType = "invalid_phone"
		default:
			statusCode = http.StatusInternalServerError
			errorType = "internal_error"
		}

		h.logger.RegistrationFailure(request.Email, ipAddress, errorType,
			zap.Error(err))
		h.metricsService.IncRegistrationFailure(c, errorType)

		c.JSON(statusCode, gin.H{
			"status":  false,
			"message": "Registration failed",
			"error":   err.Error(),
		})
		return
	}

	// Record metrics for success and duration
	h.metricsService.IncRegistrationSuccess(c)
	h.metricsService.RegistrationDuration(c, time.Since(startTime).Seconds())

	// Log successful registration
	h.logger.RegistrationSuccess(response.ID, request.Email, ipAddress)

	c.JSON(http.StatusCreated, gin.H{
		"status":  true,
		"message": "Registration initiated successfully",
		"data":    response,
	})
}
