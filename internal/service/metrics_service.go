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

// internal/service/metrics_service.go

// IncVerificationAttempt increments the verification attempt metric
func (s *NoOpMetricsService) IncVerificationAttempt(ctx context.Context) {
	// Actual implementation would use a metrics library like Prometheus
	// For now we'll just log it
}

// IncVerificationSuccess increments the verification success metric
func (s *NoOpMetricsService) IncVerificationSuccess(ctx context.Context) {
	// Actual implementation would use a metrics library like Prometheus
}

// IncVerificationFailure increments the verification failure metric
func (s *NoOpMetricsService) IncVerificationFailure(ctx context.Context, reason string) {
	// Actual implementation would use a metrics library like Prometheus
	// and include the reason as a label
}

// VerificationDuration records the duration of a verification operation
func (s *NoOpMetricsService) VerificationDuration(ctx context.Context, seconds float64) {
	// Actual implementation would use a metrics library like Prometheus
}
