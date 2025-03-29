package postgres

import (
	"log"
	"time"

	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/config"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/domain/models"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository/migrations"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes a connection to the database
func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto migrate schemas
	err = db.AutoMigrate(&models.PendingRegistration{}, &models.VerificationOTP{}, &models.User{})
	if err != nil {
		log.Printf("Failed to migrate database: %v", err)
		return nil, err
	}

	// Run custom migrations
	if err := migrations.AddIndexes(db); err != nil {
		log.Printf("Warning: Failed to apply custom migrations: %v", err)
		// Continue anyway as the indexes are also defined in the model
	}

	return db, nil
}
