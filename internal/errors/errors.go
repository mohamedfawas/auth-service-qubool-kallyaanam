// Package apperror provides a simplified error handling system for web applications.
package errors

import (
	"errors"
	"net/http"
)

// Common error types that can be used with errors.Is()
var (
	ErrValidation   = errors.New("validation error")
	ErrNotFound     = errors.New("not found error")
	ErrDuplicate    = errors.New("duplicate error")
	ErrUnauthorized = errors.New("unauthorized error")
	ErrInternal     = errors.New("internal error")
)

// Error represents an application error with HTTP status code and details.
type Error struct {
	// The original error being wrapped
	Cause error
	// HTTP status code (e.g., 400, 404, 500)
	Status int
	// User-friendly error message
	Message string
	// Optional additional context (key-value pairs)
	Details map[string]string
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "unknown error"
}

// Unwrap returns the underlying cause of the error for errors.Is() and errors.As()
func (e *Error) Unwrap() error {
	return e.Cause
}

// WithDetail adds a key-value detail to the error
func (e *Error) WithDetail(key, value string) *Error {
	if e.Details == nil {
		e.Details = make(map[string]string)
	}
	e.Details[key] = value
	return e
}

// ToResponse converts the error to a map suitable for JSON responses
func (e *Error) ToResponse() map[string]interface{} {
	response := map[string]interface{}{
		"error": e.Message,
	}

	if len(e.Details) > 0 {
		response["details"] = e.Details
	}

	return response
}

// New creates a basic application error
func New(cause error, status int, message string) *Error {
	if message == "" && cause != nil {
		message = cause.Error()
	}

	return &Error{
		Cause:   cause,
		Status:  status,
		Message: message,
		Details: make(map[string]string),
	}
}

// ValidationError creates a 400 Bad Request error for validation failures
func ValidationError(message string) *Error {
	return New(ErrValidation, http.StatusBadRequest, message)
}

// NotFoundError creates a 404 Not Found error
func NotFoundError(message string) *Error {
	return New(ErrNotFound, http.StatusNotFound, message)
}

// DuplicateError creates a 409 Conflict error for duplicate records
func DuplicateError(message string) *Error {
	return New(ErrDuplicate, http.StatusConflict, message)
}

// UnauthorizedError creates a 401 Unauthorized error
func UnauthorizedError(message string) *Error {
	return New(ErrUnauthorized, http.StatusUnauthorized, message)
}

// InternalError creates a 500 Internal Server Error
func InternalError(err error) *Error {
	message := "An internal server error occurred"
	if err != nil {
		message = err.Error()
	}
	return New(ErrInternal, http.StatusInternalServerError, message)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidation)
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsDuplicateError checks if an error is a duplicate error
func IsDuplicateError(err error) bool {
	return errors.Is(err, ErrDuplicate)
}

// IsUnauthorizedError checks if an error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsInternalError checks if an error is an internal error
func IsInternalError(err error) bool {
	return errors.Is(err, ErrInternal)
}
