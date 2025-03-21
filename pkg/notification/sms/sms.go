package sms

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Provider options
const (
	ProviderTwilio = "twilio"
	ProviderVonage = "vonage"
)

// Config holds the SMS provider configuration
type Config struct {
	Provider  string // "twilio", "vonage", etc.
	AccountID string // SID for Twilio, API Key for Vonage
	Token     string // Auth Token for Twilio, API Secret for Vonage
	From      string // Sender ID or phone number
}

// Service handles SMS operations
type Service struct {
	config Config
}

// NewService creates a new SMS service
func NewService(config Config) *Service {
	return &Service{
		config: config,
	}
}

// SendOTP sends an OTP to the specified phone number
func (s *Service) SendOTP(to, otp string) error {
	message := fmt.Sprintf("Your Qubool Kallyaanam verification code is: %s. Valid for 24 hours.", otp)

	// For testing, just log the message
	fmt.Printf("[SMS to %s]: %s\n", to, message)

	switch s.config.Provider {
	case ProviderTwilio:
		return s.sendTwilioSMS(to, message)
	case ProviderVonage:
		return s.sendVonageSMS(to, message)
	default:
		return fmt.Errorf("unsupported SMS provider: %s", s.config.Provider)
	}
}

// sendTwilioSMS sends an SMS using Twilio
func (s *Service) sendTwilioSMS(to, message string) error {
	// Ensure phone number has + prefix
	if !strings.HasPrefix(to, "+") {
		to = "+" + to
	}

	// Twilio API endpoint
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.config.AccountID)

	// Prepare form data
	data := url.Values{}
	data.Set("To", to)
	data.Set("From", s.config.From)
	data.Set("Body", message)

	// Create request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.SetBasicAuth(s.config.AccountID, s.config.Token)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("SMS API returned error status: %d", resp.StatusCode)
	}

	return nil
}

// sendVonageSMS sends an SMS using Vonage (formerly Nexmo)
func (s *Service) sendVonageSMS(to, message string) error {
	// Ensure phone number has + prefix
	if !strings.HasPrefix(to, "+") {
		to = "+" + to
	}

	// Vonage API endpoint
	apiURL := "https://rest.nexmo.com/sms/json"

	// Prepare form data
	data := url.Values{}
	data.Set("to", to)
	data.Set("from", s.config.From)
	data.Set("text", message)
	data.Set("api_key", s.config.AccountID)
	data.Set("api_secret", s.config.Token)

	// Create request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("SMS API returned error status: %d", resp.StatusCode)
	}

	return nil
}
