// Package dto contains Data Transfer Objects for API communication.
package dto

// RegisterRequest represents the data required for user registration.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Phone    string `json:"phone" binding:"required,min=10,max=20"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// EmailVerificationRequest represents the data required for email verification.
type EmailVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

// PhoneVerificationRequest represents the data required for phone verification.
type PhoneVerificationRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required,len=6"`
}
