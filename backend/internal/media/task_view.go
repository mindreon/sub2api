package media

import "time"

// TaskView 是任务 API 响应用 DTO。
type TaskView struct {
	TaskID          string     `json:"task_id"`
	UpstreamTaskID  string     `json:"upstream_task_id,omitempty"`
	UserID          int64      `json:"user_id,omitempty"`
	Model           string     `json:"model"`
	MediaType       string     `json:"media_type"`
	Status          string     `json:"status"`
	BillingMetric   string     `json:"billing_metric,omitempty"`
	ReservedCost    float64    `json:"reserved_cost"`
	ActualCost      *float64   `json:"actual_cost,omitempty"`
	BillingCurrency string     `json:"billing_currency"`
	ResultURL       string     `json:"result_url,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	ExpiresAt       time.Time  `json:"expires_at"`
	SettledAt       *time.Time `json:"settled_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ToTaskView 把领域任务转为 API 视图。
func ToTaskView(t *Task) TaskView {
	if t == nil {
		return TaskView{}
	}
	return TaskView{
		TaskID:          t.TaskID,
		UpstreamTaskID:  t.UpstreamTaskID,
		UserID:          t.UserID,
		Model:           t.Model,
		MediaType:       t.MediaType,
		Status:          string(t.Status),
		BillingMetric:   string(t.BillingMetric),
		ReservedCost:    t.ReservedCost,
		ActualCost:      t.ActualCost,
		BillingCurrency: string(t.BillingCurrency),
		ResultURL:       t.ResultURL,
		ErrorMessage:    t.ErrorMessage,
		ExpiresAt:       t.ExpiresAt,
		SettledAt:       t.SettledAt,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}
