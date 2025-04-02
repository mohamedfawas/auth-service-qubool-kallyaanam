package postgres

import (
	"fmt"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/configs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresAdapter struct {
	DB *gorm.DB // pointer to gorm.DB instance
}

func NewPostgresAdapter(config *configs.PostgresConfig) (*PostgresAdapter, error) {
	dsn := fmt.Sprintf( // dsn is the data source name
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // open a new gorm.DB instance
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	return &PostgresAdapter{DB: db}, nil
}

func (pa *PostgresAdapter) Ping() error {
	sqlDB, err := pa.DB.DB() // get the sql.DB instance from the gorm.DB instance
	if err != nil {
		return fmt.Errorf("failed to get sql.DB instance: %w", err)
	}
	return sqlDB.Ping() // ping the database
}
