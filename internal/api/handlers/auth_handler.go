package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/domain/models"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/service"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/response"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/validator"
	"gorm.io/gorm"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	authService *service.AuthService
	otpService  *service.OTPService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *gorm.DB, authService *service.AuthService, otpService *service.OTPService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		otpService:  otpService,
	}
}

// Register handles the user registration endpoint
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	// Bind the request body to the struct
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := validator.FormatValidationErrors(err)
		response.Error(c, http.StatusBadRequest, "Invalid request data", validationErrors)
		return
	}

	// Process registration
	resp, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		var statusCode int
		var errorMessage string

		// Handle specific errors
		switch {
		case errors.Is(err, service.ErrEmailExists):
			statusCode = http.StatusConflict
			errorMessage = "Email is already registered"
		case errors.Is(err, service.ErrPhoneExists):
			statusCode = http.StatusConflict
			errorMessage = "Phone number is already registered"
		case errors.Is(err, service.ErrPendingEmailExists):
			statusCode = http.StatusConflict
			errorMessage = "Email already has a pending registration"
		case errors.Is(err, service.ErrPendingPhoneExists):
			statusCode = http.StatusConflict
			errorMessage = "Phone number already has a pending registration"
		case errors.Is(err, service.ErrPasswordEncryption):
			statusCode = http.StatusInternalServerError
			errorMessage = "Failed to process registration"
			log.Printf("Password encryption error: %v", err)
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "An error occurred while processing your registration"
			log.Printf("Registration error: %v", err)
		}

		response.Error(c, statusCode, errorMessage, err.Error())
		return
	}

	// Generate and log OTPs for the pending registration
	err = h.otpService.GenerateAndLogOTPs(c.Request.Context(), resp.PendingID)
	if err != nil {
		log.Printf("Failed to generate OTPs: %v", err)
		// Continue with the registration even if OTP generation fails
		// In a production environment, you might want to handle this differently
	}

	// Return successful response
	response.Success(c, http.StatusCreated, "Registration initiated successfully. Please verify your email and phone.", resp)
}
