package media

import (
	"context"
	"sort"
	"strings"
	"sync"
)

// MemoryStore 是 TaskStore + HoldStore 的内存实现，供单测与本地开发使用。
type MemoryStore struct {
	mu    sync.Mutex
	tasks map[string]*Task
	holds map[string]*Hold // keyed by task_id
	byHoldID map[string]*Hold
}

// NewMemoryStore 创建空内存存储。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		tasks:    make(map[string]*Task),
		holds:    make(map[string]*Hold),
		byHoldID: make(map[string]*Hold),
	}
}

func (s *MemoryStore) Create(ctx context.Context, task *Task) error {
	_ = ctx
	if task == nil || task.TaskID == "" {
		return ErrTaskNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[task.TaskID]; ok {
		return ErrDuplicateTaskID
	}
	cp := *task
	s.tasks[task.TaskID] = &cp
	return nil
}

func (s *MemoryStore) GetByTaskID(ctx context.Context, taskID string) (*Task, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[taskID]
	if !ok {
		return nil, ErrTaskNotFound
	}
	cp := *t
	return &cp, nil
}

func (s *MemoryStore) Update(ctx context.Context, task *Task) error {
	_ = ctx
	if task == nil || task.TaskID == "" {
		return ErrTaskNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[task.TaskID]; !ok {
		return ErrTaskNotFound
	}
	cp := *task
	s.tasks[task.TaskID] = &cp
	return nil
}

func (s *MemoryStore) ListByStatus(ctx context.Context, statuses []TaskStatus, limit int) ([]*Task, error) {
	_ = ctx
	want := make(map[TaskStatus]bool, len(statuses))
	for _, st := range statuses {
		want[st] = true
	}

	s.mu.Lock()
	matched := make([]*Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		if want[t.Status] {
			cp := *t
			matched = append(matched, &cp)
		}
	}
	s.mu.Unlock()

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.Before(matched[j].CreatedAt)
	})
	if limit > 0 && len(matched) > limit {
		matched = matched[:limit]
	}
	return matched, nil
}

func (s *MemoryStore) List(ctx context.Context, q TaskListQuery) (*TaskListResult, error) {
	_ = ctx
	s.mu.Lock()
	all := make([]*Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		if q.UserID != nil && t.UserID != *q.UserID {
			continue
		}
		if q.Status != "" && string(t.Status) != q.Status {
			continue
		}
		if q.MediaType != "" && t.MediaType != q.MediaType {
			continue
		}
		if q.Model != "" && t.Model != q.Model {
			continue
		}
		cp := *t
		all = append(all, &cp)
	}
	s.mu.Unlock()

	sortOrder := strings.ToLower(q.SortOrder)
	if sortOrder == "" {
		sortOrder = "desc"
	}
	sort.Slice(all, func(i, j int) bool {
		less := all[i].CreatedAt.Before(all[j].CreatedAt)
		if sortOrder == "asc" {
			return less
		}
		return !less
	})

	total := len(all)
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= total {
		return &TaskListResult{Tasks: []*Task{}, Total: total}, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return &TaskListResult{Tasks: all[start:end], Total: total}, nil
}

func (s *MemoryStore) CreateHold(ctx context.Context, hold *Hold) error {
	_ = ctx
	if hold == nil || hold.TaskID == "" {
		return ErrHoldNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.holds[hold.TaskID]; ok {
		return ErrDuplicateTaskID
	}
	cp := *hold
	s.holds[hold.TaskID] = &cp
	s.byHoldID[hold.HoldID] = &cp
	return nil
}

type memoryHoldStore struct{ s *MemoryStore }

func (m memoryHoldStore) Create(ctx context.Context, hold *Hold) error {
	return m.s.CreateHold(ctx, hold)
}

func (m memoryHoldStore) GetByTaskID(ctx context.Context, taskID string) (*Hold, error) {
	_ = ctx
	m.s.mu.Lock()
	defer m.s.mu.Unlock()
	h, ok := m.s.holds[taskID]
	if !ok {
		return nil, ErrHoldNotFound
	}
	cp := *h
	return &cp, nil
}

func (m memoryHoldStore) SumHeldByUser(ctx context.Context, userID int64) (float64, error) {
	_ = ctx
	m.s.mu.Lock()
	defer m.s.mu.Unlock()
	var sum float64
	for _, h := range m.s.holds {
		if h.UserID == userID && h.Status == HoldHeld {
			sum += h.Amount
		}
	}
	return sum, nil
}

func (m memoryHoldStore) UpdateStatus(ctx context.Context, holdID string, from, to HoldStatus) error {
	_ = ctx
	m.s.mu.Lock()
	defer m.s.mu.Unlock()
	h, ok := m.s.byHoldID[holdID]
	if !ok {
		return ErrHoldNotFound
	}
	if h.Status != from {
		return ErrInvalidHoldTransition
	}
	if !h.Status.CanTransitionTo(to) {
		return ErrInvalidHoldTransition
	}
	h.Status = to
	// sync maps
	if mapped, ok := m.s.holds[h.TaskID]; ok && mapped.HoldID == holdID {
		mapped.Status = to
	}
	return nil
}

// MemoryHoldStore 返回绑定到同一 MemoryStore 的 HoldStore。
func (s *MemoryStore) MemoryHoldStore() HoldStore {
	return memoryHoldStore{s: s}
}

type memoryReservation struct{ s *MemoryStore }

// Reserve 在持锁临界区内重算 held 总额、校验余额并原子建 task+hold，
// 模拟 ent 事务路径供单测使用。
func (m memoryReservation) Reserve(ctx context.Context, task *Task, hold *Hold, upstreamAvailable float64) error {
	_ = ctx
	if task == nil || hold == nil {
		return ErrTaskNotFound
	}
	m.s.mu.Lock()
	defer m.s.mu.Unlock()

	var held float64
	for _, h := range m.s.holds {
		if h.UserID == task.UserID && h.Status == HoldHeld {
			held += h.Amount
		}
	}
	if upstreamAvailable < held+hold.Amount {
		return ErrInsufficientBalance
	}
	if _, ok := m.s.tasks[task.TaskID]; ok {
		return ErrDuplicateTaskID
	}
	taskCopy := *task
	m.s.tasks[task.TaskID] = &taskCopy
	holdCopy := *hold
	m.s.holds[hold.TaskID] = &holdCopy
	m.s.byHoldID[hold.HoldID] = &holdCopy
	return nil
}

// MemoryReservation 返回绑定到同一 MemoryStore 的原子预扣实现。
func (s *MemoryStore) MemoryReservation() Reservation {
	return memoryReservation{s: s}
}
