package service

import (
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/config"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/repository"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/logging"
	"go.uber.org/zap"
)

// Factory creates and manages service instances.
type Factory struct {
	Registration RegistrationService
	Notification NotificationService
	Session      SessionService
	logger       *zap.Logger
}

// NewFactory creates a new service factory.
func NewFactory(
	cfg *config.Config,
	repoFactory *repository.Factory,
) *Factory {
	logger := logging.Logger()

	// Create services
	sessionService := NewSessionService(cfg, repoFactory.Session)
	notificationService := NewNotificationService(cfg, repoFactory.Verification)
	registrationService := NewRegistrationService(
		cfg,
		repoFactory.PendingRegistration,
		repoFactory.User,
		notificationService,
		sessionService,
	)

	logger.Info("Service factory initialized")

	return &Factory{
		Registration: registrationService,
		Notification: notificationService,
		Session:      sessionService,
		logger:       logger,
	}
}
