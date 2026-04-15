package http

import (
	"context"
	"net/http"
	"time"

	"example.com/classic/internal/config"
	"example.com/classic/internal/handler"
	"example.com/classic/pkg/contextx"
	"example.com/classic/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Server HTTP 服务器
type Server struct {
	engine *gin.Engine
	server *http.Server
	config *config.Config
	log    logger.Logger
}

// NewServer 创建 HTTP 服务器实例
func NewServer(cfg *config.Config, log logger.Logger, userHandler *handler.UserHandler) *Server {
	// 设置 Gin 模式
	if cfg.IsDevelopment() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// 创建服务器实例
	server := &Server{
		engine: engine,
		config: cfg,
		log:    log,
		server: &http.Server{
			Addr:           cfg.HTTP.Address,
			Handler:        engine,
			ReadTimeout:    cfg.HTTP.ReadTimeout,
			WriteTimeout:   cfg.HTTP.WriteTimeout,
			IdleTimeout:    cfg.HTTP.IdleTimeout,
			MaxHeaderBytes: cfg.HTTP.MaxHeaderBytes,
		},
	}

	// 配置中间件和路由
	server.setupMiddleware()
	server.setupRoutes(userHandler)

	return server
}

// setupMiddleware 配置中间件
func (s *Server) setupMiddleware() {
	// 恢复中间件
	s.engine.Use(gin.Recovery())

	// 链路追踪中间件 (最先执行)
	s.engine.Use(s.tracingMiddleware())

	// 访问日志中间件
	s.engine.Use(s.accessLogMiddleware())

	// CORS 中间件
	if s.config.HTTP.EnableCORS {
		s.engine.Use(s.corsMiddleware())
	}
}

// setupRoutes 配置路由
func (s *Server) setupRoutes(userHandler *handler.UserHandler) {
	// 健康检查
	s.engine.GET("/health", s.healthCheck)

	// API v1 路由组
	v1 := s.engine.Group("/api/v1")
	{
		// 用户相关路由
		users := v1.Group("/users")
		{
			users.POST("", userHandler.Register)                 // 用户注册
			users.GET("", userHandler.List)                      // 用户列表
			users.GET("/:id", userHandler.GetByID)               // 获取用户
			users.PUT("/:id", userHandler.Update)                // 更新用户
			users.DELETE("/:id", userHandler.Delete)             // 删除用户
			users.PATCH("/:id/status", userHandler.ChangeStatus) // 改变用户状态
		}
	}
}

// tracingMiddleware 链路追踪中间件 (增强版)
func (s *Server) tracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取追踪信息
		traceID := c.GetHeader("X-Trace-ID")
		parentSpanID := c.GetHeader("X-Span-ID")
		userID := c.GetHeader("X-User-ID")

		// 如果没有 trace_id 则生成新的
		if traceID == "" {
			traceID = contextx.GenerateTraceID()
		}

		// 生成当前请求的 span_id
		spanID := contextx.GenerateSpanID()

		// 构建追踪上下文
		ctx := c.Request.Context()
		ctx = contextx.WithTraceID(ctx, traceID)
		ctx = contextx.WithSpanID(ctx, spanID)
		ctx = contextx.WithUserID(ctx, userID)
		ctx = contextx.WithClientIP(ctx, c.ClientIP())
		ctx = contextx.WithUserAgent(ctx, c.Request.UserAgent())
		ctx = contextx.WithServiceName(ctx, s.config.Service)
		ctx = contextx.WithOperationName(ctx, c.Request.Method+" "+c.Request.URL.Path)

		// 如果有父 span，设置 parent_span_id
		if parentSpanID != "" {
			ctx = contextx.WithParentSpanID(ctx, parentSpanID)
		}

		c.Request = c.Request.WithContext(ctx)

		// 设置响应头 (便于下游服务追踪)
		c.Header("X-Trace-ID", traceID)
		c.Header("X-Span-ID", spanID)

		c.Next()
	}
}

// accessLogMiddleware 访问日志中间件
func (s *Server) accessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 记录访问日志 (追踪信息由 traceHook 自动注入)
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method

		if raw != "" {
			path = path + "?" + raw
		}

		ctx := c.Request.Context()
		level := "info"
		if status >= 400 {
			level = "warn"
		}
		if status >= 500 {
			level = "error"
		}

		switch level {
		case "error":
			s.log.Error(ctx, "HTTP request",
				logger.String("method", method),
				logger.String("path", path),
				logger.Int("status", status),
				logger.Duration("latency", latency))
		case "warn":
			s.log.Warn(ctx, "HTTP request",
				logger.String("method", method),
				logger.String("path", path),
				logger.Int("status", status),
				logger.Duration("latency", latency))
		default:
			s.log.Info(ctx, "HTTP request",
				logger.String("method", method),
				logger.String("path", path),
				logger.Int("status", status),
				logger.Duration("latency", latency))
		}
	}
}

// corsMiddleware CORS 中间件
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Trace-ID")
		c.Header("Access-Control-Expose-Headers", "X-Trace-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// healthCheck 健康检查
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   s.config.Service,
		"version":   s.config.Version,
	})
}

// Start 启动服务器
func (s *Server) Start() error {
	s.log.Info(context.Background(), "HTTP server starting", logger.F("address", s.config.HTTP.Address))
	return s.server.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	s.log.Info(ctx, "HTTP server stopping")
	return s.server.Shutdown(ctx)
}


// GetHTTPServer 获取 HTTP 服务器实例
func (s *Server) GetHTTPServer() *http.Server {
	return s.server
}
