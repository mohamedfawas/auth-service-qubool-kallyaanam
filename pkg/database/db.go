// pkg/database/db.go
package database

import (
	"fmt"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresConnection creates a new connection to PostgreSQL database
func NewPostgresConnection(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(&model.PendingRegistration{}, &model.User{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
