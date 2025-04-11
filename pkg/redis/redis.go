// pkg/redis/redis.go
package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/config"
)

// Client wraps the Redis client
type Client struct {
	*redis.Client
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg *config.RedisConfig) (*Client, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{Client: client}, nil
}

// Healthcheck checks the Redis connection
func (c *Client) Healthcheck(ctx context.Context) error {
	if c == nil || c.Client == nil {
		return nil // Redis is disabled, so no health check needed
	}

	return c.Ping(ctx).Err()
}
