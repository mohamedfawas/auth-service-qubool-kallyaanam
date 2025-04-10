// auth-service-qubool-kallyaanam/internal/util/response/response.go
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// StandardResponse represents the standard API response format
type StandardResponse struct {
	Status  int         `json:"status"` // HTTP status code
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success returns a successful response with standard format
func Success(c *gin.Context, message string, data interface{}) {
	statusCode := http.StatusOK
	c.JSON(statusCode, StandardResponse{
		Status:  statusCode,
		Message: message,
		Data:    data,
	})
}

// Created returns a successful creation response with standard format
func Created(c *gin.Context, message string, data interface{}) {
	statusCode := http.StatusCreated
	c.JSON(statusCode, StandardResponse{
		Status:  statusCode,
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
		Status:  statusCode,
		Message: message,
		Error:   errorMessage,
	})
}

// BadRequest returns a 400 bad request error with standard format
func BadRequest(c *gin.Context, message string, err error) {
	Error(c, http.StatusBadRequest, message, err)
}

// NotFound returns a 404 not found error with standard format
func NotFound(c *gin.Context, message string, err error) {
	Error(c, http.StatusNotFound, message, err)
}

// InternalServerError returns a 500 internal server error with standard format
func InternalServerError(c *gin.Context, message string, err error) {
	Error(c, http.StatusInternalServerError, message, err)
}

// Conflict returns a 409 conflict error with standard format
func Conflict(c *gin.Context, message string, err error) {
	Error(c, http.StatusConflict, message, err)
}
