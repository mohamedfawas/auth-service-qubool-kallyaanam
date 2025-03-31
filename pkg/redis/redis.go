// pkg/database/redis/redis.go
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/config"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/logging"
	"go.uber.org/zap"
)

// Connect establishes a connection to Redis.
func Connect(cfg *config.RedisConfig) (*redis.Client, error) {
	logger := logging.Logger()

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		logger.Error("Failed to connect to Redis", zap.Error(err))
		return nil, fmt.Errorf("redis connection error: %w", err)
	}

	logger.Info("Connected to Redis",
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
	)

	return client, nil
}
