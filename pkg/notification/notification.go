package notification

import (
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/notification/email"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/notification/sms"
)

// OTPSender defines the interface for sending OTPs
type OTPSender interface {
	SendOTP(to, otp string) error
}

// Service is a combined notification service that can send both emails and SMS
type Service struct {
	EmailService *email.Service
	SMSService   *sms.Service
}

// NewService creates a new notification service
func NewService(emailConfig email.Config, smsConfig sms.Config) *Service {
	return &Service{
		EmailService: email.NewService(emailConfig),
		SMSService:   sms.NewService(smsConfig),
	}
}

// SendEmailOTP sends an OTP via email
func (s *Service) SendEmailOTP(email, otp string) error {
	return s.EmailService.SendOTP(email, otp)
}

// SendSMSOTP sends an OTP via SMS
func (s *Service) SendSMSOTP(phone, otp string) error {
	return s.SMSService.SendOTP(phone, otp)
}
