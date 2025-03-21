package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Server configuration
	ServerPort string

	// Redis configuration
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// Email configuration
	EmailHost     string
	EmailPort     string
	EmailUsername string
	EmailPassword string
	EmailFrom     string

	// SMS configuration
	SMSProviderID    string
	SMSProviderToken string
	SMSProviderFrom  string
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	// Parse Redis DB number
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		// Database config
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "qubool_kallyaanam"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// Server config
		ServerPort: getEnv("SERVER_PORT", "8080"),

		// Redis config
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		// Email config
		EmailHost:     getEnv("EMAIL_HOST", "smtp.gmail.com"),
		EmailPort:     getEnv("EMAIL_PORT", "587"),
		EmailUsername: getEnv("EMAIL_USERNAME", ""),
		EmailPassword: getEnv("EMAIL_PASSWORD", ""),
		EmailFrom:     getEnv("EMAIL_FROM", "noreply@quboolkallyaanam.com"),

		// SMS config
		SMSProviderID:    getEnv("SMS_PROVIDER_ID", ""),
		SMSProviderToken: getEnv("SMS_PROVIDER_TOKEN", ""),
		SMSProviderFrom:  getEnv("SMS_PROVIDER_FROM", "QuboolK"),
	}, nil
}
