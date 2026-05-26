package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type distributionConsumptionRepository struct {
	db *sql.DB
}

func NewDistributionConsumptionRepository(_ *dbent.Client, db *sql.DB) *distributionConsumptionRepository {
	return &distributionConsumptionRepository{db: db}
}

func (r *distributionConsumptionRepository) GetMonthlyChannelConsumption(ctx context.Context, channelOrgID int64, since time.Time) (float64, error) {
	if channelOrgID <= 0 {
		return 0, service.ErrInvalidDistributionStats
	}
	if r.db == nil {
		return 0, service.ErrInvalidDistributionStats
	}

	var total float64
	if err := r.db.QueryRowContext(ctx, `
SELECT COALESCE(SUM(COALESCE(NULLIF(ul.account_stats_cost, 0), NULLIF(ul.actual_cost, 0), ul.total_cost)), 0)
FROM usage_logs ul
JOIN user_attributions ua ON ua.user_id = ul.user_id
WHERE ua.channel_org_id = $1
  AND ul.created_at >= $2`,
		channelOrgID,
		since,
	).Scan(&total); err != nil {
		return 0, fmt.Errorf("get monthly distribution consumption: %w", err)
	}
	return total, nil
}

