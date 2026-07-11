package media

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/mediagenerationtask"
	"github.com/Wei-Shaw/sub2api/ent/mediaquotahold"
)

// EntTaskStore 用 ent 持久化 Task。
type EntTaskStore struct {
	client *dbent.Client
}

func NewEntTaskStore(client *dbent.Client) *EntTaskStore {
	return &EntTaskStore{client: client}
}

func (s *EntTaskStore) Create(ctx context.Context, task *Task) error {
	if s == nil || s.client == nil || task == nil {
		return ErrUpstreamNotWired
	}
	builder, err := mediaTaskCreateBuilder(s.client.MediaGenerationTask.Create(), task)
	if err != nil {
		return err
	}
	row, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	task.ID = row.ID
	return nil
}

// mediaTaskCreateBuilder 把领域 Task 的字段应用到 ent create 构造器。
// 供 EntTaskStore.Create 与事务化的 EntReservation 复用，避免映射逻辑重复。
func mediaTaskCreateBuilder(builder *dbent.MediaGenerationTaskCreate, task *Task) (*dbent.MediaGenerationTaskCreate, error) {
	builder = builder.
		SetTaskID(task.TaskID).
		SetUserID(task.UserID).
		SetAPIKeyID(task.APIKeyID).
		SetModel(task.Model).
		SetMediaType(task.MediaType).
		SetStatus(string(task.Status)).
		SetReservedCost(task.ReservedCost).
		SetRateMultiplier(task.RateMultiplier).
		SetBillingCurrency(string(task.BillingCurrency)).
		SetExpiresAt(task.ExpiresAt)
	if task.AccountID != nil {
		builder.SetAccountID(*task.AccountID)
	}
	if task.GroupID != nil {
		builder.SetGroupID(*task.GroupID)
	}
	if task.SubscriptionID != nil {
		builder.SetSubscriptionID(*task.SubscriptionID)
	}
	if task.BillingMetric != "" {
		builder.SetBillingMetric(string(task.BillingMetric))
	}
	if task.UpstreamTaskID != "" {
		builder.SetUpstreamTaskID(task.UpstreamTaskID)
	}
	if task.ResultURL != "" {
		builder.SetResultURL(task.ResultURL)
	}
	if task.ResultStorageKey != "" {
		builder.SetResultStorageKey(task.ResultStorageKey)
	}
	if len(task.RequestParams) > 0 {
		raw, err := json.Marshal(task.RequestParams)
		if err != nil {
			return nil, err
		}
		builder.SetRequestParams(raw)
	}
	return builder, nil
}

// mediaHoldCreateBuilder 把领域 Hold 应用到 ent create 构造器。
func mediaHoldCreateBuilder(builder *dbent.MediaQuotaHoldCreate, hold *Hold) *dbent.MediaQuotaHoldCreate {
	return builder.
		SetHoldID(hold.HoldID).
		SetTaskID(hold.TaskID).
		SetUserID(hold.UserID).
		SetAmount(hold.Amount).
		SetCurrency(string(hold.Currency)).
		SetStatus(string(hold.Status))
}

func (s *EntTaskStore) GetByTaskID(ctx context.Context, taskID string) (*Task, error) {
	if s == nil || s.client == nil {
		return nil, ErrUpstreamNotWired
	}
	row, err := s.client.MediaGenerationTask.Query().
		Where(mediagenerationtask.TaskIDEQ(taskID)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return entTaskToDomain(row), nil
}

func (s *EntTaskStore) Update(ctx context.Context, task *Task) error {
	if s == nil || s.client == nil || task == nil {
		return ErrUpstreamNotWired
	}
	upd := s.client.MediaGenerationTask.Update().
		Where(mediagenerationtask.TaskIDEQ(task.TaskID)).
		SetStatus(string(task.Status)).
		SetPollAttempts(task.PollAttempts)
	if task.UpstreamTaskID != "" {
		upd.SetUpstreamTaskID(task.UpstreamTaskID)
	}
	if task.AccountID != nil {
		upd.SetAccountID(*task.AccountID)
	}
	if task.ActualCost != nil {
		upd.SetActualCost(*task.ActualCost)
	}
	if task.BillingMetric != "" {
		upd.SetBillingMetric(string(task.BillingMetric))
	}
	if task.ErrorMessage != "" {
		upd.SetErrorMessage(task.ErrorMessage)
	}
	if task.SettledAt != nil {
		upd.SetSettledAt(*task.SettledAt)
	}
	if task.ResultURL != "" {
		upd.SetResultURL(task.ResultURL)
	}
	if task.ResultStorageKey != "" {
		upd.SetResultStorageKey(task.ResultStorageKey)
	}
	if len(task.UpstreamUsage) > 0 {
		raw, err := json.Marshal(task.UpstreamUsage)
		if err != nil {
			return err
		}
		upd.SetUpstreamUsage(raw)
	}
	n, err := upd.Save(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (s *EntTaskStore) ListByStatus(ctx context.Context, statuses []TaskStatus, limit int) ([]*Task, error) {
	if s == nil || s.client == nil {
		return nil, ErrUpstreamNotWired
	}
	strStatuses := make([]string, 0, len(statuses))
	for _, st := range statuses {
		strStatuses = append(strStatuses, string(st))
	}
	query := s.client.MediaGenerationTask.Query().
		Where(mediagenerationtask.StatusIn(strStatuses...)).
		Order(dbent.Asc(mediagenerationtask.FieldCreatedAt))
	if limit > 0 {
		query = query.Limit(limit)
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, err
	}
	tasks := make([]*Task, 0, len(rows))
	for _, row := range rows {
		tasks = append(tasks, entTaskToDomain(row))
	}
	return tasks, nil
}

func (s *EntTaskStore) List(ctx context.Context, q TaskListQuery) (*TaskListResult, error) {
	if s == nil || s.client == nil {
		return nil, ErrUpstreamNotWired
	}
	query := s.client.MediaGenerationTask.Query()
	if q.UserID != nil {
		query = query.Where(mediagenerationtask.UserIDEQ(*q.UserID))
	}
	if st := strings.TrimSpace(q.Status); st != "" {
		query = query.Where(mediagenerationtask.StatusEQ(st))
	}
	if mt := strings.TrimSpace(q.MediaType); mt != "" {
		query = query.Where(mediagenerationtask.MediaTypeEQ(mt))
	}
	if model := strings.TrimSpace(q.Model); model != "" {
		query = query.Where(mediagenerationtask.ModelEQ(model))
	}
	if q.CreatedFrom != nil {
		query = query.Where(mediagenerationtask.CreatedAtGTE(*q.CreatedFrom))
	}
	if q.CreatedTo != nil {
		query = query.Where(mediagenerationtask.CreatedAtLTE(*q.CreatedTo))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	sortOrder := strings.ToLower(q.SortOrder)
	if sortOrder == "" {
		sortOrder = "desc"
	}
	if sortOrder == "asc" {
		query = query.Order(dbent.Asc(mediagenerationtask.FieldCreatedAt))
	} else {
		query = query.Order(dbent.Desc(mediagenerationtask.FieldCreatedAt))
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	rows, err := query.
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, err
	}
	tasks := make([]*Task, 0, len(rows))
	for _, row := range rows {
		tasks = append(tasks, entTaskToDomain(row))
	}
	return &TaskListResult{Tasks: tasks, Total: total}, nil
}

func entTaskToDomain(row *dbent.MediaGenerationTask) *Task {
	t := &Task{
		ID:              row.ID,
		TaskID:          row.TaskID,
		UserID:          row.UserID,
		APIKeyID:        row.APIKeyID,
		Model:           row.Model,
		MediaType:       row.MediaType,
		Status:          TaskStatus(row.Status),
		ReservedCost:    row.ReservedCost,
		RateMultiplier:  row.RateMultiplier,
		BillingCurrency: Currency(row.BillingCurrency),
		PollAttempts:    row.PollAttempts,
		ExpiresAt:       row.ExpiresAt,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.AccountID != nil {
		t.AccountID = row.AccountID
	}
	if row.GroupID != nil {
		t.GroupID = row.GroupID
	}
	if row.SubscriptionID != nil {
		t.SubscriptionID = row.SubscriptionID
	}
	if row.UpstreamTaskID != nil {
		t.UpstreamTaskID = *row.UpstreamTaskID
	}
	if row.ResultURL != nil {
		t.ResultURL = *row.ResultURL
	}
	if row.ResultStorageKey != nil {
		t.ResultStorageKey = *row.ResultStorageKey
	}
	if row.BillingMetric != nil {
		t.BillingMetric = BillingMetric(*row.BillingMetric)
	}
	if row.ActualCost != nil {
		v := *row.ActualCost
		t.ActualCost = &v
	}
	if row.SettledAt != nil {
		t.SettledAt = row.SettledAt
	}
	if row.ErrorMessage != nil {
		t.ErrorMessage = *row.ErrorMessage
	}
	if len(row.RequestParams) > 0 {
		_ = json.Unmarshal(row.RequestParams, &t.RequestParams)
	}
	if len(row.UpstreamUsage) > 0 {
		_ = json.Unmarshal(row.UpstreamUsage, &t.UpstreamUsage)
	}
	return t
}

// EntHoldStore 用 ent 持久化 Hold。
type EntHoldStore struct {
	client *dbent.Client
}

func NewEntHoldStore(client *dbent.Client) *EntHoldStore {
	return &EntHoldStore{client: client}
}

func (s *EntHoldStore) Create(ctx context.Context, hold *Hold) error {
	if s == nil || s.client == nil || hold == nil {
		return ErrUpstreamNotWired
	}
	row, err := mediaHoldCreateBuilder(s.client.MediaQuotaHold.Create(), hold).Save(ctx)
	if err != nil {
		return err
	}
	hold.ID = row.ID
	return nil
}

func (s *EntHoldStore) GetByTaskID(ctx context.Context, taskID string) (*Hold, error) {
	if s == nil || s.client == nil {
		return nil, ErrUpstreamNotWired
	}
	row, err := s.client.MediaQuotaHold.Query().
		Where(mediaquotahold.TaskIDEQ(taskID)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrHoldNotFound
		}
		return nil, err
	}
	return entHoldToDomain(row), nil
}

func (s *EntHoldStore) SumHeldByUser(ctx context.Context, userID int64) (float64, error) {
	if s == nil || s.client == nil {
		return 0, ErrUpstreamNotWired
	}
	rows, err := s.client.MediaQuotaHold.Query().
		Where(
			mediaquotahold.UserIDEQ(userID),
			mediaquotahold.StatusEQ(string(HoldHeld)),
		).
		All(ctx)
	if err != nil {
		return 0, err
	}
	var sum float64
	for _, row := range rows {
		sum += row.Amount
	}
	return sum, nil
}

func (s *EntHoldStore) UpdateStatus(ctx context.Context, holdID string, from, to HoldStatus) error {
	if s == nil || s.client == nil {
		return ErrUpstreamNotWired
	}
	if !from.CanTransitionTo(to) {
		return ErrInvalidHoldTransition
	}
	n, err := s.client.MediaQuotaHold.Update().
		Where(
			mediaquotahold.HoldIDEQ(holdID),
			mediaquotahold.StatusEQ(string(from)),
		).
		SetStatus(string(to)).
		Save(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("%w: hold_id=%s from=%s", ErrInvalidHoldTransition, holdID, from)
	}
	return nil
}

func entHoldToDomain(row *dbent.MediaQuotaHold) *Hold {
	return &Hold{
		ID:        row.ID,
		HoldID:    row.HoldID,
		TaskID:    row.TaskID,
		UserID:    row.UserID,
		Amount:    row.Amount,
		Currency:  Currency(row.Currency),
		Status:    HoldStatus(row.Status),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
