package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the server
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	OTP      OTPConfig
	JWT      JWTConfig // Add this field
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port string
	Mode string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// OTPConfig holds OTP configuration
type OTPConfig struct {
	Length         int
	ExpiryMinutes  int
	MaxAttempts    int
	EmailOTPPrefix string
	PhoneOTPPrefix string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey string
	Issuer    string
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	otpLength, _ := strconv.Atoi(getEnv("OTP_LENGTH", "6"))
	otpExpiryMinutes, _ := strconv.Atoi(getEnv("OTP_EXPIRY_MINUTES", "15"))
	otpMaxAttempts, _ := strconv.Atoi(getEnv("OTP_MAX_ATTEMPTS", "3"))

	jwtSecret := getEnv("JWT_SECRET_KEY", "")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET_KEY environment variable is required")
	}

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "auth_service"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		OTP: OTPConfig{
			Length:         otpLength,
			ExpiryMinutes:  otpExpiryMinutes,
			MaxAttempts:    otpMaxAttempts,
			EmailOTPPrefix: "email_otp:",
			PhoneOTPPrefix: "phone_otp:",
		},
		JWT: JWTConfig{
			SecretKey: jwtSecret,
			Issuer:    getEnv("JWT_ISSUER", "qubool-kallyaanam-auth"),
		},
	}

	// Validate the config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// DSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// RedisAddr returns the Redis address (host:port)
func (c *RedisConfig) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// Helper function to get environment variables
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Get environment variable as bool
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, strconv.FormatBool(defaultValue))
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.JWT.SecretKey == "" {
		return errors.New("JWT secret key is required")
	}

	if c.Server.Port == "" {
		return errors.New("server port is required")
	}

	if c.Database.Host == "" || c.Database.Port == "" ||
		c.Database.User == "" || c.Database.DBName == "" {
		return errors.New("database configuration is incomplete")
	}

	return nil
}
