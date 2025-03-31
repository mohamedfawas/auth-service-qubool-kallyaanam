// Package dto contains Data Transfer Objects for API communication.
package dto

// Response is the standard API response structure.
type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// NewSuccessResponse creates a new Response with success status.
func NewSuccessResponse(status int, message string, data interface{}) Response {
	return Response{
		Status:  status,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates a new Response with error status.
func NewErrorResponse(status int, message string, errDetails interface{}) Response {
	return Response{
		Status:  status,
		Message: message,
		Error:   errDetails,
	}
}

// RegisterResponse contains the data returned after successful registration initiation.
type RegisterResponse struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// VerificationStatusResponse contains the verification status of email and phone.
type VerificationStatusResponse struct {
	EmailVerified bool `json:"emailVerified"`
	PhoneVerified bool `json:"phoneVerified"`
}
