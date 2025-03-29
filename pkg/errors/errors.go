package errors

import (
	"errors"
	"fmt"
)

// Custom error types for domain-specific errors
var (
	// Database errors
	ErrDatabase  = errors.New("database error")
	ErrNotFound  = errors.New("record not found")
	ErrDuplicate = errors.New("duplicate record")

	// Authentication errors
	ErrAuthentication = errors.New("authentication error")
	ErrAuthorization  = errors.New("authorization error")

	// Validation errors
	ErrValidation = errors.New("validation error")

	// OTP errors
	ErrOTP = errors.New("OTP error")
)

// Wrap wraps an error with a message
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapWithType wraps an error with a specific error type
func WrapWithType(err error, errType error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w: %w", message, errType, err)
}

// Is reports whether any error in err's chain matches target
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// SanitizeError sanitizes error messages for external consumption
// Prevents leaking sensitive details
func SanitizeError(err error) error {
	if err == nil {
		return nil
	}

	// Check for known error types and return appropriate user-friendly messages
	switch {
	case errors.Is(err, ErrDatabase):
		return errors.New("An internal error occurred. Please try again later")
	case errors.Is(err, ErrNotFound):
		return errors.New("The requested resource was not found")
	case errors.Is(err, ErrDuplicate):
		return errors.New("A duplicate record already exists")
	default:
		// For unknown errors, return a generic message
		return errors.New("An unexpected error occurred")
	}
}
