package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/config"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/handlers/http/user"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/models"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository/postgres"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository/redis"
	userService "github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/service/user"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/database"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/notification"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/notification/email"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/notification/sms"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to PostgreSQL database
	pgConfig := database.PostgresConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}

	db, err := database.NewPostgresConnection(pgConfig)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Connect to Redis
	redisConfig := database.RedisConfig{
		Host:     cfg.RedisHost,
		Port:     cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	redisClient, err := database.NewRedisClient(redisConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Auto migrate database schemas
	if err := db.AutoMigrate(&models.User{}, &models.PendingRegistration{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize notification services
	emailConfig := email.Config{
		Host:     cfg.EmailHost,
		Port:     cfg.EmailPort,
		Username: cfg.EmailUsername,
		Password: cfg.EmailPassword,
		From:     cfg.EmailFrom,
	}

	smsConfig := sms.Config{
		Provider:  "twilio", // or "vonage" depending on your preference
		AccountID: cfg.SMSProviderID,
		Token:     cfg.SMSProviderToken,
		From:      cfg.SMSProviderFrom,
	}
	notificationService := notification.NewService(emailConfig, smsConfig)

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	otpRepo := redis.NewOTPRepository(redisClient)

	// Initialize services
	authService := userService.NewAuthService(*userRepo, *otpRepo, notificationService)

	// Initialize handlers
	userHandler := user.NewHandler(authService)

	// Setup Gin router
	router := gin.Default()

	// Setup API routes
	api := router.Group("/api")
	userHandler.SetupRoutes(api)

	// Start the server
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
