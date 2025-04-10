// auth-service-qubool-kallyaanam/internal/util/response/response.go
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// StandardResponse represents the standard API response format
type StandardResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success returns a successful response with standard format
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, StandardResponse{
		Status:  true,
		Message: message,
		Data:    data,
	})
}

// Created returns a successful creation response with standard format
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, StandardResponse{
		Status:  true,
		Message: message,
		Data:    data,
	})
}

// Error returns an error response with standard format
func Error(c *gin.Context, statusCode int, message string, err error) {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}

	c.JSON(statusCode, StandardResponse{
		Status:  false,
		Message: message,
		Error:   errorMessage,
	})
}

// BadRequest returns a 400 bad request error with standard format
func BadRequest(c *gin.Context, message string, err error) {
	Error(c, http.StatusBadRequest, message, err)
}

// Unauthorized returns a 401 unauthorized error with standard format
func Unauthorized(c *gin.Context, message string, err error) {
	Error(c, http.StatusUnauthorized, message, err)
}

// Forbidden returns a 403 forbidden error with standard format
func Forbidden(c *gin.Context, message string, err error) {
	Error(c, http.StatusForbidden, message, err)
}

// NotFound returns a 404 not found error with standard format
func NotFound(c *gin.Context, message string, err error) {
	Error(c, http.StatusNotFound, message, err)
}

// InternalServerError returns a 500 internal server error with standard format
func InternalServerError(c *gin.Context, err error) {
	Error(c, http.StatusInternalServerError, "Internal server error", err)
}
