package metrics

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/interfaces"
)

// PrometheusMetrics implements MetricsService using Prometheus
type PrometheusMetrics struct {
	registrationAttempts prometheus.Counter
	registrationSuccess  prometheus.Counter
	registrationFailures *prometheus.CounterVec
	registrationDuration prometheus.Histogram
}

// NewPrometheusMetrics creates a new PrometheusMetrics instance
func NewPrometheusMetrics() interfaces.MetricsService {
	metrics := &PrometheusMetrics{
		registrationAttempts: promauto.NewCounter(prometheus.CounterOpts{
			Name: "auth_registration_attempts_total",
			Help: "Total number of registration attempts",
		}),
		registrationSuccess: promauto.NewCounter(prometheus.CounterOpts{
			Name: "auth_registration_success_total",
			Help: "Total number of successful registrations",
		}),
		registrationFailures: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "auth_registration_failures_total",
			Help: "Total number of failed registrations",
		}, []string{"reason"}),
		registrationDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "auth_registration_duration_seconds",
			Help:    "Registration operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		}),
	}

	return metrics
}

// RegisterHandler registers the metrics HTTP handler with the router
func RegisterHandler(router *gin.Engine) {
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

// IncRegistrationAttempt increments the counter for registration attempts
func (m *PrometheusMetrics) IncRegistrationAttempt(ctx context.Context) {
	m.registrationAttempts.Inc()
}

// IncRegistrationSuccess increments the counter for successful registrations
func (m *PrometheusMetrics) IncRegistrationSuccess(ctx context.Context) {
	m.registrationSuccess.Inc()
}

// IncRegistrationFailure increments the counter for failed registrations with a reason
func (m *PrometheusMetrics) IncRegistrationFailure(ctx context.Context, reason string) {
	m.registrationFailures.WithLabelValues(reason).Inc()
}

// RegistrationDuration records the duration of a registration operation
func (m *PrometheusMetrics) RegistrationDuration(ctx context.Context, duration float64) {
	m.registrationDuration.Observe(duration)
}
