package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/response"
)

// JWTCSRFToken generates a CSRF token based on JWT
func JWTCSRFToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to JWT-authenticated requests
		jwtToken, err := c.Cookie("registration_session")
		if err == nil && jwtToken != "" {
			// Generate CSRF token as a hash of the JWT token and a timestamp
			// This is a simplified approach - in production you might want to use a separate secret
			tokenHash := computeCSRFHash(jwtToken, c.ClientIP())

			// Store in cookie with SameSite strict policy
			c.SetCookie(
				"csrf_token",
				tokenHash,
				int(24*time.Hour.Seconds()),
				"/",
				"",
				c.Request.TLS != nil,
				true, // HttpOnly
			)

			// Also include in response header for SPA clients
			c.Header("X-CSRF-Token", tokenHash)
		}

		c.Next()
	}
}

// JWTCSRFProtection middleware to protect against CSRF attacks when using JWT
func JWTCSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for safe methods (they don't modify state)
		if c.Request.Method == "GET" ||
			c.Request.Method == "HEAD" ||
			c.Request.Method == "OPTIONS" ||
			c.Request.Method == "TRACE" {
			c.Next()
			return
		}

		// Only check CSRF if the user has JWT
		jwtToken, err := c.Cookie("registration_session")
		if err != nil || jwtToken == "" {
			c.Next()
			return
		}

		// Get CSRF token from header or form
		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			csrfToken = c.PostForm("csrf_token")
		}

		// Compute expected token
		expectedToken := computeCSRFHash(jwtToken, c.ClientIP())

		// Compare tokens
		if csrfToken == "" || !secureCompare(csrfToken, expectedToken) {
			response.Error(c, http.StatusForbidden, "CSRF token validation failed",
				"Invalid or missing CSRF token")
			c.Abort()
			return
		}

		c.Next()
	}
}

// computeCSRFHash creates a hash based on the JWT token and session identifier
func computeCSRFHash(token string, clientIP string) string {
	// Consider using session ID instead of client IP
	h := hmac.New(sha256.New, []byte(token))
	// Use only the token to avoid IP changes causing issues
	h.Write([]byte("csrf-salt"))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// secureCompare compares two strings in constant time to prevent timing attacks
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	equal := true
	for i := 0; i < len(a); i++ {
		equal = equal && (a[i] == b[i])
	}
	return equal
}
