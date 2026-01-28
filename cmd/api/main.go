package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/classic/internal/wire"
	"example.com/classic/pkg/logger"
)

func main() {
	// 使用 Wire 注入获取所有依赖
	server, cfg, log, err := wire.InitHTTPServer(context.Background())
	if err != nil {
		panic(fmt.Errorf("wire init: %w", err))
	}
	defer func() { _ = log.Sync() }()

	// 设置全局日志
	logger.SetGlobalLogger(log)

	// 启动 HTTP 服务器
	go func() {
		log.Info(context.Background(), "HTTP server starting", logger.F("addr", cfg.HTTP.Address))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(context.Background(), "HTTP server error", logger.F("error", err))
		}
	}()

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info(context.Background(), "shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Error(context.Background(), "server shutdown error", logger.F("error", err))
	}
	
	log.Info(context.Background(), "server exited")
}
