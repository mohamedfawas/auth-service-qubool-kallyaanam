package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/domain/models"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/service"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/response"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/validator"
	"gorm.io/gorm"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	db          *gorm.DB
	authService *service.AuthService
	otpService  *service.OTPService
	redisClient *redis.Client
	sessionRepo repository.SessionRepository
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *gorm.DB, authService *service.AuthService, otpService *service.OTPService, redisClient *redis.Client, sessionRepo repository.SessionRepository) *AuthHandler {
	return &AuthHandler{
		db:          db,
		authService: authService,
		otpService:  otpService,
		redisClient: redisClient,
		sessionRepo: sessionRepo,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Start the registration process for a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration details"
// @Success 201 {object} response.Response{data=models.RegisterResponse} "Registration initiated"
// @Failure 400 {object} response.Response{error=object} "Invalid input"
// @Failure 409 {object} response.Response{error=string} "Conflict - already exists"
// @Failure 500 {object} response.Response{error=string} "Server error"
// @Router /auth/register [post]
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

	// Create session and set cookie if Redis is available
	if h.sessionRepo != nil {
		// Generate unique session ID
		sessionID := uuid.New().String()

		// Store in Redis with same expiry as registration
		expiry := time.Until(resp.ExpiresAt)
		err = h.sessionRepo.StoreSession(c.Request.Context(), sessionID, resp.PendingID.String(), expiry)
		if err != nil {
			log.Printf("Failed to store session: %v", err)
			// Continue anyway - user will need to use pendingID
		} else {
			// Set cookie with session ID
			c.SetCookie(
				"registration_session",
				sessionID,
				int(expiry.Seconds()),
				"/",
				"",                   // Use appropriate domain in production
				c.Request.TLS != nil, // Secure if HTTPS
				true,                 // HTTP only
			)

			log.Printf("Session cookie set for pendingID: %s", resp.PendingID)
		}
	}

	// Return successful response
	response.Success(c, http.StatusCreated, "Registration initiated successfully. Please verify your email and phone.", resp)
}

// VerifyOTP handles the OTP verification endpoint
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest

	// Bind the request body to the struct
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := validator.FormatValidationErrors(err)
		response.Error(c, http.StatusBadRequest, "Invalid request data", validationErrors)
		return
	}

	// If pendingID is not provided, try to get it from session cookie
	if req.PendingID == uuid.Nil && h.sessionRepo != nil {
		// Get session cookie
		sessionID, err := c.Cookie("registration_session")
		if err == nil && sessionID != "" {
			// Retrieve pendingID from session
			pendingIDStr, err := h.sessionRepo.GetSession(c.Request.Context(), sessionID)
			if err == nil && pendingIDStr != "" {
				pendingID, err := uuid.Parse(pendingIDStr)
				if err == nil {
					req.PendingID = pendingID
					log.Printf("Retrieved pendingID %s from session cookie", pendingID)
				}
			}
		}
	}

	// If still nil, return error
	if req.PendingID == uuid.Nil {
		response.Error(c, http.StatusBadRequest, "Missing registration ID",
			"Either provide pendingID in request body or use the same browser/device where registration was initiated")
		return
	}

	// Verify the OTP
	err := h.otpService.VerifyOTP(c.Request.Context(), req.PendingID, req.OTPType, req.OTPValue)

	if err != nil {
		var statusCode int
		var errorMessage string

		switch {
		case errors.Is(err, service.ErrOTPNotFound):
			statusCode = http.StatusNotFound
			errorMessage = "OTP not found or expired"
		case errors.Is(err, service.ErrOTPMismatch):
			statusCode = http.StatusBadRequest
			errorMessage = "Invalid OTP"
		case errors.Is(err, service.ErrOTPMaxAttempts):
			statusCode = http.StatusTooManyRequests
			errorMessage = "Maximum verification attempts exceeded"
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "An error occurred while verifying the OTP"
			log.Printf("OTP verification error: %v", err)
		}

		response.Error(c, statusCode, errorMessage, err.Error())
		return
	}

	// Mark the verification as complete
	err = h.otpService.MarkVerified(c.Request.Context(), req.PendingID, req.OTPType, h.authService.GetRegistrationRepo())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update verification status", err.Error())
		return
	}

	// Return success response
	resp := models.VerifyOTPResponse{
		PendingID: req.PendingID,
		OTPType:   req.OTPType,
		Verified:  true,
		Message:   fmt.Sprintf("%s verified successfully", req.OTPType),
	}

	response.Success(c, http.StatusOK, "OTP verified successfully", resp)
}

// CompleteRegistration handles the endpoint to complete registration
func (h *AuthHandler) CompleteRegistration(c *gin.Context) {
	var req models.CompleteRegistrationRequest

	// Bind the request body to the struct
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := validator.FormatValidationErrors(err)
		response.Error(c, http.StatusBadRequest, "Invalid request data", validationErrors)
		return
	}

	// If pendingID is not provided, try to get it from session cookie
	if req.PendingID == uuid.Nil && h.sessionRepo != nil {
		// Get session cookie
		sessionID, err := c.Cookie("registration_session")
		if err == nil && sessionID != "" {
			// Retrieve pendingID from session
			pendingIDStr, err := h.sessionRepo.GetSession(c.Request.Context(), sessionID)
			if err == nil && pendingIDStr != "" {
				pendingID, err := uuid.Parse(pendingIDStr)
				if err == nil {
					req.PendingID = pendingID
					log.Printf("Retrieved pendingID %s from session cookie", pendingID)
				}
			}
		}
	}

	// If still nil, return error
	if req.PendingID == uuid.Nil {
		response.Error(c, http.StatusBadRequest, "Missing registration ID",
			"Either provide pendingID in request body or use the same browser/device where registration was initiated")
		return
	}

	// Complete registration
	resp, err := h.authService.CompleteRegistration(c.Request.Context(), req.PendingID)

	if err != nil {
		var statusCode int
		var errorMessage string

		switch {
		case errors.Is(err, service.ErrEmailNotVerified):
			statusCode = http.StatusBadRequest
			errorMessage = "Email not verified"
		case errors.Is(err, service.ErrPhoneNotVerified):
			statusCode = http.StatusBadRequest
			errorMessage = "Phone not verified"
		case errors.Is(err, service.ErrCreateUser):
			statusCode = http.StatusInternalServerError
			errorMessage = "Failed to create user"
			log.Printf("User creation error: %v", err)
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "An error occurred while completing registration"
			log.Printf("Registration completion error: %v", err)
		}

		response.Error(c, statusCode, errorMessage, err.Error())
		return
	}

	// After successful registration completion, clear the session cookie
	if h.sessionRepo != nil {
		sessionID, err := c.Cookie("registration_session")
		if err == nil && sessionID != "" {
			h.sessionRepo.DeleteSession(c.Request.Context(), sessionID)
			// Clear the cookie
			c.SetCookie("registration_session", "", -1, "/", "", c.Request.TLS != nil, true)
		}
	}

	// Return successful response
	response.Success(c, http.StatusCreated, "Registration completed successfully", resp)
}
