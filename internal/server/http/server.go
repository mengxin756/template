package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"example.com/classic/internal/config"
	"example.com/classic/internal/handler"
	"example.com/classic/pkg/logger"
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
			Addr:              cfg.HTTP.Address,
			Handler:           engine,
			ReadTimeout:       cfg.HTTP.ReadTimeout,
			WriteTimeout:      cfg.HTTP.WriteTimeout,
			IdleTimeout:       cfg.HTTP.IdleTimeout,
			MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
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

	// 请求ID中间件
	s.engine.Use(s.requestIDMiddleware())

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
			users.POST("", userHandler.Register)                    // 用户注册
			users.GET("", userHandler.List)                        // 用户列表
			users.GET("/:id", userHandler.GetByID)                 // 获取用户
			users.PUT("/:id", userHandler.Update)                  // 更新用户
			users.DELETE("/:id", userHandler.Delete)               // 删除用户
			users.PATCH("/:id/status", userHandler.ChangeStatus)   // 改变用户状态
		}
	}
}

// requestIDMiddleware 请求ID中间件
func (s *Server) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取 trace_id，如果没有则生成新的
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = generateTraceID()
		}

		// 设置到上下文
		ctx := logger.ContextWithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		// 设置响应头
		c.Header("X-Trace-ID", traceID)

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

		// 记录访问日志
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		if raw != "" {
			path = path + "?" + raw
		}

		ctx := c.Request.Context()
		s.log.Info(ctx, "HTTP request",
			logger.F("method", method),
			logger.F("path", path),
			logger.F("status", status),
			logger.F("latency", latency),
			logger.F("client_ip", clientIP),
			logger.F("user_agent", userAgent),
		)
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

// generateTraceID 生成追踪ID
func generateTraceID() string {
	// 简单的追踪ID生成，实际项目中可以使用 UUID
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// GetHTTPServer 获取 HTTP 服务器实例
func (s *Server) GetHTTPServer() *http.Server {
	return s.server
}
