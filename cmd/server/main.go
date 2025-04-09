package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/usecase"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/database"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/middleware"
	gormRepo "github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/repository/gorm"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/services/email"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/services/logger"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/services/metrics"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/services/otp"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/services/security"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/interfaces/http/handlers"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/interfaces/http/routes"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get environment settings
	isDev := getEnv("APP_ENV", "development") == "development"

	// Initialize logger
	appLogger, err := logger.NewLogger(isDev)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Get database connection string from environment
	dbDSN := getEnv("DATABASE_URL", "host=localhost user=postgres password=postgres dbname=auth_service port=5432 sslmode=disable TimeZone=UTC")

	// Get OTP config from environment
	otpLength, _ := strconv.Atoi(getEnv("OTP_LENGTH", "6"))
	otpExpiryMins, _ := strconv.Atoi(getEnv("OTP_EXPIRY_MINS", "15"))

	// Get security config from environment
	bcryptCost, _ := strconv.Atoi(getEnv("BCRYPT_COST", "12"))
	minPasswordChars, _ := strconv.Atoi(getEnv("MIN_PASSWORD_CHARS", "8"))

	// Get rate limiting config
	maxRequestsPerMin, _ := strconv.Atoi(getEnv("RATE_LIMIT_MAX_REQUESTS", "5"))
	blockDurationMins, _ := strconv.Atoi(getEnv("RATE_LIMIT_BLOCK_DURATION", "30"))

	// Get request timeout setting
	requestTimeoutSec, _ := strconv.Atoi(getEnv("REQUEST_TIMEOUT_SECONDS", "30"))

	// Initialize database connection
	db, err := database.NewDatabase(dbDSN)
	if err != nil {
		appLogger.Fatal("Failed to initialize database", appLogger.Field("error", err.Error()))
	}

	// Initialize repositories
	userRepo := gormRepo.NewUserRepository(db)

	// Initialize services
	otpService := otp.NewOTPService(otp.OTPConfig{
		Length:     otpLength,
		ExpiryMins: otpExpiryMins,
	})

	emailService, err := email.NewEmailService(email.EmailConfig{
		FromEmail:     getEnv("EMAIL_FROM_ADDRESS", "noreply@quboolkallyaanam.com"),
		FromName:      getEnv("EMAIL_FROM_NAME", "Qubool Kallyaanam"),
		OTPExpiryMins: otpExpiryMins,
		IsDevelopment: isDev,
	})
	if err != nil {
		appLogger.Fatal("Failed to initialize email service", appLogger.Field("error", err.Error()))
	}

	securityService := security.NewSecurityService(security.SecurityConfig{
		BcryptCost:       bcryptCost,
		MinPasswordChars: minPasswordChars,
	})

	metricsService := metrics.NewPrometheusMetrics()

	// Initialize use cases
	registrationUseCase := usecase.NewRegistrationUseCase(
		userRepo,
		otpService,
		emailService,
		securityService,
	)

	// Initialize Gin router
	router := gin.New() // Don't use Default() as we're adding our own middleware

	// Add middlewares
	router.Use(gin.Recovery())

	// Add timeout middleware
	router.Use(middleware.TimeoutMiddleware(
		time.Duration(requestTimeoutSec)*time.Second,
		appLogger,
	))

	// Add rate limiter for specific endpoints
	authRoutes := router.Group("/auth")
	authRoutes.Use(middleware.RateLimiterMiddleware(
		middleware.RateLimiterConfig{
			MaxRequestsPerMinute: maxRequestsPerMin,
			BlockDurationMinutes: blockDurationMins,
		},
		appLogger,
	))

	// Initialize handlers
	registrationHandler := handlers.NewRegistrationHandler(
		registrationUseCase,
		securityService,
		metricsService,
		appLogger,
	)

	// Setup routes
	routes.SetupRoutes(router, authRoutes, registrationHandler)

	// Register metrics endpoint
	metrics.RegisterHandler(router)

	// Get port from environment
	port := getEnv("PORT", "8081")

	// Start the server
	appLogger.Info("Starting auth service", appLogger.Field("port", port))
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		appLogger.Fatal("Failed to start server", appLogger.Field("error", err.Error()))
	}
}

// Helper function to get environment variables with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
