package media

import (
	"errors"
	"time"
)

// TaskStatus 异步任务状态。
type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskInProgress TaskStatus = "in_progress"
	TaskCompleted  TaskStatus = "completed"
	TaskFailed     TaskStatus = "failed"
	TaskExpired    TaskStatus = "expired"
	TaskCancelled  TaskStatus = "cancelled"
)

// HoldStatus 预扣台账状态。
type HoldStatus string

const (
	HoldHeld     HoldStatus = "held"
	HoldSettled  HoldStatus = "settled"
	HoldReleased HoldStatus = "released"
)

var (
	ErrTaskNotFound          = errors.New("media: task not found")
	ErrHoldNotFound          = errors.New("media: hold not found")
	ErrInvalidTaskTransition = errors.New("media: invalid task status transition")
	ErrInvalidHoldTransition = errors.New("media: invalid hold status transition")
	ErrTaskAlreadyTerminal   = errors.New("media: task already in terminal state")
	ErrDuplicateTaskID       = errors.New("media: duplicate task id")
)

// Task 是多模态异步生成任务领域对象。
type Task struct {
	ID               int64
	TaskID           string
	UpstreamTaskID   string
	UserID           int64
	APIKeyID         int64
	AccountID        *int64
	GroupID          *int64
	SubscriptionID   *int64
	Model            string
	MediaType        string
	Status           TaskStatus
	BillingMetric    BillingMetric
	ReservedCost     float64
	ActualCost       *float64
	RateMultiplier   float64
	BillingCurrency  Currency
	RequestParams    map[string]any
	UpstreamUsage    map[string]any
	ResultURL        string // 对客户端展示的可用视频链接（转存后为自有存储链接，否则为上游直链）
	ResultStorageKey string // 自有对象存储的对象键（转存成功时写入，用于按需重新签发链接）
	PollAttempts     int
	ExpiresAt        time.Time
	SettledAt        *time.Time
	ErrorMessage     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Hold 是额度预扣记录。
type Hold struct {
	ID        int64
	HoldID    string
	TaskID    string
	UserID    int64
	Amount    float64
	Currency  Currency
	Status    HoldStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s TaskStatus) IsTerminal() bool {
	switch s {
	case TaskCompleted, TaskFailed, TaskExpired, TaskCancelled:
		return true
	}
	return false
}

// CanTransitionTo 检查任务状态是否允许迁移到 to。
func (s TaskStatus) CanTransitionTo(to TaskStatus) bool {
	switch s {
	case TaskPending:
		return to == TaskInProgress || to == TaskFailed || to == TaskExpired || to == TaskCancelled
	case TaskInProgress:
		return to == TaskCompleted || to == TaskFailed || to == TaskExpired || to == TaskCancelled
	default:
		return false
	}
}

func (s HoldStatus) CanTransitionTo(to HoldStatus) bool {
	switch s {
	case HoldHeld:
		return to == HoldSettled || to == HoldReleased
	default:
		return false
	}
}
