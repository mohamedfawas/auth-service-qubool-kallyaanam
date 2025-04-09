// internal/handler/health_handler.go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", h.HealthCheck)
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "UP",
		"service": "auth-service",
		"version": "1.0.0",
	})
}
