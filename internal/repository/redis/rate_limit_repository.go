package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository"
)

// RateLimitRepo implements RateLimitRepository using Redis
type RateLimitRepo struct {
	client *redis.Client
}

// NewRateLimitRepository creates a new rate limit repository
func NewRateLimitRepository(client *redis.Client) repository.RateLimitRepository {
	return &RateLimitRepo{client: client}
}

// IncrementCounter increments a counter for rate limiting
func (r *RateLimitRepo) IncrementCounter(ctx context.Context, key string, expiry time.Duration) (int, error) {
	// Increment the counter
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Set expiry if this is a new key (val == 1)
	if val == 1 {
		_, err = r.client.Expire(ctx, key, expiry).Result()
		if err != nil {
			return int(val), err
		}
	}

	return int(val), nil
}

// GetCounter gets the current count for a rate limit key
func (r *RateLimitRepo) GetCounter(ctx context.Context, key string) (int, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ResetCounter resets a rate limit counter
func (r *RateLimitRepo) ResetCounter(ctx context.Context, key string) error {
	_, err := r.client.Del(ctx, key).Result()
	return err
}
