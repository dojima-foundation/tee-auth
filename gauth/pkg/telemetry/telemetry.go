package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmeter "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds configuration for telemetry setup
type Config struct {
	ServiceName            string
	ServiceVersion         string
	Environment            string
	TracingEnabled         bool
	MetricsEnabled         bool
	OTLPEndpoint           string
	OTLPInsecure           bool
	TraceSamplingRatio     float64
	MetricsReportingPeriod time.Duration
	MetricsPort            int
}

// Telemetry holds the telemetry providers
type Telemetry struct {
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
	Tracer         trace.Tracer
	Meter          metric.Meter
	Propagator     propagation.TextMapPropagator
	Shutdown       func(context.Context) error
}

// New creates a new telemetry setup
func New(ctx context.Context, cfg Config) (*Telemetry, error) {
	telemetry := &Telemetry{
		Propagator: propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	}

	// Set global propagator
	otel.SetTextMapPropagator(telemetry.Propagator)

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Setup tracing if enabled
	if cfg.TracingEnabled {
		if err := telemetry.setupTracing(ctx, cfg, res); err != nil {
			return nil, fmt.Errorf("failed to setup tracing: %w", err)
		}
	} else {
		// Use noop tracer if tracing is disabled
		telemetry.TracerProvider = trace.NewNoopTracerProvider()
		telemetry.Tracer = telemetry.TracerProvider.Tracer(cfg.ServiceName)
	}

	// Setup metrics if enabled
	if cfg.MetricsEnabled {
		if err := telemetry.setupMetrics(ctx, cfg, res); err != nil {
			return nil, fmt.Errorf("failed to setup metrics: %w", err)
		}

		// Start metrics server if metrics port is specified
		if cfg.MetricsReportingPeriod > 0 {
			if err := UpdateTelemetryWithMetrics(telemetry, cfg.MetricsPort); err != nil {
				return nil, fmt.Errorf("failed to start metrics server: %w", err)
			}
		}
	}

	return telemetry, nil
}

// setupTracing configures tracing with OpenTelemetry
func (t *Telemetry) setupTracing(ctx context.Context, cfg Config, res *resource.Resource) error {
	// Create OTLP exporter
	traceExporter, err := otlptrace.New(ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create trace provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.TraceSamplingRatio)),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter),
	)

	// Set global trace provider
	otel.SetTracerProvider(tracerProvider)

	// Create tracer
	t.TracerProvider = tracerProvider
	t.Tracer = tracerProvider.Tracer(cfg.ServiceName)

	// Add shutdown function
	originalShutdown := t.Shutdown
	t.Shutdown = func(ctx context.Context) error {
		if originalShutdown != nil {
			if err := originalShutdown(ctx); err != nil {
				return err
			}
		}
		return tracerProvider.Shutdown(ctx)
	}

	return nil
}

// setupMetrics configures metrics with OpenTelemetry
func (t *Telemetry) setupMetrics(ctx context.Context, cfg Config, res *resource.Resource) error {
	// Create Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		return fmt.Errorf("failed to create prometheus exporter: %w", err)
	}

	// Create meter provider with the exporter as reader
	t.MeterProvider = sdkmeter.NewMeterProvider(
		sdkmeter.WithReader(exporter),
		sdkmeter.WithResource(res),
	)
	t.Meter = t.MeterProvider.Meter(cfg.ServiceName)

	// Set global meter provider
	otel.SetMeterProvider(t.MeterProvider)

	return nil
}

// StartSpan starts a new span with the given name and options
func (t *Telemetry) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.Tracer.Start(ctx, name, opts...)
}

// SpanFromContext returns the current span from the context
func (t *Telemetry) SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// ContextWithSpan returns a new context with the given span
func (t *Telemetry) ContextWithSpan(ctx context.Context, span trace.Span) context.Context {
	return trace.ContextWithSpan(ctx, span)
}

// Extract extracts trace information from carrier into context
func (t *Telemetry) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	return t.Propagator.Extract(ctx, carrier)
}

// Inject injects trace information from context into carrier
func (t *Telemetry) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	t.Propagator.Inject(ctx, carrier)
}
