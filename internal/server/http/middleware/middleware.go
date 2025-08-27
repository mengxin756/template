package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"example.com/classic/internal/logger"
)

const RequestIDHeader = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(RequestIDHeader)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Writer.Header().Set(RequestIDHeader, rid)
		c.Set("trace_id", rid)
		c.Next()
	}
}

func AccessLogger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 读取部分请求体用于日志（最大 1KB）
		var bodySnippet string
		if c.Request.Body != nil {
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(c.Request.Body)
			b := buf.Bytes()
			if len(b) > 1024 {
				b = b[:1024]
			}
			bodySnippet = string(b)
			c.Request.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
		}

		c.Next()

		latency := time.Since(start)
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.Int64("bytes_in", c.Request.ContentLength),
			zap.Int("bytes_out", c.Writer.Size()),
		}
		if bodySnippet != "" {
			fields = append(fields, zap.String("body", bodySnippet))
		}

		traceID := c.GetString("trace_id")
		l := log.Logger
		if traceID != "" {
			l = l.With(zap.String("trace_id", traceID))
		}

		if c.Writer.Status() >= 500 {
			l.Error("access", fields...)
		} else {
			l.Info("access", fields...)
		}
	}
}

func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				traceID := c.GetString("trace_id")
				l := log.Logger
				if traceID != "" {
					l = l.With(zap.String("trace_id", traceID))
				}
				l.Error("panic",
					zap.Any("error", r),
					zap.String("path", c.Request.URL.Path),
				)
				c.AbortWithStatusJSON(500, gin.H{"code": 500, "msg": "internal error", "trace_id": traceID})
			}
		}()
		c.Next()
	}
}
