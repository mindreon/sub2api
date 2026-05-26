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

type distributionPromotionRepository struct {
	db *sql.DB
}

func NewDistributionPromotionRepository(_ *dbent.Client, db *sql.DB) *distributionPromotionRepository {
	return &distributionPromotionRepository{db: db}
}

func (r *distributionPromotionRepository) GetByCode(ctx context.Context, code string) (*service.DistributionPromotionLink, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, service.ErrInvalidDistributionPromotionLink
	}

	if r.db == nil {
		return nil, service.ErrInvalidDistributionPromotionLink
	}
	client := r.db
	rows, err := client.QueryContext(ctx, `
SELECT pl.id,
       pl.channel_org_id,
       pl.member_id,
       pl.code,
       pl.target_type,
       pl.status,
       u.id,
       u.email,
       COALESCE(u.username, ''),
       m.role_type,
       pl.created_at,
       pl.updated_at
FROM promotion_links pl
JOIN channel_members m ON m.id = pl.member_id
JOIN users u ON u.id = m.user_id
WHERE pl.code = $1`, code)
	if err != nil {
		return nil, fmt.Errorf("get distribution promotion link: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanDistributionPromotionLink(rows)
}

func (r *distributionPromotionRepository) Create(ctx context.Context, input service.DistributionPromotionLinkInput) (*service.DistributionPromotionLink, error) {
	if input.MemberID <= 0 {
		return nil, service.ErrInvalidDistributionPromotionLink
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionPromotionLink
	}

	code := strings.ToUpper(strings.TrimSpace(input.Code))
	if code == "" {
		return nil, service.ErrInvalidDistributionPromotionLink
	}

	rows, err := r.db.QueryContext(ctx, `
WITH inserted AS (
    INSERT INTO promotion_links (
        channel_org_id,
        member_id,
        code,
        target_type,
        status,
        created_at,
        updated_at
    )
    SELECT m.channel_org_id,
           $1,
           $2,
           $3,
           $4,
           NOW(),
           NOW()
    FROM channel_members m
    WHERE m.id = $1
    RETURNING id,
              channel_org_id,
              member_id,
              code,
              target_type,
              status,
              created_at,
              updated_at
)
SELECT i.id,
       i.channel_org_id,
       i.member_id,
       i.code,
       i.target_type,
       i.status,
       u.id,
       u.email,
       COALESCE(u.username, ''),
       m.role_type,
       i.created_at,
       i.updated_at
FROM inserted i
JOIN channel_members m ON m.id = i.member_id
JOIN users u ON u.id = m.user_id`,
		input.MemberID,
		code,
		strings.TrimSpace(input.TargetType),
		strings.TrimSpace(input.Status),
	)
	if err != nil {
		return nil, fmt.Errorf("create distribution promotion link: %w", err)
	}
	defer func() { _ = rows.Close() }()
	link, err := scanDistributionPromotionLink(rows)
	if err != nil {
		return nil, err
	}
	return link, nil
}

func (r *distributionPromotionRepository) ListByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams) ([]service.DistributionPromotionLink, *pagination.PaginationResult, error) {
	if channelOrgID <= 0 {
		return nil, nil, service.ErrInvalidDistributionPromotionLink
	}
	if r.db == nil {
		return nil, nil, service.ErrInvalidDistributionPromotionLink
	}
	client := r.db

	var total int64
	if err := client.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM promotion_links
WHERE channel_org_id = $1`, channelOrgID).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count distribution promotion links: %w", err)
	}

	rows, err := client.QueryContext(ctx, `
SELECT pl.id,
       pl.channel_org_id,
       pl.member_id,
       pl.code,
       pl.target_type,
       pl.status,
       u.id,
       u.email,
       COALESCE(u.username, ''),
       m.role_type,
       pl.created_at,
       pl.updated_at
FROM promotion_links pl
JOIN channel_members m ON m.id = pl.member_id
JOIN users u ON u.id = m.user_id
WHERE pl.channel_org_id = $1
ORDER BY pl.created_at DESC, pl.id DESC
LIMIT $2 OFFSET $3`,
		channelOrgID,
		params.Limit(),
		params.Offset(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list distribution promotion links: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionPromotionLinks(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *distributionPromotionRepository) ListAdmin(ctx context.Context, filter service.DistributionAdminListFilter, params pagination.PaginationParams) ([]service.DistributionPromotionLink, *pagination.PaginationResult, error) {
	if r.db == nil {
		return nil, nil, service.ErrInvalidDistributionPromotionLink
	}
	client := r.db

	whereSQL := "WHERE 1=1"
	args := make([]any, 0, 5)
	if filter.ChannelOrgID > 0 {
		args = append(args, filter.ChannelOrgID)
		whereSQL += fmt.Sprintf(" AND pl.channel_org_id = $%d", len(args))
	}
	if filter.UserID > 0 {
		args = append(args, filter.UserID)
		whereSQL += fmt.Sprintf(" AND m.user_id = $%d", len(args))
	}
	if roleType := strings.TrimSpace(filter.RoleType); roleType != "" {
		args = append(args, strings.ToLower(roleType))
		whereSQL += fmt.Sprintf(" AND m.role_type = $%d", len(args))
	}

	var total int64
	countQuery := `
SELECT COUNT(*)
FROM promotion_links pl
JOIN channel_members m ON m.id = pl.member_id
JOIN users u ON u.id = m.user_id
` + whereSQL
	if err := client.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count admin distribution promotion links: %w", err)
	}

	args = append(args, params.Limit(), params.Offset())
	rows, err := client.QueryContext(ctx, `
SELECT pl.id,
       pl.channel_org_id,
       pl.member_id,
       pl.code,
       pl.target_type,
       pl.status,
       u.id,
       u.email,
       COALESCE(u.username, ''),
       m.role_type,
       pl.created_at,
       pl.updated_at
FROM promotion_links pl
JOIN channel_members m ON m.id = pl.member_id
JOIN users u ON u.id = m.user_id
`+whereSQL+`
ORDER BY pl.created_at DESC, pl.id DESC
LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args)),
		args...,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list admin distribution promotion links: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionPromotionLinks(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func scanDistributionPromotionLink(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}) (*service.DistributionPromotionLink, error) {
	var out service.DistributionPromotionLink
	var (
		userID   sql.NullInt64
		email    sql.NullString
		username sql.NullString
		roleType sql.NullString
	)
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrDistributionPromotionLinkNotFound
	}
	if err := rows.Scan(
		&out.ID,
		&out.ChannelOrgID,
		&out.MemberID,
		&out.Code,
		&out.TargetType,
		&out.Status,
		&userID,
		&email,
		&username,
		&roleType,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if userID.Valid {
		out.UserID = userID.Int64
	}
	if email.Valid {
		out.UserEmail = email.String
	}
	if username.Valid {
		out.Username = username.String
	}
	if roleType.Valid {
		out.RoleType = roleType.String
	}
	return &out, nil
}

func scanDistributionPromotionLinks(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}) ([]service.DistributionPromotionLink, error) {
	items := make([]service.DistributionPromotionLink, 0)
	for rows.Next() {
		var (
			item     service.DistributionPromotionLink
			userID   sql.NullInt64
			email    sql.NullString
			username sql.NullString
			roleType sql.NullString
		)
		if err := rows.Scan(
			&item.ID,
			&item.ChannelOrgID,
			&item.MemberID,
			&item.Code,
			&item.TargetType,
			&item.Status,
			&userID,
			&email,
			&username,
			&roleType,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if userID.Valid {
			item.UserID = userID.Int64
		}
		if email.Valid {
			item.UserEmail = email.String
		}
		if username.Valid {
			item.Username = username.String
		}
		if roleType.Valid {
			item.RoleType = roleType.String
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
