// Package tracer provides span tracing utilities for distributed tracing.
package tracer

import (
	"context"
	"time"

	"example.com/classic/pkg/contextx"
	"example.com/classic/pkg/logger"
)

// Span represents a unit of work in distributed tracing.
type Span struct {
	operationName string
	startTime     time.Time
	ctx           context.Context
	log           logger.Logger
}

// StartSpan creates a new span for the given operation.
// Use this for important operations like DB queries, external calls, etc.
func StartSpan(ctx context.Context, log logger.Logger, operationName string) (*Span, context.Context) {
	// Create child span context
	childCtx := contextx.ChildSpan(ctx)
	childCtx = contextx.WithOperationName(childCtx, operationName)

	span := &Span{
		operationName: operationName,
		startTime:     time.Now(),
		ctx:           childCtx,
		log:           log,
	}

	span.log.Debug(childCtx, "span started",
		logger.String("operation", operationName))

	return span, childCtx
}

// End completes the span and logs the duration.
func (s *Span) End() {
	duration := time.Since(s.startTime)
	s.log.Debug(s.ctx, "span completed",
		logger.String("operation", s.operationName),
		logger.Duration("duration", duration))
}

// EndWithError completes the span with an error.
func (s *Span) EndWithError(err error) {
	duration := time.Since(s.startTime)
	s.log.Error(s.ctx, "span failed",
		logger.String("operation", s.operationName),
		logger.Err(err),
		logger.Duration("duration", duration))
}

// Context returns the span context.
func (s *Span) Context() context.Context {
	return s.ctx
}

// ---  convenient functions for common operations ---

// DBSpan starts a span for database operations.
func DBSpan(ctx context.Context, log logger.Logger, query string) (*Span, context.Context) {
	return StartSpan(ctx, log, "db:"+truncateOperation(query, 50))
}

// CacheSpan starts a span for cache operations.
func CacheSpan(ctx context.Context, log logger.Logger, operation string) (*Span, context.Context) {
	return StartSpan(ctx, log, "cache:"+operation)
}

// ExternalSpan starts a span for external service calls.
func ExternalSpan(ctx context.Context, log logger.Logger, service string) (*Span, context.Context) {
	return StartSpan(ctx, log, "external:"+service)
}

// QueueSpan starts a span for message queue operations.
func QueueSpan(ctx context.Context, log logger.Logger, queueName string) (*Span, context.Context) {
	return StartSpan(ctx, log, "queue:"+queueName)
}

// ServiceSpan starts a span for service layer operations.
func ServiceSpan(ctx context.Context, log logger.Logger, method string) (*Span, context.Context) {
	return StartSpan(ctx, log, "service:"+method)
}

// truncateOperation truncates operation name for readability.
func truncateOperation(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// --- SpanFromContext helper ---

// SpanLogger wraps logger with automatic span context.
type SpanLogger struct {
	log logger.Logger
	ctx context.Context
}

// NewSpanLogger creates a span logger that automatically uses span context.
func NewSpanLogger(log logger.Logger, ctx context.Context) *SpanLogger {
	return &SpanLogger{log: log, ctx: ctx}
}

// Debug logs with span context.
func (l *SpanLogger) Debug(msg string, fields ...logger.Field) {
	l.log.Debug(l.ctx, msg, fields...)
}

// Info logs with span context.
func (l *SpanLogger) Info(msg string, fields ...logger.Field) {
	l.log.Info(l.ctx, msg, fields...)
}

// Warn logs with span context.
func (l *SpanLogger) Warn(msg string, fields ...logger.Field) {
	l.log.Warn(l.ctx, msg, fields...)
}

// Error logs with span context.
func (l *SpanLogger) Error(msg string, fields ...logger.Field) {
	l.log.Error(l.ctx, msg, fields...)
}
