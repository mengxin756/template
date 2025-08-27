package asynq

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

// 任务类型常量
const (
	TaskTypeWelcomeEmail             = "welcome_email"
	TaskTypeStatusChangeNotification = "status_change_notification"
	TaskTypeDataCleanup              = "data_cleanup"
)

// WelcomeEmailPayload 欢迎邮件任务载荷
type WelcomeEmailPayload struct {
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	UserName  string `json:"user_name"`
	Timestamp int64  `json:"timestamp"`
}

// StatusChangeNotificationPayload 状态变更通知任务载荷
type StatusChangeNotificationPayload struct {
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	UserName  string `json:"user_name"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
	ChangedBy string `json:"changed_by"`
	Timestamp int64  `json:"timestamp"`
}

// DataCleanupPayload 数据清理任务载荷
type DataCleanupPayload struct {
	CleanupType string `json:"cleanup_type"` // logs, temp_files, old_records
	Retention   int    `json:"retention"`    // 保留天数
	Timestamp   int64  `json:"timestamp"`
}

// NewWelcomeEmailTask 创建欢迎邮件任务
func NewWelcomeEmailTask(userID int, email, userName string) *asynq.Task {
	payload := WelcomeEmailPayload{
		UserID:    userID,
		Email:     email,
		UserName:  userName,
		Timestamp: time.Now().Unix(),
	}

	data, _ := json.Marshal(payload)
	return asynq.NewTask(TaskTypeWelcomeEmail, data)
}

// NewStatusChangeNotificationTask 创建状态变更通知任务
func NewStatusChangeNotificationTask(userID int, email, userName, oldStatus, newStatus, changedBy string) *asynq.Task {
	payload := StatusChangeNotificationPayload{
		UserID:    userID,
		Email:     email,
		UserName:  userName,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		ChangedBy: changedBy,
		Timestamp: time.Now().Unix(),
	}

	data, _ := json.Marshal(payload)
	return asynq.NewTask(TaskTypeStatusChangeNotification, data)
}

// NewDataCleanupTask 创建数据清理任务
func NewDataCleanupTask(cleanupType string, retention int) *asynq.Task {
	payload := DataCleanupPayload{
		CleanupType: cleanupType,
		Retention:   retention,
		Timestamp:   time.Now().Unix(),
	}

	data, _ := json.Marshal(payload)
	return asynq.NewTask(TaskTypeDataCleanup, data)
}

// NewDelayedWelcomeEmailTask 创建延迟的欢迎邮件任务
func NewDelayedWelcomeEmailTask(userID int, email, userName string, delay time.Duration) *asynq.Task {
	task := NewWelcomeEmailTask(userID, email, userName)
	return task
}

// NewRetryableTask 创建可重试的任务
func NewRetryableTask(taskType string, payload interface{}, maxRetries int) *asynq.Task {
	data, _ := json.Marshal(payload)
	task := asynq.NewTask(taskType, data)

	// 注意：Asynq 的任务选项需要在入队时设置，不是在任务创建时
	return task
}

// NewPriorityTask 创建优先级任务
func NewPriorityTask(taskType string, payload interface{}, priority int) *asynq.Task {
	data, _ := json.Marshal(payload)
	task := asynq.NewTask(taskType, data)

	// 注意：Asynq 的任务选项需要在入队时设置，不是在任务创建时
	return task
}
