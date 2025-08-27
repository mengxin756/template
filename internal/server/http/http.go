package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"example.com/classic/internal/config"
	"example.com/classic/internal/logger"
	"example.com/classic/internal/server/http/middleware"
	"example.com/classic/internal/server/http/router"
)

func BuildEngine(cfg *config.Config, log *logger.Logger, userHandler router.UserHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// 中间件：请求ID、访问日志、恢复
	engine.Use(middleware.RequestID())
	engine.Use(middleware.AccessLogger(log))
	engine.Use(middleware.Recovery(log))

	// 健康检查
	engine.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	// 示例路由
	engine.GET("/api/v1/ping", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"pong": true})
	})

	// 用户路由
	if userHandler != nil {
		router.Register(engine, userHandler)
	}

	return engine
}

type noopUserHandler struct{}

func (n *noopUserHandler) Get(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"msg": "not implemented"})
}
func (n *noopUserHandler) Create(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"msg": "not implemented"})
}
