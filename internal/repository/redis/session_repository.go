package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository"
)

// SessionRepo implements SessionRepository using Redis
type SessionRepo struct {
	client *redis.Client
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(client *redis.Client) repository.SessionRepository {
	return &SessionRepo{client: client}
}

// StoreSession stores a mapping between a session ID and a pending registration ID
func (r *SessionRepo) StoreSession(ctx context.Context, sessionID string, pendingID string, expiry time.Duration) error {
	key := "session:" + sessionID
	return r.client.Set(ctx, key, pendingID, expiry).Err()
}

// GetSession retrieves a pending registration ID by session ID
func (r *SessionRepo) GetSession(ctx context.Context, sessionID string) (string, error) {
	key := "session:" + sessionID
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // Session not found
		}
		return "", err
	}
	return val, nil
}

// DeleteSession removes a session
func (r *SessionRepo) DeleteSession(ctx context.Context, sessionID string) error {
	key := "session:" + sessionID
	return r.client.Del(ctx, key).Err()
}
