package logger

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"example.com/classic/pkg/contextx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger 日志接口
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field) // 添加 Fatal 级别
	Panic(ctx context.Context, msg string, fields ...Field)
	WithContext(ctx context.Context) Logger
	With(fields ...Field) Logger
	Sync() error
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// F 创建日志字段 (通用，优先使用类型安全函数)
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// String 创建字符串字段
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int 创建整数字段
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 创建 int64 字段
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 创建 float64 字段
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool 创建布尔字段
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Err 创建错误字段
func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

// Duration 创建时间间隔字段
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
}

// Time 创建时间字段
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value.Format(time.RFC3339)}
}

// traceHook 从 context 中提取追踪信息并添加到日志
type traceHook struct{}

func (h traceHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	if ctx == nil {
		return
	}

	// 提取所有追踪字段
	if traceID := contextx.GetTraceID(ctx); traceID != "" {
		e.Str("trace_id", traceID)
	}
	if spanID := contextx.GetSpanID(ctx); spanID != "" {
		e.Str("span_id", spanID)
	}
	if parentSpanID := contextx.GetParentSpanID(ctx); parentSpanID != "" {
		e.Str("parent_span_id", parentSpanID)
	}
	if userID := contextx.GetUserID(ctx); userID != "" {
		e.Str("user_id", userID)
	}
	if requestID := contextx.GetRequestID(ctx); requestID != "" {
		e.Str("request_id", requestID)
	}
	if clientIP := contextx.GetClientIP(ctx); clientIP != "" {
		e.Str("client_ip", clientIP)
	}
	if operation := contextx.GetOperationName(ctx); operation != "" {
		e.Str("operation", operation)
	}
}

// logger 日志实现
type logger struct {
	log zerolog.Logger
}

// New creates a new logger instance.
// If logDir is provided, logs will be written to both console and file.
func New(service string, level string, isDevelopment bool, logDir ...string) Logger {
	// Set log level
	logLevel := zerolog.InfoLevel
	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	case "fatal":
		logLevel = zerolog.FatalLevel
	case "panic":
		logLevel = zerolog.PanicLevel
	}

	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(logLevel)

	// Determine output writer
	var writer io.Writer
	if isDevelopment {
		// Development: console with colors
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	} else {
		// Production: JSON to stdout
		writer = os.Stdout
	}

	// If logDir is provided, also write to file
	if len(logDir) > 0 && logDir[0] != "" {
		logPath := filepath.Join(logDir[0], "app.log")
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			// MultiWriter: write to both console and file
			writer = io.MultiWriter(writer, file)
		}
	}

	log.Logger = log.Output(writer)

	// Create base logger instance
	baseLogger := log.With().
		Str("service", service).
		Timestamp().
		Logger()

	// Add trace hook
	baseLogger = baseLogger.Hook(traceHook{})

	return &logger{log: baseLogger}
}

// Debug 调试日志
func (l *logger) Debug(ctx context.Context, msg string, fields ...Field) {
	event := l.log.Debug()
	if ctx != nil {
		event = event.Ctx(ctx)
	}
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Info 信息日志
func (l *logger) Info(ctx context.Context, msg string, fields ...Field) {
	event := l.log.Info()
	if ctx != nil {
		event = event.Ctx(ctx)
	}
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Warn 警告日志
func (l *logger) Warn(ctx context.Context, msg string, fields ...Field) {
	event := l.log.Warn()
	if ctx != nil {
		event = event.Ctx(ctx)
	}
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Error 错误日志
func (l *logger) Error(ctx context.Context, msg string, fields ...Field) {
	event := l.log.Error()
	if ctx != nil {
		event = event.Ctx(ctx)
	}
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Fatal 致命错误日志 (调用后程序退出)
func (l *logger) Fatal(ctx context.Context, msg string, fields ...Field) {
	event := l.log.Fatal()
	if ctx != nil {
		event = event.Ctx(ctx)
	}
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Panic 严重错误日志
func (l *logger) Panic(ctx context.Context, msg string, fields ...Field) {
	event := l.log.Panic()
	if ctx != nil {
		event = event.Ctx(ctx)
	}
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// WithContext creates a logger instance bound to context
func (l *logger) WithContext(ctx context.Context) Logger {
	return &logger{log: l.log.With().Ctx(ctx).Logger()}
}

// With 添加固定字段到日志实例
func (l *logger) With(fields ...Field) Logger {
	ctx := l.log.With()
	for _, field := range fields {
		ctx = ctx.Interface(field.Key, field.Value)
	}
	return &logger{log: ctx.Logger()}
}

// Sync 同步日志
func (l *logger) Sync() error {
	// zerolog 不需要同步，返回 nil
	return nil
}

// ContextWithTraceID 在上下文中设置 trace_id (兼容旧代码，推荐使用 contextx 包)
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return contextx.WithTraceID(ctx, traceID)
}

// 便捷函数
func Debug(ctx context.Context, msg string, fields ...Field) {
	GetLogger().Debug(ctx, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...Field) {
	GetLogger().Info(ctx, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...Field) {
	GetLogger().Warn(ctx, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...Field) {
	GetLogger().Error(ctx, msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...Field) {
	GetLogger().Fatal(ctx, msg, fields...)
}

func Panic(ctx context.Context, msg string, fields ...Field) {
	GetLogger().Panic(ctx, msg, fields...)
}

// 全局日志实例
var globalLogger Logger

// SetGlobalLogger 设置全局日志实例
func SetGlobalLogger(l Logger) {
	globalLogger = l
}

// GetLogger 获取全局日志实例
func GetLogger() Logger {
	if globalLogger == nil {
		// 默认日志实例
		globalLogger = New("unknown", "info", false)
	}
	return globalLogger
}
