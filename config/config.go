// config/config.go
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
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
	Redis        RedisConfig
	JWT          JWTConfig
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

	// Validate Redis config
	if err := c.Redis.Validate(); err != nil {
		return err
	}

	// Validate JWT config
	if err := c.JWT.Validate(); err != nil {
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
	BcryptCost                   int    `mapstructure:"bcrypt_cost"`
	MinPasswordChars             int    `mapstructure:"min_password_chars"`
	JWTSecret                    string `mapstructure:"jwt_secret"`
	AccessTokenExpiryMinutes     int    `mapstructure:"access_token_expiry_minutes"`
	RefreshTokenExpiryHours      int    `mapstructure:"refresh_token_expiry_hours"`
	TokenIssuer                  string `mapstructure:"token_issuer"`
	LoginAttemptsThreshold       int    `mapstructure:"login_attempts_threshold"`
	LoginThrottleDurationMinutes int    `mapstructure:"login_throttle_duration_minutes"`
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

// RedisConfig holds configuration for Redis connection
type RedisConfig struct {
	Address  string
	Password string
	DB       int
	Enabled  bool
}

// Parse parses a Redis URI into RedisConfig
func (c *RedisConfig) Parse(uri string) error {
	if uri == "" {
		return nil
	}

	// Parse the Redis URI
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}

	// Extract host and port
	c.Address = u.Host

	// Extract password if present
	if u.User != nil {
		c.Password, _ = u.User.Password()
	}

	// Extract database number
	if len(u.Path) > 1 {
		dbStr := strings.TrimPrefix(u.Path, "/")
		db, err := strconv.Atoi(dbStr)
		if err == nil {
			c.DB = db
		}
	}

	c.Enabled = true
	return nil
}

// Validate checks if the redis configuration is valid
func (c *RedisConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.Address == "" {
		return &ValidationError{Field: "Redis.Address", Message: "cannot be empty when Redis is enabled"}
	}

	return nil
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	TokenExpiry   time.Duration `mapstructure:"token_expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry"`
	Issuer        string        `mapstructure:"issuer"`
}

// Validate checks if JWT configuration is valid
func (c *JWTConfig) Validate() error {
	if c.Secret == "" {
		return &ValidationError{Field: "JWT.Secret", Message: "cannot be empty"}
	}

	return nil
}

// LoadConfig loads configuration using Viper
func LoadConfig() (*Config, error) {
	// Load environment variables from .env file if it exists
	_ = godotenv.Load()

	// Initialize Viper
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AutomaticEnv()

	// Read the config file if it exists (optional)
	if err := v.ReadInConfig(); err != nil {
		// We'll just use environment variables if the config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Server config
	v.SetDefault("PORT", "8081")
	v.SetDefault("REQUEST_TIMEOUT_SECONDS", 30)

	// Database config
	v.SetDefault("DATABASE_URL", "host=localhost user=postgres password=postgres dbname=auth_service port=5432 sslmode=disable TimeZone=UTC")

	// Email config
	v.SetDefault("EMAIL_FROM_ADDRESS", "noreply@quboolkallyaanam.com")
	v.SetDefault("EMAIL_FROM_NAME", "Qubool Kallyaanam")
	v.SetDefault("OTP_EXPIRY_MINS", 15)
	v.SetDefault("APP_ENV", "development")

	// OTP config
	v.SetDefault("OTP_LENGTH", 6)

	// Security config
	v.SetDefault("BCRYPT_COST", 12)
	v.SetDefault("MIN_PASSWORD_CHARS", 8)
	v.SetDefault("ACCESS_TOKEN_EXPIRY_MINUTES", 15)
	v.SetDefault("REFRESH_TOKEN_EXPIRY_HOURS", 24)
	v.SetDefault("TOKEN_ISSUER", "qubool-kallyaanam-auth")
	v.SetDefault("LOGIN_ATTEMPTS_THRESHOLD", 5)
	v.SetDefault("LOGIN_THROTTLE_DURATION_MINUTES", 15)

	// Rate limiting config
	v.SetDefault("RATE_LIMIT_MAX_REQUESTS", 5)
	v.SetDefault("RATE_LIMIT_BLOCK_DURATION", 30)

	// Logging config
	v.SetDefault("LOG_LEVEL", "info")

	// Redis config
	v.SetDefault("REDIS_ADDRESS", "localhost:6379")
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)

	// JWT config
	v.SetDefault("JWT_TOKEN_EXPIRY", "15m")
	v.SetDefault("JWT_REFRESH_EXPIRY", "24h")
	v.SetDefault("JWT_ISSUER", "qubool-kallyaanam-auth")

	// Create Redis config
	redisConfig := RedisConfig{
		Address:  v.GetString("REDIS_ADDRESS"),
		Password: v.GetString("REDIS_PASSWORD"),
		DB:       v.GetInt("REDIS_DB"),
		Enabled:  true,
	}

	// Try to parse from REDIS_URI if available
	redisURI := v.GetString("REDIS_URI")
	if redisURI != "" {
		if err := redisConfig.Parse(redisURI); err != nil {
			// Just log the error, don't return
			fmt.Printf("Failed to parse REDIS_URI: %v, using default settings\n", err)
		}
	}

	// Parse token expiry durations
	tokenExpiry, err := time.ParseDuration(v.GetString("JWT_TOKEN_EXPIRY"))
	if err != nil {
		tokenExpiry = 15 * time.Minute
	}

	refreshExpiry, err := time.ParseDuration(v.GetString("JWT_REFRESH_EXPIRY"))
	if err != nil {
		refreshExpiry = 24 * time.Hour
	}

	// Create config with defaults and environment variable overrides
	config := &Config{
		Server: ServerConfig{
			Port:              v.GetString("PORT"),
			RequestTimeoutSec: v.GetInt("REQUEST_TIMEOUT_SECONDS"),
		},
		Database: DatabaseConfig{
			DSN: v.GetString("DATABASE_URL"),
		},
		Email: EmailConfig{
			FromEmail:     v.GetString("EMAIL_FROM_ADDRESS"),
			FromName:      v.GetString("EMAIL_FROM_NAME"),
			OTPExpiryMins: v.GetInt("OTP_EXPIRY_MINS"),
			IsDevelopment: v.GetString("APP_ENV") == "development",
		},
		OTP: OTPConfig{
			Length:     v.GetInt("OTP_LENGTH"),
			ExpiryMins: v.GetInt("OTP_EXPIRY_MINS"),
		},
		Security: SecurityConfig{
			BcryptCost:                   v.GetInt("BCRYPT_COST"),
			MinPasswordChars:             v.GetInt("MIN_PASSWORD_CHARS"),
			JWTSecret:                    v.GetString("JWT_SECRET"),
			AccessTokenExpiryMinutes:     v.GetInt("ACCESS_TOKEN_EXPIRY_MINUTES"),
			RefreshTokenExpiryHours:      v.GetInt("REFRESH_TOKEN_EXPIRY_HOURS"),
			TokenIssuer:                  v.GetString("TOKEN_ISSUER"),
			LoginAttemptsThreshold:       v.GetInt("LOGIN_ATTEMPTS_THRESHOLD"),
			LoginThrottleDurationMinutes: v.GetInt("LOGIN_THROTTLE_DURATION_MINUTES"),
		},
		RateLimiting: RateLimitingConfig{
			MaxRequestsPerMinute: v.GetInt("RATE_LIMIT_MAX_REQUESTS"),
			BlockDurationMinutes: v.GetInt("RATE_LIMIT_BLOCK_DURATION"),
		},
		Logging: LoggingConfig{
			IsDevelopment: v.GetString("APP_ENV") == "development",
			LogLevel:      v.GetString("LOG_LEVEL"),
		},
		Redis: redisConfig,
		JWT: JWTConfig{
			Secret:        v.GetString("JWT_SECRET"),
			TokenExpiry:   tokenExpiry,
			RefreshExpiry: refreshExpiry,
			Issuer:        v.GetString("JWT_ISSUER"),
		},
	}

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Helper functions below are kept for backward compatibility
// but will be deprecated in favor of Viper

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
