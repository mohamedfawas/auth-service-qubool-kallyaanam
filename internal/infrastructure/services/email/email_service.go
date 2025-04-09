package email

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"text/template"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/interfaces"
)

// Templates
const verificationEmailTemplate = `
Hello,

Welcome to Qubool Kallyaanam! Please verify your email by entering the following OTP:

OTP: {{.OTP}}

This OTP is valid for {{.ExpiryMins}} minutes.

Thank you,
Qubool Kallyaanam Team
`

type EmailConfig struct {
	FromEmail     string
	FromName      string
	OTPExpiryMins int
	IsDevelopment bool
}

type EmailService struct {
	config    EmailConfig
	templates map[string]*template.Template
}

// NewEmailService creates a new email service
func NewEmailService(config EmailConfig) (interfaces.EmailService, error) {
	// Compile templates
	verificationTmpl, err := template.New("verification").Parse(verificationEmailTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse verification email template: %w", err)
	}

	templates := map[string]*template.Template{
		"verification": verificationTmpl,
	}

	return &EmailService{
		config:    config,
		templates: templates,
	}, nil
}

// SendVerificationEmail sends a verification email with OTP
func (s *EmailService) SendVerificationEmail(ctx context.Context, to string, otp string) error {
	// Prepare template data
	data := map[string]interface{}{
		"OTP":        otp,
		"ExpiryMins": s.config.OTPExpiryMins,
	}

	// Execute template
	var body bytes.Buffer
	if err := s.templates["verification"].Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// In development mode, just log the email content
	if s.config.IsDevelopment {
		log.Printf("=========== EMAIL WOULD BE SENT ==========")
		log.Printf("To: %s", to)
		log.Printf("From: %s <%s>", s.config.FromName, s.config.FromEmail)
		log.Printf("Subject: Verify Your Email - Qubool Kallyaanam")
		log.Printf("Body: \n%s", body.String())
		log.Printf("=========================================")
		return nil
	}

	// In production mode, this would integrate with an actual SMTP provider
	// For now, we just log and return success
	return nil
}
