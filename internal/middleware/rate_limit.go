package middleware

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository"
	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/pkg/response"
)

// RateLimiter middleware limits requests based on client IP or user ID
func RateLimiter(rateLimitRepo repository.RateLimitRepository, endpoint string, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting if repository is nil (Redis not available)
		if rateLimitRepo == nil {
			c.Next()
			return
		}

		// Try to get user ID from context if authenticated
		var identifier string
		if userID, exists := c.Get("user_id"); exists {
			identifier = fmt.Sprintf("user:%s", userID)
		} else {
			// Fall back to client IP for unauthenticated users
			identifier = fmt.Sprintf("ip:%s", c.ClientIP())
		}

		key := fmt.Sprintf("rate_limit:%s:%s", endpoint, identifier)

		// Use circuit breaker pattern for Redis operations
		count, err := rateLimitRepo.IncrementCounter(c.Request.Context(), key, window)
		if err != nil {
			// Log the error
			log.Printf("Rate limiting error: %v", err)

			// If too many errors, start failing closed for safety
			if rateLimitErrors.Increment() > 10 {
				response.Error(c, http.StatusServiceUnavailable,
					"Rate limiting service unavailable",
					"Please try again later")
				c.Abort()
				return
			}

			// Otherwise continue (fail open for lower error rates)
			c.Next()
			return
		}

		// Reset error counter on success
		rateLimitErrors.Reset()

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

// Simple circuit breaker counter
type errorCounter struct {
	count int
	mu    sync.Mutex
}

func (c *errorCounter) Increment() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
	return c.count
}

func (c *errorCounter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count = 0
}

var rateLimitErrors = &errorCounter{}
