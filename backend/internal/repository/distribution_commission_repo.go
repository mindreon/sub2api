package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type distributionCommissionRepository struct {
	db *sql.DB
}

func NewDistributionCommissionRepository(_ *dbent.Client, db *sql.DB) *distributionCommissionRepository {
	return &distributionCommissionRepository{db: db}
}

func (r *distributionCommissionRepository) Create(ctx context.Context, input service.DistributionCommissionInput) (*service.DistributionCommissionLedger, error) {
	if input.ChannelOrgID <= 0 || input.MemberID <= 0 || input.UserID <= 0 || input.BaseAmount <= 0 || input.Rate <= 0 || input.Amount <= 0 {
		return nil, service.ErrInvalidDistributionCommission
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionCommission
	}

	rows, err := r.db.QueryContext(ctx, `
INSERT INTO commission_ledger (
    channel_org_id,
    member_id,
    user_id,
    usage_log_id,
    commission_type,
    base_amount,
    rate,
    amount,
    status,
    settlement_method,
    settlement_reference_no,
    settlement_note,
    frozen_until,
    settled_at,
    settled_by_user_id,
    reversed_from_id,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
RETURNING id,
          channel_org_id,
          member_id,
          user_id,
          usage_log_id,
          commission_type,
          base_amount,
          rate,
          amount,
          status,
          settlement_method,
          settlement_reference_no,
          settlement_note,
          frozen_until,
          settled_at,
          settled_by_user_id,
          reversed_from_id,
          created_at,
          updated_at`,
		input.ChannelOrgID,
		input.MemberID,
		input.UserID,
		nullableInt64Arg(input.UsageLogID),
		input.CommissionType,
		input.BaseAmount,
		input.Rate,
		input.Amount,
		input.Status,
		input.SettlementMethod,
		input.SettlementReferenceNo,
		input.SettlementNote,
		nullableTimeArg(input.FrozenUntil),
		nullableTimeArg(input.SettledAt),
		nullableInt64Arg(input.SettledByUserID),
		nullableInt64Arg(input.ReversedFromID),
	)
	if err != nil {
		return nil, fmt.Errorf("create distribution commission: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanDistributionCommissionLedger(rows)
}

func (r *distributionCommissionRepository) ListByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams) ([]service.DistributionCommissionLedgerView, *pagination.PaginationResult, error) {
	if channelOrgID <= 0 {
		return nil, nil, service.ErrInvalidDistributionAttribution
	}
	if r.db == nil {
		return nil, nil, service.ErrInvalidDistributionAttribution
	}
	client := r.db
	if err := thawDistributionCommissionsByChannelOrgID(ctx, client, channelOrgID); err != nil {
		return nil, nil, err
	}

	var total int64
	if err := client.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM commission_ledger
WHERE channel_org_id = $1`, channelOrgID).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count distribution commissions: %w", err)
	}

	rows, err := client.QueryContext(ctx, `
SELECT cl.id,
       cl.channel_org_id,
       cl.member_id,
       cl.user_id,
       u.email,
       COALESCE(u.username, ''),
       cl.usage_log_id,
       cl.commission_type,
       cl.base_amount,
       cl.rate,
       cl.amount,
       cl.status,
       cl.settlement_method,
       cl.settlement_reference_no,
       cl.settlement_note,
       cl.frozen_until,
       cl.settled_at,
       cl.settled_by_user_id,
       cl.reversed_from_id,
       cl.created_at,
       cl.updated_at
FROM commission_ledger cl
JOIN users u ON u.id = cl.user_id
WHERE cl.channel_org_id = $1
ORDER BY cl.created_at DESC, cl.id DESC
LIMIT $2 OFFSET $3`,
		channelOrgID,
		params.Limit(),
		params.Offset(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list distribution commissions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionCommissionLedgerViews(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *distributionCommissionRepository) ListAdmin(ctx context.Context, filter service.DistributionAdminListFilter, params pagination.PaginationParams) ([]service.DistributionCommissionLedgerView, *pagination.PaginationResult, error) {
	if r.db == nil {
		return nil, nil, service.ErrInvalidDistributionCommission
	}
	client := r.db
	if err := thawDistributionCommissionsByChannelOrgID(ctx, client, filter.ChannelOrgID); err != nil {
		return nil, nil, err
	}

	whereSQL := "WHERE 1=1"
	args := make([]any, 0, 4)
	if filter.ChannelOrgID > 0 {
		args = append(args, filter.ChannelOrgID)
		whereSQL += fmt.Sprintf(" AND cl.channel_org_id = $%d", len(args))
	}
	if filter.UserID > 0 {
		args = append(args, filter.UserID)
		whereSQL += fmt.Sprintf(" AND cl.user_id = $%d", len(args))
	}

	var total int64
	countQuery := `
SELECT COUNT(*)
FROM commission_ledger cl
` + whereSQL
	if err := client.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count admin distribution commissions: %w", err)
	}

	args = append(args, params.Limit(), params.Offset())
	rows, err := client.QueryContext(ctx, `
SELECT cl.id,
       cl.channel_org_id,
       cl.member_id,
       cl.user_id,
       u.email,
       COALESCE(u.username, ''),
       cl.usage_log_id,
       cl.commission_type,
       cl.base_amount,
       cl.rate,
       cl.amount,
       cl.status,
       cl.settlement_method,
       cl.settlement_reference_no,
       cl.settlement_note,
       cl.frozen_until,
       cl.settled_at,
       cl.settled_by_user_id,
       cl.reversed_from_id,
       cl.created_at,
       cl.updated_at
FROM commission_ledger cl
JOIN users u ON u.id = cl.user_id
`+whereSQL+`
ORDER BY cl.created_at DESC, cl.id DESC
LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args)),
		args...,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list admin distribution commissions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionCommissionLedgerViews(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *distributionCommissionRepository) GetByID(ctx context.Context, commissionID int64) (*service.DistributionCommissionLedger, error) {
	if commissionID <= 0 {
		return nil, service.ErrInvalidDistributionCommission
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionCommission
	}
	if err := thawDistributionCommissionByID(ctx, r.db, commissionID); err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT id,
       channel_org_id,
       member_id,
       user_id,
       usage_log_id,
       commission_type,
       base_amount,
       rate,
       amount,
       status,
       settlement_method,
       settlement_reference_no,
       settlement_note,
       frozen_until,
       settled_at,
       settled_by_user_id,
       reversed_from_id,
       created_at,
       updated_at
FROM commission_ledger
WHERE id = $1`, commissionID)
	if err != nil {
		return nil, fmt.Errorf("get distribution commission: %w", err)
	}
	defer func() { _ = rows.Close() }()
	ledger, err := scanDistributionCommissionLedger(rows)
	if err != nil {
		return nil, err
	}
	return ledger, nil
}

func (r *distributionCommissionRepository) ListAutoSettleCommissionIDs(ctx context.Context, limit int) ([]int64, error) {
	if r.db == nil {
		return nil, service.ErrInvalidDistributionCommission
	}
	if limit <= 0 {
		limit = 100
	}
	if err := thawDistributionCommissionsByChannelOrgID(ctx, r.db, 0); err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT id
FROM commission_ledger
WHERE status = 'available'
  AND settlement_method = 'auto'
ORDER BY id ASC
LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list auto-settle distribution commissions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	ids := make([]int64, 0, limit)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan auto-settle distribution commission id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate auto-settle distribution commission ids: %w", err)
	}
	return ids, nil
}

func thawDistributionCommissionsByChannelOrgID(ctx context.Context, execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}, channelOrgID int64) error {
	query := `
UPDATE commission_ledger
SET status = 'available',
    frozen_until = NULL,
    updated_at = NOW()
WHERE status = 'frozen'
  AND frozen_until IS NOT NULL
  AND frozen_until <= NOW()`
	args := make([]any, 0, 1)
	if channelOrgID > 0 {
		query += `
  AND channel_org_id = $1`
		args = append(args, channelOrgID)
	}
	if _, err := execer.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("thaw matured distribution commissions: %w", err)
	}
	return nil
}

func thawDistributionCommissionByID(ctx context.Context, execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}, commissionID int64) error {
	if commissionID <= 0 {
		return service.ErrInvalidDistributionCommission
	}
	if _, err := execer.ExecContext(ctx, `
UPDATE commission_ledger
SET status = 'available',
    frozen_until = NULL,
    updated_at = NOW()
WHERE id = $1
  AND status = 'frozen'
  AND frozen_until IS NOT NULL
  AND frozen_until <= NOW()`, commissionID); err != nil {
		return fmt.Errorf("thaw matured distribution commission: %w", err)
	}
	return nil
}

func (r *distributionCommissionRepository) HasCommissionTypeSince(ctx context.Context, channelOrgID int64, commissionType string, since time.Time) (bool, error) {
	if channelOrgID <= 0 || strings.TrimSpace(commissionType) == "" {
		return false, service.ErrInvalidDistributionCommission
	}
	if r.db == nil {
		return false, service.ErrInvalidDistributionCommission
	}

	var exists bool
	if err := r.db.QueryRowContext(ctx, `
SELECT EXISTS(
	SELECT 1
	FROM commission_ledger
	WHERE channel_org_id = $1
	  AND commission_type = $2
	  AND created_at >= $3
)`, channelOrgID, strings.TrimSpace(commissionType), since).Scan(&exists); err != nil {
		return false, fmt.Errorf("check distribution commission type since: %w", err)
	}
	return exists, nil
}

func (r *distributionCommissionRepository) GetTotalCommissionByUserID(ctx context.Context, channelOrgID int64, userID int64) (float64, error) {
	if channelOrgID <= 0 || userID <= 0 {
		return 0, service.ErrInvalidDistributionCommission
	}
	if r.db == nil {
		return 0, service.ErrInvalidDistributionCommission
	}

	var total float64
	if err := r.db.QueryRowContext(ctx, `
SELECT COALESCE(SUM(amount), 0)::double precision
FROM commission_ledger
WHERE channel_org_id = $1
  AND user_id = $2
  AND status <> 'cancelled'`, channelOrgID, userID).Scan(&total); err != nil {
		return 0, fmt.Errorf("get total distribution commission by user: %w", err)
	}
	return total, nil
}

func (r *distributionCommissionRepository) GetTotalCommissionByMemberID(ctx context.Context, channelOrgID int64, memberID int64) (float64, error) {
	if channelOrgID <= 0 || memberID <= 0 {
		return 0, service.ErrInvalidDistributionCommission
	}
	if r.db == nil {
		return 0, service.ErrInvalidDistributionCommission
	}

	var total float64
	if err := r.db.QueryRowContext(ctx, `
SELECT COALESCE(SUM(amount), 0)::double precision
FROM commission_ledger
WHERE channel_org_id = $1
  AND member_id = $2
  AND status <> 'cancelled'`, channelOrgID, memberID).Scan(&total); err != nil {
		return 0, fmt.Errorf("get total distribution commission by member: %w", err)
	}
	return total, nil
}

func (r *distributionCommissionRepository) Settle(ctx context.Context, commissionID int64, input service.DistributionCommissionSettlementInput) (*service.DistributionCommissionLedger, error) {
	if commissionID <= 0 {
		return nil, service.ErrInvalidDistributionCommission
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionCommission
	}
	settlementMethod := strings.TrimSpace(input.SettlementMethod)
	if settlementMethod == "" {
		settlementMethod = "manual"
	}

	rows, err := r.db.QueryContext(ctx, `
UPDATE commission_ledger
SET status = 'settled',
    settlement_method = $2,
    settlement_reference_no = $3,
    settlement_note = $4,
    settled_by_user_id = $5,
    settled_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING id,
          channel_org_id,
          member_id,
          user_id,
          usage_log_id,
          commission_type,
          base_amount,
          rate,
          amount,
          status,
          settlement_method,
          settlement_reference_no,
          settlement_note,
          frozen_until,
          settled_at,
          settled_by_user_id,
          reversed_from_id,
          created_at,
          updated_at`,
		commissionID,
		settlementMethod,
		strings.TrimSpace(input.SettlementReferenceNo),
		strings.TrimSpace(input.SettlementNote),
		nullableInt64Arg(input.SettledByUserID),
	)
	if err != nil {
		return nil, fmt.Errorf("settle distribution commission: %w", err)
	}
	defer func() { _ = rows.Close() }()
	ledger, err := scanDistributionCommissionLedger(rows)
	if err != nil {
		return nil, err
	}
	return ledger, nil
}

func (r *distributionCommissionRepository) SettleToBalance(ctx context.Context, commissionID int64, input service.DistributionCommissionSettlementInput) (*service.DistributionCommissionLedger, error) {
	if commissionID <= 0 {
		return nil, service.ErrInvalidDistributionCommission
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionCommission
	}

	settlementMethod := strings.TrimSpace(input.SettlementMethod)
	if settlementMethod == "" {
		settlementMethod = "balance"
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin settle distribution commission to balance tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := thawDistributionCommissionByID(ctx, tx, commissionID); err != nil {
		return nil, err
	}

	original, err := getDistributionCommissionByIDForUpdate(ctx, tx, commissionID)
	if err != nil {
		return nil, err
	}
	if original == nil {
		return nil, service.ErrInvalidDistributionCommission
	}
	if strings.EqualFold(strings.TrimSpace(original.Status), "settled") {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit settled distribution commission to balance tx: %w", err)
		}
		return original, nil
	}

	if _, err := tx.ExecContext(ctx, `UPDATE users SET balance = balance + $2 WHERE id = $1`, original.UserID, original.Amount); err != nil {
		return nil, fmt.Errorf("credit distribution commission to user balance: %w", err)
	}
	if err := insertDistributionBalanceRedeemRecord(ctx, tx, original.UserID, original.Amount, commissionID); err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
UPDATE commission_ledger
SET status = 'settled',
    settlement_method = $2,
    settlement_reference_no = $3,
    settlement_note = $4,
    settled_by_user_id = $5,
    settled_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING id,
          channel_org_id,
          member_id,
          user_id,
          usage_log_id,
          commission_type,
          base_amount,
          rate,
          amount,
          status,
          settlement_method,
          settlement_reference_no,
          settlement_note,
          frozen_until,
          settled_at,
          settled_by_user_id,
          reversed_from_id,
          created_at,
          updated_at`,
		commissionID,
		settlementMethod,
		strings.TrimSpace(input.SettlementReferenceNo),
		strings.TrimSpace(input.SettlementNote),
		nullableInt64Arg(input.SettledByUserID),
	)
	if err != nil {
		return nil, fmt.Errorf("settle distribution commission to balance: %w", err)
	}
	defer func() { _ = rows.Close() }()

	ledger, err := scanDistributionCommissionLedger(rows)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit settle distribution commission to balance tx: %w", err)
	}
	return ledger, nil
}

func (r *distributionCommissionRepository) Reverse(ctx context.Context, commissionID int64) (*service.DistributionCommissionLedger, error) {
	original, err := r.GetByID(ctx, commissionID)
	if err != nil {
		return nil, err
	}
	if original == nil {
		return nil, service.ErrInvalidDistributionCommission
	}

	if strings.EqualFold(strings.TrimSpace(original.Status), "settled") {
		rows, err := r.db.QueryContext(ctx, `
INSERT INTO commission_ledger (
    channel_org_id,
    member_id,
    user_id,
    usage_log_id,
    commission_type,
    base_amount,
    rate,
    amount,
    status,
    settlement_method,
    settlement_reference_no,
    settlement_note,
    frozen_until,
    settled_at,
    settled_by_user_id,
    reversed_from_id,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'reversed', $9, $10, $11, $12, NOW(), $13, $14, NOW(), NOW())
RETURNING id,
          channel_org_id,
          member_id,
          user_id,
          usage_log_id,
          commission_type,
          base_amount,
          rate,
          amount,
          status,
          settlement_method,
          settlement_reference_no,
          settlement_note,
          frozen_until,
          settled_at,
          settled_by_user_id,
          reversed_from_id,
          created_at,
          updated_at`,
			original.ChannelOrgID,
			original.MemberID,
			original.UserID,
			nullableInt64Arg(original.UsageLogID),
			original.CommissionType,
			original.BaseAmount,
			original.Rate,
			-original.Amount,
			original.SettlementMethod,
			original.SettlementReferenceNo,
			original.SettlementNote,
			nullableTimeArg(original.FrozenUntil),
			nullableInt64Arg(original.SettledByUserID),
			nullableInt64Arg(&original.ID),
		)
		if err != nil {
			return nil, fmt.Errorf("reverse distribution commission: %w", err)
		}
		defer func() { _ = rows.Close() }()
		return scanDistributionCommissionLedger(rows)
	}

	rows, err := r.db.QueryContext(ctx, `
UPDATE commission_ledger
SET status = 'cancelled',
    updated_at = NOW()
WHERE id = $1
RETURNING id,
          channel_org_id,
          member_id,
          user_id,
          usage_log_id,
          commission_type,
          base_amount,
          rate,
          amount,
          status,
          settlement_method,
          settlement_reference_no,
          settlement_note,
          frozen_until,
          settled_at,
          settled_by_user_id,
          reversed_from_id,
          created_at,
          updated_at`,
		commissionID,
	)
	if err != nil {
		return nil, fmt.Errorf("cancel distribution commission: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanDistributionCommissionLedger(rows)
}

func (r *distributionCommissionRepository) ReverseBalanceSettlement(ctx context.Context, commissionID int64) (*service.DistributionCommissionLedger, error) {
	if commissionID <= 0 {
		return nil, service.ErrInvalidDistributionCommission
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionCommission
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin reverse distribution balance settlement tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	original, err := getDistributionCommissionByIDForUpdate(ctx, tx, commissionID)
	if err != nil {
		return nil, err
	}
	if original == nil {
		return nil, service.ErrInvalidDistributionCommission
	}

	if _, err := tx.ExecContext(ctx, `UPDATE users SET balance = balance - $2 WHERE id = $1`, original.UserID, original.Amount); err != nil {
		return nil, fmt.Errorf("deduct reversed distribution commission from user balance: %w", err)
	}
	if err := insertDistributionBalanceRedeemRecord(ctx, tx, original.UserID, -original.Amount, commissionID); err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
INSERT INTO commission_ledger (
    channel_org_id,
    member_id,
    user_id,
    usage_log_id,
    commission_type,
    base_amount,
    rate,
    amount,
    status,
    settlement_method,
    settlement_reference_no,
    settlement_note,
    frozen_until,
    settled_at,
    settled_by_user_id,
    reversed_from_id,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'reversed', $9, $10, $11, $12, NOW(), $13, $14, NOW(), NOW())
RETURNING id,
          channel_org_id,
          member_id,
          user_id,
          usage_log_id,
          commission_type,
          base_amount,
          rate,
          amount,
          status,
          settlement_method,
          settlement_reference_no,
          settlement_note,
          frozen_until,
          settled_at,
          settled_by_user_id,
          reversed_from_id,
          created_at,
          updated_at`,
		original.ChannelOrgID,
		original.MemberID,
		original.UserID,
		nullableInt64Arg(original.UsageLogID),
		original.CommissionType,
		original.BaseAmount,
		original.Rate,
		-original.Amount,
		original.SettlementMethod,
		original.SettlementReferenceNo,
		original.SettlementNote,
		nullableTimeArg(original.FrozenUntil),
		nullableInt64Arg(original.SettledByUserID),
		nullableInt64Arg(&original.ID),
	)
	if err != nil {
		return nil, fmt.Errorf("reverse distribution balance settlement: %w", err)
	}
	defer func() { _ = rows.Close() }()

	ledger, err := scanDistributionCommissionLedger(rows)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit reverse distribution balance settlement tx: %w", err)
	}
	return ledger, nil
}

func nullableTimeArg(v *time.Time) any {
	if v == nil {
		return nil
	}
	return *v
}

func getDistributionCommissionByIDForUpdate(ctx context.Context, tx *sql.Tx, commissionID int64) (*service.DistributionCommissionLedger, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id,
       channel_org_id,
       member_id,
       user_id,
       usage_log_id,
       commission_type,
       base_amount,
       rate,
       amount,
       status,
       settlement_method,
       settlement_reference_no,
       settlement_note,
       frozen_until,
       settled_at,
       settled_by_user_id,
       reversed_from_id,
       created_at,
       updated_at
FROM commission_ledger
WHERE id = $1
FOR UPDATE`, commissionID)
	if err != nil {
		return nil, fmt.Errorf("get distribution commission for update: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanDistributionCommissionLedger(rows)
}

func insertDistributionBalanceRedeemRecord(ctx context.Context, tx *sql.Tx, userID int64, amount float64, commissionID int64) error {
	code, err := service.GenerateRedeemCode()
	if err != nil {
		return fmt.Errorf("generate distribution balance redeem code: %w", err)
	}
	notes := fmt.Sprintf("distribution commission settlement #%d", commissionID)
	if _, err := tx.ExecContext(ctx, `
INSERT INTO redeem_codes (
    code,
    type,
    value,
    status,
    used_by,
    used_at,
    notes,
    created_at,
    group_id,
    validity_days
)
VALUES ($1, $2, $3, 'used', $4, NOW(), $5, NOW(), NULL, 0)`,
		code,
		service.RedeemTypeDistributionBalance,
		amount,
		userID,
		notes,
	); err != nil {
		return fmt.Errorf("insert distribution balance redeem record: %w", err)
	}
	return nil
}

func scanDistributionCommissionLedger(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}) (*service.DistributionCommissionLedger, error) {
	var (
		item            service.DistributionCommissionLedger
		usageLogID      sql.NullInt64
		frozenUntil     sql.NullTime
		settledAt       sql.NullTime
		settledByUserID sql.NullInt64
		reversedFromID  sql.NullInt64
	)
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrInvalidDistributionCommission
	}
	if err := rows.Scan(
		&item.ID,
		&item.ChannelOrgID,
		&item.MemberID,
		&item.UserID,
		&usageLogID,
		&item.CommissionType,
		&item.BaseAmount,
		&item.Rate,
		&item.Amount,
		&item.Status,
		&item.SettlementMethod,
		&item.SettlementReferenceNo,
		&item.SettlementNote,
		&frozenUntil,
		&settledAt,
		&settledByUserID,
		&reversedFromID,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if usageLogID.Valid {
		item.UsageLogID = &usageLogID.Int64
	}
	if frozenUntil.Valid {
		t := frozenUntil.Time
		item.FrozenUntil = &t
	}
	if settledAt.Valid {
		t := settledAt.Time
		item.SettledAt = &t
	}
	if settledByUserID.Valid {
		item.SettledByUserID = &settledByUserID.Int64
	}
	if reversedFromID.Valid {
		item.ReversedFromID = &reversedFromID.Int64
	}
	return &item, nil
}

func scanDistributionCommissionLedgerViews(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}) ([]service.DistributionCommissionLedgerView, error) {
	items := make([]service.DistributionCommissionLedgerView, 0)
	for rows.Next() {
		var (
			item            service.DistributionCommissionLedgerView
			usageLogID      sql.NullInt64
			frozenUntil     sql.NullTime
			settledAt       sql.NullTime
			settledByUserID sql.NullInt64
			reversedFromID  sql.NullInt64
		)
		if err := rows.Scan(
			&item.ID,
			&item.ChannelOrgID,
			&item.MemberID,
			&item.UserID,
			&item.UserEmail,
			&item.Username,
			&usageLogID,
			&item.CommissionType,
			&item.BaseAmount,
			&item.Rate,
			&item.Amount,
			&item.Status,
			&item.SettlementMethod,
			&item.SettlementReferenceNo,
			&item.SettlementNote,
			&frozenUntil,
			&settledAt,
			&settledByUserID,
			&reversedFromID,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if usageLogID.Valid {
			item.UsageLogID = &usageLogID.Int64
		}
		if frozenUntil.Valid {
			t := frozenUntil.Time
			item.FrozenUntil = &t
		}
		if settledAt.Valid {
			t := settledAt.Time
			item.SettledAt = &t
		}
		if settledByUserID.Valid {
			item.SettledByUserID = &settledByUserID.Int64
		}
		if reversedFromID.Valid {
			item.ReversedFromID = &reversedFromID.Int64
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
