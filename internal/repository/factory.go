// internal/repository/factory.go
package repository

import (
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/config"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/repository/postgres"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/repository/redis"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/logging"
	pgdb "github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/postgres"
	redisdb "github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/redis"
	"go.uber.org/zap"
)

// Factory creates and manages repository instances.
type Factory struct {
	PendingRegistration PendingRegistrationRepository
	User                UserRepository
	Session             SessionRepository
	Verification        VerificationRepository
	logger              *zap.Logger
}

// NewFactory creates a new repository factory with all repositories.
func NewFactory(cfg *config.Config) (*Factory, error) {
	logger := logging.Logger()

	// Connect to PostgreSQL using the pkg utility
	db, err := pgdb.Connect(&cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL", zap.Error(err))
		return nil, err
	}

	// Connect to Redis using the pkg utility
	redisClient, err := redisdb.Connect(&cfg.Redis)
	if err != nil {
		logger.Error("Failed to connect to Redis", zap.Error(err))
		return nil, err
	}

	// Create repositories
	pendingRepo := postgres.NewPendingRegistrationRepo(db)
	userRepo := postgres.NewUserRepo(db)
	sessionRepo := redis.NewSessionRepo(redisClient, &cfg.Auth)
	verificationRepo := redis.NewVerificationRepo(redisClient)

	return &Factory{
		PendingRegistration: pendingRepo,
		User:                userRepo,
		Session:             sessionRepo,
		Verification:        verificationRepo,
		logger:              logger,
	}, nil
}
