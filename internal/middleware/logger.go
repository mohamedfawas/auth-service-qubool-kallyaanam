package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// StructuredLogger implements structured logging middleware
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get client IP
		clientIP := c.ClientIP()

		// Get response status
		status := c.Writer.Status()

		// Get request method
		method := c.Request.Method

		// Build query string
		query := ""
		if raw != "" {
			query = "?" + raw
		}

		// Create structured log entry
		logger := log.With().
			Str("method", method).
			Str("path", path).
			Str("query", query).
			Int("status", status).
			Str("ip", clientIP).
			Dur("latency", latency).
			Logger()

		// Log based on status code
		switch {
		case status >= 500:
			logger.Error().Msg("Server error")
		case status >= 400:
			logger.Warn().Msg("Client error")
		default:
			logger.Info().Msg("Request processed")
		}
	}
}
