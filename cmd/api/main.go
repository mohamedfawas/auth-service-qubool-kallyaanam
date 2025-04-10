// cmd/api/main.go
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/config"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/handler"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/middleware"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/repository"
	postgreRepo "github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/repository/postgres"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/service"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/util/logger"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/pkg/database"
)

func createDBIfNotExists() error {
	// Connect to default postgres database first
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "postgres"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "auth_db"
	}

	// Check if database exists
	var count int64
	db.Raw("SELECT COUNT(*) FROM pg_database WHERE datname = ?", dbName).Scan(&count)

	// Create database if it doesn't exist
	if count == 0 {
		log.Printf("Creating database: %s", dbName)
		createSQL := fmt.Sprintf("CREATE DATABASE %s", dbName)
		if err := db.Exec(createSQL).Error; err != nil {
			return err
		}
	}

	return nil
}

func main() {
	// Create database if it doesn't exist
	if err := createDBIfNotExists(); err != nil {
		log.Printf("Error creating database: %v", err)
	}

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

	// Setup routes
	healthHandler.RegisterRoutes(router)
	authHandler.RegisterRoutes(authRoutes)

	// Start the server
	appLogger.Info("Starting auth service", appLogger.Field("port", cfg.Server.Port))
	if err := router.Run(fmt.Sprintf(":%s", cfg.Server.Port)); err != nil {
		appLogger.Fatal("Failed to start server", appLogger.Field("error", err.Error()))
	}
}
