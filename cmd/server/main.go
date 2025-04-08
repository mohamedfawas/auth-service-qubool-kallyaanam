package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/usecase"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/database"
	gormRepo "github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/repository/gorm"
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

	// Initialize database connection
	db, err := database.NewDatabase(dbDSN)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repositories
	userRepo := gormRepo.NewUserRepository(db)

	// Initialize use cases
	registrationUseCase := usecase.NewRegistrationUseCase(userRepo)

	// Initialize Gin router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, registrationUseCase)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Start the server
	log.Printf("Starting auth service on port %s", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
