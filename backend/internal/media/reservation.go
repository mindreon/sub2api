package media

import "context"

// Reservation 在单个事务内原子地创建 task + hold，并在用户维度校验余额上限，
// 消除"检查余额→建任务→建预扣"三步非原子导致的并发超发。
//
// upstreamAvailable 为**未扣除本包预扣**的上游可用余额；held 总额由实现自身
// 在事务内（尽力加行锁）重算，避免与 HoldAwareBalance 双重扣减。
// 当 upstreamAvailable < 现有 held 总额 + hold.Amount 时返回 ErrInsufficientBalance。
type Reservation interface {
	Reserve(ctx context.Context, task *Task, hold *Hold, upstreamAvailable float64) error
}
