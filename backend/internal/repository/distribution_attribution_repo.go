package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type distributionAttributionRepository struct {
	client *dbent.Client
	sql    *sql.DB
}

func NewDistributionAttributionRepository(client *dbent.Client, sqlDB *sql.DB) *distributionAttributionRepository {
	return &distributionAttributionRepository{client: client, sql: sqlDB}
}

func (r *distributionAttributionRepository) GetByUserID(ctx context.Context, userID int64) (*service.DistributionAttribution, error) {
	if userID <= 0 {
		return nil, service.ErrInvalidDistributionAttribution
	}
	client := r.sql
	if client == nil {
		return nil, service.ErrInvalidDistributionAttribution
	}
	rows, err := client.QueryContext(ctx, `
SELECT user_id,
       channel_org_id,
       referrer_member_id,
       promotion_link_id,
       bound_at,
       bound_source,
       bound_by,
       audit_id,
       created_at,
       updated_at
FROM user_attributions
WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("get distribution attribution: %w", err)
	}
	defer func() { _ = rows.Close() }()
	attribution, err := scanDistributionAttribution(rows)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrDistributionAttributionNotFound
		}
		return nil, err
	}
	return attribution, nil
}

func (r *distributionAttributionRepository) ListByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams) ([]service.DistributionAttributionView, *pagination.PaginationResult, error) {
	if channelOrgID <= 0 {
		return nil, nil, service.ErrInvalidDistributionAttribution
	}
	client := r.sql
	if client == nil {
		return nil, nil, service.ErrInvalidDistributionAttribution
	}

	var total int64
	if err := client.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM user_attributions
WHERE channel_org_id = $1`, channelOrgID).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count distribution attributions: %w", err)
	}

	rows, err := client.QueryContext(ctx, `
SELECT ua.user_id,
       u.email,
       COALESCE(u.username, ''),
       ua.channel_org_id,
       ua.referrer_member_id,
       ua.promotion_link_id,
       ua.bound_at,
       ua.bound_source,
       ua.bound_by,
       ua.created_at,
       ua.updated_at
FROM user_attributions ua
JOIN users u ON u.id = ua.user_id
WHERE ua.channel_org_id = $1
ORDER BY ua.bound_at DESC, ua.user_id DESC
LIMIT $2 OFFSET $3`,
		channelOrgID,
		params.Limit(),
		params.Offset(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list distribution attributions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionAttributionViews(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *distributionAttributionRepository) ListAdmin(ctx context.Context, filter service.DistributionAdminListFilter, params pagination.PaginationParams) ([]service.DistributionAttributionView, *pagination.PaginationResult, error) {
	client := r.sql
	if client == nil {
		return nil, nil, service.ErrInvalidDistributionAttribution
	}

	whereSQL := "WHERE 1=1"
	args := make([]any, 0, 4)
	if filter.ChannelOrgID > 0 {
		args = append(args, filter.ChannelOrgID)
		whereSQL += fmt.Sprintf(" AND ua.channel_org_id = $%d", len(args))
	}
	if filter.UserID > 0 {
		args = append(args, filter.UserID)
		whereSQL += fmt.Sprintf(" AND ua.user_id = $%d", len(args))
	}

	var total int64
	countQuery := `
SELECT COUNT(*)
FROM user_attributions ua
` + whereSQL
	if err := client.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count admin distribution attributions: %w", err)
	}

	args = append(args, params.Limit(), params.Offset())
	rows, err := client.QueryContext(ctx, `
SELECT ua.user_id,
       u.email,
       COALESCE(u.username, ''),
       ua.channel_org_id,
       ua.referrer_member_id,
       ua.promotion_link_id,
       ua.bound_at,
       ua.bound_source,
       ua.bound_by,
       ua.created_at,
       ua.updated_at
FROM user_attributions ua
JOIN users u ON u.id = ua.user_id
`+whereSQL+`
ORDER BY ua.bound_at DESC, ua.user_id DESC
LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args)),
		args...,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list admin distribution attributions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionAttributionViews(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *distributionAttributionRepository) ListAuditsAdmin(ctx context.Context, filter service.DistributionAdminListFilter, params pagination.PaginationParams) ([]service.DistributionAttributionAuditView, *pagination.PaginationResult, error) {
	client := r.sql
	if client == nil {
		return nil, nil, service.ErrInvalidDistributionAttribution
	}

	whereSQL := "WHERE 1=1"
	args := make([]any, 0, 4)
	if filter.ChannelOrgID > 0 {
		args = append(args, filter.ChannelOrgID)
		whereSQL += fmt.Sprintf(" AND da.new_channel_org_id = $%d", len(args))
	}
	if filter.UserID > 0 {
		args = append(args, filter.UserID)
		whereSQL += fmt.Sprintf(" AND da.user_id = $%d", len(args))
	}

	var total int64
	countQuery := `
SELECT COUNT(*)
FROM distribution_attribution_audits da
` + whereSQL
	if err := client.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count admin distribution attribution audits: %w", err)
	}

	args = append(args, params.Limit(), params.Offset())
	rows, err := client.QueryContext(ctx, `
SELECT da.id,
       da.user_id,
       u.email,
       COALESCE(u.username, ''),
       da.previous_channel_org_id,
       da.previous_referrer_member_id,
       da.previous_promotion_link_id,
       COALESCE(da.previous_bound_source, ''),
       COALESCE(da.previous_bound_by, ''),
       da.new_channel_org_id,
       da.new_referrer_member_id,
       da.new_promotion_link_id,
       da.new_bound_source,
       da.new_bound_by,
       da.note,
       da.operator_user_id,
       COALESCE(op.email, ''),
       COALESCE(op.username, ''),
       da.created_at
FROM distribution_attribution_audits da
JOIN users u ON u.id = da.user_id
LEFT JOIN users op ON op.id = da.operator_user_id
`+whereSQL+`
ORDER BY da.created_at DESC, da.id DESC
LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args)),
		args...,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list admin distribution attribution audits: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionAttributionAuditViews(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *distributionAttributionRepository) Create(ctx context.Context, input service.DistributionAttributionInput) (*service.DistributionAttribution, error) {
	client := r.sql
	if client == nil {
		return nil, service.ErrInvalidDistributionAttribution
	}
	rows, err := client.QueryContext(ctx, `
INSERT INTO user_attributions (
    user_id,
    channel_org_id,
    referrer_member_id,
    promotion_link_id,
    bound_at,
    bound_source,
    bound_by,
    audit_id,
    created_at,
    updated_at
)
SELECT $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM user_attributions WHERE user_id = $1
)
RETURNING user_id,
          channel_org_id,
          referrer_member_id,
          promotion_link_id,
          bound_at,
          bound_source,
          bound_by,
          audit_id,
          created_at,
          updated_at`,
		input.UserID,
		input.ChannelOrgID,
		nullableInt64Arg(input.ReferrerMemberID),
		nullableInt64Arg(input.PromotionLinkID),
		input.BoundAt,
		strings.TrimSpace(input.BoundSource),
		strings.TrimSpace(input.BoundBy),
		nullableInt64Arg(input.AuditID),
	)
	if err != nil {
		return nil, fmt.Errorf("create distribution attribution: %w", err)
	}
	defer func() { _ = rows.Close() }()
	attribution, err := scanDistributionAttribution(rows)
	if err == nil {
		return attribution, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("create distribution attribution: %w", err)
	}

	return r.GetByUserID(ctx, input.UserID)
}

func (r *distributionAttributionRepository) UpdateByAdmin(ctx context.Context, input service.DistributionAttributionAdminUpdateInput) (*service.DistributionAttribution, error) {
	if input.UserID <= 0 || input.ChannelOrgID <= 0 {
		return nil, service.ErrInvalidDistributionAttribution
	}
	if r.sql == nil {
		return nil, service.ErrInvalidDistributionAttribution
	}

	tx, err := r.sql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin distribution attribution admin update: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	existing, err := getDistributionAttributionByUserIDTx(ctx, tx, input.UserID)
	if err != nil && !errors.Is(err, service.ErrDistributionAttributionNotFound) {
		return nil, err
	}
	if err := validateDistributionAttributionAdminReferences(ctx, tx, input.ChannelOrgID, input.ReferrerMemberID, input.PromotionLinkID); err != nil {
		return nil, err
	}

	resolvedReferrerMemberID := input.ReferrerMemberID
	if input.PromotionLinkID != nil && *input.PromotionLinkID > 0 {
		memberID, err := getDistributionPromotionLinkMemberIDTx(ctx, tx, *input.PromotionLinkID, input.ChannelOrgID)
		if err != nil {
			return nil, err
		}
		if resolvedReferrerMemberID == nil {
			resolvedReferrerMemberID = &memberID
		} else if *resolvedReferrerMemberID != memberID {
			return nil, service.ErrInvalidDistributionAttribution
		}
	}

	boundSource := strings.ToLower(strings.TrimSpace(input.BoundSource))
	if boundSource == "" {
		boundSource = "manual"
	}
	boundBy := strings.ToLower(strings.TrimSpace(input.BoundBy))
	if boundBy == "" {
		boundBy = "admin"
	}

	auditID, err := insertDistributionAttributionAuditTx(ctx, tx, existing, input, resolvedReferrerMemberID, boundSource, boundBy)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
INSERT INTO user_attributions (
    user_id,
    channel_org_id,
    referrer_member_id,
    promotion_link_id,
    bound_at,
    bound_source,
    bound_by,
    audit_id,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, NOW(), $5, $6, $7, NOW(), NOW())
ON CONFLICT (user_id) DO UPDATE
SET channel_org_id = EXCLUDED.channel_org_id,
    referrer_member_id = EXCLUDED.referrer_member_id,
    promotion_link_id = EXCLUDED.promotion_link_id,
    bound_at = EXCLUDED.bound_at,
    bound_source = EXCLUDED.bound_source,
    bound_by = EXCLUDED.bound_by,
    audit_id = EXCLUDED.audit_id,
    updated_at = NOW()
RETURNING user_id,
          channel_org_id,
          referrer_member_id,
          promotion_link_id,
          bound_at,
          bound_source,
          bound_by,
          audit_id,
          created_at,
          updated_at`,
		input.UserID,
		input.ChannelOrgID,
		nullableInt64Arg(resolvedReferrerMemberID),
		nullableInt64Arg(input.PromotionLinkID),
		boundSource,
		boundBy,
		auditID,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert distribution attribution by admin: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out, err := scanDistributionAttribution(rows)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit distribution attribution admin update: %w", err)
	}
	committed = true
	return out, nil
}

func getDistributionAttributionByUserIDTx(ctx context.Context, tx *sql.Tx, userID int64) (*service.DistributionAttribution, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT user_id,
       channel_org_id,
       referrer_member_id,
       promotion_link_id,
       bound_at,
       bound_source,
       bound_by,
       audit_id,
       created_at,
       updated_at
FROM user_attributions
WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("get distribution attribution in tx: %w", err)
	}
	defer func() { _ = rows.Close() }()

	attribution, err := scanDistributionAttribution(rows)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrDistributionAttributionNotFound
		}
		return nil, err
	}
	return attribution, nil
}

func validateDistributionAttributionAdminReferences(ctx context.Context, tx *sql.Tx, channelOrgID int64, referrerMemberID *int64, promotionLinkID *int64) error {
	if referrerMemberID != nil && *referrerMemberID > 0 {
		var memberChannelOrgID int64
		var status string
		if err := tx.QueryRowContext(ctx, `
SELECT channel_org_id, status
FROM channel_members
WHERE id = $1`, *referrerMemberID).Scan(&memberChannelOrgID, &status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return service.ErrInvalidDistributionAttribution
			}
			return fmt.Errorf("validate distribution attribution member: %w", err)
		}
		if memberChannelOrgID != channelOrgID || !strings.EqualFold(strings.TrimSpace(status), "active") {
			return service.ErrInvalidDistributionAttribution
		}
	}
	if promotionLinkID != nil && *promotionLinkID > 0 {
		var linkChannelOrgID int64
		var status string
		if err := tx.QueryRowContext(ctx, `
SELECT channel_org_id, status
FROM promotion_links
WHERE id = $1`, *promotionLinkID).Scan(&linkChannelOrgID, &status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return service.ErrInvalidDistributionAttribution
			}
			return fmt.Errorf("validate distribution attribution promotion link: %w", err)
		}
		if linkChannelOrgID != channelOrgID || !strings.EqualFold(strings.TrimSpace(status), "active") {
			return service.ErrInvalidDistributionAttribution
		}
	}
	return nil
}

func getDistributionPromotionLinkMemberIDTx(ctx context.Context, tx *sql.Tx, promotionLinkID int64, channelOrgID int64) (int64, error) {
	var memberID int64
	if err := tx.QueryRowContext(ctx, `
SELECT member_id
FROM promotion_links
WHERE id = $1
  AND channel_org_id = $2`, promotionLinkID, channelOrgID).Scan(&memberID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, service.ErrInvalidDistributionAttribution
		}
		return 0, fmt.Errorf("get distribution promotion link member: %w", err)
	}
	return memberID, nil
}

func insertDistributionAttributionAuditTx(
	ctx context.Context,
	tx *sql.Tx,
	existing *service.DistributionAttribution,
	input service.DistributionAttributionAdminUpdateInput,
	resolvedReferrerMemberID *int64,
	boundSource string,
	boundBy string,
) (int64, error) {
	var (
		previousChannelOrgID     any
		previousReferrerMemberID any
		previousPromotionLinkID  any
		previousBoundSource      any
		previousBoundBy          any
	)
	if existing != nil {
		previousChannelOrgID = existing.ChannelOrgID
		previousReferrerMemberID = nullableInt64Arg(existing.ReferrerMemberID)
		previousPromotionLinkID = nullableInt64Arg(existing.PromotionLinkID)
		previousBoundSource = existing.BoundSource
		previousBoundBy = existing.BoundBy
	}

	var auditID int64
	if err := tx.QueryRowContext(ctx, `
INSERT INTO distribution_attribution_audits (
    user_id,
    previous_channel_org_id,
    previous_referrer_member_id,
    previous_promotion_link_id,
    previous_bound_source,
    previous_bound_by,
    new_channel_org_id,
    new_referrer_member_id,
    new_promotion_link_id,
    new_bound_source,
    new_bound_by,
    note,
    operator_user_id,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW())
RETURNING id`,
		input.UserID,
		previousChannelOrgID,
		previousReferrerMemberID,
		previousPromotionLinkID,
		previousBoundSource,
		previousBoundBy,
		input.ChannelOrgID,
		nullableInt64Arg(resolvedReferrerMemberID),
		nullableInt64Arg(input.PromotionLinkID),
		boundSource,
		boundBy,
		strings.TrimSpace(input.Note),
		nullableInt64Arg(input.OperatorUserID),
	).Scan(&auditID); err != nil {
		return 0, fmt.Errorf("insert distribution attribution audit: %w", err)
	}
	return auditID, nil
}

func scanDistributionAttribution(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}) (*service.DistributionAttribution, error) {
	var (
		out              service.DistributionAttribution
		referrerMemberID sql.NullInt64
		promotionLinkID  sql.NullInt64
		auditID          sql.NullInt64
	)
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, sql.ErrNoRows
	}
	if err := rows.Scan(
		&out.UserID,
		&out.ChannelOrgID,
		&referrerMemberID,
		&promotionLinkID,
		&out.BoundAt,
		&out.BoundSource,
		&out.BoundBy,
		&auditID,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if referrerMemberID.Valid {
		out.ReferrerMemberID = &referrerMemberID.Int64
	}
	if promotionLinkID.Valid {
		out.PromotionLinkID = &promotionLinkID.Int64
	}
	if auditID.Valid {
		out.AuditID = &auditID.Int64
	}
	return &out, nil
}

func scanDistributionAttributionViews(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}) ([]service.DistributionAttributionView, error) {
	items := make([]service.DistributionAttributionView, 0)
	for rows.Next() {
		var (
			item             service.DistributionAttributionView
			referrerMemberID sql.NullInt64
			promotionLinkID  sql.NullInt64
		)
		if err := rows.Scan(
			&item.UserID,
			&item.UserEmail,
			&item.Username,
			&item.ChannelOrgID,
			&referrerMemberID,
			&promotionLinkID,
			&item.BoundAt,
			&item.BoundSource,
			&item.BoundBy,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if referrerMemberID.Valid {
			item.ReferrerMemberID = &referrerMemberID.Int64
		}
		if promotionLinkID.Valid {
			item.PromotionLinkID = &promotionLinkID.Int64
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func scanDistributionAttributionAuditViews(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}) ([]service.DistributionAttributionAuditView, error) {
	items := make([]service.DistributionAttributionAuditView, 0)
	for rows.Next() {
		var (
			item                     service.DistributionAttributionAuditView
			previousChannelOrgID     sql.NullInt64
			previousReferrerMemberID sql.NullInt64
			previousPromotionLinkID  sql.NullInt64
			newReferrerMemberID      sql.NullInt64
			newPromotionLinkID       sql.NullInt64
			operatorUserID           sql.NullInt64
		)
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.UserEmail,
			&item.Username,
			&previousChannelOrgID,
			&previousReferrerMemberID,
			&previousPromotionLinkID,
			&item.PreviousBoundSource,
			&item.PreviousBoundBy,
			&item.NewChannelOrgID,
			&newReferrerMemberID,
			&newPromotionLinkID,
			&item.NewBoundSource,
			&item.NewBoundBy,
			&item.Note,
			&operatorUserID,
			&item.OperatorUserEmail,
			&item.OperatorUsername,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if previousChannelOrgID.Valid {
			item.PreviousChannelOrgID = &previousChannelOrgID.Int64
		}
		if previousReferrerMemberID.Valid {
			item.PreviousReferrerMemberID = &previousReferrerMemberID.Int64
		}
		if previousPromotionLinkID.Valid {
			item.PreviousPromotionLinkID = &previousPromotionLinkID.Int64
		}
		if newReferrerMemberID.Valid {
			item.NewReferrerMemberID = &newReferrerMemberID.Int64
		}
		if newPromotionLinkID.Valid {
			item.NewPromotionLinkID = &newPromotionLinkID.Int64
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
