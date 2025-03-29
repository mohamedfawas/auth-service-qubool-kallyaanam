// @title Auth Service API
// @version 1.0
// @description Authentication service for Qubool Kallyaanam
// @contact.name API Support
// @contact.email support@example.com
// @host localhost:8080
// @BasePath /
// @schemes http https
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/api/handlers"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/config"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/middleware"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository"
	redisRepo "github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository/redis"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/service"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/postgres"
	redisClient "github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/redis"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/validator"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Setup validation
	validator.Setup()

	// Connect to database
	db, err := postgres.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Connect to Redis
	var redis *redis.Client
	var rateLimitRepo repository.RateLimitRepository

	redis, err = redisClient.Connect(&cfg.Redis)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Continuing without Redis - rate limiting and session management will be disabled")
	} else {
		rateLimitRepo = redisRepo.NewRateLimitRepository(redis)

	}
	// Create repositories
	userRepo := repository.NewUserRepository(db)
	regRepo := repository.NewRegistrationRepository(db)
	otpRepo := repository.NewOTPRepository(db)

	// Create services
	otpService := service.NewOTPService(otpRepo)
	authService := service.NewAuthService(userRepo, regRepo, otpService)
	cleanupService := service.NewCleanupService(regRepo)
	tokenService := service.NewTokenService(cfg.JWT.SecretKey, cfg.JWT.Issuer)

	// Start cleanup jobs
	cleanupService.StartCleanupJobs()
	defer cleanupService.StopCleanupJobs()

	// Create handlers
	authHandler := handlers.NewAuthHandler(db, authService, otpService, redis, tokenService)

	// Create router with default logger and recovery middleware
	router := gin.New()

	// Apply global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.CORS())
	router.Use(middleware.JWTCSRFToken())

	// Setup routes
	auth := router.Group("/auth")

	// Apply rate limiting to auth endpoints - 5 requests per minute
	if rateLimitRepo != nil {
		auth.Use(middleware.RateLimiter(rateLimitRepo, "auth", 5, time.Minute))
	}

	// Apply CSRF protection to all non-GET auth endpoints
	auth.Use(middleware.JWTCSRFProtection())

	// Set up routes with the auth group
	auth.POST("/register", authHandler.Register)
	auth.POST("/verify-otp", authHandler.VerifyOTP)
	auth.POST("/complete-registration", authHandler.CompleteRegistration)
	auth.POST("/refresh-token", authHandler.RefreshToken)

	healthHandler := handlers.NewHealthHandler(db, redis)
	router.GET("/health", healthHandler.Health)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting server on %s", serverAddr)

	go func() {
		if err := router.Run(serverAddr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server...")
}
