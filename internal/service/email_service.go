// internal/service/email_service.go
package service

import "context"

// EmailData contains data needed for sending emails
type EmailData struct {
	To       string
	Subject  string
	Template string
	Data     map[string]interface{}
}

// EmailConfig holds email service configuration
type EmailConfig struct {
	FromEmail     string
	FromName      string
	OTPExpiryMins int
	IsDevelopment bool
}

// Implementation of the EmailService interface
type emailService struct {
	config EmailConfig
}

// NewEmailService creates a new email service instance
func NewEmailService(config EmailConfig) (EmailService, error) {
	return &emailService{
		config: config,
	}, nil
}

// SendVerificationEmail sends an email with verification OTP
func (s *emailService) SendVerificationEmail(ctx context.Context, to string, otp string) error {
	// In development mode, just log the OTP
	if s.config.IsDevelopment {
		// Log the OTP instead of sending an email
		// In a real implementation, you'd use a proper logger
		return nil
	}

	// TODO: Implement actual email sending
	// This is typically done using a library like mailgun, sendgrid, etc.
	return nil
}
