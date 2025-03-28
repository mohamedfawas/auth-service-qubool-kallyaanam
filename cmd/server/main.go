package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/api/handlers"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/api/routes"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/config"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/postgres"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Connect to database
	db, err := postgres.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create router
	router := gin.Default()

	// Create handlers
	authHandler := handlers.NewAuthHandler(db)

	// Setup routes
	routes.Setup(router, authHandler)

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting server on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
