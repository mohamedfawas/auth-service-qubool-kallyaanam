package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/api/handlers"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/api/routes"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/config"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository"
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
	_, err = redisClient.Connect(&cfg.Redis)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Continuing without Redis - rate limiting will be disabled")
	}

	// Create repositories
	userRepo := repository.NewUserRepository(db)
	regRepo := repository.NewRegistrationRepository(db)
	otpRepo := repository.NewOTPRepository(db)

	// Create services
	otpService := service.NewOTPService(otpRepo)
	authService := service.NewAuthService(userRepo, regRepo, otpService)
	cleanupService := service.NewCleanupService(regRepo)

	// Start cleanup jobs
	cleanupService.StartCleanupJobs()
	defer cleanupService.StopCleanupJobs()

	// Create handlers
	authHandler := handlers.NewAuthHandler(db, authService, otpService)

	// Create router
	router := gin.Default()

	// Setup routes
	routes.Setup(router, authHandler)

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
