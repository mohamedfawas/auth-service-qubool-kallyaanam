package service

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/config"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/repository"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/logging"
	"go.uber.org/zap"
)

// sessionService implements SessionService.
type sessionService struct {
	config      *config.Config
	sessionRepo repository.SessionRepository
	logger      *zap.Logger
}

// NewSessionService creates a new session service.
func NewSessionService(
	config *config.Config,
	sessionRepo repository.SessionRepository,
) SessionService {
	return &sessionService{
		config:      config,
		sessionRepo: sessionRepo,
		logger:      logging.Logger(),
	}
}

// CreateRegistrationSession creates a new registration session.
func (s *sessionService) CreateRegistrationSession(ctx context.Context, data interface{}) (string, error) {
	// Generate a unique session ID
	sessionID := uuid.New().String()

	// Store session in Redis with expiration
	expiration := s.config.Auth.RegistrationExpiry
	err := s.sessionRepo.StoreRegistrationSession(ctx, sessionID, data, expiration)
	if err != nil {
		s.logger.Error("Failed to store registration session",
			zap.Error(err),
		)
		return "", err
	}

	s.logger.Debug("Created registration session",
		zap.String("sessionID", sessionID),
		zap.Duration("expiration", expiration),
	)

	return sessionID, nil
}

// SetRegistrationSessionCookie sets the registration session cookie.
func (s *sessionService) SetRegistrationSessionCookie(c *gin.Context, sessionID string) {
	// Convert duration to seconds for cookie max age
	maxAge := int(s.config.Auth.RegistrationExpiry / time.Second)

	// Set the cookie on the response
	s.sessionRepo.SetSessionCookie(c, sessionID, maxAge)

	s.logger.Debug("Set registration session cookie",
		zap.String("sessionID", sessionID),
		zap.Int("maxAge", maxAge),
	)
}
