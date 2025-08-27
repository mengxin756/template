package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/classic/internal/logger"
	"example.com/classic/internal/wire"
)

func main() {
	// 使用 Wire 注入获取所有依赖
	server, cfg, log, err := wire.InitHTTPServer(context.Background())
	if err != nil {
		panic(fmt.Errorf("wire init: %w", err))
	}
	defer func() { _ = log.Sync() }()

	go func() {
		log.Info("http server starting", logger.Field("addr", cfg.HTTP.Address))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http server error", logger.Err(err))
		}
	}()

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("server shutdown error", logger.Err(err))
	}
	log.Info("server exited")
}
