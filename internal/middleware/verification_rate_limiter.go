// internal/middleware/verification_rate_limiter.go
package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/util/logger"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/util/response"
	"golang.org/x/time/rate"
)

// VerificationRateLimiterConfig contains specific configuration for verification rate limiting
type VerificationRateLimiterConfig struct {
	// Maximum verification attempts per time period
	MaxAttemptsPerPeriod int
	// Time period in minutes
	PeriodMinutes int
	// Block duration in minutes after rate limit exceeded
	BlockDurationMinutes int
}

// VerificationRateLimiter creates a rate limiter specifically for email verification
func VerificationRateLimiter(config VerificationRateLimiterConfig, logger *logger.Logger) gin.HandlerFunc {
	// Use sensible defaults if not provided
	if config.MaxAttemptsPerPeriod <= 0 {
		config.MaxAttemptsPerPeriod = 5 // Default: 5 attempts per period
	}
	if config.PeriodMinutes <= 0 {
		config.PeriodMinutes = 15 // Default: 15 minute period
	}
	if config.BlockDurationMinutes <= 0 {
		config.BlockDurationMinutes = 60 // Default: block for 60 minutes
	}

	type emailLimiter struct {
		limiter      *rate.Limiter
		lastSeen     time.Time
		blocked      bool
		blockedUntil time.Time
	}

	var (
		// Track limiters by both IP and email to prevent multiple IPs trying same email
		ipLimiters    = make(map[string]*emailLimiter)
		emailLimiters = make(map[string]*emailLimiter)
		mu            sync.Mutex
	)

	// Start a goroutine to clean up old limiters
	go func() {
		for {
			time.Sleep(5 * time.Minute)

			mu.Lock()
			now := time.Now()

			// Clean up IP limiters
			for ip, limiter := range ipLimiters {
				if now.After(limiter.lastSeen.Add(3 * time.Hour)) {
					delete(ipLimiters, ip)
				}
			}

			// Clean up email limiters
			for email, limiter := range emailLimiters {
				if now.After(limiter.lastSeen.Add(24 * time.Hour)) {
					delete(emailLimiters, email)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		// Only apply to POST /auth/verify-email
		// Note: Update the path to match your actual API endpoint path
		if (c.FullPath() == "/auth/verify-email" || c.FullPath() == "/api/v1/auth/verify-email") && c.Request.Method == "POST" {
			// Extract email from request
			var requestData struct {
				Email string `json:"email"`
			}

			// Store body content
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				// If we can't read the body, just use IP-based limiting
				checkIPLimit(c, ip, ipLimiters, &mu, config, logger)
				return
			}

			// Reset the body for binding
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			if err := c.ShouldBindJSON(&requestData); err != nil {
				// If we can't parse the email, just use IP-based limiting
				checkIPLimit(c, ip, ipLimiters, &mu, config, logger)
				return
			}

			email := requestData.Email

			// Reset the body again for downstream handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Check email-based rate limit
			mu.Lock()
			emailLimiterInfo, exists := emailLimiters[email]

			if !exists {
				// Create a new limiter for this email
				emailLimiterInfo = &emailLimiter{
					limiter: rate.NewLimiter(
						rate.Limit(float64(config.MaxAttemptsPerPeriod)/(float64(config.PeriodMinutes)*60)),
						config.MaxAttemptsPerPeriod,
					),
					lastSeen: time.Now(),
					blocked:  false,
				}
				emailLimiters[email] = emailLimiterInfo
			}

			// Update last seen
			emailLimiterInfo.lastSeen = time.Now()

			// Check if currently blocked
			if emailLimiterInfo.blocked && time.Now().Before(emailLimiterInfo.blockedUntil) {
				mu.Unlock()
				blockedResponse(c, logger, email, ip, emailLimiterInfo.blockedUntil)
				return
			}

			// If block expired, reset it
			if emailLimiterInfo.blocked && time.Now().After(emailLimiterInfo.blockedUntil) {
				emailLimiterInfo.blocked = false
			}

			// Check rate limit
			allow := emailLimiterInfo.limiter.Allow()

			if !allow {
				// Apply block
				emailLimiterInfo.blocked = true
				emailLimiterInfo.blockedUntil = time.Now().Add(time.Duration(config.BlockDurationMinutes) * time.Minute)
				mu.Unlock()

				blockedResponse(c, logger, email, ip, emailLimiterInfo.blockedUntil)
				return
			}

			mu.Unlock()

			// Also apply IP-based rate limiting as a second layer
			checkIPLimit(c, ip, ipLimiters, &mu, config, logger)
			return
		}

		c.Next()
	}
}

// Helper functions
func checkIPLimit(c *gin.Context, ip string, limiters map[string]*emailLimiter, mu *sync.Mutex,
	config VerificationRateLimiterConfig, logger *logger.Logger) {

	mu.Lock()
	limiterInfo, exists := limiters[ip]

	if !exists {
		// Create a new limiter for this IP (slightly more permissive than email-based)
		limiterInfo = &emailLimiter{
			limiter: rate.NewLimiter(
				rate.Limit(float64(config.MaxAttemptsPerPeriod+2)/(float64(config.PeriodMinutes)*60)),
				config.MaxAttemptsPerPeriod+2,
			),
			lastSeen: time.Now(),
			blocked:  false,
		}
		limiters[ip] = limiterInfo
	}

	limiterInfo.lastSeen = time.Now()

	// Check if currently blocked
	if limiterInfo.blocked && time.Now().Before(limiterInfo.blockedUntil) {
		mu.Unlock()
		blockedResponse(c, logger, "", ip, limiterInfo.blockedUntil)
		return
	}

	// If block expired, reset it
	if limiterInfo.blocked && time.Now().After(limiterInfo.blockedUntil) {
		limiterInfo.blocked = false
	}

	// Check rate limit
	allow := limiterInfo.limiter.Allow()

	if !allow {
		// Apply block
		limiterInfo.blocked = true
		limiterInfo.blockedUntil = time.Now().Add(time.Duration(config.BlockDurationMinutes) * time.Minute)
		mu.Unlock()

		blockedResponse(c, logger, "", ip, limiterInfo.blockedUntil)
		return
	}

	mu.Unlock()
	c.Next()
}

func blockedResponse(c *gin.Context, logger *logger.Logger, email string, ip string, blockedUntil time.Time) {
	remainingSeconds := int(blockedUntil.Sub(time.Now()).Seconds())

	logger.SecurityEvent("Verification rate limit exceeded",
		logger.Field("ip_address", ip),
		logger.Field("email", email),
		logger.Field("blocked_until", blockedUntil),
	)

	response.Error(c, http.StatusTooManyRequests, "Too many verification attempts, please try again later",
		&RateLimitError{
			Message:    "Rate limit exceeded",
			RetryAfter: remainingSeconds,
		},
	)
	c.Abort()
}

// RateLimitError implements error interface
type RateLimitError struct {
	Message    string
	RetryAfter int
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("%s. Try again after %d seconds", e.Message, e.RetryAfter)
}
