package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type distributionAnalyticsRepository struct {
	db *sql.DB
}

func NewDistributionAnalyticsRepository(_ *dbent.Client, db *sql.DB) *distributionAnalyticsRepository {
	return &distributionAnalyticsRepository{db: db}
}

func (r *distributionAnalyticsRepository) GetChannelAnalyticsSummary(ctx context.Context, channelOrgID int64, filter service.DistributionAnalyticsFilter) (*service.DistributionAnalyticsSummary, error) {
	if err := validateDistributionAnalyticsChannel(channelOrgID, r.db); err != nil {
		return nil, err
	}

	balanceStatuses := []string{payment.OrderStatusPaid, payment.OrderStatusRecharging, payment.OrderStatusCompleted}
	eventTimeExpr := distributionRechargeEventTimeExpression()
	query := fmt.Sprintf(`
SELECT
	COALESCE((SELECT COUNT(*) FROM user_attributions ua WHERE ua.channel_org_id = $1 AND ua.bound_at >= $2 AND ua.bound_at < $3), 0),
	COALESCE((SELECT SUM(po.amount)::double precision
	          FROM payment_orders po
	          JOIN user_attributions ua ON ua.user_id = po.user_id
	          WHERE ua.channel_org_id = $1
	            AND po.order_type = $4
	            AND po.status = ANY($5)
	            AND %s >= $2
	            AND %s < $3), 0),
	COALESCE((SELECT SUM(COALESCE(NULLIF(ul.account_stats_cost, 0), NULLIF(ul.actual_cost, 0), ul.total_cost))::double precision
	          FROM usage_logs ul
	          JOIN user_attributions ua ON ua.user_id = ul.user_id
	          WHERE ua.channel_org_id = $1
	            AND ul.created_at >= $2
	            AND ul.created_at < $3), 0),
	COALESCE((SELECT SUM(CASE WHEN cl.status NOT IN ('reversed', 'cancelled') THEN cl.amount ELSE 0 END)::double precision
	          FROM commission_ledger cl
	          WHERE cl.channel_org_id = $1
	            AND cl.created_at >= $2
	            AND cl.created_at < $3), 0),
	COALESCE((SELECT SUM(CASE WHEN cl.status = 'settled' THEN cl.amount ELSE 0 END)::double precision
	          FROM commission_ledger cl
	          WHERE cl.channel_org_id = $1
	            AND COALESCE(cl.settled_at, cl.updated_at, cl.created_at) >= $2
	            AND COALESCE(cl.settled_at, cl.updated_at, cl.created_at) < $3), 0),
	COALESCE((SELECT COUNT(*) FROM channel_members m WHERE m.channel_org_id = $1 AND m.status = 'active' AND m.role_type IN ('agent', 'kol1', 'kol2')), 0),
	COALESCE((SELECT COUNT(*) FROM channel_members m WHERE m.channel_org_id = $1 AND m.status = 'active' AND m.role_type = 'agent'), 0),
	COALESCE((SELECT COUNT(*) FROM channel_members m WHERE m.channel_org_id = $1 AND m.status = 'active' AND m.role_type = 'kol1'), 0),
	COALESCE((SELECT COUNT(*) FROM channel_members m WHERE m.channel_org_id = $1 AND m.status = 'active' AND m.role_type = 'kol2'), 0)
`, eventTimeExpr, eventTimeExpr)

	var summary service.DistributionAnalyticsSummary
	if err := r.db.QueryRowContext(ctx, query, channelOrgID, filter.StartTime, filter.EndTime, payment.OrderTypeBalance, pq.Array(balanceStatuses)).Scan(
		&summary.RegisteredUsers,
		&summary.RechargeAmount,
		&summary.ConsumptionAmount,
		&summary.CommissionAmount,
		&summary.SettledCommissionAmount,
		&summary.MemberCount,
		&summary.AgentCount,
		&summary.Kol1Count,
		&summary.Kol2Count,
	); err != nil {
		return nil, fmt.Errorf("get distribution channel analytics summary: %w", err)
	}
	summary.CommissionExpenseRatio = distributionCommissionExpenseRatio(summary.CommissionAmount, summary.ConsumptionAmount)
	if org, err := getDistributionAnalyticsOrganization(ctx, r.db, channelOrgID); err == nil && org != nil {
		summary.CommissionUpperRatio = distributionAnalyticsCommissionUpperRatio(org.Config)
	}
	return &summary, nil
}

func (r *distributionAnalyticsRepository) ListChannelTrend(ctx context.Context, channelOrgID int64, filter service.DistributionAnalyticsFilter) ([]service.DistributionAnalyticsTrendPoint, error) {
	if err := validateDistributionAnalyticsChannel(channelOrgID, r.db); err != nil {
		return nil, err
	}

	dateFormat := safeDateFormat(filter.Granularity)
	dateTruncUnit := distributionAnalyticsDateTruncUnit(filter.Granularity)
	balanceStatuses := []string{payment.OrderStatusPaid, payment.OrderStatusRecharging, payment.OrderStatusCompleted}
	eventTimeExpr := distributionRechargeEventTimeExpression()
	query := fmt.Sprintf(`
WITH registrations AS (
	SELECT DATE_TRUNC('%s', ua.bound_at) AS bucket_start,
	       COUNT(*)::bigint AS registered_users
	FROM user_attributions ua
	WHERE ua.channel_org_id = $1
	  AND ua.bound_at >= $2
	  AND ua.bound_at < $3
	GROUP BY 1
),
recharges AS (
	SELECT DATE_TRUNC('%s', %s) AS bucket_start,
	       COALESCE(SUM(po.amount), 0)::double precision AS recharge_amount
	FROM payment_orders po
	JOIN user_attributions ua ON ua.user_id = po.user_id
	WHERE ua.channel_org_id = $1
	  AND po.order_type = $4
	  AND po.status = ANY($5)
	  AND %s >= $2
	  AND %s < $3
	GROUP BY 1
),
consumptions AS (
	SELECT DATE_TRUNC('%s', ul.created_at) AS bucket_start,
	       COALESCE(SUM(COALESCE(NULLIF(ul.account_stats_cost, 0), NULLIF(ul.actual_cost, 0), ul.total_cost)), 0)::double precision AS consumption_amount
	FROM usage_logs ul
	JOIN user_attributions ua ON ua.user_id = ul.user_id
	WHERE ua.channel_org_id = $1
	  AND ul.created_at >= $2
	  AND ul.created_at < $3
	GROUP BY 1
),
commissions AS (
	SELECT DATE_TRUNC('%s', cl.created_at) AS bucket_start,
	       COALESCE(SUM(CASE WHEN cl.status NOT IN ('reversed', 'cancelled') THEN cl.amount ELSE 0 END), 0)::double precision AS commission_amount
	FROM commission_ledger cl
	WHERE cl.channel_org_id = $1
	  AND cl.created_at >= $2
	  AND cl.created_at < $3
	GROUP BY 1
),
settled AS (
	SELECT DATE_TRUNC('%s', COALESCE(cl.settled_at, cl.updated_at, cl.created_at)) AS bucket_start,
	       COALESCE(SUM(CASE WHEN cl.status = 'settled' THEN cl.amount ELSE 0 END), 0)::double precision AS settled_commission_amount
	FROM commission_ledger cl
	WHERE cl.channel_org_id = $1
	  AND COALESCE(cl.settled_at, cl.updated_at, cl.created_at) >= $2
	  AND COALESCE(cl.settled_at, cl.updated_at, cl.created_at) < $3
	GROUP BY 1
),
buckets AS (
	SELECT bucket_start FROM registrations
	UNION
	SELECT bucket_start FROM recharges
	UNION
	SELECT bucket_start FROM consumptions
	UNION
	SELECT bucket_start FROM commissions
	UNION
	SELECT bucket_start FROM settled
)
SELECT
	TO_CHAR(b.bucket_start, '%s') AS date,
	COALESCE(registrations.registered_users, 0),
	COALESCE(recharges.recharge_amount, 0),
	COALESCE(consumptions.consumption_amount, 0),
	COALESCE(commissions.commission_amount, 0),
	COALESCE(settled.settled_commission_amount, 0)
FROM buckets b
LEFT JOIN registrations ON registrations.bucket_start = b.bucket_start
LEFT JOIN recharges ON recharges.bucket_start = b.bucket_start
LEFT JOIN consumptions ON consumptions.bucket_start = b.bucket_start
LEFT JOIN commissions ON commissions.bucket_start = b.bucket_start
LEFT JOIN settled ON settled.bucket_start = b.bucket_start
ORDER BY b.bucket_start ASC
`, dateTruncUnit, dateTruncUnit, eventTimeExpr, eventTimeExpr, eventTimeExpr, dateTruncUnit, dateTruncUnit, dateTruncUnit, dateFormat)

	rows, err := r.db.QueryContext(ctx, query, channelOrgID, filter.StartTime, filter.EndTime, payment.OrderTypeBalance, pq.Array(balanceStatuses))
	if err != nil {
		return nil, fmt.Errorf("list distribution channel trend: %w", err)
	}
	defer func() { _ = rows.Close() }()

	points := make([]service.DistributionAnalyticsTrendPoint, 0)
	for rows.Next() {
		var item service.DistributionAnalyticsTrendPoint
		if err := rows.Scan(
			&item.Date,
			&item.RegisteredUsers,
			&item.RechargeAmount,
			&item.ConsumptionAmount,
			&item.CommissionAmount,
			&item.SettledCommissionAmount,
		); err != nil {
			return nil, fmt.Errorf("scan distribution channel trend: %w", err)
		}
		points = append(points, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate distribution channel trend: %w", err)
	}
	return points, nil
}

func (r *distributionAnalyticsRepository) ListChannelMemberRanking(ctx context.Context, channelOrgID int64, filter service.DistributionAnalyticsFilter, limit int) ([]service.DistributionAnalyticsRankingItem, error) {
	if err := validateDistributionAnalyticsChannel(channelOrgID, r.db); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = filter.Limit
	}
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.db.QueryContext(ctx, distributionAnalyticsRankingQuery("m.channel_org_id = $1"), channelOrgID, filter.StartTime, filter.EndTime, payment.OrderTypeBalance, pq.Array([]string{payment.OrderStatusPaid, payment.OrderStatusRecharging, payment.OrderStatusCompleted}), limit)
	if err != nil {
		return nil, fmt.Errorf("list distribution channel member ranking: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanDistributionAnalyticsRankingRows(rows, "scan distribution channel member ranking")
}

func (r *distributionAnalyticsRepository) GetMemberAnalyticsSummary(ctx context.Context, memberIDs []int64, filter service.DistributionAnalyticsFilter) (*service.DistributionAnalyticsSummary, error) {
	if err := validateDistributionAnalyticsMemberIDs(memberIDs, r.db); err != nil {
		return nil, err
	}

	balanceStatuses := []string{payment.OrderStatusPaid, payment.OrderStatusRecharging, payment.OrderStatusCompleted}
	eventTimeExpr := distributionRechargeEventTimeExpression()
	query := fmt.Sprintf(`
SELECT
	COALESCE((SELECT COUNT(*) FROM user_attributions ua WHERE ua.referrer_member_id = ANY($1) AND ua.bound_at >= $2 AND ua.bound_at < $3), 0),
	COALESCE((SELECT SUM(po.amount)::double precision
	          FROM payment_orders po
	          JOIN user_attributions ua ON ua.user_id = po.user_id
	          WHERE ua.referrer_member_id = ANY($1)
	            AND po.order_type = $4
	            AND po.status = ANY($5)
	            AND %s >= $2
	            AND %s < $3), 0),
	COALESCE((SELECT SUM(COALESCE(NULLIF(ul.account_stats_cost, 0), NULLIF(ul.actual_cost, 0), ul.total_cost))::double precision
	          FROM usage_logs ul
	          JOIN user_attributions ua ON ua.user_id = ul.user_id
	          WHERE ua.referrer_member_id = ANY($1)
	            AND ul.created_at >= $2
	            AND ul.created_at < $3), 0),
	COALESCE((SELECT SUM(CASE WHEN cl.status NOT IN ('reversed', 'cancelled') THEN cl.amount ELSE 0 END)::double precision
	          FROM commission_ledger cl
	          WHERE cl.member_id = ANY($1)
	            AND cl.created_at >= $2
	            AND cl.created_at < $3), 0),
	COALESCE((SELECT SUM(CASE WHEN cl.status = 'settled' THEN cl.amount ELSE 0 END)::double precision
	          FROM commission_ledger cl
	          WHERE cl.member_id = ANY($1)
	            AND COALESCE(cl.settled_at, cl.updated_at, cl.created_at) >= $2
	            AND COALESCE(cl.settled_at, cl.updated_at, cl.created_at) < $3), 0)
`, eventTimeExpr, eventTimeExpr)

	var summary service.DistributionAnalyticsSummary
	if err := r.db.QueryRowContext(ctx, query, pq.Array(memberIDs), filter.StartTime, filter.EndTime, payment.OrderTypeBalance, pq.Array(balanceStatuses)).Scan(
		&summary.RegisteredUsers,
		&summary.RechargeAmount,
		&summary.ConsumptionAmount,
		&summary.CommissionAmount,
		&summary.SettledCommissionAmount,
	); err != nil {
		return nil, fmt.Errorf("get distribution member analytics summary: %w", err)
	}
	return &summary, nil
}

func (r *distributionAnalyticsRepository) ListChildMemberRanking(ctx context.Context, parentMemberIDs []int64, filter service.DistributionAnalyticsFilter, limit int) ([]service.DistributionAnalyticsRankingItem, error) {
	if err := validateDistributionAnalyticsMemberIDs(parentMemberIDs, r.db); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = filter.Limit
	}
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.db.QueryContext(ctx, distributionAnalyticsRankingQuery("m.parent_member_id = ANY($1)"), pq.Array(parentMemberIDs), filter.StartTime, filter.EndTime, payment.OrderTypeBalance, pq.Array([]string{payment.OrderStatusPaid, payment.OrderStatusRecharging, payment.OrderStatusCompleted}), limit)
	if err != nil {
		return nil, fmt.Errorf("list distribution child member ranking: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanDistributionAnalyticsRankingRows(rows, "scan distribution child member ranking")
}

func (r *distributionAnalyticsRepository) GetAttributedUserStats(ctx context.Context, memberIDs []int64, filter service.DistributionAnalyticsFilter) (*service.DistributionAttributedUserStats, error) {
	if err := validateDistributionAnalyticsMemberIDs(memberIDs, r.db); err != nil {
		return nil, err
	}
	query := `
SELECT
	COALESCE((SELECT COUNT(*) FROM user_attributions ua WHERE ua.referrer_member_id = ANY($1) AND ua.bound_at < $3), 0) AS total_users,
	COALESCE((SELECT COUNT(*) FROM user_attributions ua WHERE ua.referrer_member_id = ANY($1) AND ua.bound_at >= $2 AND ua.bound_at < $3), 0) AS new_users
`
	stats := &service.DistributionAttributedUserStats{}
	if err := r.db.QueryRowContext(ctx, query, pq.Array(memberIDs), filter.StartTime, filter.EndTime).Scan(&stats.TotalUsers, &stats.NewUsers); err != nil {
		return nil, fmt.Errorf("get distribution attributed user stats: %w", err)
	}
	return stats, nil
}

func distributionAnalyticsRankingQuery(memberFilterCondition string) string {
	return fmt.Sprintf(`
WITH attributed_users AS (
	SELECT
		ua.referrer_member_id AS member_id,
		COUNT(*)::bigint AS registered_users
	FROM user_attributions ua
	WHERE ua.referrer_member_id IS NOT NULL
	  AND ua.bound_at >= $2
	  AND ua.bound_at < $3
	GROUP BY ua.referrer_member_id
),
recharge_stats AS (
	SELECT
		ua.referrer_member_id AS member_id,
		COALESCE(SUM(po.amount), 0)::double precision AS recharge_amount
	FROM payment_orders po
	JOIN user_attributions ua ON ua.user_id = po.user_id
	WHERE ua.referrer_member_id IS NOT NULL
	  AND po.order_type = $4
	  AND po.status = ANY($5)
	  AND COALESCE(po.completed_at, po.paid_at, po.created_at) >= $2
	  AND COALESCE(po.completed_at, po.paid_at, po.created_at) < $3
	GROUP BY ua.referrer_member_id
),
consumption_stats AS (
	SELECT
		ua.referrer_member_id AS member_id,
		COALESCE(SUM(COALESCE(NULLIF(ul.account_stats_cost, 0), NULLIF(ul.actual_cost, 0), ul.total_cost)), 0)::double precision AS consumption_amount
	FROM usage_logs ul
	JOIN user_attributions ua ON ua.user_id = ul.user_id
	WHERE ua.referrer_member_id IS NOT NULL
	  AND ul.created_at >= $2
	  AND ul.created_at < $3
	GROUP BY ua.referrer_member_id
),
commission_stats AS (
	SELECT
		cl.member_id,
		COALESCE(SUM(CASE WHEN cl.status NOT IN ('reversed', 'cancelled') THEN cl.amount ELSE 0 END), 0)::double precision AS commission_amount,
		COALESCE(SUM(CASE WHEN cl.status = 'settled' THEN cl.amount ELSE 0 END), 0)::double precision AS settled_commission_amount
	FROM commission_ledger cl
	WHERE cl.created_at >= $2
	  AND cl.created_at < $3
	GROUP BY cl.member_id
)
SELECT
	m.id,
	m.user_id,
	COALESCE(u.email, ''),
	COALESCE(u.username, ''),
	m.role_type,
	COALESCE(attributed_users.registered_users, 0),
	COALESCE(recharge_stats.recharge_amount, 0),
	COALESCE(consumption_stats.consumption_amount, 0),
	COALESCE(commission_stats.commission_amount, 0),
	COALESCE(commission_stats.settled_commission_amount, 0)
FROM channel_members m
LEFT JOIN users u ON u.id = m.user_id
LEFT JOIN attributed_users ON attributed_users.member_id = m.id
LEFT JOIN recharge_stats ON recharge_stats.member_id = m.id
LEFT JOIN consumption_stats ON consumption_stats.member_id = m.id
LEFT JOIN commission_stats ON commission_stats.member_id = m.id
WHERE %s
  AND m.status = 'active'
  AND m.role_type IN ('agent', 'kol1', 'kol2')
ORDER BY
	COALESCE(consumption_stats.consumption_amount, 0) DESC,
	COALESCE(recharge_stats.recharge_amount, 0) DESC,
	COALESCE(attributed_users.registered_users, 0) DESC,
	m.id ASC
LIMIT $6
`, memberFilterCondition)
}

func getDistributionAnalyticsOrganization(ctx context.Context, db *sql.DB, channelOrgID int64) (*service.DistributionOrganization, error) {
	if db == nil || channelOrgID <= 0 {
		return nil, nil
	}
	rows, err := db.QueryContext(ctx, `
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
		return nil, fmt.Errorf("get distribution analytics organization: %w", err)
	}
	defer func() { _ = rows.Close() }()
	org, err := scanDistributionOrganization(rows)
	if err != nil {
		return nil, err
	}
	return org, nil
}

func distributionAnalyticsCommissionUpperRatio(config map[string]any) float64 {
	if config == nil {
		return service.DistributionCommissionUpperRatioDefault / 100
	}
	raw, ok := config["commission_upper_ratio"]
	if !ok {
		raw = config["total_commission_ratio"]
	}
	if raw == nil {
		raw = config["commission_limit_ratio"]
	}
	value := 0.0
	switch typed := raw.(type) {
	case float64:
		value = typed
	case float32:
		value = float64(typed)
	case int:
		value = float64(typed)
	case int64:
		value = float64(typed)
	case string:
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64); err == nil {
			value = parsed
		}
	}
	if value <= 0 {
		return service.DistributionCommissionUpperRatioDefault / 100
	}
	if value > 1 && value <= 100 {
		value = value / 100
	}
	if value > 1 {
		return service.DistributionCommissionUpperRatioDefault / 100
	}
	return value
}

func scanDistributionAnalyticsRankingRows(rows *sql.Rows, contextLabel string) ([]service.DistributionAnalyticsRankingItem, error) {
	items := make([]service.DistributionAnalyticsRankingItem, 0)
	for rows.Next() {
		var item service.DistributionAnalyticsRankingItem
		if err := rows.Scan(
			&item.MemberID,
			&item.UserID,
			&item.UserEmail,
			&item.Username,
			&item.RoleType,
			&item.RegisteredUsers,
			&item.RechargeAmount,
			&item.ConsumptionAmount,
			&item.CommissionAmount,
			&item.SettledCommissionAmount,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", contextLabel, err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", contextLabel, err)
	}
	return items, nil
}

func validateDistributionAnalyticsChannel(channelOrgID int64, db *sql.DB) error {
	if channelOrgID <= 0 || db == nil {
		return service.ErrInvalidDistributionStats
	}
	return nil
}

func validateDistributionAnalyticsMemberIDs(memberIDs []int64, db *sql.DB) error {
	if db == nil || len(memberIDs) == 0 {
		return service.ErrInvalidDistributionStats
	}
	for _, memberID := range memberIDs {
		if memberID <= 0 {
			return service.ErrInvalidDistributionStats
		}
	}
	return nil
}

func distributionAnalyticsDateTruncUnit(granularity string) string {
	switch strings.ToLower(strings.TrimSpace(granularity)) {
	case "hour":
		return "hour"
	case "week":
		return "week"
	case "month":
		return "month"
	default:
		return "day"
	}
}

func distributionRechargeEventTimeExpression() string {
	return "COALESCE(po.completed_at, po.paid_at, po.created_at)"
}
