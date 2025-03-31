// Package middleware contains HTTP middleware functions.
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/internal/domain/dto"
	"github.com/mohamedfawas/auth-service-qubool-kallyanam/pkg/logging"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// IPRateLimiter stores rate limiters for different IPs.
type IPRateLimiter struct {
	ips    map[string]*rate.Limiter
	mu     *sync.RWMutex
	rate   rate.Limit
	burst  int
	logger *zap.Logger
}

// NewIPRateLimiter creates a new rate limiter for IPs.
func NewIPRateLimiter(r rate.Limit, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		ips:    make(map[string]*rate.Limiter),
		mu:     &sync.RWMutex{},
		rate:   r,
		burst:  burst,
		logger: logging.Logger(),
	}
}

// GetLimiter returns the rate limiter for an IP.
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		i.mu.Lock()
		// Check again to handle race condition
		limiter, stillNotExists := i.ips[ip]
		if stillNotExists {
			limiter = rate.NewLimiter(i.rate, i.burst)
			i.ips[ip] = limiter
		}
		i.mu.Unlock()
	}

	return limiter
}

// RateLimit creates a middleware for rate limiting based on IP address.
func RateLimit(requestsPerMinute int, burst int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rate.Limit(float64(requestsPerMinute)/60.0), burst)
	logger := logging.Logger()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = c.Request.RemoteAddr
		}

		if !limiter.GetLimiter(ip).Allow() {
			logger.Warn("Rate limit exceeded",
				zap.String("ip", ip),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)

			c.JSON(http.StatusTooManyRequests, dto.NewErrorResponse(
				http.StatusTooManyRequests,
				"Rate limit exceeded",
				"Too many requests, please try again later",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// CleanupLimiters periodically cleans up old limiters to prevent memory leaks.
func (i *IPRateLimiter) CleanupLimiters() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.Lock()
		currentSize := len(i.ips)
		i.ips = make(map[string]*rate.Limiter) // Reset the map
		i.mu.Unlock()

		i.logger.Info("Cleaned up rate limiters",
			zap.Int("previous_size", currentSize),
		)
	}
}
