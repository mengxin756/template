// Package tracer provides OpenTelemetry integration for distributed tracing.
package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracerProvider *sdktrace.TracerProvider
	tracerInstance trace.Tracer
)

// Config holds tracer configuration.
type Config struct {
	ServiceName  string
	OTLPEndpoint string // e.g., "localhost:4317"
	SampleRate   float64 // 0.0 - 1.0
}

// InitOTEL initializes OpenTelemetry tracing.
// This enables tree-view traces in Grafana Tempo.
func InitOTEL(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	// Create OTLP exporter (sends traces to Tempo)
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Create resource with service info
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create tracer provider with sampling
	sampler := sdktrace.ParentBased(
		sdktrace.TraceIDRatioBased(cfg.SampleRate),
	)
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tracerProvider)

	// Set text map propagator for context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer instance
	tracerInstance = tracerProvider.Tracer(cfg.ServiceName)

	// Return shutdown function
	return tracerProvider.Shutdown, nil
}

// StartOTELSpan starts an OpenTelemetry span.
// Use this instead of the simple tracer.Span for tree-view traces.
func StartOTELSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if tracerInstance == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return tracerInstance.Start(ctx, name, opts...)
}

// SpanFromContext returns the current span from context.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddSpanEvent adds an event to the current span.
func AddSpanEvent(ctx context.Context, name string, attrs ...trace.EventOption) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, attrs...)
	}
}

// SetSpanError records an error on the span.
func SetSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err)
	}
}

// InjectTraceContext injects trace context into a carrier (for HTTP/gRPC propagation).
func InjectTraceContext(ctx context.Context, carrier propagation.TextMapCarrier) {
	otel.GetTextMapPropagator().Inject(ctx, carrier)
}

// ExtractTraceContext extracts trace context from a carrier.
func ExtractTraceContext(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}
