// cmd/api/main.go
package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/di"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/pkg/migrations"
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

	// Run database migrations
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	if err := migrations.RunMigrations(dsn); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize the dependency injection container
	container, err := di.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize DI container: %v", err)
	}

	// Setup routes
	container.SetupRoutes()

	// Start the server
	container.Logger.Info("Starting auth service", container.Logger.Field("port", container.Config.Server.Port))
	if err := container.Router.Run(fmt.Sprintf(":%s", container.Config.Server.Port)); err != nil {
		container.Logger.Fatal("Failed to start server", container.Logger.Field("error", err.Error()))
	}
}
