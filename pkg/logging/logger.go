// Package logging provides logging utilities.
package logging

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log is the global logger instance
	Log *zap.Logger
	// Sugar is the global sugared logger instance
	Sugar *zap.SugaredLogger
	once  sync.Once
)

// Initialize sets up the logger for the application.
func Initialize(environment string) {
	once.Do(func() {
		var config zap.Config

		if environment == "production" {
			config = zap.NewProductionConfig()
			config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		} else {
			config = zap.NewDevelopmentConfig()
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}

		var err error
		Log, err = config.Build()
		if err != nil {
			// If logger initialization fails, set up a basic logger
			// that writes to stderr
			Log = zap.New(
				zapcore.NewCore(
					zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
					zapcore.AddSync(os.Stderr),
					zapcore.DebugLevel,
				),
			)
			Log.Error("Failed to initialize logger, using fallback", zap.Error(err))
		}

		Sugar = Log.Sugar()

		// Make sure to sync on application exit
		zap.RedirectStdLog(Log)
	})
}

// Logger returns the global logger instance.
func Logger() *zap.Logger {
	if Log == nil {
		Initialize("development")
	}
	return Log
}

// SugaredLogger returns the global sugared logger instance.
func SugaredLogger() *zap.SugaredLogger {
	if Sugar == nil {
		Initialize("development")
	}
	return Sugar
}

// With creates a child logger with additional fields.
func With(fields ...zap.Field) *zap.Logger {
	return Logger().With(fields...)
}

// WithFields creates a child sugared logger with additional fields.
func WithFields(fields map[string]interface{}) *zap.SugaredLogger {
	return SugaredLogger().With(fields)
}
