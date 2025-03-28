package response

import (
	"github.com/gin-gonic/gin"
)

// Response is a standard API response format
type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// Success sends a successful response
func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, Response{
		Status:  status,
		Message: message,
		Data:    data,
	})
}

// Error sends an error response
func Error(c *gin.Context, status int, message string, err interface{}) {
	c.JSON(status, Response{
		Status:  status,
		Message: message,
		Error:   err,
	})
}
