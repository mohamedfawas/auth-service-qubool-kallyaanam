// config/config.go
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("config validation error: %s - %s", e.Field, e.Message)
}

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

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate Server config
	if err := c.Server.Validate(); err != nil {
		return err
	}

	// Validate Database config
	if err := c.Database.Validate(); err != nil {
		return err
	}

	// Validate Email config
	if err := c.Email.Validate(); err != nil {
		return err
	}

	// Validate OTP config
	if err := c.OTP.Validate(); err != nil {
		return err
	}

	// Validate Security config
	if err := c.Security.Validate(); err != nil {
		return err
	}

	// Validate RateLimiting config
	if err := c.RateLimiting.Validate(); err != nil {
		return err
	}

	return nil
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port              string
	RequestTimeoutSec int
}

// Validate checks if server configuration is valid
func (c *ServerConfig) Validate() error {
	if c.Port == "" {
		return &ValidationError{Field: "Server.Port", Message: "cannot be empty"}
	}

	if c.RequestTimeoutSec <= 0 {
		return &ValidationError{Field: "Server.RequestTimeoutSec", Message: "must be greater than 0"}
	}

	return nil
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	DSN string
}

// Validate checks if database configuration is valid
func (c *DatabaseConfig) Validate() error {
	if c.DSN == "" {
		return &ValidationError{Field: "Database.DSN", Message: "cannot be empty"}
	}
	return nil
}

// EmailConfig holds email service configuration
type EmailConfig struct {
	FromEmail     string
	FromName      string
	OTPExpiryMins int
	IsDevelopment bool
}

// Validate checks if email configuration is valid
func (c *EmailConfig) Validate() error {
	if c.FromEmail == "" {
		return &ValidationError{Field: "Email.FromEmail", Message: "cannot be empty"}
	}

	if !strings.Contains(c.FromEmail, "@") {
		return &ValidationError{Field: "Email.FromEmail", Message: "invalid email format"}
	}

	if c.FromName == "" {
		return &ValidationError{Field: "Email.FromName", Message: "cannot be empty"}
	}

	if c.OTPExpiryMins <= 0 {
		return &ValidationError{Field: "Email.OTPExpiryMins", Message: "must be greater than 0"}
	}

	return nil
}

// OTPConfig holds OTP service configuration
type OTPConfig struct {
	Length     int
	ExpiryMins int
}

// Validate checks if OTP configuration is valid
func (c *OTPConfig) Validate() error {
	if c.Length <= 0 {
		return &ValidationError{Field: "OTP.Length", Message: "must be greater than 0"}
	}

	if c.ExpiryMins <= 0 {
		return &ValidationError{Field: "OTP.ExpiryMins", Message: "must be greater than 0"}
	}

	return nil
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	BcryptCost       int
	MinPasswordChars int
}

// Validate checks if security configuration is valid
func (c *SecurityConfig) Validate() error {
	if c.BcryptCost < 10 || c.BcryptCost > 31 {
		return &ValidationError{Field: "Security.BcryptCost", Message: "must be between 10 and 31"}
	}

	if c.MinPasswordChars < 8 {
		return &ValidationError{Field: "Security.MinPasswordChars", Message: "must be at least 8"}
	}

	return nil
}

// RateLimitingConfig holds rate limiting configuration
type RateLimitingConfig struct {
	MaxRequestsPerMinute int
	BlockDurationMinutes int
}

// Validate checks if rate limiting configuration is valid
func (c *RateLimitingConfig) Validate() error {
	if c.MaxRequestsPerMinute <= 0 {
		return &ValidationError{Field: "RateLimiting.MaxRequestsPerMinute", Message: "must be greater than 0"}
	}

	if c.BlockDurationMinutes <= 0 {
		return &ValidationError{Field: "RateLimiting.BlockDurationMinutes", Message: "must be greater than 0"}
	}

	return nil
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	IsDevelopment bool
	LogLevel      string
}

// Validate checks if logging configuration is valid
func (c *LoggingConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}

	if c.LogLevel != "" && !validLevels[strings.ToLower(c.LogLevel)] {
		return &ValidationError{Field: "Logging.LogLevel", Message: "invalid log level"}
	}

	return nil
}

// NewConfig creates and initializes a new Config instance
func NewConfig() (*Config, error) {
	// Load environment variables from .env file if it exists
	// Only load in development mode and don't fail if file is missing
	_ = godotenv.Load()

	// Create config with defaults and environment variable overrides
	config := &Config{
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
			LogLevel:      getEnv("LOG_LEVEL", "info"),
		},
	}

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// LoadConfig loads configuration from environment variables with validation
func LoadConfig() (*Config, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return config, nil
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

// Helper function to get a required environment variable
func getRequiredEnv(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return "", errors.New("required environment variable not set: " + key)
	}
	return value, nil
}

// Helper function to get a required integer environment variable
func getRequiredIntEnv(key string) (int, error) {
	strValue, err := getRequiredEnv(key)
	if err != nil {
		return 0, err
	}

	intValue, err := strconv.Atoi(strValue)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value for %s: %w", key, err)
	}
	return intValue, nil
}
