package asynq

import (
	"context"
	"time"

	"example.com/classic/internal/config"
	"example.com/classic/pkg/logger"
	"github.com/hibiken/asynq"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	scheduler *asynq.Scheduler
	config    *config.Config
	log       logger.Logger
}

// NewScheduler 创建调度器
func NewScheduler(cfg *config.Config, log logger.Logger) (*Scheduler, error) {
	// 复用 Asynq Redis 配置
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.Asynq.RedisAddr,
		Password: cfg.Asynq.RedisPassword,
		DB:       cfg.Asynq.RedisDB,
	}

	sch := asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{
		// 可根据需要设置 Location 等
		Location: time.Local,
	})

	s := &Scheduler{
		scheduler: sch,
		config:    cfg,
		log:       log,
	}

	// 注册默认示例任务
	s.registerDefaultCrons()

	return s, nil
}

// registerDefaultCrons 注册示例 Cron 任务
func (s *Scheduler) registerDefaultCrons() {
	// 每天 03:30 进行数据清理示例
	cleanupTask := NewDataCleanupTask("logs", 30)
	if _, err := s.scheduler.Register("30 3 * * *", cleanupTask); err != nil {
		s.log.Warn(context.Background(), "register cleanup cron failed", logger.F("error", err))
	}

	// 每小时检查一次状态变更通知示例（仅演示）
	noticeTask := NewStatusChangeNotificationTask(0, "", "", "", "", "system")
	if _, err := s.scheduler.Register("0 * * * *", noticeTask); err != nil {
		s.log.Warn(context.Background(), "register status notice cron failed", logger.F("error", err))
	}
}

// Start 启动 Scheduler（非阻塞）
func (s *Scheduler) Start() error {
	s.log.Info(context.Background(), "asynq scheduler starting")
	return s.scheduler.Start()
}

// Stop 停止 Scheduler（无返回值）
func (s *Scheduler) Stop() {
	s.log.Info(context.Background(), "asynq scheduler stopping")
	s.scheduler.Shutdown()
}
