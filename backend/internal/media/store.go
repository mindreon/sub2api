package media

import "context"

// TaskStore 持久化异步任务。
type TaskStore interface {
	Create(ctx context.Context, task *Task) error
	GetByTaskID(ctx context.Context, taskID string) (*Task, error)
	Update(ctx context.Context, task *Task) error
	// ListByStatus 按状态集合列出任务（按创建时间升序，先进先处理），供轮询 Worker 使用。
	ListByStatus(ctx context.Context, statuses []TaskStatus, limit int) ([]*Task, error)
	// List 分页列出任务（控制台/管理端）；实现见 EntTaskStore。
	List(ctx context.Context, q TaskListQuery) (*TaskListResult, error)
}

// HoldStore 持久化额度预扣台账。
type HoldStore interface {
	Create(ctx context.Context, hold *Hold) error
	GetByTaskID(ctx context.Context, taskID string) (*Hold, error)
	SumHeldByUser(ctx context.Context, userID int64) (float64, error)
	UpdateStatus(ctx context.Context, holdID string, from, to HoldStatus) error
}
