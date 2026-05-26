package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type distributionMemberRepository struct {
	db *sql.DB
}

func NewDistributionMemberRepository(_ *dbent.Client, db *sql.DB) *distributionMemberRepository {
	return &distributionMemberRepository{db: db}
}

func (r *distributionMemberRepository) Create(ctx context.Context, input service.DistributionMemberInput) (*service.DistributionMemberView, error) {
	if input.ChannelOrgID <= 0 || input.UserID <= 0 || input.RoleType == "" || input.Status == "" || input.CommissionRate < 0 {
		return nil, service.ErrInvalidDistributionMember
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionMember
	}

	rows, err := r.db.QueryContext(ctx, `
WITH inserted AS (
    INSERT INTO channel_members (
        channel_org_id,
        user_id,
        role_type,
        parent_member_id,
        level_code,
        commission_rate,
        status,
        created_at,
        updated_at
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
    RETURNING id
)
SELECT m.id,
       m.user_id,
       u.email,
       COALESCE(u.username, ''),
       m.channel_org_id,
       m.role_type,
       m.parent_member_id,
       m.level_code,
       m.commission_rate,
       m.status,
       m.created_at,
       m.updated_at
FROM channel_members m
JOIN inserted i ON i.id = m.id
JOIN users u ON u.id = m.user_id`,
		input.ChannelOrgID,
		input.UserID,
		input.RoleType,
		nullableInt64Arg(input.ParentMemberID),
		input.LevelCode,
		input.CommissionRate,
		input.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("create distribution member: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionMemberViews(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, service.ErrInvalidDistributionMember
	}
	return &items[0], nil
}

func (r *distributionMemberRepository) GetByID(ctx context.Context, memberID int64) (*service.DistributionMemberView, error) {
	if memberID <= 0 {
		return nil, service.ErrInvalidDistributionMember
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionMember
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT m.id,
       m.user_id,
       u.email,
       COALESCE(u.username, ''),
       m.channel_org_id,
       m.role_type,
       m.parent_member_id,
       m.level_code,
       m.commission_rate,
       m.status,
       m.created_at,
       m.updated_at
FROM channel_members m
JOIN users u ON u.id = m.user_id
WHERE m.id = $1`, memberID)
	if err != nil {
		return nil, fmt.Errorf("get distribution member: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionMemberViews(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, service.ErrDistributionMemberNotFound
	}
	return &items[0], nil
}

func (r *distributionMemberRepository) GetByChannelOrgIDAndRole(ctx context.Context, channelOrgID int64, roleType string) (*service.DistributionMemberView, error) {
	if channelOrgID <= 0 || strings.TrimSpace(roleType) == "" {
		return nil, service.ErrInvalidDistributionMember
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionMember
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT m.id,
       m.user_id,
       u.email,
       COALESCE(u.username, ''),
       m.channel_org_id,
       m.role_type,
       m.parent_member_id,
       m.level_code,
       m.commission_rate,
       m.status,
       m.created_at,
       m.updated_at
FROM channel_members m
JOIN users u ON u.id = m.user_id
WHERE m.channel_org_id = $1
  AND m.role_type = $2
  AND m.status = 'active'
ORDER BY m.created_at ASC, m.id ASC
LIMIT 1`, channelOrgID, strings.TrimSpace(roleType))
	if err != nil {
		return nil, fmt.Errorf("get distribution member by channel role: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionMemberViews(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, service.ErrDistributionMemberNotFound
	}
	return &items[0], nil
}

func (r *distributionMemberRepository) ListByUserID(ctx context.Context, userID int64) ([]service.DistributionMemberView, error) {
	if userID <= 0 {
		return nil, service.ErrInvalidDistributionMember
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionMember
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT m.id,
       m.user_id,
       u.email,
       COALESCE(u.username, ''),
       m.channel_org_id,
       m.role_type,
       m.parent_member_id,
       m.level_code,
       m.commission_rate,
       m.status,
       m.created_at,
       m.updated_at
FROM channel_members m
JOIN users u ON u.id = m.user_id
WHERE m.user_id = $1
ORDER BY m.created_at DESC, m.id DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list distribution members by user: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanDistributionMemberViews(rows)
}

func (r *distributionMemberRepository) CountByChannelOrgIDAndRole(ctx context.Context, channelOrgID int64, roleType string) (int64, error) {
	if channelOrgID <= 0 || strings.TrimSpace(roleType) == "" {
		return 0, service.ErrInvalidDistributionMember
	}
	if r.db == nil {
		return 0, service.ErrInvalidDistributionMember
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM channel_members
WHERE channel_org_id = $1
  AND role_type = $2`,
		channelOrgID,
		strings.TrimSpace(roleType),
	).Scan(&total); err != nil {
		return 0, fmt.Errorf("count distribution members by role: %w", err)
	}
	return total, nil
}

func (r *distributionMemberRepository) ListByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams, roleType string) ([]service.DistributionMemberView, *pagination.PaginationResult, error) {
	if channelOrgID <= 0 {
		return nil, nil, service.ErrInvalidDistributionAttribution
	}
	if r.db == nil {
		return nil, nil, service.ErrInvalidDistributionAttribution
	}
	client := r.db

	filterSQL := ""
	args := []any{channelOrgID}
	if strings.TrimSpace(roleType) != "" {
		filterSQL = " AND m.role_type = $2"
		args = append(args, strings.TrimSpace(roleType))
	}

	var total int64
	countQuery := `
SELECT COUNT(*)
FROM channel_members m
WHERE m.channel_org_id = $1` + filterSQL
	if err := client.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count distribution members: %w", err)
	}

	query := `
SELECT m.id,
       m.user_id,
       u.email,
       COALESCE(u.username, ''),
       m.channel_org_id,
       m.role_type,
       m.parent_member_id,
       m.level_code,
       m.commission_rate,
       m.status,
       m.created_at,
       m.updated_at
FROM channel_members m
JOIN users u ON u.id = m.user_id
WHERE m.channel_org_id = $1` + filterSQL + `
ORDER BY m.created_at DESC, m.id DESC
LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, params.Limit(), params.Offset())

	rows, err := client.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list distribution members: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionMemberViews(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *distributionMemberRepository) ListAdmin(ctx context.Context, filter service.DistributionAdminListFilter, params pagination.PaginationParams) ([]service.DistributionMemberView, *pagination.PaginationResult, error) {
	if r.db == nil {
		return nil, nil, service.ErrInvalidDistributionMember
	}
	client := r.db

	whereSQL := "WHERE 1=1"
	args := make([]any, 0, 6)
	if filter.ChannelOrgID > 0 {
		args = append(args, filter.ChannelOrgID)
		whereSQL += fmt.Sprintf(" AND m.channel_org_id = $%d", len(args))
	}
	if filter.UserID > 0 {
		args = append(args, filter.UserID)
		whereSQL += fmt.Sprintf(" AND m.user_id = $%d", len(args))
	}
	roleType := strings.TrimSpace(filter.RoleType)
	if roleType != "" {
		args = append(args, roleType)
		whereSQL += fmt.Sprintf(" AND m.role_type = $%d", len(args))
	}
	if filter.Q != "" {
		args = append(args, "%"+filter.Q+"%")
		whereSQL += fmt.Sprintf(" AND (u.username ILIKE $%d OR u.email ILIKE $%d)", len(args), len(args))
	}

	var total int64
	countQuery := `
SELECT COUNT(*)
FROM channel_members m
JOIN users u ON u.id = m.user_id
` + whereSQL
	if err := client.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count admin distribution members: %w", err)
	}

	args = append(args, params.Limit(), params.Offset())
	query := `
SELECT m.id,
       m.user_id,
       u.email,
       COALESCE(u.username, ''),
       m.channel_org_id,
       m.role_type,
       m.parent_member_id,
       m.level_code,
       m.commission_rate,
       m.status,
       m.created_at,
       m.updated_at
FROM channel_members m
JOIN users u ON u.id = m.user_id
` + whereSQL + `
ORDER BY m.created_at DESC, m.id DESC
LIMIT $` + fmt.Sprintf("%d", len(args)-1) + ` OFFSET $` + fmt.Sprintf("%d", len(args))

	rows, err := client.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list admin distribution members: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionMemberViews(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func scanDistributionMemberViews(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}) ([]service.DistributionMemberView, error) {
	items := make([]service.DistributionMemberView, 0)
	for rows.Next() {
		var (
			item           service.DistributionMemberView
			parentMemberID sql.NullInt64
		)
		if err := rows.Scan(
			&item.MemberID,
			&item.UserID,
			&item.UserEmail,
			&item.Username,
			&item.ChannelOrgID,
			&item.RoleType,
			&parentMemberID,
			&item.LevelCode,
			&item.CommissionRate,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if parentMemberID.Valid {
			item.ParentMemberID = &parentMemberID.Int64
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
