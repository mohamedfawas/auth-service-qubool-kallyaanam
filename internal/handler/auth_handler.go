// Package handler contains HTTP handlers for the API.
// internal/handler/auth_handler.go
package handler

import (
	"errors" // Add the standard library errors package
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/domain/dto" // Rename custom errors package import
	apperrors "github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/errors"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/service"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/logging"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/validator"
	"go.uber.org/zap"
)

// AuthHandler handles authentication-related routes.
type AuthHandler struct {
	registrationService service.RegistrationService
	sessionService      service.SessionService
	logger              *zap.Logger
	validator           *validator.Validator
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(
	registrationService service.RegistrationService,
	sessionService service.SessionService,
) *AuthHandler {
	return &AuthHandler{
		registrationService: registrationService,
		sessionService:      sessionService,
		logger:              logging.Logger(),
		validator:           validator.NewValidator(),
	}
}

// Register handles user registration.
// @Summary Register a new user
// @Description Initiates the registration process for a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration details"
// @Success 202 {object} dto.Response "Registration initiated successfully"
// @Failure 400 {object} dto.Response "Invalid input data"
// @Failure 409 {object} dto.Response "Email or phone already registered"
// @Failure 429 {object} dto.Response "Rate limit exceeded"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Validate struct
	if validationErrors := h.validator.ValidateStruct(req); validationErrors != nil {
		h.logger.Warn("Validation errors", zap.Any("errors", validationErrors))
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			http.StatusBadRequest,
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Register user
	response, sessionID, err := h.registrationService.Register(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Set registration session cookie
	h.sessionService.SetRegistrationSessionCookie(c, sessionID)

	// Return success response
	c.JSON(http.StatusAccepted, dto.NewSuccessResponse(
		http.StatusAccepted,
		"Registration initiated successfully. Please verify your email and phone.",
		response,
	))
}

// handleError handles different types of errors and returns appropriate HTTP responses.
func (h *AuthHandler) handleError(c *gin.Context, err error) {
	var appErr *apperrors.Error
	if errors.As(err, &appErr) {
		// Handle application error
		c.JSON(appErr.Status, dto.NewErrorResponse(
			appErr.Status,
			appErr.Message,
			appErr.Details,
		))
	} else {
		// Handle unknown error
		h.logger.Error("Unexpected error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			http.StatusInternalServerError,
			"An unexpected error occurred",
			nil,
		))
	}
}
