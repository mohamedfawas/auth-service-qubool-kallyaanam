package cache

import (
	"context"
	"fmt"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/configs"
	"github.com/redis/go-redis/v9"
)

type RedisAdapter struct {
	Client *redis.Client // pointer to redis.Client instance
}

func NewRedisAdapter(config *configs.RedisConfig) (*RedisAdapter, error) {
	// Create a new Redis client with the provided configuration settings
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port), // Set Redis server address (host:port)
		Password: config.Password,                                // Set password (empty if no authentication is needed)
		DB:       config.DB,                                      // Set the database number (0-15), 0 is the default database
	})

	// Create a new context without timeout or cancellation.
	// This means the operation will keep running until it completes
	ctx := context.Background()

	// test the connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisAdapter{Client: client}, nil
}

func (ra *RedisAdapter) Ping() error {
	ctx := context.Background()

	// Send a PING command to Redis to check if it's responding.
	// If the command succeeds, it returns a "PONG" message.
	// If there's an error (e.g., Redis is down), it returns the error.
	return ra.Client.Ping(ctx).Err()
}
