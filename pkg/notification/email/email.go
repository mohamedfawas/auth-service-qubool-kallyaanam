package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Config holds the SMTP configuration
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// Service handles email operations
type Service struct {
	config Config
}

// NewService creates a new email service
func NewService(config Config) *Service {
	return &Service{
		config: config,
	}
}

// SendOTP sends an OTP to the specified email address
func (s *Service) SendOTP(to, otp string) error {
	// Get the OTP email template
	body, err := s.renderOTPTemplate(to, otp)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Compose email
	subject := "Your Verification Code for Qubool Kallyaanam"
	err = s.sendEmail(to, subject, body)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// sendEmail sends an email with the provided details
func (s *Service) sendEmail(to, subject, body string) error {
	// Set up authentication
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Compose the message
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", s.config.From, to, subject, body))

	// Send the email
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	return smtp.SendMail(addr, auth, s.config.From, []string{to}, msg)
}

// renderOTPTemplate generates the HTML content for OTP emails
func (s *Service) renderOTPTemplate(to, otp string) (string, error) {
	// Parse the template
	tmpl, err := template.New("otp_email").Parse(otpEmailTemplate)
	if err != nil {
		return "", err
	}

	// Extract the first name (if available)
	name := strings.Split(to, "@")[0]
	name = strings.ReplaceAll(name, ".", " ")
	name = cases.Title(language.English).String(name)

	// Template data
	data := struct {
		Name string
		OTP  string
	}{
		Name: name,
		OTP:  otp,
	}

	// Render the template
	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
