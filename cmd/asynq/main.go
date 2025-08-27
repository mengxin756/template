package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"example.com/classic/internal/config"
	"example.com/classic/internal/job/asynq"
	"example.com/classic/pkg/logger"
)

func main() {
	ctx := context.Background()

	mode := flag.String("mode", "worker", "asynq run mode: worker | scheduler")
	flag.Parse()

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

	switch *mode {
	case "worker":
		runWorker(ctx, cfg, log)
	case "scheduler":
		runScheduler(ctx, cfg, log)
	default:
		log.Error(ctx, "invalid mode", logger.F("mode", *mode))
		os.Exit(1)
	}
}

func runWorker(ctx context.Context, cfg *config.Config, log logger.Logger) {
	queue, err := asynq.New(cfg, log)
	if err != nil {
		log.Error(ctx, "failed to init asynq queue", logger.F("error", err))
		os.Exit(1)
	}

	go func() {
		if err := queue.Start(); err != nil {
			log.Error(ctx, "asynq server exited with error", logger.F("error", err))
			os.Exit(1)
		}
	}()

	log.Info(ctx, "asynq worker started")
	waitForSignal()
	log.Info(ctx, "asynq worker stopping...")
	_ = queue.Stop()
	log.Info(ctx, "asynq worker stopped")
}

func runScheduler(ctx context.Context, cfg *config.Config, log logger.Logger) {
	sch, err := asynq.NewScheduler(cfg, log)
	if err != nil {
		log.Error(ctx, "failed to init asynq scheduler", logger.F("error", err))
		os.Exit(1)
	}
	if err := sch.Start(); err != nil {
		log.Error(ctx, "failed to start scheduler", logger.F("error", err))
		os.Exit(1)
	}

	log.Info(ctx, "asynq scheduler started")
	waitForSignal()
	log.Info(ctx, "asynq scheduler stopping...")
	sch.Stop()
	log.Info(ctx, "asynq scheduler stopped")
}

func waitForSignal() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
