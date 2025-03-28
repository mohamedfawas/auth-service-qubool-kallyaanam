package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/config"
)

// Connect creates a new Redis client connection
func Connect(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}
