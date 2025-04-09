package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/usecase"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/database"
	gormRepo "github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/repository/gorm"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/services/email"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/services/otp"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/interfaces/http/routes"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get database connection string from environment
	dbDSN := os.Getenv("DATABASE_URL")
	if dbDSN == "" {
		dbDSN = "host=localhost user=postgres password=postgres dbname=auth_service port=5432 sslmode=disable TimeZone=UTC"
	}

	// Get OTP config from environment
	otpLength, _ := strconv.Atoi(getEnv("OTP_LENGTH", "6"))
	otpExpiryMins, _ := strconv.Atoi(getEnv("OTP_EXPIRY_MINS", "15"))

	// Get email config from environment
	isDev := getEnv("APP_ENV", "development") == "development"

	// Initialize database connection
	db, err := database.NewDatabase(dbDSN)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
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
		log.Fatalf("Failed to initialize email service: %v", err)
	}

	// Initialize use cases
	registrationUseCase := usecase.NewRegistrationUseCase(userRepo, otpService, emailService)

	// Initialize Gin router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, registrationUseCase)

	// Get port from environment
	port := getEnv("PORT", "8081")

	// Start the server
	log.Printf("Starting auth service on port %s", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Helper function to get environment variables with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
