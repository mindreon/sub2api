package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type distributionAlertEventRepository struct {
	db *sql.DB
}

type distributionAlertRuleState struct {
	alertType string
	severity  string
	details   map[string]any
}

type distributionActiveAlertRow struct {
	ID        int64
	AlertType string
	Severity  string
	Status    string
	Details   map[string]any
}

func NewDistributionAlertEventRepository(_ *dbent.Client, db *sql.DB) *distributionAlertEventRepository {
	return &distributionAlertEventRepository{db: db}
}

func (r *distributionAlertEventRepository) ListAlertEvents(ctx context.Context, filter service.DistributionAlertEventListFilter, params pagination.PaginationParams) ([]service.DistributionAlertEvent, *pagination.PaginationResult, error) {
	if r.db == nil {
		return nil, nil, service.ErrInvalidDistributionAlertEvent
	}
	switch strings.ToLower(strings.TrimSpace(filter.AlertType)) {
	case "low_balance", "balance_exhausted", "consumption_warning", "consumption_exhausted":
		filter.AlertType = strings.ToLower(strings.TrimSpace(filter.AlertType))
	default:
		filter.AlertType = ""
	}
	switch strings.ToLower(strings.TrimSpace(filter.Severity)) {
	case "info", "warning", "critical":
		filter.Severity = strings.ToLower(strings.TrimSpace(filter.Severity))
	default:
		filter.Severity = ""
	}
	switch strings.ToLower(strings.TrimSpace(filter.Status)) {
	case "active", "resolved":
		filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	default:
		filter.Status = ""
	}
	if filter.ChannelOrgID < 0 {
		filter.ChannelOrgID = 0
	}
	whereSQL, args := buildDistributionAlertEventWhere(filter)

	var total int64
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM channel_alert_events e
`+whereSQL, args...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count distribution alert events: %w", err)
	}

	args = append(args, params.Limit(), params.Offset())
	rows, err := r.db.QueryContext(ctx, `
SELECT e.id,
       e.channel_org_id,
       o.name,
       o.type,
       e.alert_type,
       e.severity,
       e.status,
       e.details,
       e.triggered_at,
       e.resolved_at,
       e.last_observed_at,
       e.created_at,
       e.updated_at
FROM channel_alert_events e
JOIN channel_organizations o ON o.id = e.channel_org_id
`+whereSQL+`
ORDER BY e.triggered_at DESC, e.id DESC
LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args)), args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list distribution alert events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionAlertEvents(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func buildDistributionAlertEventWhere(filter service.DistributionAlertEventListFilter) (string, []any) {
	whereSQL := "WHERE 1=1"
	args := make([]any, 0, 4)
	if filter.ChannelOrgID > 0 {
		args = append(args, filter.ChannelOrgID)
		whereSQL += fmt.Sprintf(" AND e.channel_org_id = $%d", len(args))
	}
	if filter.AlertType != "" {
		args = append(args, filter.AlertType)
		whereSQL += fmt.Sprintf(" AND e.alert_type = $%d", len(args))
	}
	if filter.Severity != "" {
		args = append(args, filter.Severity)
		whereSQL += fmt.Sprintf(" AND e.severity = $%d", len(args))
	}
	if filter.Status != "" {
		args = append(args, filter.Status)
		whereSQL += fmt.Sprintf(" AND e.status = $%d", len(args))
	}
	return whereSQL, args
}

func syncDistributionAlertEvents(ctx context.Context, tx *sql.Tx, org *service.DistributionOrganization, wallet *service.DistributionWallet) error {
	if tx == nil || org == nil || wallet == nil || wallet.ChannelOrgID <= 0 {
		return nil
	}
	desired := evaluateDistributionAlertStates(org, wallet)

	existing, err := listActiveDistributionAlertRows(ctx, tx, wallet.ChannelOrgID)
	if err != nil {
		return err
	}

	existingByType := make(map[string]distributionActiveAlertRow, len(existing))
	for _, item := range existing {
		existingByType[item.AlertType] = item
	}

	for _, state := range desired {
		current, ok := existingByType[state.alertType]
		if ok {
			if err := updateDistributionAlertEvent(ctx, tx, current.ID, state); err != nil {
				return err
			}
			delete(existingByType, state.alertType)
			continue
		}
		if err := insertDistributionAlertEvent(ctx, tx, wallet.ChannelOrgID, state); err != nil {
			return err
		}
	}

	for _, stale := range existingByType {
		if err := resolveDistributionAlertEvent(ctx, tx, stale.ID); err != nil {
			return err
		}
	}
	return nil
}

func evaluateDistributionAlertStates(org *service.DistributionOrganization, wallet *service.DistributionWallet) []distributionAlertRuleState {
	if org == nil || wallet == nil {
		return nil
	}
	states := make([]distributionAlertRuleState, 0, 4)
	availableBalance := roundDistributionAlertValue(wallet.PrepaidBalance - wallet.CommissionReserved)
	rechargeLeadDays := distributionAlertConfigInt(org.Config, "recharge_lead_time_days")
	rechargeDeadlineNote := distributionAlertConfigString(org.Config, "recharge_deadline_note")

	if availableBalance <= 0 {
		states = append(states, distributionAlertRuleState{
			alertType: "balance_exhausted",
			severity:  "critical",
			details: map[string]any{
				"prepaid_balance":        wallet.PrepaidBalance,
				"commission_reserved":    wallet.CommissionReserved,
				"available_balance":      availableBalance,
				"recharge_lead_days":     rechargeLeadDays,
				"recharge_deadline_note": rechargeDeadlineNote,
			},
		})
	}

	if wallet.WarningThreshold > 0 && roundDistributionAlertValue(wallet.PrepaidBalance) <= roundDistributionAlertValue(wallet.WarningThreshold) && availableBalance > 0 {
		states = append(states, distributionAlertRuleState{
			alertType: "low_balance",
			severity:  "warning",
			details: map[string]any{
				"prepaid_balance":        wallet.PrepaidBalance,
				"warning_threshold":      wallet.WarningThreshold,
				"commission_reserved":    wallet.CommissionReserved,
				"available_balance":      availableBalance,
				"recharge_lead_days":     rechargeLeadDays,
				"recharge_deadline_note": rechargeDeadlineNote,
			},
		})
	}

	consumptionLimit := distributionWalletConsumptionLimitFromConfig(org.Config)
	consumptionWarningThreshold := distributionAlertConfigFloat(org.Config, "consumption_warning_threshold")
	if consumptionLimit > 0 {
		remaining := roundDistributionAlertValue(consumptionLimit - wallet.TotalConsumed)
		if roundDistributionAlertValue(wallet.TotalConsumed) >= roundDistributionAlertValue(consumptionLimit) {
			states = append(states, distributionAlertRuleState{
				alertType: "consumption_exhausted",
				severity:  "critical",
				details: map[string]any{
					"consumption_limit":      consumptionLimit,
					"total_consumed":         wallet.TotalConsumed,
					"remaining_consumption":  remaining,
					"warning_threshold":      consumptionWarningThreshold,
					"recharge_lead_days":     rechargeLeadDays,
					"recharge_deadline_note": rechargeDeadlineNote,
				},
			})
		} else if consumptionWarningThreshold > 0 && remaining <= roundDistributionAlertValue(consumptionWarningThreshold) {
			states = append(states, distributionAlertRuleState{
				alertType: "consumption_warning",
				severity:  "warning",
				details: map[string]any{
					"consumption_limit":      consumptionLimit,
					"total_consumed":         wallet.TotalConsumed,
					"remaining_consumption":  remaining,
					"warning_threshold":      consumptionWarningThreshold,
					"recharge_lead_days":     rechargeLeadDays,
					"recharge_deadline_note": rechargeDeadlineNote,
				},
			})
		}
	}

	return states
}

func listActiveDistributionAlertRows(ctx context.Context, tx *sql.Tx, channelOrgID int64) ([]distributionActiveAlertRow, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id, alert_type, severity, status, details
FROM channel_alert_events
WHERE channel_org_id = $1
  AND status = 'active'
FOR UPDATE`, channelOrgID)
	if err != nil {
		return nil, fmt.Errorf("list active distribution alert events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items := make([]distributionActiveAlertRow, 0)
	for rows.Next() {
		var (
			item        distributionActiveAlertRow
			detailsJSON []byte
		)
		if err := rows.Scan(&item.ID, &item.AlertType, &item.Severity, &item.Status, &detailsJSON); err != nil {
			return nil, err
		}
		if len(detailsJSON) > 0 {
			if err := json.Unmarshal(detailsJSON, &item.Details); err != nil {
				return nil, err
			}
		}
		if item.Details == nil {
			item.Details = map[string]any{}
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func insertDistributionAlertEvent(ctx context.Context, tx *sql.Tx, channelOrgID int64, state distributionAlertRuleState) error {
	detailsJSON, err := json.Marshal(state.details)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO channel_alert_events (
    channel_org_id,
    alert_type,
    severity,
    status,
    details,
    triggered_at,
    last_observed_at,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, 'active', $4, NOW(), NOW(), NOW(), NOW())`,
		channelOrgID,
		state.alertType,
		state.severity,
		detailsJSON,
	)
	if err != nil {
		return fmt.Errorf("insert distribution alert event: %w", err)
	}
	return nil
}

func updateDistributionAlertEvent(ctx context.Context, tx *sql.Tx, eventID int64, state distributionAlertRuleState) error {
	detailsJSON, err := json.Marshal(state.details)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
UPDATE channel_alert_events
SET severity = $2,
    details = $3,
    last_observed_at = NOW(),
    updated_at = NOW()
WHERE id = $1`, eventID, state.severity, detailsJSON)
	if err != nil {
		return fmt.Errorf("update distribution alert event: %w", err)
	}
	return nil
}

func resolveDistributionAlertEvent(ctx context.Context, tx *sql.Tx, eventID int64) error {
	_, err := tx.ExecContext(ctx, `
UPDATE channel_alert_events
SET status = 'resolved',
    resolved_at = NOW(),
    last_observed_at = NOW(),
    updated_at = NOW()
WHERE id = $1`, eventID)
	if err != nil {
		return fmt.Errorf("resolve distribution alert event: %w", err)
	}
	return nil
}

func scanDistributionAlertEvents(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}) ([]service.DistributionAlertEvent, error) {
	items := make([]service.DistributionAlertEvent, 0)
	for rows.Next() {
		var (
			item        service.DistributionAlertEvent
			resolvedAt  sql.NullTime
			detailsJSON []byte
		)
		if err := rows.Scan(
			&item.ID,
			&item.ChannelOrgID,
			&item.OrganizationName,
			&item.OrganizationType,
			&item.AlertType,
			&item.Severity,
			&item.Status,
			&detailsJSON,
			&item.TriggeredAt,
			&resolvedAt,
			&item.LastObservedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if len(detailsJSON) > 0 {
			if err := json.Unmarshal(detailsJSON, &item.Details); err != nil {
				return nil, err
			}
		}
		if item.Details == nil {
			item.Details = map[string]any{}
		}
		if resolvedAt.Valid {
			item.ResolvedAt = &resolvedAt.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func roundDistributionAlertValue(value float64) float64 {
	return math.Round(value*1e8) / 1e8
}

func distributionAlertConfigFloat(config map[string]any, key string) float64 {
	value, ok := config[key]
	if !ok {
		return 0
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
	case json.Number:
		if parsed, err := typed.Float64(); err == nil {
			return parsed
		}
	case string:
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64); err == nil {
			return parsed
		}
	}
	return 0
}

func distributionAlertConfigInt(config map[string]any, key string) int64 {
	return int64(distributionAlertConfigFloat(config, key))
}

func distributionAlertConfigString(config map[string]any, key string) string {
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
