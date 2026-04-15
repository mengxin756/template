package asynq

import (
	"context"
	"fmt"
	"time"

	"example.com/classic/internal/config"
	"example.com/classic/internal/taskqueue"
	"example.com/classic/pkg/logger"
	"github.com/hibiken/asynq"
)

// Queue 任务队列 (实现 taskqueue.TaskQueue 接口)
var _ taskqueue.TaskQueue = (*Queue)(nil)

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

// Start 启动任务队列服务器 (实现 taskqueue.TaskQueue 接口)
func (q *Queue) Start(ctx context.Context) error {
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

// Stop 停止任务队列服务器 (实现 taskqueue.TaskQueue 接口)
func (q *Queue) Stop(ctx context.Context) error {
	q.log.Info(context.Background(), "stopping Asynq server")

	// 优雅关闭客户端
	q.client.Close()

	// 优雅关闭服务器
	q.server.Shutdown()

	return nil
}

// Enqueue 入队任务 (实现 taskqueue.TaskQueue 接口)
func (q *Queue) Enqueue(ctx context.Context, task *taskqueue.Task, opts ...taskqueue.Option) (*taskqueue.TaskResult, error) {
	options := taskqueue.DefaultEnqueueOptions()
	for _, opt := range opts {
		opt(options)
	}

	asynqTask := asynq.NewTask(task.Type, task.Payload)
	asynqOpts := q.convertOptions(options)

	info, err := q.client.Enqueue(asynqTask, asynqOpts...)
	if err != nil {
		q.log.Error(ctx, "failed to enqueue task",
			logger.Err(err),
			logger.String("task_type", task.Type))
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	q.log.Debug(ctx, "task enqueued successfully",
		logger.String("task_id", info.ID),
		logger.String("task_type", task.Type),
		logger.String("queue", info.Queue))

	return &taskqueue.TaskResult{
		ID:    info.ID,
		Queue: info.Queue,
		Type:  task.Type,
	}, nil
}

// EnqueueIn 延迟入队任务 (实现 taskqueue.TaskQueue 接口)
func (q *Queue) EnqueueIn(ctx context.Context, task *taskqueue.Task, delay time.Duration, opts ...taskqueue.Option) (*taskqueue.TaskResult, error) {
	options := taskqueue.DefaultEnqueueOptions()
	for _, opt := range opts {
		opt(options)
	}

	asynqTask := asynq.NewTask(task.Type, task.Payload)
	asynqOpts := q.convertOptions(options)
	asynqOpts = append(asynqOpts, asynq.ProcessIn(delay))

	info, err := q.client.Enqueue(asynqTask, asynqOpts...)
	if err != nil {
		q.log.Error(ctx, "failed to enqueue delayed task",
			logger.Err(err),
			logger.String("task_type", task.Type),
			logger.Duration("delay", delay))
		return nil, fmt.Errorf("failed to enqueue delayed task: %w", err)
	}

	return &taskqueue.TaskResult{
		ID:    info.ID,
		Queue: info.Queue,
		Type:  task.Type,
	}, nil
}

// EnqueueAt 定时入队任务 (实现 taskqueue.TaskQueue 接口)
func (q *Queue) EnqueueAt(ctx context.Context, task *taskqueue.Task, processAt time.Time, opts ...taskqueue.Option) (*taskqueue.TaskResult, error) {
	options := taskqueue.DefaultEnqueueOptions()
	for _, opt := range opts {
		opt(options)
	}

	asynqTask := asynq.NewTask(task.Type, task.Payload)
	asynqOpts := q.convertOptions(options)
	asynqOpts = append(asynqOpts, asynq.ProcessAt(processAt))

	info, err := q.client.Enqueue(asynqTask, asynqOpts...)
	if err != nil {
		q.log.Error(ctx, "failed to enqueue scheduled task",
			logger.Err(err),
			logger.String("task_type", task.Type),
			logger.Time("process_at", processAt))
		return nil, fmt.Errorf("failed to enqueue scheduled task: %w", err)
	}

	return &taskqueue.TaskResult{
		ID:    info.ID,
		Queue: info.Queue,
		Type:  task.Type,
	}, nil
}

// RegisterHandler 注册任务处理器 (实现 taskqueue.TaskQueue 接口)
func (q *Queue) RegisterHandler(taskType string, handler taskqueue.Handler) error {
	q.handlers[taskType] = func(ctx context.Context, t *asynq.Task) error {
		return handler.Process(ctx, &taskqueue.Task{
			Type:    t.Type(),
			Payload: t.Payload(),
		})
	}
	q.log.Debug(context.Background(), "registered task handler", logger.String("task_type", taskType))
	return nil
}

// convertOptions 转换选项为 asynq 选项
func (q *Queue) convertOptions(opts *taskqueue.EnqueueOptions) []asynq.Option {
	var asynqOpts []asynq.Option

	if opts.Queue != "" && opts.Queue != "default" {
		asynqOpts = append(asynqOpts, asynq.Queue(opts.Queue))
	}
	if opts.MaxRetry > 0 {
		asynqOpts = append(asynqOpts, asynq.MaxRetry(opts.MaxRetry))
	}
	if opts.Timeout > 0 {
		asynqOpts = append(asynqOpts, asynq.Timeout(opts.Timeout))
	}
	if opts.Unique {
		asynqOpts = append(asynqOpts, asynq.Unique(time.Hour)) // 默认 1 小时内唯一
	}
	if opts.TaskID != "" {
		asynqOpts = append(asynqOpts, asynq.TaskID(opts.TaskID))
	}

	return asynqOpts
}

// GetClient 获取 Asynq 客户端 (保留用于高级用法)
func (q *Queue) GetClient() *asynq.Client {
	return q.client
}

// GetServer 获取 Asynq 服务器 (保留用于高级用法)
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
