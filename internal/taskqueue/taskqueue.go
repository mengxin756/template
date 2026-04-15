// Package taskqueue defines the task queue interface for decoupling concrete implementations.
// It follows the Dependency Inversion Principle - high-level modules should not depend on
// low-level modules; both should depend on abstractions.
package taskqueue

import (
	"context"
	"time"
)

// Task represents a unit of work to be processed asynchronously.
type Task struct {
	// Type identifies the kind of task (e.g., "email:welcome", "cleanup:data")
	Type string
	// Payload contains the task data as bytes
	Payload []byte
}

// TaskResult contains information about a queued task.
type TaskResult struct {
	// ID is the unique identifier of the queued task
	ID string
	// Queue is the name of the queue the task was enqueued to
	Queue string
	// Type is the task type
	Type string
}

// Handler processes tasks of a specific type.
type Handler interface {
	// Process handles the task and returns an error if processing failed.
	Process(ctx context.Context, task *Task) error
}

// HandlerFunc is an adapter to allow using functions as handlers.
type HandlerFunc func(ctx context.Context, task *Task) error

// Process implements Handler interface.
func (h HandlerFunc) Process(ctx context.Context, task *Task) error {
	return h(ctx, task)
}

// TaskQueue defines the interface for asynchronous task processing.
// Implementations can use different backends (Asynq, Machinery, RabbitMQ, etc.).
type TaskQueue interface {
	// Enqueue adds a task to the queue for immediate processing.
	Enqueue(ctx context.Context, task *Task, opts ...Option) (*TaskResult, error)

	// EnqueueIn adds a task to the queue for processing after the specified delay.
	EnqueueIn(ctx context.Context, task *Task, delay time.Duration, opts ...Option) (*TaskResult, error)

	// EnqueueAt adds a task to the queue for processing at the specified time.
	EnqueueAt(ctx context.Context, task *Task, processAt time.Time, opts ...Option) (*TaskResult, error)

	// RegisterHandler registers a handler for a specific task type.
	RegisterHandler(taskType string, handler Handler) error

	// Start starts the task queue server to process tasks.
	Start(ctx context.Context) error

	// Stop gracefully stops the task queue server.
	Stop(ctx context.Context) error
}

// Option configures task enqueue options.
type Option func(*EnqueueOptions)

// EnqueueOptions contains options for enqueueing tasks.
type EnqueueOptions struct {
	// Queue specifies which queue to use (default: "default")
	Queue string
	// MaxRetry specifies the maximum number of retry attempts
	MaxRetry int
	// Timeout specifies the task processing timeout
	Timeout time.Duration
	// Priority specifies the task priority (higher = more important)
	Priority int
	// Unique ensures only one task of this type+payload exists
	Unique bool
	// TaskID specifies a custom task ID
	TaskID string
}

// WithQueue sets the queue name.
func WithQueue(queue string) Option {
	return func(o *EnqueueOptions) {
		o.Queue = queue
	}
}

// WithMaxRetry sets the maximum retry count.
func WithMaxRetry(maxRetry int) Option {
	return func(o *EnqueueOptions) {
		o.MaxRetry = maxRetry
	}
}

// WithTimeout sets the processing timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(o *EnqueueOptions) {
		o.Timeout = timeout
	}
}

// WithPriority sets the task priority.
func WithPriority(priority int) Option {
	return func(o *EnqueueOptions) {
		o.Priority = priority
	}
}

// WithUnique ensures task uniqueness.
func WithUnique() Option {
	return func(o *EnqueueOptions) {
		o.Unique = true
	}
}

// WithTaskID sets a custom task ID.
func WithTaskID(id string) Option {
	return func(o *EnqueueOptions) {
		o.TaskID = id
	}
}

// DefaultEnqueueOptions returns the default enqueue options.
func DefaultEnqueueOptions() *EnqueueOptions {
	return &EnqueueOptions{
		Queue:    "default",
		MaxRetry: 3,
		Timeout:  30 * time.Minute,
		Priority: 0,
		Unique:   false,
	}
}
