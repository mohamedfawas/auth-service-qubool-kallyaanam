package service

import (
	"context"
	"fmt"
	"net/smtp"
	"time"

	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/config"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/repository"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/logging"
	"go.uber.org/zap"
)

// notificationService implements NotificationService.
type notificationService struct {
	config           *config.Config
	verificationRepo repository.VerificationRepository
	logger           *zap.Logger
}

// NewNotificationService creates a new notification service.
func NewNotificationService(
	config *config.Config,
	verificationRepo repository.VerificationRepository,
) NotificationService {
	return &notificationService{
		config:           config,
		verificationRepo: verificationRepo,
		logger:           logging.Logger(),
	}
}

// SendEmailVerification sends an email verification code and stores it.
func (s *notificationService) SendEmailVerification(ctx context.Context, email, code string) error {
	// Store verification code in Redis
	expiration := s.config.Auth.VerificationExpiry
	if err := s.verificationRepo.StoreEmailCode(ctx, email, code, expiration); err != nil {
		s.logger.Error("Failed to store email verification code",
			zap.String("email", email),
			zap.Error(err),
		)
		return err
	}

	// In a real implementation, you would send an actual email
	// For now, we'll just log it and return success
	s.logger.Info("Sending email verification code",
		zap.String("email", email),
		zap.String("code", code),
	)

	// If email configuration is available, send actual email
	if s.config.Email.SMTPHost != "" {
		if err := s.sendEmail(email, code); err != nil {
			s.logger.Error("Failed to send email verification code",
				zap.String("email", email),
				zap.Error(err),
			)
			return err
		}
	}

	return nil
}

// SendSMSVerification sends an SMS verification code and stores it.
func (s *notificationService) SendSMSVerification(ctx context.Context, phone, code string) error {
	// Store verification code in Redis
	expiration := s.config.Auth.VerificationExpiry
	if err := s.verificationRepo.StorePhoneCode(ctx, phone, code, expiration); err != nil {
		s.logger.Error("Failed to store phone verification code",
			zap.String("phone", phone),
			zap.Error(err),
		)
		return err
	}

	// In a real implementation, you would send an actual SMS
	// For now, we'll just log it and return success
	s.logger.Info("Sending SMS verification code",
		zap.String("phone", phone),
		zap.String("code", code),
	)

	// If SMS configuration is available, send actual SMS
	if s.config.SMS.APIKey != "" {
		if err := s.sendSMS(phone, code); err != nil {
			s.logger.Error("Failed to send SMS verification code",
				zap.String("phone", phone),
				zap.Error(err),
			)
			return err
		}
	}

	return nil
}

// sendEmail sends an actual email with the verification code.
func (s *notificationService) sendEmail(email, code string) error {
	// Authentication information
	auth := smtp.PlainAuth("", s.config.Email.SMTPUser, s.config.Email.SMTPPassword, s.config.Email.SMTPHost)

	// Compose the message
	to := []string{email}
	subject := "Qubool Kallyanam - Verify Your Email"
	body := fmt.Sprintf("Your email verification code is: %s\nThis code will expire in %s.",
		code, s.config.Auth.VerificationExpiry.String())
	message := fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", s.config.Email.FromName, s.config.Email.FromEmail, email, subject, body)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	addr := fmt.Sprintf("%s:%d", s.config.Email.SMTPHost, s.config.Email.SMTPPort)
	err := smtp.SendMail(addr, auth, s.config.Email.FromEmail, to, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// sendSMS sends an actual SMS with the verification code.
// This is a simplified implementation - in a real app, you'd use a proper SMS gateway.
func (s *notificationService) sendSMS(phone, code string) error {
	// In a real implementation, this would use an SMS provider API
	// For now, we'll just log it as if it was sent
	s.logger.Info("SMS would be sent in production",
		zap.String("phone", phone),
		zap.String("code", code),
		zap.String("message", fmt.Sprintf("Your verification code is: %s", code)),
	)

	// Simulate API call delay
	time.Sleep(100 * time.Millisecond)

	return nil
}
