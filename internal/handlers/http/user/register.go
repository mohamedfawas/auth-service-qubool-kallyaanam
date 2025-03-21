package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/service/user"
)

// Handler manages HTTP requests for user authentication
type Handler struct {
	authService *user.AuthService
}

// NewHandler creates a new instance of user Handler
func NewHandler(authService *user.AuthService) *Handler {
	return &Handler{
		authService: authService,
	}
}

// Register handles the POST /auth/register endpoint
func (h *Handler) Register(c *gin.Context) {
	var req user.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   response,
	})
}

// SetupRoutes registers all user authentication routes
func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		// These will be implemented later
		// auth.POST("/verify-email", h.VerifyEmail)
		// auth.POST("/verify-phone", h.VerifyPhone)
		// auth.POST("/login", h.Login)
		// auth.POST("/logout", h.Logout)
		// auth.POST("/forgot-password", h.ForgotPassword)
	}
}
