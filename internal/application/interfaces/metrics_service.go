package interfaces

import "context"

// MetricsService defines the interface for collecting metrics
type MetricsService interface {
	// IncRegistrationAttempt increments the counter for registration attempts
	IncRegistrationAttempt(ctx context.Context)

	// IncRegistrationSuccess increments the counter for successful registrations
	IncRegistrationSuccess(ctx context.Context)

	// IncRegistrationFailure increments the counter for failed registrations with a reason
	IncRegistrationFailure(ctx context.Context, reason string)

	// RegistrationDuration records the duration of a registration operation
	RegistrationDuration(ctx context.Context, duration float64)
}
