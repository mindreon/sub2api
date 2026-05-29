package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type distributionOrganizationRepository struct {
	db *sql.DB
}

func NewDistributionOrganizationRepository(_ *dbent.Client, db *sql.DB) *distributionOrganizationRepository {
	return &distributionOrganizationRepository{db: db}
}

func (r *distributionOrganizationRepository) List(ctx context.Context, params pagination.PaginationParams) ([]service.DistributionOrganization, *pagination.PaginationResult, error) {
	if r.db == nil {
		return nil, nil, service.ErrInvalidDistributionOrganization
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM channel_organizations`).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count distribution organizations: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
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
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2`,
		params.Limit(),
		params.Offset(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list distribution organizations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionOrganizations(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *distributionOrganizationRepository) GetByID(ctx context.Context, id int64) (*service.DistributionOrganization, error) {
	if id <= 0 {
		return nil, service.ErrInvalidDistributionOrganization
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionOrganization
	}

	rows, err := r.db.QueryContext(ctx, `
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
WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("get distribution organization: %w", err)
	}
	defer func() { _ = rows.Close() }()

	org, err := scanDistributionOrganization(rows)
	if err != nil {
		return nil, err
	}
	return org, nil
}

func (r *distributionOrganizationRepository) GetByOwnerUserID(ctx context.Context, userID int64) (*service.DistributionOrganization, error) {
	if userID <= 0 {
		return nil, service.ErrInvalidDistributionOrganization
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionOrganization
	}

	rows, err := r.db.QueryContext(ctx, `
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
WHERE owner_user_id = $1
ORDER BY created_at ASC, id ASC
LIMIT 1`, userID)
	if err != nil {
		return nil, fmt.Errorf("get distribution organization by owner: %w", err)
	}
	defer func() { _ = rows.Close() }()

	org, err := scanDistributionOrganization(rows)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, service.ErrInvalidDistributionOrganization) {
			return nil, nil
		}
		return nil, err
	}
	return org, nil
}

func (r *distributionOrganizationRepository) GetByBrandHost(ctx context.Context, host string) (*service.DistributionOrganization, error) {
	host = service.NormalizeDistributionBrandHost(host)
	if host == "" {
		return nil, nil
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionOrganization
	}

	rows, err := r.db.QueryContext(ctx, `
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
WHERE status = 'active'
ORDER BY created_at ASC, id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list distribution organizations by brand host: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionOrganizations(rows)
	if err != nil {
		return nil, err
	}

	var fallback *service.DistributionOrganization
	for i := range items {
		org := items[i]
		domain := service.NormalizeDistributionBrandHost(brandConfigString(org.BrandConfig, "domain"))
		apiDomain := service.NormalizeDistributionBrandHost(brandConfigString(org.BrandConfig, "api_domain"))
		switch {
		case domain != "" && domain == host:
			return &org, nil
		case fallback == nil && apiDomain != "" && apiDomain == host:
			orgCopy := org
			fallback = &orgCopy
		}
	}

	return fallback, nil
}

func (r *distributionOrganizationRepository) Create(ctx context.Context, input service.DistributionOrganizationInput) (*service.DistributionOrganization, error) {
	if input.Type == "" || input.Name == "" || input.Status == "" {
		return nil, service.ErrInvalidDistributionOrganization
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionOrganization
	}

	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, service.ErrInvalidDistributionOrganization.WithCause(err)
	}
	brandConfigJSON, err := json.Marshal(input.BrandConfig)
	if err != nil {
		return nil, service.ErrInvalidDistributionOrganization.WithCause(err)
	}

	rows, err := r.db.QueryContext(ctx, `
WITH inserted AS (
    INSERT INTO channel_organizations (
        type,
        name,
        owner_user_id,
        status,
        config,
        brand_config,
        created_at,
        updated_at
    )
    VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
    RETURNING id,
              type,
              name,
              owner_user_id,
              status,
              config,
              brand_config,
              created_at,
              updated_at
),
wallet AS (
    INSERT INTO channel_wallets (
        channel_org_id,
        prepaid_balance,
        commission_reserved,
        total_recharged,
        total_consumed,
        warning_threshold,
        status,
        created_at,
        updated_at
    )
    SELECT id, 0, 0, 0, 0, 0, 'active', NOW(), NOW()
    FROM inserted
    ON CONFLICT (channel_org_id) DO NOTHING
)
SELECT id,
       type,
       name,
       owner_user_id,
       status,
       config,
       brand_config,
       created_at,
       updated_at
FROM inserted`,
		input.Type,
		input.Name,
		nullableInt64Arg(input.OwnerUserID),
		input.Status,
		configJSON,
		brandConfigJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("create distribution organization: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanDistributionOrganization(rows)
}

func (r *distributionOrganizationRepository) Update(ctx context.Context, id int64, input service.DistributionOrganizationInput) (*service.DistributionOrganization, error) {
	if id <= 0 || input.Type == "" || input.Name == "" || input.Status == "" {
		return nil, service.ErrInvalidDistributionOrganization
	}
	if r.db == nil {
		return nil, service.ErrInvalidDistributionOrganization
	}

	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, service.ErrInvalidDistributionOrganization.WithCause(err)
	}
	brandConfigJSON, err := json.Marshal(input.BrandConfig)
	if err != nil {
		return nil, service.ErrInvalidDistributionOrganization.WithCause(err)
	}

	rows, err := r.db.QueryContext(ctx, `
UPDATE channel_organizations
SET type = $2,
    name = $3,
    owner_user_id = $4,
    status = $5,
    config = $6,
    brand_config = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING id,
          type,
          name,
          owner_user_id,
          status,
          config,
          brand_config,
          created_at,
          updated_at`,
		id,
		input.Type,
		input.Name,
		nullableInt64Arg(input.OwnerUserID),
		input.Status,
		configJSON,
		brandConfigJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("update distribution organization: %w", err)
	}
	defer func() { _ = rows.Close() }()

	org, err := scanDistributionOrganization(rows)
	if err != nil {
		return nil, err
	}
	return org, nil
}

func scanDistributionOrganizations(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}) ([]service.DistributionOrganization, error) {
	items := make([]service.DistributionOrganization, 0)
	for rows.Next() {
		item, err := scanDistributionOrganizationRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func scanDistributionOrganization(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}) (*service.DistributionOrganization, error) {
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrInvalidDistributionOrganization
	}
	return scanDistributionOrganizationRow(rows)
}

func scanDistributionOrganizationRow(rows interface {
	Scan(dest ...any) error
}) (*service.DistributionOrganization, error) {
	var (
		out             service.DistributionOrganization
		ownerUserID     sql.NullInt64
		configJSON      []byte
		brandConfigJSON []byte
	)
	if err := rows.Scan(
		&out.ID,
		&out.Type,
		&out.Name,
		&ownerUserID,
		&out.Status,
		&configJSON,
		&brandConfigJSON,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if ownerUserID.Valid {
		out.OwnerUserID = &ownerUserID.Int64
	}
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &out.Config); err != nil {
			return nil, err
		}
	}
	if out.Config == nil {
		out.Config = map[string]any{}
	}
	if len(brandConfigJSON) > 0 {
		if err := json.Unmarshal(brandConfigJSON, &out.BrandConfig); err != nil {
			return nil, err
		}
	}
	if out.BrandConfig == nil {
		out.BrandConfig = map[string]any{}
	}
	return &out, nil
}

func brandConfigString(config map[string]any, key string) string {
	if len(config) == 0 {
		return ""
	}
	value, ok := config[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}
