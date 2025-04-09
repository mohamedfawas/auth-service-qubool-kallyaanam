package interfaces

import "context"

// EmailData contains data needed for sending emails
type EmailData struct {
	To       string
	Subject  string
	Template string
	Data     map[string]interface{}
}

// EmailService defines the interface for email operations
type EmailService interface {
	// SendVerificationEmail sends verification email with OTP
	SendVerificationEmail(ctx context.Context, to string, otp string) error
}
