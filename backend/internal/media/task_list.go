package media

import "time"

// TaskListQuery 是任务列表查询参数。
type TaskListQuery struct {
	UserID      *int64
	Status      string
	MediaType   string
	Model       string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Page        int
	PageSize    int
	SortBy      string
	SortOrder   string
}

// TaskListResult 是分页任务列表结果。
type TaskListResult struct {
	Tasks []*Task
	Total int
}
