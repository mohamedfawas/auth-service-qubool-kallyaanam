// internal/repository/redis/session_repository.go
package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/config"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/logging"
	"go.uber.org/zap"
)

// SessionRepo handles session operations in Redis.
type SessionRepo struct {
	client       *redis.Client
	logger       *zap.Logger
	cookieConfig *config.AuthConfig
}

// NewSessionRepo creates a new session repository.
func NewSessionRepo(client *redis.Client, cookieConfig *config.AuthConfig) *SessionRepo {
	return &SessionRepo{
		client:       client,
		logger:       logging.Logger(),
		cookieConfig: cookieConfig,
	}
}

const (
	// registrationSessionPrefix is used as key prefix for registration sessions
	registrationSessionPrefix = "registration_session:"
	// defaultSessionCookiePath is the default path for session cookies
	defaultSessionCookiePath = "/"
	// registrationSessionCookieName is the name of the registration session cookie
	registrationSessionCookieName = "registration_session"
)

// StoreRegistrationSession stores a registration session in Redis.
func (r *SessionRepo) StoreRegistrationSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error {
	key := registrationSessionPrefix + sessionID

	jsonData, err := json.Marshal(data)
	if err != nil {
		r.logger.Error("Failed to marshal session data", zap.Error(err))
		return err
	}

	err = r.client.Set(ctx, key, jsonData, expiration).Err()
	if err != nil {
		r.logger.Error("Failed to store session in Redis",
			zap.String("sessionID", sessionID),
			zap.Error(err),
		)
		return err
	}

	r.logger.Debug("Registration session stored",
		zap.String("sessionID", sessionID),
		zap.Duration("expiration", expiration),
	)
	return nil
}

// SetSessionCookie sets a session cookie in the HTTP response.
func (r *SessionRepo) SetSessionCookie(c *gin.Context, sessionID string, maxAge int) {
	c.SetCookie(
		registrationSessionCookieName,
		sessionID,
		maxAge,
		defaultSessionCookiePath,
		r.cookieConfig.CookieDomain,
		r.cookieConfig.CookieSecure,
		true, // HTTP only
	)

	r.logger.Debug("Session cookie set",
		zap.String("sessionID", sessionID),
		zap.Int("maxAge", maxAge),
	)
}
