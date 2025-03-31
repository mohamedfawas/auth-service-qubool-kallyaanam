// Package config contains configuration structures and loading functions.
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config is the main configuration structure.
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Auth       AuthConfig       `mapstructure:"auth"`
	Email      EmailConfig      `mapstructure:"email"`
	SMS        SMSConfig        `mapstructure:"sms"`
	Validation ValidationConfig `mapstructure:"validation"`
}

// ServerConfig contains server-specific configuration.
type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	Env          string        `mapstructure:"env"`
}

// DatabaseConfig contains database connection configuration.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	MaxIdle  int    `mapstructure:"max_idle_connections"`
	MaxOpen  int    `mapstructure:"max_open_connections"`
}

// RedisConfig contains Redis connection configuration.
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// AuthConfig contains authentication configuration.
type AuthConfig struct {
	JWTSecret          string        `mapstructure:"jwt_secret"`
	AccessTokenExpiry  time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry"`
	RegistrationExpiry time.Duration `mapstructure:"registration_expiry"`
	VerificationExpiry time.Duration `mapstructure:"verification_expiry"`
	BcryptCost         int           `mapstructure:"bcrypt_cost"`
	CookieSecure       bool          `mapstructure:"cookie_secure"`
	CookieDomain       string        `mapstructure:"cookie_domain"`
	EnableRateLimiting bool          `mapstructure:"enable_rate_limiting"`
	RateLimitPerMinute int           `mapstructure:"rate_limit_per_minute"`
}

// EmailConfig contains email service configuration.
type EmailConfig struct {
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port"`
	SMTPUser     string `mapstructure:"smtp_user"`
	SMTPPassword string `mapstructure:"smtp_password"`
	FromEmail    string `mapstructure:"from_email"`
	FromName     string `mapstructure:"from_name"`
}

// SMSConfig contains SMS service configuration.
type SMSConfig struct {
	Provider  string `mapstructure:"provider"`
	APIKey    string `mapstructure:"api_key"`
	APISecret string `mapstructure:"api_secret"`
	FromPhone string `mapstructure:"from_phone"`
}

// ValidationConfig contains validation rules.
type ValidationConfig struct {
	PasswordMinLength int `mapstructure:"password_min_length"`
	EmailCodeLength   int `mapstructure:"email_code_length"`
	PhoneCodeLength   int `mapstructure:"phone_code_length"`
}

// LoadConfig loads configuration from file and environment variables.
func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.AutomaticEnv()

	// Default values
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, using defaults and env variables
			return loadFromEnv(v)
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for configuration.
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.read_timeout", 10*time.Second)
	v.SetDefault("server.write_timeout", 10*time.Second)
	v.SetDefault("server.idle_timeout", 120*time.Second)

	// Auth defaults
	v.SetDefault("auth.access_token_expiry", 1*time.Hour)
	v.SetDefault("auth.refresh_token_expiry", 7*24*time.Hour)
	v.SetDefault("auth.registration_expiry", 24*time.Hour)
	v.SetDefault("auth.verification_expiry", 15*time.Minute)
	v.SetDefault("auth.bcrypt_cost", 12)
	v.SetDefault("auth.cookie_secure", true)
	v.SetDefault("auth.enable_rate_limiting", true)
	v.SetDefault("auth.rate_limit_per_minute", 10)

	// Validation defaults
	v.SetDefault("validation.password_min_length", 8)
	v.SetDefault("validation.email_code_length", 6)
	v.SetDefault("validation.phone_code_length", 6)

	// Database defaults
	v.SetDefault("database.max_idle_connections", 10)
	v.SetDefault("database.max_open_connections", 100)
	v.SetDefault("database.sslmode", "disable")
}

// loadFromEnv loads configuration from environment variables.
func loadFromEnv(v *viper.Viper) (*Config, error) {
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}
	return &config, nil
}
