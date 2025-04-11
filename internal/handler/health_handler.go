// internal/handler/health_handler.go
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/pkg/redis"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHealthHandler(db *gorm.DB, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redisClient,
	}
}

// RegisterRoutes registers health check routes
// The router parameter can be either *gin.Engine or *gin.RouterGroup
func (h *HealthHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/health", h.HealthCheck)
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	healthStatus := map[string]interface{}{
		"status":    "up",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Check database connection
	if err := h.db.Raw("SELECT 1").Error; err != nil {
		healthStatus["status"] = "down"
		healthStatus["database"] = "error: " + err.Error()
	} else {
		healthStatus["database"] = "up"
	}

	// Check Redis connection if enabled
	if h.redis != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := h.redis.Healthcheck(ctx); err != nil {
			healthStatus["status"] = "down"
			healthStatus["redis"] = "error: " + err.Error()
		} else {
			healthStatus["redis"] = "up"
		}
	} else {
		healthStatus["redis"] = "disabled"
	}

	c.JSON(http.StatusOK, healthStatus)
}
