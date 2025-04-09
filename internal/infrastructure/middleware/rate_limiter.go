package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/infrastructure/services/logger"
)

// RateLimiterConfig contains configuration for the rate limiter
type RateLimiterConfig struct {
	MaxRequestsPerMinute int
	BlockDurationMinutes int
}

// RateLimiterMiddleware creates a middleware that limits requests by IP address
func RateLimiterMiddleware(config RateLimiterConfig, logger *logger.Logger) gin.HandlerFunc {
	// Use default values if not provided
	if config.MaxRequestsPerMinute <= 0 {
		config.MaxRequestsPerMinute = 5 // Default: 5 requests per minute
	}
	if config.BlockDurationMinutes <= 0 {
		config.BlockDurationMinutes = 30 // Default: block for 30 minutes
	}

	type ipLimiter struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		limiters = make(map[string]*ipLimiter)
		mu       sync.Mutex
	)

	// Start a goroutine to clean up old limiters
	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()
			for ip, limiter := range limiters {
				if time.Since(limiter.lastSeen) > time.Hour {
					delete(limiters, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()

		limiterInfo, exists := limiters[ip]
		if !exists {
			// Create a new limiter for this IP
			limiterInfo = &ipLimiter{
				limiter:  rate.NewLimiter(rate.Limit(config.MaxRequestsPerMinute)/60, config.MaxRequestsPerMinute),
				lastSeen: time.Now(),
			}
			limiters[ip] = limiterInfo
		}

		limiterInfo.lastSeen = time.Now()
		allow := limiterInfo.limiter.Allow()

		mu.Unlock()

		if !allow {
			logger.SecurityEvent("Rate limit exceeded",
				logger.Field("ip_address", ip),
				logger.Field("path", c.FullPath()),
				logger.Field("method", c.Request.Method),
			)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  false,
				"message": "Too many requests, please try again later",
				"error":   "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
