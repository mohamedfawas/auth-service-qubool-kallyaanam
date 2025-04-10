// internal/handler/health_handler.go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// RegisterRoutes registers health check routes
// The router parameter can be either *gin.Engine or *gin.RouterGroup
func (h *HealthHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/health", h.HealthCheck)
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	// Add database connection check
	sqlDB, err := h.db.DB() // Get the underlying SQL DB
	dbStatus := "UP"
	dbDetails := map[string]string{}

	if err != nil {
		dbStatus = "DOWN"
		dbDetails["error"] = err.Error()
	} else {
		// Check DB connection
		if err := sqlDB.Ping(); err != nil {
			dbStatus = "DOWN"
			dbDetails["error"] = err.Error()
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "UP",
		"service": "auth-service",
		"version": "1.0.0",
		"dependencies": map[string]interface{}{
			"postgres": map[string]interface{}{
				"status":  dbStatus,
				"details": dbDetails,
			},
		},
	})
}
