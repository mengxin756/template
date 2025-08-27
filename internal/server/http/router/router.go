package router

import (
	"github.com/gin-gonic/gin"
)

func Register(e *gin.Engine, userHandler UserHandler) {
	e.GET("/api/v1/users/:id", userHandler.Get)
	e.POST("/api/v1/users", userHandler.Create)
}

// UserHandler 定义处理器接口，便于依赖注入
// 由 handler 模块实现
