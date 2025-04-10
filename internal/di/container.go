// internal/di/container.go
package di

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/config"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/handler"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/middleware"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/repository"
	postgreRepo "github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/repository/postgres"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/service"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/util/logger"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/pkg/database"
)

// Container holds all application dependencies
type Container struct {
	Config     *config.Config
	Router     *gin.Engine
	AuthRoutes *gin.RouterGroup
	Logger     *logger.Logger

	// Services
	AuthService     service.AuthService
	OTPService      service.OTPService
	EmailService    service.EmailService
	SecurityService service.SecurityService
	MetricsService  service.MetricsService

	// Handlers
	AuthHandler   *handler.AuthHandler
	HealthHandler *handler.HealthHandler
}

// Initialize creates a new dependency injection container with all dependencies wired up
func Initialize() (*Container, error) {
	// Load configuration with validation
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	appLogger, err := logger.NewLogger(cfg.Logging.IsDevelopment)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize database connection
	db, err := database.NewPostgresConnection(cfg.Database.DSN)
	if err != nil {
		appLogger.Fatal("Failed to initialize database", appLogger.Field("error", err.Error()))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize repositories
	var userRepo repository.UserRepository
	userRepo = postgreRepo.NewUserRepository(db)

	// Initialize services
	var otpService service.OTPService
	otpService = service.NewOTPService(service.OTPConfig{
		Length:     cfg.OTP.Length,
		ExpiryMins: cfg.OTP.ExpiryMins,
	})

	var emailService service.EmailService
	emailService, err = service.NewEmailService(service.EmailConfig{
		FromEmail:     cfg.Email.FromEmail,
		FromName:      cfg.Email.FromName,
		OTPExpiryMins: cfg.Email.OTPExpiryMins,
		IsDevelopment: cfg.Email.IsDevelopment,
	})
	if err != nil {
		appLogger.Fatal("Failed to initialize email service", appLogger.Field("error", err.Error()))
		return nil, fmt.Errorf("failed to initialize email service: %w", err)
	}

	var securityService service.SecurityService
	securityService = service.NewSecurityService(service.SecurityConfig{
		BcryptCost:       cfg.Security.BcryptCost,
		MinPasswordChars: cfg.Security.MinPasswordChars,
	})

	// Use no-op metrics service
	var metricsService service.MetricsService
	metricsService = service.NewNoOpMetricsService()

	// Initialize auth service
	var authService service.AuthService
	authService = service.NewAuthService(
		userRepo,
		otpService,
		emailService,
		securityService,
	)

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	// Add timeout middleware
	router.Use(middleware.TimeoutMiddleware(
		time.Duration(cfg.Server.RequestTimeoutSec)*time.Second,
		appLogger,
	))

	// Add rate limiter for specific endpoints
	authRoutes := router.Group("/auth")
	authRoutes.Use(middleware.RateLimiterMiddleware(
		middleware.RateLimiterConfig{
			MaxRequestsPerMinute: cfg.RateLimiting.MaxRequestsPerMinute,
			BlockDurationMinutes: cfg.RateLimiting.BlockDurationMinutes,
		},
		appLogger,
	))

	// Initialize handlers
	authHandler := handler.NewAuthHandler(
		authService,
		securityService,
		metricsService,
		appLogger,
	)

	// Health check handler
	healthHandler := handler.NewHealthHandler(db)

	return &Container{
		Config:     cfg,
		Router:     router,
		AuthRoutes: authRoutes,
		Logger:     appLogger,

		// Services
		AuthService:     authService,
		OTPService:      otpService,
		EmailService:    emailService,
		SecurityService: securityService,
		MetricsService:  metricsService,

		// Handlers
		AuthHandler:   authHandler,
		HealthHandler: healthHandler,
	}, nil
}

// SetupRoutes configures all routes for the application
func (c *Container) SetupRoutes() {
	// Register health routes at the root level
	c.HealthHandler.RegisterRoutes(c.Router)
	// Register auth routes in the auth group
	c.AuthHandler.RegisterRoutes(c.AuthRoutes)
}
