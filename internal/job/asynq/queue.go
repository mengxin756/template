package asynq

import (
	"context"
	"fmt"
	"time"

	"example.com/classic/internal/config"
	"example.com/classic/pkg/logger"
	"github.com/hibiken/asynq"
)

// Queue 任务队列
type Queue struct {
	client   *asynq.Client
	server   *asynq.Server
	config   *config.Config
	log      logger.Logger
	handlers map[string]asynq.HandlerFunc
}

// New 创建任务队列
func New(cfg *config.Config, log logger.Logger) (*Queue, error) {
	// 创建 Redis 连接选项
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.Asynq.RedisAddr,
		Password: cfg.Asynq.RedisPassword,
		DB:       cfg.Asynq.RedisDB,
	}

	// 创建客户端
	client := asynq.NewClient(redisOpt)

	// 创建服务器
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency:              cfg.Asynq.Concurrency,
			StrictPriority:           cfg.Asynq.StrictPriority,
			ShutdownTimeout:          cfg.Asynq.ShutdownTimeout,
			HealthCheckFunc:          func(error) {}, // 简单的健康检查
			DelayedTaskCheckInterval: time.Second,
		},
	)

	queue := &Queue{
		client:   client,
		server:   server,
		config:   cfg,
		log:      log,
		handlers: make(map[string]asynq.HandlerFunc),
	}

	// 注册默认任务处理器
	queue.registerDefaultHandlers()

	return queue, nil
}

// Start 启动任务队列服务器
func (q *Queue) Start() error {
	q.log.Info(context.Background(), "starting Asynq server",
		logger.F("concurrency", q.config.Asynq.Concurrency),
		logger.F("redis_addr", q.config.Asynq.RedisAddr))

	// 创建多路复用器
	mux := asynq.NewServeMux()

	// 注册任务处理器
	for taskType, handler := range q.handlers {
		mux.HandleFunc(taskType, handler)
		q.log.Debug(context.Background(), "registered task handler", logger.F("task_type", taskType))
	}

	// 启动服务器
	return q.server.Run(mux)
}

// Stop 停止任务队列服务器
func (q *Queue) Stop() error {
	q.log.Info(context.Background(), "stopping Asynq server")

	// 优雅关闭客户端
	q.client.Close()

	// 优雅关闭服务器
	q.server.Shutdown()

	return nil
}

// Enqueue 入队任务
func (q *Queue) Enqueue(task *asynq.Task, opts ...asynq.Option) error {
	info, err := q.client.Enqueue(task, opts...)
	if err != nil {
		q.log.Error(context.Background(), "failed to enqueue task",
			logger.F("error", err),
			logger.F("task_type", task.Type()))
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	q.log.Debug(context.Background(), "task enqueued successfully",
		logger.F("task_id", info.ID),
		logger.F("task_type", task.Type()),
		logger.F("queue", info.Queue))

	return nil
}

// EnqueueDelay 延迟入队任务
func (q *Queue) EnqueueDelay(delay time.Duration, task *asynq.Task, opts ...asynq.Option) error {
	opts = append(opts, asynq.ProcessIn(delay))
	return q.Enqueue(task, opts...)
}

// GetClient 获取 Asynq 客户端
func (q *Queue) GetClient() *asynq.Client {
	return q.client
}

// GetServer 获取 Asynq 服务器
func (q *Queue) GetServer() *asynq.Server {
	return q.server
}

// registerDefaultHandlers 注册默认任务处理器
func (q *Queue) registerDefaultHandlers() {
	// 用户注册欢迎邮件任务
	q.handlers[TaskTypeWelcomeEmail] = q.handleWelcomeEmail

	// 用户状态变更通知任务
	q.handlers[TaskTypeStatusChangeNotification] = q.handleStatusChangeNotification

	// 数据清理任务
	q.handlers[TaskTypeDataCleanup] = q.handleDataCleanup
}

// handleWelcomeEmail 处理欢迎邮件任务
func (q *Queue) handleWelcomeEmail(ctx context.Context, t *asynq.Task) error {
	q.log.Info(ctx, "processing welcome email task",
		logger.F("task_id", t.ResultWriter().TaskID()),
		logger.F("payload", string(t.Payload())))

	// 这里实现发送欢迎邮件的逻辑
	// 例如：解析任务数据，调用邮件服务等

	q.log.Info(ctx, "welcome email task completed successfully")
	return nil
}

// handleStatusChangeNotification 处理状态变更通知任务
func (q *Queue) handleStatusChangeNotification(ctx context.Context, t *asynq.Task) error {
	q.log.Info(ctx, "processing status change notification task",
		logger.F("task_id", t.ResultWriter().TaskID()),
		logger.F("payload", string(t.Payload())))

	// 这里实现状态变更通知的逻辑
	// 例如：发送邮件、短信、推送通知等

	q.log.Info(ctx, "status change notification task completed successfully")
	return nil
}

// handleDataCleanup 处理数据清理任务
func (q *Queue) handleDataCleanup(ctx context.Context, t *asynq.Task) error {
	q.log.Info(ctx, "processing data cleanup task",
		logger.F("task_id", t.ResultWriter().TaskID()),
		logger.F("payload", string(t.Payload())))

	// 这里实现数据清理的逻辑
	// 例如：清理过期日志、临时文件等

	q.log.Info(ctx, "data cleanup task completed successfully")
	return nil
}
