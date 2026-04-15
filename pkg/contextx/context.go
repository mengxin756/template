// Package contextx provides enhanced context utilities for distributed tracing.
package contextx

import (
	"context"

	"github.com/google/uuid"
)

// Key types for context values to avoid collisions.
type key string

const (
	traceIDKey       key = "trace_id"
	spanIDKey        key = "span_id"
	parentSpanIDKey  key = "parent_span_id"
	userIDKey        key = "user_id"
	requestIDKey     key = "request_id"
	clientIPKey      key = "client_ip"
	userAgentKey     key = "user_agent"
	serviceNameKey   key = "service_name"
	operationNameKey key = "operation_name"
)

// TraceContext contains tracing information.
type TraceContext struct {
	TraceID       string
	SpanID        string
	ParentSpanID  string
	UserID        string
	RequestID     string
	ClientIP      string
	UserAgent     string
	ServiceName   string
	OperationName string
}

// NewTraceContext creates a new TraceContext with a generated trace ID.
func NewTraceContext() *TraceContext {
	return &TraceContext{
		TraceID:   GenerateTraceID(),
		SpanID:    GenerateSpanID(),
		RequestID: GenerateRequestID(),
	}
}

// GenerateTraceID generates a new trace ID.
func GenerateTraceID() string {
	return uuid.New().String()
}

// GenerateSpanID generates a new span ID.
func GenerateSpanID() string {
	return uuid.New().String()[:16]
}

// GenerateRequestID generates a new request ID.
func GenerateRequestID() string {
	return uuid.New().String()[:8]
}

// --- Context Setters ---

// WithTraceID sets trace ID in context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// WithSpanID sets span ID in context.
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, spanIDKey, spanID)
}

// WithParentSpanID sets parent span ID in context.
func WithParentSpanID(ctx context.Context, parentSpanID string) context.Context {
	return context.WithValue(ctx, parentSpanIDKey, parentSpanID)
}

// WithUserID sets user ID in context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// WithRequestID sets request ID in context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithClientIP sets client IP in context.
func WithClientIP(ctx context.Context, clientIP string) context.Context {
	return context.WithValue(ctx, clientIPKey, clientIP)
}

// WithUserAgent sets user agent in context.
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, userAgentKey, userAgent)
}

// WithServiceName sets service name in context.
func WithServiceName(ctx context.Context, serviceName string) context.Context {
	return context.WithValue(ctx, serviceNameKey, serviceName)
}

// WithOperationName sets operation name in context.
func WithOperationName(ctx context.Context, operationName string) context.Context {
	return context.WithValue(ctx, operationNameKey, operationName)
}

// WithTraceContext sets all trace context values at once.
func WithTraceContext(ctx context.Context, tc *TraceContext) context.Context {
	if tc.TraceID != "" {
		ctx = WithTraceID(ctx, tc.TraceID)
	}
	if tc.SpanID != "" {
		ctx = WithSpanID(ctx, tc.SpanID)
	}
	if tc.ParentSpanID != "" {
		ctx = WithParentSpanID(ctx, tc.ParentSpanID)
	}
	if tc.UserID != "" {
		ctx = WithUserID(ctx, tc.UserID)
	}
	if tc.RequestID != "" {
		ctx = WithRequestID(ctx, tc.RequestID)
	}
	if tc.ClientIP != "" {
		ctx = WithClientIP(ctx, tc.ClientIP)
	}
	if tc.UserAgent != "" {
		ctx = WithUserAgent(ctx, tc.UserAgent)
	}
	if tc.ServiceName != "" {
		ctx = WithServiceName(ctx, tc.ServiceName)
	}
	if tc.OperationName != "" {
		ctx = WithOperationName(ctx, tc.OperationName)
	}
	return ctx
}

// --- Context Getters ---

// GetTraceID retrieves trace ID from context.
func GetTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}

// GetSpanID retrieves span ID from context.
func GetSpanID(ctx context.Context) string {
	if v, ok := ctx.Value(spanIDKey).(string); ok {
		return v
	}
	return ""
}

// GetParentSpanID retrieves parent span ID from context.
func GetParentSpanID(ctx context.Context) string {
	if v, ok := ctx.Value(parentSpanIDKey).(string); ok {
		return v
	}
	return ""
}

// GetUserID retrieves user ID from context.
func GetUserID(ctx context.Context) string {
	if v, ok := ctx.Value(userIDKey).(string); ok {
		return v
	}
	return ""
}

// GetRequestID retrieves request ID from context.
func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// GetClientIP retrieves client IP from context.
func GetClientIP(ctx context.Context) string {
	if v, ok := ctx.Value(clientIPKey).(string); ok {
		return v
	}
	return ""
}

// GetUserAgent retrieves user agent from context.
func GetUserAgent(ctx context.Context) string {
	if v, ok := ctx.Value(userAgentKey).(string); ok {
		return v
	}
	return ""
}

// GetServiceName retrieves service name from context.
func GetServiceName(ctx context.Context) string {
	if v, ok := ctx.Value(serviceNameKey).(string); ok {
		return v
	}
	return ""
}

// GetOperationName retrieves operation name from context.
func GetOperationName(ctx context.Context) string {
	if v, ok := ctx.Value(operationNameKey).(string); ok {
		return v
	}
	return ""
}

// GetTraceContext retrieves all trace context from context.
func GetTraceContext(ctx context.Context) *TraceContext {
	return &TraceContext{
		TraceID:       GetTraceID(ctx),
		SpanID:        GetSpanID(ctx),
		ParentSpanID:  GetParentSpanID(ctx),
		UserID:        GetUserID(ctx),
		RequestID:     GetRequestID(ctx),
		ClientIP:      GetClientIP(ctx),
		UserAgent:     GetUserAgent(ctx),
		ServiceName:   GetServiceName(ctx),
		OperationName: GetOperationName(ctx),
	}
}

// ChildSpan creates a child span context from current context.
// The current span becomes the parent of the new span.
func ChildSpan(ctx context.Context) context.Context {
	parentSpanID := GetSpanID(ctx)
	if parentSpanID != "" {
		ctx = WithParentSpanID(ctx, parentSpanID)
	}
	return WithSpanID(ctx, GenerateSpanID())
}
