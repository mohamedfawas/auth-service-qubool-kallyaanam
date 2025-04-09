// cmd/api/main.go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/config"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/handler"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/middleware"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/repository/postgres"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/service"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/util/logger"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/pkg/database"
)

func main() {
	// Load configuration
	cfg := config.NewConfig()

	// Initialize logger
	appLogger, err := logger.NewLogger(cfg.Logging.IsDevelopment)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize database connection
	db, err := database.NewPostgresConnection(cfg.Database.DSN)
	if err != nil {
		appLogger.Fatal("Failed to initialize database", appLogger.Field("error", err.Error()))
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)

	// Initialize services
	otpService := service.NewOTPService(service.OTPConfig{
		Length:     cfg.OTP.Length,
		ExpiryMins: cfg.OTP.ExpiryMins,
	})

	emailService, err := service.NewEmailService(service.EmailConfig{
		FromEmail:     cfg.Email.FromEmail,
		FromName:      cfg.Email.FromName,
		OTPExpiryMins: cfg.Email.OTPExpiryMins,
		IsDevelopment: cfg.Email.IsDevelopment,
	})
	if err != nil {
		appLogger.Fatal("Failed to initialize email service", appLogger.Field("error", err.Error()))
	}

	securityService := service.NewSecurityService(service.SecurityConfig{
		BcryptCost:       cfg.Security.BcryptCost,
		MinPasswordChars: cfg.Security.MinPasswordChars,
	})

	// Use no-op metrics service instead of Prometheus
	metricsService := service.NewNoOpMetricsService()

	// Initialize auth service
	authService := service.NewAuthService(
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
	healthHandler := handler.NewHealthHandler()

	// Setup routes
	healthHandler.RegisterRoutes(router)
	authHandler.RegisterRoutes(authRoutes)

	// Start the server
	appLogger.Info("Starting auth service", appLogger.Field("port", cfg.Server.Port))
	if err := router.Run(fmt.Sprintf(":%s", cfg.Server.Port)); err != nil {
		appLogger.Fatal("Failed to start server", appLogger.Field("error", err.Error()))
	}
}
