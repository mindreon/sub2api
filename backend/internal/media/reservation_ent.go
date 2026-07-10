package media

import (
	"context"
	"fmt"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/mediaquotahold"
)

// EntReservation 用 ent 事务实现严格预扣：单事务内锁定重算用户 held 总额、
// 校验上游余额，通过则原子创建 task+hold，否则回滚。
type EntReservation struct {
	client *dbent.Client
}

// NewEntReservation 构造事务化预扣器。
func NewEntReservation(client *dbent.Client) *EntReservation {
	return &EntReservation{client: client}
}

func (r *EntReservation) Reserve(ctx context.Context, task *Task, hold *Hold, upstreamAvailable float64) error {
	if r == nil || r.client == nil || task == nil || hold == nil {
		return ErrUpstreamNotWired
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("media reservation: begin tx: %w", err)
	}
	// 出错统一回滚；成功路径显式 Commit 后 committed=true 跳过回滚。
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	held, err := sumHeldTx(ctx, tx, task.UserID)
	if err != nil {
		return err
	}
	if upstreamAvailable < held+hold.Amount {
		return ErrInsufficientBalance
	}

	taskBuilder, err := mediaTaskCreateBuilder(tx.MediaGenerationTask.Create(), task)
	if err != nil {
		return err
	}
	taskRow, err := taskBuilder.Save(ctx)
	if err != nil {
		return fmt.Errorf("media reservation: create task: %w", err)
	}
	holdRow, err := mediaHoldCreateBuilder(tx.MediaQuotaHold.Create(), hold).Save(ctx)
	if err != nil {
		return fmt.Errorf("media reservation: create hold: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("media reservation: commit: %w", err)
	}
	committed = true
	task.ID = taskRow.ID
	hold.ID = holdRow.ID
	return nil
}

// sumHeldTx 在事务内求某用户 held 预扣总额，尽力加行锁（ForUpdate）。
// 不支持行锁的方言（如 SQLite）退回无锁查询，靠单写事务保证一致。
func sumHeldTx(ctx context.Context, tx *dbent.Tx, userID int64) (float64, error) {
	query := tx.MediaQuotaHold.Query().Where(
		mediaquotahold.UserIDEQ(userID),
		mediaquotahold.StatusEQ(string(HoldHeld)),
	)
	rows, err := query.Clone().ForUpdate().All(ctx)
	if err != nil {
		rows, err = query.All(ctx)
		if err != nil {
			return 0, fmt.Errorf("media reservation: sum held: %w", err)
		}
	}
	var sum float64
	for _, row := range rows {
		sum += row.Amount
	}
	return sum, nil
}
