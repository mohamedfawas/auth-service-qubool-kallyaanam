package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/response"
)

// CSRFToken generates a new CSRF token
func CSRFToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to cookie-based authenticated requests
		token, err := c.Cookie("registration_session")
		if err == nil && token != "" {
			// Generate CSRF token
			csrfToken := uuid.New().String()

			// Store in cookie with SameSite strict policy
			c.SetCookie(
				"csrf_token",
				csrfToken,
				int(24*time.Hour.Seconds()),
				"/",
				"",
				c.Request.TLS != nil,
				true,
			)

			// Also include token in response header for JS clients
			c.Header("X-CSRF-Token", csrfToken)
		}

		c.Next()
	}
}

// CSRFProtection middleware to protect against CSRF attacks
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for GET, HEAD, OPTIONS, TRACE requests (they should be safe)
		if c.Request.Method == "GET" ||
			c.Request.Method == "HEAD" ||
			c.Request.Method == "OPTIONS" ||
			c.Request.Method == "TRACE" {
			c.Next()
			return
		}

		// Only check if user has an active session
		sessionToken, err := c.Cookie("registration_session")
		if err != nil || sessionToken == "" {
			c.Next()
			return
		}

		// Get CSRF token from header then from form
		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			csrfToken = c.PostForm("csrf_token")
		}

		// Get the expected token from cookie
		expectedToken, err := c.Cookie("csrf_token")
		if err != nil || expectedToken == "" || csrfToken != expectedToken {
			response.Error(c, http.StatusForbidden, "CSRF token validation failed",
				"Invalid or missing CSRF token")
			c.Abort()
			return
		}

		c.Next()
	}
}
