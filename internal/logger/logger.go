package logger

import (
	"context"
	"strings"

	"example.com/classic/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
    *zap.Logger
}

func New(cfg *config.Config) *Logger {
    level := zap.InfoLevel
    if err := level.UnmarshalText([]byte(strings.ToLower(cfg.Log.Level))); err != nil {
        level = zap.InfoLevel
    }

    zapCfg := zap.Config{
        Level:            zap.NewAtomicLevelAt(level),
        Development:      cfg.Log.Development,
        Encoding:         cfg.Log.Encoding,
        EncoderConfig:    encoderConfig(),
        OutputPaths:      []string{"stdout"},
        ErrorOutputPaths: []string{"stderr"},
    }
    l, _ := zapCfg.Build()
    return &Logger{Logger: l}
}

func encoderConfig() zapcore.EncoderConfig {
    return zapcore.EncoderConfig{
        TimeKey:        "ts",
        LevelKey:       "level",
        NameKey:        "logger",
        CallerKey:      "caller",
        MessageKey:     "msg",
        StacktraceKey:  "stack",
        EncodeTime:     zapcore.ISO8601TimeEncoder,
        EncodeLevel:    zapcore.LowercaseLevelEncoder,
        EncodeDuration: zapcore.SecondsDurationEncoder,
        EncodeCaller:   zapcore.ShortCallerEncoder,
    }
}

// 统一字段辅助
func Field(key string, val any) zap.Field { return zap.Any(key, val) }
func Err(err error) zap.Field                   { return zap.Error(err) }

// 从 context 中提取 trace_id 并加入日志
type ctxKey string

const traceIDKey ctxKey = "trace_id"

func WithTrace(ctx context.Context, l *Logger) *Logger {
    if ctx == nil {
        return l
    }
    if v := ctx.Value(traceIDKey); v != nil {
        if s, ok := v.(string); ok && s != "" {
            return &Logger{Logger: l.With(zap.String("trace_id", s))}
        }
    }
    return l
}

func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
    return context.WithValue(ctx, traceIDKey, traceID)
}


