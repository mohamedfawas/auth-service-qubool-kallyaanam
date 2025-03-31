// Package app initializes and configures the application.
package app

import (
	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/config"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/handler"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/middleware"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/repository"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/service"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/logging"
	"go.uber.org/zap"
)

// App represents the application.
type App struct {
	Router         *gin.Engine
	Config         *config.Config
	RepoFactory    *repository.Factory
	ServiceFactory *service.Factory
	HandlerFactory *handler.Factory
	Logger         *zap.Logger
}

// NewApp creates a new application.
func NewApp(cfg *config.Config) (*App, error) {
	logger := logging.Logger()

	// Set Gin mode based on environment
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := gin.New()

	// Initialize repository factory
	repoFactory, err := repository.NewFactory(cfg)
	if err != nil {
		logger.Error("Failed to create repository factory", zap.Error(err))
		return nil, err
	}

	// Initialize service factory
	serviceFactory := service.NewFactory(cfg, repoFactory)

	// Initialize handler factory
	handlerFactory := handler.NewFactory(serviceFactory)

	return &App{
		Router:         router,
		Config:         cfg,
		RepoFactory:    repoFactory,
		ServiceFactory: serviceFactory,
		HandlerFactory: handlerFactory,
		Logger:         logger,
	}, nil
}

// SetupRoutes configures the routes for the application.
func (a *App) SetupRoutes() {
	// Global middleware
	a.Router.Use(middleware.Logger())
	a.Router.Use(gin.Recovery())

	// Health check
	a.Router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes
	authRoutes := a.Router.Group("/auth")
	{
		// Rate limit for registration
		registerGroup := authRoutes.Group("")
		registerGroup.Use(middleware.RateLimit(a.Config.Auth.RateLimitPerMinute, 5))

		// Register endpoint
		registerGroup.POST("/register", a.HandlerFactory.Auth.Register)

		// Add other auth endpoints here as needed
	}

	a.Logger.Info("Routes configured successfully")
}

// Run starts the application.
func (a *App) Run() error {
	a.Logger.Info("Starting server", zap.String("port", a.Config.Server.Port))
	return a.Router.Run(":" + a.Config.Server.Port)
}
