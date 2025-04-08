package dto

// RegistrationRequest defines the input data for registration
type RegistrationRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// RegistrationResponse defines the response data after registration
type RegistrationResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}
