package dto

// Add a struct for registration request and response
type RegistrationRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type RegistrationResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// VerifyEmailRequest represents the request to verify email with OTP
type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required"`
}

// VerifyEmailResponse represents the response after email verification
type VerifyEmailResponse struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Message string `json:"message"`
}
