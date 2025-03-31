package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/logging"
	"go.uber.org/zap"
)

// MaxBodyLogSize is the maximum body size to log.
const MaxBodyLogSize = 1024 * 10 // 10KB

// Logger middleware logs request and response details.
func Logger() gin.HandlerFunc {
	logger := logging.Logger()

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method

		// Log request body for specific endpoints (like registration)
		var requestBody []byte
		if shouldLogBody(path, method) {
			if c.Request.Body != nil {
				bodyBytes, _ := io.ReadAll(c.Request.Body)
				// Truncate if too large
				if len(bodyBytes) > MaxBodyLogSize {
					bodyBytes = bodyBytes[:MaxBodyLogSize]
				}
				requestBody = bodyBytes
				// Replace the body for downstream handlers
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Create a custom response writer to capture the response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Get response details
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Determine log level based on status code
		logFunc := logger.Info
		if statusCode >= 400 && statusCode < 500 {
			logFunc = logger.Warn
		} else if statusCode >= 500 {
			logFunc = logger.Error
		}

		// Log the request/response details
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
		}

		if raw != "" {
			fields = append(fields, zap.String("query", raw))
		}

		if len(requestBody) > 0 {
			// Sanitize sensitive data before logging
			sanitizedBody := sanitizeBody(requestBody, path)
			fields = append(fields, zap.String("request", sanitizedBody))
		}

		if errorMessage != "" {
			fields = append(fields, zap.String("error", errorMessage))
		}

		if shouldLogBody(path, method) && statusCode != 204 {
			// For successful responses, log a summary of the response body
			if statusCode >= 200 && statusCode < 300 && blw.body.Len() > 0 {
				responseBody := blw.body.String()
				if len(responseBody) > MaxBodyLogSize {
					responseBody = responseBody[:MaxBodyLogSize] + "... (truncated)"
				}
				fields = append(fields, zap.String("response", responseBody))
			}
		}

		logFunc("HTTP Request", fields...)
	}
}

// bodyLogWriter is a custom gin.ResponseWriter that captures the response body.
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body while writing it to the underlying writer.
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// shouldLogBody determines if the request/response body should be logged.
func shouldLogBody(path, method string) bool {
	// Only log bodies for certain endpoints and methods
	if method == "POST" || method == "PUT" || method == "PATCH" {
		if path == "/auth/register" {
			return true
		}
	}
	return false
}

// sanitizeBody removes sensitive information from request bodies.
func sanitizeBody(body []byte, path string) string {
	// Convert to string for processing
	bodyStr := string(body)

	// Different sanitization logic based on the path
	if path == "/auth/register" {
		// Replace password with asterisks while keeping the structure
		// This is a simple regex-like approach, in production you might want to use proper JSON parsing
		// passwordRegex := `"password"\s*:\s*"[^"]*"`
		// passwordReplacement := `"password":"********"`
		// Basic find and replace - in a real implementation, use a proper JSON parser or regex
		for i := 0; i < len(bodyStr)-11; i++ {
			if bodyStr[i:i+10] == `"password"` {
				// Find the end of the password value
				valueStart := i + 10
				for valueStart < len(bodyStr) && (bodyStr[valueStart] == ' ' || bodyStr[valueStart] == ':') {
					valueStart++
				}
				if valueStart < len(bodyStr) && bodyStr[valueStart] == '"' {
					valueStart++ // Skip the opening quote
					valueEnd := valueStart
					for valueEnd < len(bodyStr) && bodyStr[valueEnd] != '"' {
						valueEnd++
					}
					if valueEnd < len(bodyStr) {
						// Replace the password value with asterisks
						bodyStr = bodyStr[:valueStart] + "********" + bodyStr[valueEnd:]
					}
				}
			}
		}
	}

	return bodyStr
}
