package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"example.com/classic/internal/config"
	"example.com/classic/internal/job/asynq"
	"example.com/classic/pkg/logger"
)

func main() {
	ctx := context.Background()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fallback := logger.New("fallback", "error", true)
		fallback.Error(ctx, "failed to load config", logger.F("error", err))
		os.Exit(1)
	}

	// 初始化日志
	log := logger.New(cfg.Service+"-asynq", cfg.Log.Level, cfg.IsDevelopment())
	logger.SetGlobalLogger(log)

	// 创建任务队列
	queue, err := asynq.New(cfg, log)
	if err != nil {
		log.Error(ctx, "failed to init asynq queue", logger.F("error", err))
		os.Exit(1)
	}

	// 启动服务器（阻塞式）
	go func() {
		if err := queue.Start(); err != nil {
			log.Error(ctx, "asynq server exited with error", logger.F("error", err))
			os.Exit(1)
		}
	}()

	log.Info(ctx, "asynq server started")

	// 监听退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Info(ctx, "asynq server stopping...")
	_ = queue.Stop()
	log.Info(ctx, "asynq server stopped")
}
