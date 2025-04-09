// internal/middleware/timeout.go
package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/util/logger"
)

// TimeoutMiddleware creates a middleware that adds a timeout to the request context
func TimeoutMiddleware(timeout time.Duration, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace the request context
		c.Request = c.Request.WithContext(ctx)

		// Create channel to listen for context cancellation
		done := make(chan struct{})

		// Process request in goroutine
		go func() {
			c.Next()
			close(done)
		}()

		// Wait for either timeout or request completion
		select {
		case <-done:
			// Request completed normally
			return
		case <-ctx.Done():
			// Timeout occurred
			if ctx.Err() == context.DeadlineExceeded {
				logger.Warn("Request timed out",
					logger.Field("path", c.FullPath()),
					logger.Field("method", c.Request.Method),
					logger.Field("ip", c.ClientIP()),
				)

				c.Abort()
				c.JSON(http.StatusGatewayTimeout, gin.H{
					"status":  false,
					"message": "Request timed out",
					"error":   "The server took too long to process your request",
				})
			}
		}
	}
}
