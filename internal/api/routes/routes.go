package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/api/handlers"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/middleware"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository"
)

// Setup configures all the routes for the server
func Setup(r *gin.Engine, authHandler *handlers.AuthHandler, rateLimitRepo repository.RateLimitRepository) {
	// Health check route
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Auth routes
	auth := r.Group("/auth")

	// Apply rate limiting to auth endpoints if Redis is available
	if rateLimitRepo != nil {
		auth.Use(middleware.RateLimiter(rateLimitRepo, "auth", 5, time.Minute))
	}

	auth.POST("/register", authHandler.Register)
}
