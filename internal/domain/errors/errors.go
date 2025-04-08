package errors

import "errors"

// Domain errors
var (
	ErrEmailAlreadyExists     = errors.New("email already exists")
	ErrPhoneAlreadyExists     = errors.New("phone number already exists")
	ErrInvalidEmail           = errors.New("invalid email format")
	ErrInvalidPhone           = errors.New("invalid phone format")
	ErrWeakPassword           = errors.New("password does not meet strength requirements")
	ErrRequiredFieldMissing   = errors.New("required field is missing")
	ErrUserRegistrationFailed = errors.New("user registration failed")
)
