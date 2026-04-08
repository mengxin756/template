package messaging

import (
	"fmt"

	"example.com/classic/internal/domain"
	"example.com/classic/internal/job/asynq"
	"example.com/classic/pkg/logger"
)

// AsynqEventPublisher 基于 Asynq 的事件发布器实现
type AsynqEventPublisher struct {
	handlers map[string][]domain.EventProcessor
	log      logger.Logger
}

// NewAsynqEventPublisher 创建 Asynq 事件发布器
func NewAsynqEventPublisher(taskQueue *asynq.Queue, log logger.Logger) domain.EventPublisher {
	publisher := &AsynqEventPublisher{
		handlers: make(map[string][]domain.EventProcessor),
		log:      log,
	}

	// 自动注册默认处理器
	publisher.RegisterHandler(NewUserCreatedHandler(taskQueue))
	publisher.RegisterHandler(NewUserStatusChangedHandler(taskQueue))
	publisher.RegisterHandler(NewUserUpdatedHandler(taskQueue))
	publisher.RegisterHandler(NewUserDeletedHandler(taskQueue))

	return publisher
}

// handlerBase 处理器基类
type handlerBase struct {
	taskQueue *asynq.Queue
	eventType string
}

func (h *handlerBase) EventType() string {
	return h.eventType
}

// UserCreatedHandler 用户创建事件处理器
type UserCreatedHandler struct {
	*handlerBase
}

func NewUserCreatedHandler(taskQueue *asynq.Queue) domain.EventProcessor {
	return &UserCreatedHandler{
		handlerBase: &handlerBase{
			taskQueue: taskQueue,
			eventType: "user.created",
		},
	}
}

func (h *UserCreatedHandler) Process(event domain.DomainEvent) error {
	if userEvent, ok := event.(*domain.UserCreatedEvent); ok {
		task := asynq.NewWelcomeEmailTask(userEvent.UserID, userEvent.Email, userEvent.Name)
		const delaySeconds = 10
		if err := h.taskQueue.EnqueueDelay(delaySeconds*1000000000, task); err != nil {
			return fmt.Errorf("failed to enqueue welcome email task: %w", err)
		}
		return nil
	}
	return fmt.Errorf("unexpected event type: %T", event)
}

// UserStatusChangedHandler 用户状态变更事件处理器
type UserStatusChangedHandler struct {
	*handlerBase
}

func NewUserStatusChangedHandler(taskQueue *asynq.Queue) domain.EventProcessor {
	return &UserStatusChangedHandler{
		handlerBase: &handlerBase{
			taskQueue: taskQueue,
			eventType: "user.status_changed",
		},
	}
}

func (h *UserStatusChangedHandler) Process(event domain.DomainEvent) error {
	if userEvent, ok := event.(*domain.UserStatusChangedEvent); ok {
		task := asynq.NewStatusChangeNotificationTask(
			userEvent.UserID,
			userEvent.Email,
			userEvent.Name,
			string(userEvent.OldStatus),
			string(userEvent.NewStatus),
			"system",
		)
		if err := h.taskQueue.Enqueue(task); err != nil {
			return fmt.Errorf("failed to enqueue status change notification task: %w", err)
		}
		return nil
	}
	return fmt.Errorf("unexpected event type: %T", event)
}

// UserUpdatedHandler 用户更新事件处理器
type UserUpdatedHandler struct {
	*handlerBase
}

func NewUserUpdatedHandler(taskQueue *asynq.Queue) domain.EventProcessor {
	return &UserUpdatedHandler{
		handlerBase: &handlerBase{
			taskQueue: taskQueue,
			eventType: "user.updated",
		},
	}
}

func (h *UserUpdatedHandler) Process(event domain.DomainEvent) error {
	if _, ok := event.(*domain.UserUpdatedEvent); ok {
		// TODO: 实现用户更新通知
		// task := NewUserUpdatedNotificationTask(...)
		// h.taskQueue.Enqueue(task)
		return nil
	}
	return fmt.Errorf("unexpected event type: %T", event)
}

// UserDeletedHandler 用户删除事件处理器
type UserDeletedHandler struct {
	*handlerBase
}

func NewUserDeletedHandler(taskQueue *asynq.Queue) domain.EventProcessor {
	return &UserDeletedHandler{
		handlerBase: &handlerBase{
			taskQueue: taskQueue,
			eventType: "user.deleted",
		},
	}
}

func (h *UserDeletedHandler) Process(event domain.DomainEvent) error {
	if _, ok := event.(*domain.UserDeletedEvent); ok {
		// TODO: 实现用户删除通知
		// task := NewUserDeletedNotificationTask(...)
		// h.taskQueue.Enqueue(task)
		return nil
	}
	return fmt.Errorf("unexpected event type: %T", event)
}

// RegisterHandler 注册事件处理器
func (p *AsynqEventPublisher) RegisterHandler(processor domain.EventProcessor) {
	eventType := processor.EventType()
	p.handlers[eventType] = append(p.handlers[eventType], processor)
	p.log.Info(nil, "event handler registered",
		logger.F("event_type", eventType),
		logger.F("handler_count", len(p.handlers[eventType])))
}

// Publish 发布单个事件
func (p *AsynqEventPublisher) Publish(event domain.DomainEvent) error {
	p.log.Debug(nil, "publishing domain event",
		logger.F("event_type", event.EventType()),
		logger.F("aggregate_id", event.AggregateID()))

	// 获取该事件类型的所有处理器
	processors, exists := p.handlers[event.EventType()]
	if !exists {
		p.log.Debug(nil, "no handlers registered for event type",
			logger.F("event_type", event.EventType()))
		return nil
	}

	// 调用所有处理器
	for _, processor := range processors {
		if err := processor.Process(event); err != nil {
			p.log.Error(nil, "failed to process event",
				logger.F("event_type", event.EventType()),
				logger.F("error", err))
			// 继续处理其他处理器，不中断
		}
	}

	return nil
}

// PublishBatch 批量发布事件
func (p *AsynqEventPublisher) PublishBatch(events []domain.DomainEvent) error {
	for _, event := range events {
		if err := p.Publish(event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", event.EventType(), err)
		}
	}
	return nil
}