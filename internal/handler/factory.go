package handler

import (
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/service"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/logging"
	"go.uber.org/zap"
)

// Factory creates and manages handler instances.
type Factory struct {
	Auth   *AuthHandler
	logger *zap.Logger
}

// NewFactory creates a new handler factory.
func NewFactory(serviceFactory *service.Factory) *Factory {
	logger := logging.Logger()

	authHandler := NewAuthHandler(
		serviceFactory.Registration,
		serviceFactory.Session,
	)

	logger.Info("Handler factory initialized")

	return &Factory{
		Auth:   authHandler,
		logger: logger,
	}
}
