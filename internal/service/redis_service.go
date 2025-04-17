package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/util/logger"
)

// Common Redis key prefixes
const (
	RefreshTokenPrefix  = "refresh_token:"
	BlacklistPrefix     = "blacklist:"
	LoginAttemptsPrefix = "login_attempts:"
	LoginHistoryPrefix  = "login_history:"
)

// TokenData represents data stored with a refresh token
type TokenData struct {
	UserID    string    `json:"user_id"`
	TokenID   string    `json:"token_id"`
	UserRole  string    `json:"user_role"`
	IssuedAt  time.Time `json:"issued_at"`
	UserAgent string    `json:"user_agent"`
	ClientIP  string    `json:"client_ip"`
}

// RedisServiceConfig holds Redis configuration
type RedisServiceConfig struct {
	Address      string
	Password     string
	DB           int
	TokenExpiry  time.Duration
	ThrottleRate int           // Max attempts per minute
	ThrottleTTL  time.Duration // How long throttling lasts
}

// RedisService provides Redis operations
type redisService struct {
	client *redis.Client
	config RedisServiceConfig
	logger *logger.Logger
}

// NewRedisService creates a new RedisService
func NewRedisService(config RedisServiceConfig, logger *logger.Logger) RedisService {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Password: config.Password,
		DB:       config.DB,
	})

	return &redisService{
		client: client,
		config: config,
		logger: logger,
	}
}

// StoreRefreshToken stores a refresh token with associated data
func (s *redisService) StoreRefreshToken(ctx context.Context, tokenID, token string, data TokenData) error {
	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	// Store token data with expiry
	key := RefreshTokenPrefix + tokenID
	if err := s.client.Set(ctx, key, jsonData, s.config.TokenExpiry).Err(); err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

// GetRefreshTokenData retrieves data associated with a refresh token
func (s *redisService) GetRefreshTokenData(ctx context.Context, tokenID string) (*TokenData, error) {
	key := RefreshTokenPrefix + tokenID
	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Token not found
		}
		return nil, fmt.Errorf("failed to get refresh token data: %w", err)
	}

	var tokenData TokenData
	if err := json.Unmarshal([]byte(data), &tokenData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token data: %w", err)
	}

	return &tokenData, nil
}

// DeleteRefreshToken removes a refresh token
func (s *redisService) DeleteRefreshToken(ctx context.Context, tokenID string) error {
	key := RefreshTokenPrefix + tokenID
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

// BlacklistToken adds a token to the blacklist
func (s *redisService) BlacklistToken(ctx context.Context, tokenID string, expiry time.Duration) error {
	key := BlacklistPrefix + tokenID
	if err := s.client.Set(ctx, key, "1", expiry).Err(); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	return nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *redisService) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := BlacklistPrefix + tokenID
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}
	return exists > 0, nil
}

// IncrementLoginAttempts increments login attempts for an IP address
func (s *redisService) IncrementLoginAttempts(ctx context.Context, ip string) (int64, error) {
	key := LoginAttemptsPrefix + ip

	// Increment the counter
	count, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment login attempts: %w", err)
	}

	// Set expiry if this is the first attempt
	if count == 1 {
		if err := s.client.Expire(ctx, key, time.Minute).Err(); err != nil {
			s.logger.Warn("Failed to set expiry for login attempts",
				s.logger.Field("ip", ip),
				s.logger.Field("error", err.Error()))
		}
	}

	return count, nil
}

// IsLoginThrottled checks if login attempts should be throttled
func (s *redisService) IsLoginThrottled(ctx context.Context, ip string) (bool, error) {
	key := LoginAttemptsPrefix + ip

	count, err := s.client.Get(ctx, key).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil // No attempts recorded
		}
		return false, fmt.Errorf("failed to get login attempts: %w", err)
	}

	// Check if attempts exceed threshold
	return count >= int64(s.config.ThrottleRate), nil
}

// StoreLoginHistory stores login history for a user
func (s *redisService) StoreLoginHistory(ctx context.Context, userID, userAgent, ip string) error {
	key := LoginHistoryPrefix + userID
	now := time.Now().UTC().Format(time.RFC3339)

	// Create history entry
	entry := map[string]interface{}{
		"time":       now,
		"user_agent": userAgent,
		"ip":         ip,
	}

	// Convert to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal login history: %w", err)
	}

	// Add to list with a maximum size (LPUSH + LTRIM)
	pipe := s.client.Pipeline()
	pipe.LPush(ctx, key, string(data))
	pipe.LTrim(ctx, key, 0, 9) // Keep last 10 entries

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to store login history: %w", err)
	}

	return nil
}
