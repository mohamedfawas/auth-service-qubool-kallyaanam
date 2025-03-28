package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/domain/models"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/response"
	"gorm.io/gorm"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	db *gorm.DB
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{
		db: db,
	}
}

// Register handles the user registration endpoint
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	// Bind the request body to the struct
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Phase 1: Just acknowledge the request and return a dummy response
	pendingID := uuid.New()
	resp := models.RegisterResponse{
		PendingID: pendingID,
		Email:     req.Email,
		Phone:     req.Phone,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	response.Success(c, http.StatusCreated, "Registration request received. Implementation in progress.", resp)
}
