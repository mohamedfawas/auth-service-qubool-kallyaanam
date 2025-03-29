package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/response"
)

// RateLimiter middleware limits requests based on client IP
func RateLimiter(rateLimitRepo repository.RateLimitRepository, endpoint string, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting if repository is nil (Redis not available)
		if rateLimitRepo == nil {
			c.Next()
			return
		}

		// Use client IP as identifier
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s:%s", endpoint, clientIP)

		// Increment counter
		count, err := rateLimitRepo.IncrementCounter(c.Request.Context(), key, window)
		if err != nil {
			// Log error but continue (fail open for availability)
			c.Next()
			return
		}

		// If count exceeds limit, reject the request
		if count > limit {
			response.Error(c, http.StatusTooManyRequests,
				"Rate limit exceeded. Please try again later.",
				gin.H{"retry_after": window.Seconds()})
			c.Abort()
			return
		}

		c.Next()
	}
}
