package logger

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger 日志接口
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Panic(ctx context.Context, msg string, fields ...Field)
	WithContext(ctx context.Context) Logger
	Sync() error
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// F 创建日志字段
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// logger 日志实现
type logger struct {
	log zerolog.Logger
}

// New 创建新的日志实例
func New(service string, level string, isDevelopment bool) Logger {
	// 设置日志级别
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
	case "panic":
		logLevel = zerolog.PanicLevel
	}

	// 配置 zerolog
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(logLevel)

	// 开发环境使用彩色输出
	if isDevelopment {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	} else {
		log.Logger = log.Output(os.Stdout)
	}

	// 创建基础日志实例
	baseLogger := log.With().
		Str("service", service).
		Timestamp().
		Logger()

	return &logger{log: baseLogger}
}

// withTraceID 添加 trace_id 到日志上下文
func (l *logger) withTraceID(ctx context.Context) zerolog.Context {
	logCtx := l.log.With()

	if ctx != nil {
		if traceID := getTraceID(ctx); traceID != "" {
			logCtx = logCtx.Str("trace_id", traceID)
		}
	}

	return logCtx
}

// Debug 调试日志
func (l *logger) Debug(ctx context.Context, msg string, fields ...Field) {
	logCtx := l.withTraceID(ctx)
	for _, field := range fields {
		logCtx = logCtx.Interface(field.Key, field.Value)
	}
	logger := logCtx.Logger()
	logger.Debug().Msg(msg)
}

// Info 信息日志
func (l *logger) Info(ctx context.Context, msg string, fields ...Field) {
	logCtx := l.withTraceID(ctx)
	for _, field := range fields {
		logCtx = logCtx.Interface(field.Key, field.Value)
	}
	logger := logCtx.Logger()
	logger.Info().Msg(msg)
}

// Warn 警告日志
func (l *logger) Warn(ctx context.Context, msg string, fields ...Field) {
	logCtx := l.withTraceID(ctx)
	for _, field := range fields {
		logCtx = logCtx.Interface(field.Key, field.Value)
	}
	logger := logCtx.Logger()
	logger.Warn().Msg(msg)
}

// Error 错误日志
func (l *logger) Error(ctx context.Context, msg string, fields ...Field) {
	logCtx := l.withTraceID(ctx)
	for _, field := range fields {
		logCtx = logCtx.Interface(field.Key, field.Value)
	}
	logger := logCtx.Logger()
	logger.Error().Msg(msg)
}

// Panic 严重错误日志
func (l *logger) Panic(ctx context.Context, msg string, fields ...Field) {
	logCtx := l.withTraceID(ctx)
	for _, field := range fields {
		logCtx = logCtx.Interface(field.Key, field.Value)
	}
	logger := logCtx.Logger()
	logger.Panic().Msg(msg)
}

// WithContext 创建带上下文的日志实例
func (l *logger) WithContext(ctx context.Context) Logger {
	return &logger{log: l.withTraceID(ctx).Logger()}
}

// Sync 同步日志
func (l *logger) Sync() error {
	// zerolog 不需要同步，返回 nil
	return nil
}

// 上下文相关
type contextKey string

const traceIDKey contextKey = "trace_id"

// ContextWithTraceID 在上下文中设置 trace_id
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// getTraceID 从上下文中获取 trace_id
func getTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
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
