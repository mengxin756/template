package domain

// EventPublisher 事件发布接口（领域层定义）
type EventPublisher interface {
	Publish(event DomainEvent) error
	PublishBatch(events []DomainEvent) error
}

// EventProcessor 事件处理器接口（用于解耦）
type EventProcessor interface {
	Process(event DomainEvent) error
	EventType() string // 处理器关心的事件类型
}