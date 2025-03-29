package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/response"
	"gorm.io/gorm"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *gorm.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// Health checks the health of the service and its dependencies
func (h *HealthHandler) Health(c *gin.Context) {
	status := "ok"
	dbStatus := "ok"
	redisStatus := "ok"

	// Check database
	sqlDB, err := h.db.DB()
	if err != nil {
		dbStatus = "error: " + err.Error()
		status = "degraded"
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err = sqlDB.PingContext(ctx)
		if err != nil {
			dbStatus = "error: " + err.Error()
			status = "degraded"
		}
	}

	// Check redis if available
	if h.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, err := h.redis.Ping(ctx).Result()
		if err != nil {
			redisStatus = "error: " + err.Error()
			status = "degraded"
		}
	} else {
		redisStatus = "disabled"
	}

	// Return health status
	response.Success(c, http.StatusOK, "Health check", gin.H{
		"status": status,
		"dependencies": gin.H{
			"database": dbStatus,
			"redis":    redisStatus,
		},
		"version": "1.0.0",
	})
}
