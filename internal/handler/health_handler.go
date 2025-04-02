package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/adapter/cache"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/adapter/database/postgres"
)

type HealthHandler struct {
	db    *postgres.PostgresAdapter
	cache *cache.RedisAdapter
}

func NewHealthHandler(db *postgres.PostgresAdapter, cache *cache.RedisAdapter) *HealthHandler {
	return &HealthHandler{
		db:    db,
		cache: cache,
	}
}

type HealthResponse struct {
	Status    string `json:"status"`
	Postgres  string `json:"postgres"`
	Redis     string `json:"redis"`
	Timestamp string `json:"timestamp"`
}

func (h *HealthHandler) Check(c *gin.Context) {
	response := HealthResponse{
		Status:    "ok",
		Postgres:  "up",
		Redis:     "up",
		Timestamp: time.Now().UTC().Format(time.RFC3339), // get the current time in UTC and format it as RFC3339
		// RFC3339 is a standard format for representing date and time
	}

	if err := h.db.Ping(); err != nil {
		response.Status = "degraded"
		response.Postgres = "down"
	}

	if err := h.cache.Ping(); err != nil {
		response.Status = "degraded"
		response.Redis = "down"
	}

	c.JSON(http.StatusOK, response)
}
