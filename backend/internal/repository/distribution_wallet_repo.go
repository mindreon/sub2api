package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type distributionWalletRepository struct {
	db *sql.DB
}

func NewDistributionWalletRepository(_ *dbent.Client, db *sql.DB) *distributionWalletRepository {
	return &distributionWalletRepository{db: db}
}

func (r *distributionWalletRepository) List(ctx context.Context, filter service.DistributionAdminListFilter, params pagination.PaginationParams) ([]service.DistributionWallet, *pagination.PaginationResult, error) {
	if r.db == nil {
		return nil, nil, service.ErrInvalidDistributionWallet
	}
	client := r.db

	whereSQL := "WHERE 1=1"
	args := make([]any, 0, 3)
	if filter.ChannelOrgID > 0 {
		args = append(args, filter.ChannelOrgID)
		whereSQL += fmt.Sprintf(" AND w.channel_org_id = $%d", len(args))
	}

	var total int64
	countQuery := `
SELECT COUNT(*)
FROM channel_wallets w
` + whereSQL
	if err := client.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count distribution wallets: %w", err)
	}

	args = append(args, params.Limit(), params.Offset())
	rows, err := client.QueryContext(ctx, `
SELECT w.channel_org_id,
       o.name,
       o.type,
       w.prepaid_balance,
       w.commission_reserved,
       w.total_recharged,
       w.total_consumed,
       w.warning_threshold,
       w.status,
       w.created_at,
       w.updated_at
FROM channel_wallets w
JOIN channel_organizations o ON o.id = w.channel_org_id
`+whereSQL+`
ORDER BY w.updated_at DESC, w.channel_org_id DESC
LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args)),
		args...,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list distribution wallets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionWalletViews(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *distributionWalletRepository) ListTransactions(ctx context.Context, filter service.DistributionAdminListFilter, params pagination.PaginationParams) ([]service.DistributionWalletTransaction, *pagination.PaginationResult, error) {
	if r.db == nil {
		return nil, nil, service.ErrInvalidDistributionWalletTransaction
	}

	whereSQL := "WHERE 1=1"
	args := make([]any, 0, 4)
	if filter.ChannelOrgID > 0 {
		args = append(args, filter.ChannelOrgID)
		whereSQL += fmt.Sprintf(" AND t.channel_org_id = $%d", len(args))
	}
	if filter.TransactionType != "" {
		args = append(args, filter.TransactionType)
		whereSQL += fmt.Sprintf(" AND t.transaction_type = $%d", len(args))
	}

	var total int64
	countQuery := `
SELECT COUNT(*)
FROM channel_wallet_transactions t
` + whereSQL
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count distribution wallet transactions: %w", err)
	}

	args = append(args, params.Limit(), params.Offset())
	rows, err := r.db.QueryContext(ctx, `
SELECT t.id,
       t.channel_org_id,
       o.name,
       o.type,
       t.transaction_type,
       t.amount,
       t.prepaid_balance_before,
       t.prepaid_balance_after,
       t.commission_reserved_before,
       t.commission_reserved_after,
       t.reference_no,
       t.note,
       t.operator_user_id,
       t.created_at
FROM channel_wallet_transactions t
JOIN channel_organizations o ON o.id = t.channel_org_id
`+whereSQL+`
ORDER BY t.created_at DESC, t.id DESC
LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args)),
		args...,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list distribution wallet transactions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionWalletTransactions(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *distributionWalletRepository) ListTransactionsByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams, transactionType string) ([]service.DistributionWalletTransaction, *pagination.PaginationResult, error) {
	if channelOrgID <= 0 {
		return nil, nil, service.ErrInvalidDistributionWalletTransaction
	}
	return r.ListTransactions(ctx, service.DistributionAdminListFilter{
		ChannelOrgID:    channelOrgID,
		TransactionType: transactionType,
	}, params)
}

func (r *distributionWalletRepository) GetByChannelOrgID(ctx context.Context, channelOrgID int64) (*service.DistributionWallet, error) {
	if channelOrgID <= 0 {
		return nil, service.ErrInvalidDistributionWallet
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionWallet
	}
	return getDistributionWalletByChannelOrgID(ctx, r.db, channelOrgID, false)
}

func (r *distributionWalletRepository) SyncStatus(ctx context.Context, channelOrgID int64, consumptionLimit float64) (*service.DistributionWallet, error) {
	if channelOrgID <= 0 {
		return nil, service.ErrInvalidDistributionWallet
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionWallet
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin distribution wallet status sync: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	current, err := getDistributionWalletByChannelOrgID(ctx, tx, channelOrgID, true)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, service.ErrInvalidDistributionWallet
	}
	org, err := getDistributionOrganizationByID(ctx, tx, channelOrgID)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, service.ErrInvalidDistributionOrganization
	}
	if consumptionLimit <= 0 {
		consumptionLimit = distributionWalletConsumptionLimitFromConfig(org.Config)
	}

	nextStatus := service.ResolveDistributionWalletStatus(
		current.Status,
		current.PrepaidBalance,
		current.CommissionReserved,
		current.TotalConsumed,
		consumptionLimit,
	)
	if nextStatus == service.NormalizeDistributionWalletStatus(current.Status) {
		current.Status = nextStatus
		if err := syncDistributionAlertEvents(ctx, tx, org, current); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit distribution wallet status sync: %w", err)
		}
		return current, nil
	}

	updated, err := updateDistributionWalletStatus(ctx, tx, channelOrgID, nextStatus)
	if err != nil {
		return nil, err
	}
	if err := syncDistributionAlertEvents(ctx, tx, org, updated); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit distribution wallet status sync: %w", err)
	}
	return updated, nil
}

func (r *distributionWalletRepository) UpdateWarningThreshold(ctx context.Context, channelOrgID int64, warningThreshold float64) (*service.DistributionWallet, error) {
	if channelOrgID <= 0 {
		return nil, service.ErrInvalidDistributionWallet
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionWallet
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin distribution wallet warning threshold transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.QueryContext(ctx, `
UPDATE channel_wallets
SET warning_threshold = $2,
    updated_at = NOW()
WHERE channel_org_id = $1
RETURNING channel_org_id,
          (SELECT name FROM channel_organizations WHERE id = channel_org_id),
          (SELECT type FROM channel_organizations WHERE id = channel_org_id),
          prepaid_balance,
          commission_reserved,
          total_recharged,
          total_consumed,
          warning_threshold,
          status,
          created_at,
          updated_at`,
		channelOrgID,
		warningThreshold,
	)
	if err != nil {
		return nil, fmt.Errorf("update distribution wallet warning threshold: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionWalletViews(rows)
	if err != nil {
		return nil, err
	}
	_ = rows.Close()
	if len(items) == 0 {
		return nil, service.ErrInvalidDistributionWallet
	}
	org, err := getDistributionOrganizationByID(ctx, tx, channelOrgID)
	if err != nil {
		return nil, err
	}
	if err := syncDistributionAlertEvents(ctx, tx, org, &items[0]); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit distribution wallet warning threshold transaction: %w", err)
	}
	return &items[0], nil
}

func (r *distributionWalletRepository) Recharge(ctx context.Context, channelOrgID int64, input service.DistributionWalletRechargeInput) (*service.DistributionWallet, error) {
	return r.mutateWallet(ctx, channelOrgID, input.Amount, func(current *service.DistributionWallet) (*distributionWalletMutationState, error) {
		return &distributionWalletMutationState{
			transactionType: "recharge",
			prepaidBalance:  current.PrepaidBalance + input.Amount,
			totalRecharged:  current.TotalRecharged + input.Amount,
			totalConsumed:   current.TotalConsumed,
			reservedBalance: current.CommissionReserved,
			referenceNo:     input.ReferenceNo,
			note:            input.Note,
			operatorUserID:  input.OperatorUserID,
		}, nil
	})
}

func (r *distributionWalletRepository) RefundPrepaidBalance(ctx context.Context, channelOrgID int64, input service.DistributionWalletRefundInput) (*service.DistributionWallet, error) {
	return r.mutateWallet(ctx, channelOrgID, input.Amount, func(current *service.DistributionWallet) (*distributionWalletMutationState, error) {
		if current.PrepaidBalance-current.CommissionReserved < input.Amount {
			return nil, service.ErrDistributionWalletInsufficientBalance
		}
		return &distributionWalletMutationState{
			transactionType: "refund",
			prepaidBalance:  current.PrepaidBalance - input.Amount,
			reservedBalance: current.CommissionReserved,
			totalRecharged:  current.TotalRecharged,
			totalConsumed:   current.TotalConsumed,
			referenceNo:     input.ReferenceNo,
			note:            input.Note,
			operatorUserID:  input.OperatorUserID,
		}, nil
	})
}

func (r *distributionWalletRepository) ReserveCommission(ctx context.Context, channelOrgID int64, amount float64) (*service.DistributionWallet, error) {
	return r.mutateWallet(ctx, channelOrgID, amount, func(current *service.DistributionWallet) (*distributionWalletMutationState, error) {
		if current.PrepaidBalance-current.CommissionReserved < amount {
			return nil, service.ErrDistributionWalletInsufficientBalance
		}
		return &distributionWalletMutationState{
			transactionType: "commission_reserve",
			prepaidBalance:  current.PrepaidBalance,
			reservedBalance: current.CommissionReserved + amount,
			totalRecharged:  current.TotalRecharged,
			totalConsumed:   current.TotalConsumed,
		}, nil
	})
}

func (r *distributionWalletRepository) ReleaseCommission(ctx context.Context, channelOrgID int64, amount float64) (*service.DistributionWallet, error) {
	return r.mutateWallet(ctx, channelOrgID, amount, func(current *service.DistributionWallet) (*distributionWalletMutationState, error) {
		if current.CommissionReserved < amount {
			return nil, service.ErrDistributionWalletInsufficientReserved
		}
		return &distributionWalletMutationState{
			transactionType: "commission_release",
			prepaidBalance:  current.PrepaidBalance,
			reservedBalance: current.CommissionReserved - amount,
			totalRecharged:  current.TotalRecharged,
			totalConsumed:   current.TotalConsumed,
		}, nil
	})
}

func (r *distributionWalletRepository) SettleReservedCommission(ctx context.Context, channelOrgID int64, amount float64) (*service.DistributionWallet, error) {
	return r.mutateWallet(ctx, channelOrgID, amount, func(current *service.DistributionWallet) (*distributionWalletMutationState, error) {
		if current.PrepaidBalance < amount {
			return nil, service.ErrDistributionWalletInsufficientBalance
		}
		if current.CommissionReserved < amount {
			return nil, service.ErrDistributionWalletInsufficientReserved
		}
		return &distributionWalletMutationState{
			transactionType: "commission_settle",
			prepaidBalance:  current.PrepaidBalance - amount,
			reservedBalance: current.CommissionReserved - amount,
			totalRecharged:  current.TotalRecharged,
			totalConsumed:   current.TotalConsumed,
		}, nil
	})
}

func (r *distributionWalletRepository) DeductCommission(ctx context.Context, channelOrgID int64, amount float64) (*service.DistributionWallet, error) {
	return r.mutateWallet(ctx, channelOrgID, amount, func(current *service.DistributionWallet) (*distributionWalletMutationState, error) {
		if current.PrepaidBalance < amount {
			return nil, service.ErrDistributionWalletInsufficientBalance
		}
		return &distributionWalletMutationState{
			transactionType: "commission_deduct",
			prepaidBalance:  current.PrepaidBalance - amount,
			reservedBalance: current.CommissionReserved,
			totalRecharged:  current.TotalRecharged,
			totalConsumed:   current.TotalConsumed,
		}, nil
	})
}

func (r *distributionWalletRepository) RefundCommission(ctx context.Context, channelOrgID int64, amount float64) (*service.DistributionWallet, error) {
	return r.mutateWallet(ctx, channelOrgID, amount, func(current *service.DistributionWallet) (*distributionWalletMutationState, error) {
		return &distributionWalletMutationState{
			transactionType: "commission_refund",
			prepaidBalance:  current.PrepaidBalance + amount,
			reservedBalance: current.CommissionReserved,
			totalRecharged:  current.TotalRecharged,
			totalConsumed:   current.TotalConsumed,
		}, nil
	})
}

func (r *distributionWalletRepository) ConsumeUsage(ctx context.Context, channelOrgID int64, amount float64, consumptionLimit float64) (*service.DistributionWallet, error) {
	return r.mutateWallet(ctx, channelOrgID, amount, func(current *service.DistributionWallet) (*distributionWalletMutationState, error) {
		if consumptionLimit > 0 && current.TotalConsumed+amount > consumptionLimit {
			return nil, service.ErrDistributionChannelConsumptionLimitExceeded
		}
		if current.PrepaidBalance-current.CommissionReserved < amount {
			return nil, service.ErrDistributionWalletInsufficientBalance
		}
		return &distributionWalletMutationState{
			transactionType: "consume",
			prepaidBalance:  current.PrepaidBalance - amount,
			reservedBalance: current.CommissionReserved,
			totalRecharged:  current.TotalRecharged,
			totalConsumed:   current.TotalConsumed + amount,
		}, nil
	})
}

func (r *distributionWalletRepository) GetAdminStats(ctx context.Context) (*service.DistributionAdminStats, error) {
	if r.db == nil {
		return nil, service.ErrInvalidDistributionStats
	}
	client := r.db
	if err := thawDistributionCommissionsByChannelOrgID(ctx, client, 0); err != nil {
		return nil, err
	}

	var out service.DistributionAdminStats
	if err := client.QueryRowContext(ctx, `
SELECT
	COUNT(*) AS organization_count,
	COUNT(*) FILTER (WHERE type = 'platform') AS platform_count,
	COUNT(*) FILTER (WHERE type = 'reseller') AS reseller_count,
	COUNT(*) FILTER (WHERE type = 'oem') AS oem_count
FROM channel_organizations`).Scan(&out.OrganizationCount, &out.PlatformCount, &out.ResellerCount, &out.OemCount); err != nil {
		return nil, fmt.Errorf("get distribution admin org stats: %w", err)
	}

	if err := client.QueryRowContext(ctx, `
SELECT
	COUNT(*) AS member_count,
	COUNT(*) FILTER (WHERE role_type = 'agent') AS agent_count,
	COUNT(*) FILTER (WHERE role_type = 'kol1') AS kol1_count,
	COUNT(*) FILTER (WHERE role_type = 'kol2') AS kol2_count
FROM channel_members`).Scan(&out.MemberCount, &out.AgentCount, &out.Kol1Count, &out.Kol2Count); err != nil {
		return nil, fmt.Errorf("get distribution admin member stats: %w", err)
	}

	if err := client.QueryRowContext(ctx, `SELECT COUNT(*) FROM promotion_links`).Scan(&out.PromotionLinkCount); err != nil {
		return nil, fmt.Errorf("get distribution admin promotion stats: %w", err)
	}
	if err := client.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_attributions`).Scan(&out.AttributionCount); err != nil {
		return nil, fmt.Errorf("get distribution admin attribution stats: %w", err)
	}
	if err := client.QueryRowContext(ctx, `
SELECT
	COUNT(*) AS commission_count,
	COALESCE(SUM(CASE WHEN status = 'frozen' THEN amount ELSE 0 END), 0) AS frozen_amount,
	COALESCE(SUM(CASE WHEN status = 'available' THEN amount ELSE 0 END), 0) AS available_amount,
	COALESCE(SUM(CASE WHEN status = 'settled' THEN amount ELSE 0 END), 0) AS settled_amount
FROM commission_ledger`).Scan(&out.CommissionCount, &out.FrozenCommissionAmount, &out.AvailableCommissionAmount, &out.SettledCommissionAmount); err != nil {
		return nil, fmt.Errorf("get distribution admin commission stats: %w", err)
	}
	if err := client.QueryRowContext(ctx, `
SELECT
	COUNT(*) AS wallet_count,
	COALESCE(SUM(prepaid_balance), 0) AS prepaid_balance_total,
	COALESCE(SUM(commission_reserved), 0) AS commission_reserved_total,
	COALESCE(SUM(total_recharged), 0) AS total_recharged,
	COALESCE(SUM(total_consumed), 0) AS total_consumed
	FROM channel_wallets`).Scan(&out.WalletCount, &out.PrepaidBalanceTotal, &out.CommissionReservedTotal, &out.TotalRecharged, &out.TotalConsumed); err != nil {
		return nil, fmt.Errorf("get distribution admin wallet stats: %w", err)
	}
	out.CommissionExpenseRatio = distributionCommissionExpenseRatio(
		out.FrozenCommissionAmount+out.AvailableCommissionAmount+out.SettledCommissionAmount,
		out.TotalConsumed,
	)
	upperRatio, err := distributionGlobalCommissionUpperRatio(ctx, client)
	if err != nil {
		return nil, err
	}
	out.CommissionUpperRatio = upperRatio

	return &out, nil
}

type distributionWalletMutationState struct {
	transactionType string
	prepaidBalance  float64
	reservedBalance float64
	totalRecharged  float64
	totalConsumed   float64
	status          string
	referenceNo     string
	note            string
	operatorUserID  *int64
}

type distributionWalletMutationBuilder func(current *service.DistributionWallet) (*distributionWalletMutationState, error)

type distributionWalletQueryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func (r *distributionWalletRepository) mutateWallet(ctx context.Context, channelOrgID int64, amount float64, build distributionWalletMutationBuilder) (*service.DistributionWallet, error) {
	if channelOrgID <= 0 || amount <= 0 {
		return nil, service.ErrInvalidDistributionWallet
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionWallet
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin distribution wallet transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	current, err := getDistributionWalletByChannelOrgID(ctx, tx, channelOrgID, true)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, service.ErrInvalidDistributionWallet
	}
	org, err := getDistributionOrganizationByID(ctx, tx, channelOrgID)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, service.ErrInvalidDistributionOrganization
	}

	next, err := build(current)
	if err != nil {
		return nil, err
	}
	if next == nil {
		return nil, service.ErrInvalidDistributionWallet
	}
	next.status = resolveDistributionWalletMutationStatus(current, next, org)

	updated, err := updateDistributionWalletState(ctx, tx, channelOrgID, next)
	if err != nil {
		return nil, err
	}
	if err := syncDistributionAlertEvents(ctx, tx, org, updated); err != nil {
		return nil, err
	}
	if err := insertDistributionWalletTransaction(ctx, tx, channelOrgID, amount, current, updated, next); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit distribution wallet mutation: %w", err)
	}
	return updated, nil
}

func (r *distributionWalletRepository) GetByChannelOrgIDForUpdate(ctx context.Context, channelOrgID int64) (*service.DistributionWallet, error) {
	return getDistributionWalletByChannelOrgID(ctx, r.db, channelOrgID, true)
}

func (r *distributionWalletRepository) GetChannelSummary(ctx context.Context, channelOrgID int64) (*service.DistributionChannelSummary, error) {
	if channelOrgID <= 0 {
		return nil, service.ErrInvalidDistributionWallet
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionWallet
	}
	client := r.db
	if err := thawDistributionCommissionsByChannelOrgID(ctx, client, channelOrgID); err != nil {
		return nil, err
	}

	var summary service.DistributionChannelSummary
	orgRows, err := client.QueryContext(ctx, `
SELECT id,
       type,
       name,
       owner_user_id,
       status,
       config,
       brand_config,
       created_at,
       updated_at
FROM channel_organizations
WHERE id = $1`, channelOrgID)
	if err != nil {
		return nil, fmt.Errorf("get distribution channel organization: %w", err)
	}
	defer func() { _ = orgRows.Close() }()
	org, err := scanDistributionOrganization(orgRows)
	if err != nil {
		return nil, err
	}
	summary.Organization = *org

	wallet, err := r.SyncStatus(ctx, channelOrgID, distributionWalletConsumptionLimitFromConfig(org.Config))
	if err != nil {
		return nil, err
	}
	summary.Wallet = *wallet

	if err := client.QueryRowContext(ctx, `
SELECT
	COUNT(*) AS member_count,
	COUNT(*) FILTER (WHERE role_type = 'agent') AS agent_count,
	COUNT(*) FILTER (WHERE role_type = 'kol1') AS kol1_count,
	COUNT(*) FILTER (WHERE role_type = 'kol2') AS kol2_count
FROM channel_members
WHERE channel_org_id = $1`, channelOrgID).Scan(&summary.MemberCount, &summary.AgentCount, &summary.Kol1Count, &summary.Kol2Count); err != nil {
		return nil, fmt.Errorf("get distribution channel member stats: %w", err)
	}

	if err := client.QueryRowContext(ctx, `SELECT COUNT(*) FROM promotion_links WHERE channel_org_id = $1`, channelOrgID).Scan(&summary.PromotionLinkCount); err != nil {
		return nil, fmt.Errorf("get distribution channel promotion stats: %w", err)
	}
	if err := client.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_attributions WHERE channel_org_id = $1`, channelOrgID).Scan(&summary.AttributionCount); err != nil {
		return nil, fmt.Errorf("get distribution channel attribution stats: %w", err)
	}
	if err := client.QueryRowContext(ctx, `
SELECT
	COUNT(*) AS commission_count,
	COALESCE(SUM(CASE WHEN status = 'frozen' THEN amount ELSE 0 END), 0) AS frozen_amount,
	COALESCE(SUM(CASE WHEN status = 'available' THEN amount ELSE 0 END), 0) AS available_amount,
	COALESCE(SUM(CASE WHEN status = 'settled' THEN amount ELSE 0 END), 0) AS settled_amount
FROM commission_ledger
WHERE channel_org_id = $1`, channelOrgID).Scan(&summary.CommissionCount, &summary.FrozenCommissionAmount, &summary.AvailableCommissionAmount, &summary.SettledCommissionAmount); err != nil {
		return nil, fmt.Errorf("get distribution channel commission stats: %w", err)
	}

	return &summary, nil
}

func getDistributionWalletByChannelOrgID(ctx context.Context, q distributionWalletQueryer, channelOrgID int64, forUpdate bool) (*service.DistributionWallet, error) {
	if channelOrgID <= 0 {
		return nil, service.ErrInvalidDistributionWallet
	}
	rows, err := q.QueryContext(ctx, `
SELECT w.channel_org_id,
       o.name,
       o.type,
       w.prepaid_balance,
       w.commission_reserved,
       w.total_recharged,
       w.total_consumed,
       w.warning_threshold,
       w.status,
       w.created_at,
       w.updated_at
FROM channel_wallets w
JOIN channel_organizations o ON o.id = w.channel_org_id
WHERE w.channel_org_id = $1`+distributionWalletForUpdateSQL(forUpdate), channelOrgID)
	if err != nil {
		return nil, fmt.Errorf("get distribution wallet: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionWalletViews(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, service.ErrInvalidDistributionWallet
	}
	return &items[0], nil
}

func distributionWalletForUpdateSQL(forUpdate bool) string {
	if forUpdate {
		return ` FOR UPDATE`
	}
	return ""
}

func updateDistributionWalletState(ctx context.Context, tx *sql.Tx, channelOrgID int64, next *distributionWalletMutationState) (*service.DistributionWallet, error) {
	rows, err := tx.QueryContext(ctx, `
UPDATE channel_wallets
SET prepaid_balance = $2,
    commission_reserved = $3,
    total_recharged = $4,
    total_consumed = $5,
    status = $6,
    updated_at = NOW()
WHERE channel_org_id = $1
RETURNING channel_org_id,
          (SELECT name FROM channel_organizations WHERE id = channel_org_id),
          (SELECT type FROM channel_organizations WHERE id = channel_org_id),
          prepaid_balance,
          commission_reserved,
          total_recharged,
          total_consumed,
          warning_threshold,
          status,
          created_at,
          updated_at`,
		channelOrgID,
		next.prepaidBalance,
		next.reservedBalance,
		next.totalRecharged,
		next.totalConsumed,
		next.status,
	)
	if err != nil {
		return nil, fmt.Errorf("update distribution wallet state: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionWalletViews(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, service.ErrInvalidDistributionWallet
	}
	return &items[0], nil
}

func updateDistributionWalletStatus(ctx context.Context, tx *sql.Tx, channelOrgID int64, status string) (*service.DistributionWallet, error) {
	rows, err := tx.QueryContext(ctx, `
UPDATE channel_wallets
SET status = $2,
    updated_at = NOW()
WHERE channel_org_id = $1
RETURNING channel_org_id,
          (SELECT name FROM channel_organizations WHERE id = channel_org_id),
          (SELECT type FROM channel_organizations WHERE id = channel_org_id),
          prepaid_balance,
          commission_reserved,
          total_recharged,
          total_consumed,
          warning_threshold,
          status,
          created_at,
          updated_at`,
		channelOrgID,
		status,
	)
	if err != nil {
		return nil, fmt.Errorf("update distribution wallet status: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionWalletViews(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, service.ErrInvalidDistributionWallet
	}
	return &items[0], nil
}

func resolveDistributionWalletMutationStatus(
	current *service.DistributionWallet,
	next *distributionWalletMutationState,
	org *service.DistributionOrganization,
) string {
	if current == nil || next == nil {
		return "active"
	}
	if service.NormalizeDistributionWalletStatus(current.Status) == "disabled" {
		return "disabled"
	}
	return service.ResolveDistributionWalletStatus(
		current.Status,
		next.prepaidBalance,
		next.reservedBalance,
		next.totalConsumed,
		distributionWalletConsumptionLimitFromConfig(distributionOrganizationConfigMap(org)),
	)
}

func distributionWalletConsumptionLimitFromConfig(config map[string]any) float64 {
	for _, key := range []string{"consumption_limit", "credit_limit", "usage_limit"} {
		value, ok := config[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return typed
		case float32:
			return float64(typed)
		case int:
			return float64(typed)
		case int64:
			return float64(typed)
		case string:
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func distributionGlobalCommissionUpperRatio(ctx context.Context, q distributionWalletQueryer) (float64, error) {
	rows, err := q.QueryContext(ctx, `SELECT value FROM settings WHERE key = $1`, service.SettingKeyDistributionCommissionUpperRatio)
	if err != nil {
		return 0, fmt.Errorf("get distribution commission upper ratio: %w", err)
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, err
		}
		return service.DistributionCommissionUpperRatioDefault / 100, nil
	}

	var raw sql.NullString
	if err := rows.Scan(&raw); err != nil {
		return 0, err
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if !raw.Valid {
		return service.DistributionCommissionUpperRatioDefault / 100, nil
	}
	value, err := strconv.ParseFloat(strings.TrimSpace(raw.String), 64)
	if err != nil || value < 0 {
		return service.DistributionCommissionUpperRatioDefault / 100, nil
	}
	if value > 100 {
		return service.DistributionCommissionUpperRatioDefault / 100, nil
	}
	return value / 100, nil
}

func distributionCommissionExpenseRatio(commissionAmount float64, consumptionAmount float64) float64 {
	if consumptionAmount <= 0 {
		return 0
	}
	return math.Round((commissionAmount/consumptionAmount)*1e8) / 1e8
}

func getDistributionOrganizationByID(ctx context.Context, q distributionWalletQueryer, channelOrgID int64) (*service.DistributionOrganization, error) {
	rows, err := q.QueryContext(ctx, `
SELECT id,
       type,
       name,
       owner_user_id,
       status,
       config,
       brand_config,
       created_at,
       updated_at
FROM channel_organizations
WHERE id = $1`, channelOrgID)
	if err != nil {
		return nil, fmt.Errorf("get distribution organization: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanDistributionOrganization(rows)
}

func distributionOrganizationConfigMap(org *service.DistributionOrganization) map[string]any {
	if org == nil || org.Config == nil {
		return map[string]any{}
	}
	return org.Config
}

func insertDistributionWalletTransaction(
	ctx context.Context,
	tx *sql.Tx,
	channelOrgID int64,
	amount float64,
	before *service.DistributionWallet,
	after *service.DistributionWallet,
	state *distributionWalletMutationState,
) error {
	if before == nil || after == nil || state == nil {
		return service.ErrInvalidDistributionWalletTransaction
	}
	state.transactionType = strings.TrimSpace(state.transactionType)
	if state.transactionType == "" {
		return service.ErrInvalidDistributionWalletTransaction
	}
	rows, err := tx.QueryContext(ctx, `
INSERT INTO channel_wallet_transactions (
    channel_org_id,
    transaction_type,
    amount,
    prepaid_balance_before,
    prepaid_balance_after,
    commission_reserved_before,
    commission_reserved_after,
    reference_no,
    note,
    operator_user_id,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
RETURNING id`,
		channelOrgID,
		state.transactionType,
		amount,
		before.PrepaidBalance,
		after.PrepaidBalance,
		before.CommissionReserved,
		after.CommissionReserved,
		strings.TrimSpace(state.referenceNo),
		strings.TrimSpace(state.note),
		nullableInt64Arg(state.operatorUserID),
	)
	if err != nil {
		return fmt.Errorf("insert distribution wallet transaction: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var id int64
	if !rows.Next() {
		return service.ErrInvalidDistributionWalletTransaction
	}
	if err := rows.Scan(&id); err != nil {
		return err
	}
	return rows.Err()
}

func scanDistributionWalletViews(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}) ([]service.DistributionWallet, error) {
	items := make([]service.DistributionWallet, 0)
	for rows.Next() {
		var item service.DistributionWallet
		if err := rows.Scan(
			&item.ChannelOrgID,
			&item.OrganizationName,
			&item.OrganizationType,
			&item.PrepaidBalance,
			&item.CommissionReserved,
			&item.TotalRecharged,
			&item.TotalConsumed,
			&item.WarningThreshold,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func scanDistributionWalletTransactions(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}) ([]service.DistributionWalletTransaction, error) {
	items := make([]service.DistributionWalletTransaction, 0)
	for rows.Next() {
		var (
			item           service.DistributionWalletTransaction
			operatorUserID sql.NullInt64
		)
		if err := rows.Scan(
			&item.ID,
			&item.ChannelOrgID,
			&item.OrganizationName,
			&item.OrganizationType,
			&item.TransactionType,
			&item.Amount,
			&item.PrepaidBalanceBefore,
			&item.PrepaidBalanceAfter,
			&item.CommissionReservedBefore,
			&item.CommissionReservedAfter,
			&item.ReferenceNo,
			&item.Note,
			&operatorUserID,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if operatorUserID.Valid {
			item.OperatorUserID = &operatorUserID.Int64
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
