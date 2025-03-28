package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/api/handlers"
)

// Setup configures all the routes for the server
func Setup(r *gin.Engine, authHandler *handlers.AuthHandler) {
	// Health check route
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
	}
}
