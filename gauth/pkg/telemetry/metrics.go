package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/metric"
)

// MetricsServer represents a Prometheus metrics HTTP server
type MetricsServer struct {
	server   *http.Server
	registry *prometheus.Registry
	meter    metric.Meter
}

// Common metrics that can be used across the application
var (
	// HTTP metrics
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	// gRPC metrics
	GRPCRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	GRPCRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"method", "status"},
	)

	// Database metrics
	DatabaseQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "status"},
	)

	DatabaseQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"operation", "status"},
	)

	// Renclave client metrics
	RenclaveRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "renclave_requests_total",
			Help: "Total number of requests to Renclave service",
		},
		[]string{"method", "status"},
	)

	RenclaveRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "renclave_request_duration_seconds",
			Help:    "Renclave request duration in seconds",
			Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"method", "status"},
	)

	// Business metrics
	WalletsCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "wallets_created_total",
			Help: "Total number of wallets created",
		},
	)

	PrivateKeysCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "private_keys_created_total",
			Help: "Total number of private keys created",
		},
	)

	OrganizationsCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "organizations_created_total",
			Help: "Total number of organizations created",
		},
	)

	UsersCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "users_created_total",
			Help: "Total number of users created",
		},
	)

	ActivitiesCreatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "activities_created_total",
			Help: "Total number of activities created",
		},
		[]string{"type"},
	)
)

// NewMetricsServer creates a new metrics server
func NewMetricsServer(port int, meter metric.Meter) (*MetricsServer, error) {
	registry := prometheus.NewRegistry()

	// Register metrics
	registry.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		GRPCRequestsTotal,
		GRPCRequestDuration,
		DatabaseQueriesTotal,
		DatabaseQueryDuration,
		RenclaveRequestsTotal,
		RenclaveRequestDuration,
		WalletsCreatedTotal,
		PrivateKeysCreatedTotal,
		OrganizationsCreatedTotal,
		UsersCreatedTotal,
		ActivitiesCreatedTotal,
	)

	// Create HTTP server for metrics
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
		ReadHeaderTimeout: 30 * time.Second, // Prevent Slowloris attacks
	}

	return &MetricsServer{
		server:   server,
		registry: registry,
		meter:    meter,
	}, nil
}

// Start starts the metrics server
func (m *MetricsServer) Start() error {
	return m.server.ListenAndServe()
}

// Stop stops the metrics server
func (m *MetricsServer) Stop(ctx context.Context) error {
	return m.server.Shutdown(ctx)
}

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(method, path string, status int, duration time.Duration) {
	statusStr := fmt.Sprintf("%d", status)
	HTTPRequestsTotal.WithLabelValues(method, path, statusStr).Inc()
	HTTPRequestDuration.WithLabelValues(method, path, statusStr).Observe(duration.Seconds())
}

// RecordGRPCRequest records gRPC request metrics
func RecordGRPCRequest(method string, status string, duration time.Duration) {
	GRPCRequestsTotal.WithLabelValues(method, status).Inc()
	GRPCRequestDuration.WithLabelValues(method, status).Observe(duration.Seconds())
}

// RecordDatabaseQuery records database query metrics
func RecordDatabaseQuery(operation string, success bool, duration time.Duration) {
	status := "success"
	if !success {
		status = "error"
	}
	DatabaseQueriesTotal.WithLabelValues(operation, status).Inc()
	DatabaseQueryDuration.WithLabelValues(operation, status).Observe(duration.Seconds())
}

// RecordRenclaveRequest records Renclave request metrics
func RecordRenclaveRequest(method string, success bool, duration time.Duration) {
	status := "success"
	if !success {
		status = "error"
	}
	RenclaveRequestsTotal.WithLabelValues(method, status).Inc()
	RenclaveRequestDuration.WithLabelValues(method, status).Observe(duration.Seconds())
}

// RecordWalletCreated increments the wallet created counter
func RecordWalletCreated() {
	WalletsCreatedTotal.Inc()
}

// RecordPrivateKeyCreated increments the private key created counter
func RecordPrivateKeyCreated() {
	PrivateKeysCreatedTotal.Inc()
}

// RecordOrganizationCreated increments the organization created counter
func RecordOrganizationCreated() {
	OrganizationsCreatedTotal.Inc()
}

// RecordUserCreated increments the user created counter
func RecordUserCreated() {
	UsersCreatedTotal.Inc()
}

// RecordActivityCreated increments the activity created counter
func RecordActivityCreated(activityType string) {
	ActivitiesCreatedTotal.WithLabelValues(activityType).Inc()
}

// HTTPMetricsMiddleware returns a gin middleware that records HTTP metrics
func HTTPMetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer that captures the status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call the next handler
			next.ServeHTTP(rw, r)

			// Record metrics
			duration := time.Since(start)
			RecordHTTPRequest(r.Method, r.URL.Path, rw.statusCode, duration)
		})
	}
}

// responseWriter is a wrapper around http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// UpdateTelemetryWithMetrics adds metrics to the telemetry
func UpdateTelemetryWithMetrics(t *Telemetry, metricsPort int) error {
	// Create metrics server
	metricsServer, err := NewMetricsServer(metricsPort, t.Meter)
	if err != nil {
		return fmt.Errorf("failed to create metrics server: %w", err)
	}

	// Start metrics server in a goroutine
	go func() {
		if err := metricsServer.Start(); err != nil && err != http.ErrServerClosed {
			// Log error but don't fail the application
			fmt.Printf("Failed to start metrics server: %v\n", err)
		}
	}()

	// Add shutdown function for metrics server
	originalShutdown := t.Shutdown
	t.Shutdown = func(ctx context.Context) error {
		if originalShutdown != nil {
			if err := originalShutdown(ctx); err != nil {
				return err
			}
		}

		return metricsServer.Stop(ctx)
	}

	return nil
}
