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

// traceIDHook 从 context 中提取 trace_id 并添加到日志
type traceIDHook struct{}

func (h traceIDHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if ctx := e.GetCtx(); ctx != nil {
		if traceID := getTraceID(ctx); traceID != "" {
			e.Str("trace_id", traceID)
		}
	}
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
	case "fatal":
		logLevel = zerolog.FatalLevel
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

	// 添加 trace_id hook
	baseLogger = baseLogger.Hook(traceIDHook{})

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
