package domain

import (
	"strconv"
	"time"
)

// AggregateID generates aggregate ID from type and ID
func AggregateID(aggregateType string, id int) string {
	return aggregateType + "-" + strconv.Itoa(id)
}

// DomainEvent 领域事件接口
type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
	AggregateID() string
}

// EventRecorder 事件记录器，用于在聚合根中收集事件
type EventRecorder struct {
	events []DomainEvent
}

// NewEventRecorder 创建新的事件记录器
func NewEventRecorder() *EventRecorder {
	return &EventRecorder{
		events: make([]DomainEvent, 0),
	}
}

// AddEvent 添加事件
func (r *EventRecorder) AddEvent(event DomainEvent) {
	r.events = append(r.events, event)
}

// Events 返回所有事件
func (r *EventRecorder) Events() []DomainEvent {
	return r.events
}

// ClearEvents 清除所有事件
func (r *EventRecorder) ClearEvents() {
	r.events = make([]DomainEvent, 0)
}

// HasEvents 检查是否有事件
func (r *EventRecorder) HasEvents() bool {
	return len(r.events) > 0
}

// UserCreatedEvent 用户创建事件
type UserCreatedEvent struct {
	UserID    int
	Email     string
	Name      string
	occurredAt time.Time
}

func NewUserCreatedEvent(userID int, email, name string) *UserCreatedEvent {
	return &UserCreatedEvent{
		UserID:    userID,
		Email:     email,
		Name:      name,
		occurredAt: time.Now(),
	}
}

func (e *UserCreatedEvent) EventType() string {
	return "user.created"
}

func (e *UserCreatedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *UserCreatedEvent) AggregateID() string {
	return AggregateID("user", e.UserID)
}

// UserUpdatedEvent 用户更新事件
type UserUpdatedEvent struct {
	UserID    int
	Email     string
	Name      string
	occurredAt time.Time
}

func NewUserUpdatedEvent(userID int, email, name string) *UserUpdatedEvent {
	return &UserUpdatedEvent{
		UserID:    userID,
		Email:     email,
		Name:      name,
		occurredAt: time.Now(),
	}
}

func (e *UserUpdatedEvent) EventType() string {
	return "user.updated"
}

func (e *UserUpdatedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *UserUpdatedEvent) AggregateID() string {
	return AggregateID("user", e.UserID)
}

// UserStatusChangedEvent 用户状态变更事件
type UserStatusChangedEvent struct {
	UserID     int
	Email      string
	Name       string
	OldStatus  Status
	NewStatus  Status
	occurredAt time.Time
}

func NewUserStatusChangedEvent(userID int, email, name string, oldStatus, newStatus Status) *UserStatusChangedEvent {
	return &UserStatusChangedEvent{
		UserID:     userID,
		Email:      email,
		Name:       name,
		OldStatus:  oldStatus,
		NewStatus:  newStatus,
		occurredAt: time.Now(),
	}
}

func (e *UserStatusChangedEvent) EventType() string {
	return "user.status_changed"
}

func (e *UserStatusChangedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *UserStatusChangedEvent) AggregateID() string {
	return AggregateID("user", e.UserID)
}

// UserDeletedEvent 用户删除事件
type UserDeletedEvent struct {
	UserID     int
	Email      string
	Name       string
	occurredAt time.Time
}

func NewUserDeletedEvent(userID int, email, name string) *UserDeletedEvent {
	return &UserDeletedEvent{
		UserID:     userID,
		Email:      email,
		Name:       name,
		occurredAt: time.Now(),
	}
}

func (e *UserDeletedEvent) EventType() string {
	return "user.deleted"
}

func (e *UserDeletedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *UserDeletedEvent) AggregateID() string {
	return AggregateID("user", e.UserID)
}