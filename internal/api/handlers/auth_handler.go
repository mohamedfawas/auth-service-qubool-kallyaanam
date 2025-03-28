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
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/service"
	appErrors "github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/errors"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/response"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/validator"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	authService  *service.AuthService
	otpService   *service.OTPService
	redisClient  *redis.Client
	tokenService *service.TokenService
}

// Update constructor to remove db parameter
func NewAuthHandler(
	authService *service.AuthService,
	otpService *service.OTPService,
	redisClient *redis.Client,
	tokenService *service.TokenService,
) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		otpService:   otpService,
		redisClient:  redisClient,
		tokenService: tokenService,
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

		// Use appErrors.SanitizeError to ensure no sensitive details are leaked
		response.Error(c, statusCode, errorMessage, appErrors.SanitizeError(err).Error())
		return
	}

	// Generate and log OTPs for the pending registration
	err = h.otpService.GenerateAndLogOTPs(c.Request.Context(), resp.PendingID)
	if err != nil {
		log.Printf("Failed to generate OTPs: %v", err)
		// Continue with the registration even if OTP generation fails
	}

	// JWT-based approach (new method)
	if h.tokenService != nil {
		// Generate JWT token
		token, err := h.tokenService.GenerateToken(resp.PendingID, time.Now().Add(15*time.Minute))
		if err != nil {
			log.Printf("Failed to generate JWT token: %v", err)
		} else {
			// Set cookie with JWT token
			c.SetCookie(
				h.tokenService.GetCookieName(),
				token,
				int((15 * time.Minute).Seconds()),
				"/",
				"",
				c.Request.TLS != nil,
				true,
			)

			// Generate and set refresh token
			refreshToken, err := h.tokenService.GenerateRefreshToken(resp.PendingID)
			if err != nil {
				log.Printf("Failed to generate refresh token: %v", err)
			} else {
				c.SetCookie(
					"refresh_token",
					refreshToken,
					int((24 * time.Hour * 7).Seconds()), // 7 days
					"/",
					"",
					c.Request.TLS != nil,
					true,
				)
			}
		}
	}

	// Return successful response without pendingID
	cleanResp := &models.RegisterResponse{
		Email:     resp.Email,
		Phone:     resp.Phone,
		ExpiresAt: resp.ExpiresAt,
	}

	response.Success(c, http.StatusCreated, "Registration initiated successfully. Please verify your email and phone.", cleanResp)
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

	// If pendingID is not provided, try JWT token
	if req.PendingID == uuid.Nil && h.tokenService != nil {
		// Try to get token from cookie
		token, err := c.Cookie(h.tokenService.GetCookieName())
		if err == nil && token != "" {
			// Parse JWT token to extract pendingID
			pendingID, err := h.tokenService.ParseToken(token)
			if err == nil {
				req.PendingID = pendingID
				log.Printf("Retrieved pendingID %s from JWT token", pendingID)
			} else {
				log.Printf("Failed to parse JWT token: %v", err)
			}
		}
	}

	// If still nil, return more user-friendly error
	if req.PendingID == uuid.Nil {
		response.Error(c, http.StatusBadRequest, "Session not found",
			"Please use the same browser/device where you started registration or restart the registration process")
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

	// Rotate JWT token for enhanced security
	if h.tokenService != nil {
		rotatedToken, err := h.tokenService.GenerateToken(req.PendingID, time.Now().Add(24*time.Hour))
		if err == nil {
			c.SetCookie(
				h.tokenService.GetCookieName(),
				rotatedToken,
				int((24 * time.Hour).Seconds()),
				"/",
				"",
				c.Request.TLS != nil,
				true,
			)
		}
	}

	// Return success response without pendingID
	resp := models.VerifyOTPResponse{
		OTPType:  req.OTPType,
		Verified: true,
		Message:  fmt.Sprintf("%s verified successfully", req.OTPType),
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

	// If pendingID is not provided, try JWT token
	if req.PendingID == uuid.Nil && h.tokenService != nil {
		token, err := c.Cookie(h.tokenService.GetCookieName())
		if err == nil && token != "" {
			pendingID, err := h.tokenService.ParseToken(token)
			if err == nil {
				req.PendingID = pendingID
				log.Printf("Retrieved pendingID %s from JWT token", pendingID)
			}
		}
	}

	// If still nil, return improved error
	if req.PendingID == uuid.Nil {
		response.Error(c, http.StatusBadRequest, "Session not found",
			"Please use the same browser/device where you started registration or restart the registration process")
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

	// Clear JWT token cookies
	if h.tokenService != nil {
		// Get the token before clearing it
		token, err := c.Cookie(h.tokenService.GetCookieName())
		if err == nil && token != "" {
			// Add to invalidation list (blacklist)
			h.tokenService.InvalidateToken(token)
		}

		// Clear access and refresh token cookies
		c.SetCookie(h.tokenService.GetCookieName(), "", -1, "/", "", c.Request.TLS != nil, true)
		c.SetCookie("refresh_token", "", -1, "/", "", c.Request.TLS != nil, true)
		c.SetCookie("csrf_token", "", -1, "/", "", c.Request.TLS != nil, true)
	}

	// Return successful response
	response.Success(c, http.StatusCreated, "Registration completed successfully", resp)
}

// RefreshToken handles the token refresh endpoint
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	// Bind the request body to the struct
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := validator.FormatValidationErrors(err)
		response.Error(c, http.StatusBadRequest, "Invalid request data", validationErrors)
		return
	}

	// If no refresh token in body, try to get it from cookie
	if req.RefreshToken == "" {
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil || refreshToken == "" {
			response.Error(c, http.StatusBadRequest, "Refresh token required", "Missing refresh token")
			return
		}
		req.RefreshToken = refreshToken
	}

	// Refresh the token
	newAccessToken, newRefreshToken, err := h.tokenService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid refresh token", err.Error())
		return
	}

	// Set cookies for both tokens
	c.SetCookie(
		h.tokenService.GetCookieName(),
		newAccessToken,
		int((15 * time.Minute).Seconds()),
		"/",
		"",
		c.Request.TLS != nil,
		true,
	)

	c.SetCookie(
		"refresh_token",
		newRefreshToken,
		int((24 * time.Hour * 7).Seconds()), // Refresh token valid for a week
		"/",
		"",
		c.Request.TLS != nil,
		true,
	)

	// Also return tokens in response for clients that don't use cookies
	response.Success(c, http.StatusOK, "Token refreshed successfully", gin.H{
		"access_token": newAccessToken,
		"expires_in":   int((15 * time.Minute).Seconds()),
	})
}
