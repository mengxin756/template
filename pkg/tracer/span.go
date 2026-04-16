// Package tracer provides unified span management.
package tracer

import (
	"context"

	"example.com/classic/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// UnifiedSpan combines simple logging span with OTEL span.
type UnifiedSpan struct {
	logSpan    *Span
	otelSpan   trace.Span
	operation  string
}

// StartUnifiedSpan creates a span that logs AND exports to OTEL.
// This gives you both: log-based tracing AND tree-view in Grafana.
func StartUnifiedSpan(ctx context.Context, log logger.Logger, operation string) (*UnifiedSpan, context.Context) {
	// 1. Start OTEL span (for tree-view in Tempo)
	otelCtx, otelSpan := StartOTELSpan(ctx, operation,
		trace.WithAttributes(
			attribute.String("operation", operation),
		),
	)

	// 2. Start log span (for log aggregation in Loki)
	logSpan, logCtx := StartSpan(otelCtx, log, operation)

	return &UnifiedSpan{
		logSpan:   logSpan,
		otelSpan:  otelSpan,
		operation: operation,
	}, logCtx
}

// End completes the span.
func (s *UnifiedSpan) End() {
	if s.logSpan != nil {
		s.logSpan.End()
	}
	if s.otelSpan != nil {
		s.otelSpan.End()
	}
}

// EndWithError completes the span with an error.
func (s *UnifiedSpan) EndWithError(err error) {
	if s.logSpan != nil {
		s.logSpan.EndWithError(err)
	}
	if s.otelSpan != nil {
		s.otelSpan.RecordError(err)
		s.otelSpan.SetStatus(codes.Error, err.Error())
		s.otelSpan.End()
	}
}

// SetAttributes sets attributes on the OTEL span.
func (s *UnifiedSpan) SetAttributes(attrs ...attribute.KeyValue) {
	if s.otelSpan != nil {
		s.otelSpan.SetAttributes(attrs...)
	}
}

// AddEvent adds an event to the span.
func (s *UnifiedSpan) AddEvent(name string, attrs ...attribute.KeyValue) {
	if s.otelSpan != nil {
		s.otelSpan.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// Context returns the span context.
func (s *UnifiedSpan) Context() context.Context {
	if s.otelSpan != nil {
		return trace.ContextWithSpan(context.Background(), s.otelSpan)
	}
	return context.Background()
}

// SpanContext returns the OTEL span context.
func (s *UnifiedSpan) SpanContext() trace.SpanContext {
	if s.otelSpan != nil {
		return s.otelSpan.SpanContext()
	}
	return trace.SpanContext{}
}

// --- Convenience functions using UnifiedSpan ---

// UnifiedDBSpan creates a unified span for database operations.
func UnifiedDBSpan(ctx context.Context, log logger.Logger, query string) (*UnifiedSpan, context.Context) {
	span, ctx := StartUnifiedSpan(ctx, log, "db:"+truncateOperation(query, 50))
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.statement", query),
	)
	return span, ctx
}

// UnifiedCacheSpan creates a unified span for cache operations.
func UnifiedCacheSpan(ctx context.Context, log logger.Logger, operation string) (*UnifiedSpan, context.Context) {
	return StartUnifiedSpan(ctx, log, "cache:"+operation)
}

// UnifiedExternalSpan creates a unified span for external service calls.
func UnifiedExternalSpan(ctx context.Context, log logger.Logger, service string) (*UnifiedSpan, context.Context) {
	span, ctx := StartUnifiedSpan(ctx, log, "external:"+service)
	span.SetAttributes(
		attribute.String("peer.service", service),
	)
	return span, ctx
}

// UnifiedServiceSpan creates a unified span for service layer operations.
func UnifiedServiceSpan(ctx context.Context, log logger.Logger, method string) (*UnifiedSpan, context.Context) {
	return StartUnifiedSpan(ctx, log, "service:"+method)
}

// UnifiedQueueSpan creates a unified span for message queue operations.
func UnifiedQueueSpan(ctx context.Context, log logger.Logger, queueName string) (*UnifiedSpan, context.Context) {
	span, ctx := StartUnifiedSpan(ctx, log, "queue:"+queueName)
	span.SetAttributes(
		attribute.String("messaging.system", "asynq"),
		attribute.String("messaging.destination", queueName),
	)
	return span, ctx
}
