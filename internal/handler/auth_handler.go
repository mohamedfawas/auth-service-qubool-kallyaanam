// internal/handler/auth_handler.go
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/model"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/service"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/util/logger"
)

type AuthHandler struct {
	authService     *service.AuthService
	securityService service.SecurityService
	metricsService  service.MetricsService
	logger          *logger.Logger
}

func NewAuthHandler(
	authService *service.AuthService,
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
	start := time.Now()

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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Sanitize inputs
	request.Email = h.securityService.SanitizeInput(c, request.Email)
	request.Phone = h.securityService.SanitizeInput(c, request.Phone)

	// Validate password strength
	if isValid, reason := h.securityService.ValidatePassword(c, request.Password); !isValid {
		h.metricsService.IncRegistrationFailure(c, "weak_password")
		h.logger.RegistrationFailure(request.Email, clientIP, reason)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password validation failed: " + reason,
		})
		return
	}

	// Register the user
	response, err := h.authService.Register(c, &request)

	// Record metrics for the duration
	duration := time.Since(start).Seconds()
	h.metricsService.RegistrationDuration(c, duration)

	// Handle response based on result
	if err != nil {
		var statusCode int
		var errorType string

		// Determine appropriate status code and error type based on error
		switch err.Error() {
		case "email already exists", "phone already exists":
			statusCode = http.StatusConflict
			errorType = "duplicate_user"
		default:
			statusCode = http.StatusInternalServerError
			errorType = "server_error"
		}

		h.metricsService.IncRegistrationFailure(c, errorType)
		h.logger.RegistrationFailure(request.Email, clientIP, err.Error())

		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful registration
	h.metricsService.IncRegistrationSuccess(c)
	h.logger.RegistrationSuccess(response.ID, request.Email, clientIP)

	c.JSON(http.StatusCreated, response)
}
