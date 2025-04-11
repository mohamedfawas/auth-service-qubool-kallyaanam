package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger with additional methods
type Logger struct {
	*zap.Logger
}

// NewLogger creates a new Logger instance
func NewLogger(isDevelopment bool) (*Logger, error) {
	var config zap.Config

	if isDevelopment {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err := config.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &Logger{Logger: logger}, nil
}

// SecurityEvent logs a security-related event with high visibility
func (l *Logger) SecurityEvent(message string, fields ...zap.Field) {
	l.With(zap.String("event_type", "security")).Warn(message, fields...)
}

// RegistrationAttempt logs a registration attempt
func (l *Logger) RegistrationAttempt(email string, ipAddress string, userAgent string, fields ...zap.Field) {
	// Mask email for privacy (show only first 3 chars and domain)
	maskedEmail := maskEmail(email)

	l.Info("Registration attempt",
		append(fields,
			zap.String("event", "registration_attempt"),
			zap.String("email", maskedEmail),
			zap.String("ip_address", ipAddress),
			zap.String("user_agent", userAgent),
		)...,
	)
}

// RegistrationSuccess logs a successful registration
func (l *Logger) RegistrationSuccess(id string, email string, ipAddress string, fields ...zap.Field) {
	maskedEmail := maskEmail(email)

	l.Info("Registration successful",
		append(fields,
			zap.String("event", "registration_success"),
			zap.String("id", id),
			zap.String("email", maskedEmail),
			zap.String("ip_address", ipAddress),
		)...,
	)
}

// RegistrationFailure logs a failed registration
func (l *Logger) RegistrationFailure(email string, ipAddress string, reason string, fields ...zap.Field) {
	maskedEmail := maskEmail(email)

	l.Warn("Registration failed",
		append(fields,
			zap.String("event", "registration_failure"),
			zap.String("email", maskedEmail),
			zap.String("ip_address", ipAddress),
			zap.String("reason", reason),
		)...,
	)
}

// Add these methods to the Logger struct
func (l *Logger) VerificationAttempt(email, ip, userAgent string) {
	l.Info("Email verification attempt",
		zap.String("email", email),
		zap.String("ip", ip),
		zap.String("userAgent", userAgent))
}

func (l *Logger) VerificationSuccess(email, ip string) {
	l.Info("Email verification successful",
		zap.String("email", email),
		zap.String("ip", ip))
}

func (l *Logger) VerificationFailure(email, ip, reason string, fields ...zap.Field) {
	allFields := []zap.Field{
		zap.String("email", email),
		zap.String("ip", ip),
		zap.String("reason", reason),
	}
	allFields = append(allFields, fields...)
	l.Error("Email verification failed", allFields...)
}

// maskEmail partially masks an email for privacy
func maskEmail(email string) string {
	// Only implement basic masking in this version
	if len(email) <= 3 {
		return "***@***"
	}

	atIndex := -1
	for i, char := range email {
		if char == '@' {
			atIndex = i
			break
		}
	}

	if atIndex <= 3 || atIndex == -1 {
		return "***@***"
	}

	return email[:3] + "***" + email[atIndex:]
}
func (l *Logger) Field(key string, value interface{}) zap.Field {
	switch v := value.(type) {
	case string:
		return zap.String(key, v)
	case int:
		return zap.Int(key, v)
	case bool:
		return zap.Bool(key, v)
	case error:
		return zap.Error(v)
	default:
		return zap.Any(key, v)
	}
}

// Warn logs a message at warn level
func (l *Logger) Warn(message string, fields ...zap.Field) {
	l.Logger.Warn(message, fields...)
}

// Fatal logs a message at fatal level and then calls os.Exit(1)
func (l *Logger) Fatal(message string, fields ...zap.Field) {
	l.Logger.Fatal(message, fields...)
}

// Info logs a message at info level
func (l *Logger) Info(message string, fields ...zap.Field) {
	l.Logger.Info(message, fields...)
}
