// internal/service/metrics_service.go
package service

import (
	"context"
)

// NoOpMetricsService implements MetricsService with no-op functions
type NoOpMetricsService struct{}

// NewNoOpMetricsService creates a new no-op metrics service
func NewNoOpMetricsService() MetricsService {
	return &NoOpMetricsService{}
}

// IncRegistrationAttempt is a no-op implementation
func (s *NoOpMetricsService) IncRegistrationAttempt(ctx context.Context) {
	// No-op
}

// IncRegistrationSuccess is a no-op implementation
func (s *NoOpMetricsService) IncRegistrationSuccess(ctx context.Context) {
	// No-op
}

// IncRegistrationFailure is a no-op implementation
func (s *NoOpMetricsService) IncRegistrationFailure(ctx context.Context, reason string) {
	// No-op
}

// RegistrationDuration is a no-op implementation
func (s *NoOpMetricsService) RegistrationDuration(ctx context.Context, duration float64) {
	// No-op
}
