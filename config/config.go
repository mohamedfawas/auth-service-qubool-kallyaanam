// config/config.go
package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	Email        EmailConfig
	OTP          OTPConfig
	Security     SecurityConfig
	RateLimiting RateLimitingConfig
	Logging      LoggingConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port              string
	RequestTimeoutSec int
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	DSN string
}

// EmailConfig holds email service configuration
type EmailConfig struct {
	FromEmail     string
	FromName      string
	OTPExpiryMins int
	IsDevelopment bool
}

// OTPConfig holds OTP service configuration
type OTPConfig struct {
	Length     int
	ExpiryMins int
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	BcryptCost       int
	MinPasswordChars int
}

// RateLimitingConfig holds rate limiting configuration
type RateLimitingConfig struct {
	MaxRequestsPerMinute int
	BlockDurationMinutes int
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	IsDevelopment bool
}

// NewConfig creates and initializes a new Config instance
func NewConfig() *Config {
	// Load environment variables if .env file exists
	// This is kept in the config initialization for backward compatibility
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port:              getEnv("PORT", "8081"),
			RequestTimeoutSec: getIntEnv("REQUEST_TIMEOUT_SECONDS", 30),
		},
		Database: DatabaseConfig{
			DSN: getEnv("DATABASE_URL", "host=localhost user=postgres password=postgres dbname=auth_service port=5432 sslmode=disable TimeZone=UTC"),
		},
		Email: EmailConfig{
			FromEmail:     getEnv("EMAIL_FROM_ADDRESS", "noreply@quboolkallyaanam.com"),
			FromName:      getEnv("EMAIL_FROM_NAME", "Qubool Kallyaanam"),
			OTPExpiryMins: getIntEnv("OTP_EXPIRY_MINS", 15),
			IsDevelopment: getEnv("APP_ENV", "development") == "development",
		},
		OTP: OTPConfig{
			Length:     getIntEnv("OTP_LENGTH", 6),
			ExpiryMins: getIntEnv("OTP_EXPIRY_MINS", 15),
		},
		Security: SecurityConfig{
			BcryptCost:       getIntEnv("BCRYPT_COST", 12),
			MinPasswordChars: getIntEnv("MIN_PASSWORD_CHARS", 8),
		},
		RateLimiting: RateLimitingConfig{
			MaxRequestsPerMinute: getIntEnv("RATE_LIMIT_MAX_REQUESTS", 5),
			BlockDurationMinutes: getIntEnv("RATE_LIMIT_BLOCK_DURATION", 30),
		},
		Logging: LoggingConfig{
			IsDevelopment: getEnv("APP_ENV", "development") == "development",
		},
	}
}

// Helper function to get an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// Helper function to get an integer environment variable with a fallback value
func getIntEnv(key string, fallback int) int {
	strValue := getEnv(key, "")
	if strValue == "" {
		return fallback
	}

	intValue, err := strconv.Atoi(strValue)
	if err != nil {
		return fallback
	}
	return intValue
}

// Helper function to get a boolean environment variable with a fallback value
func getBoolEnv(key string, fallback bool) bool {
	strValue := getEnv(key, "")
	if strValue == "" {
		return fallback
	}

	return strings.ToLower(strValue) == "true"
}
