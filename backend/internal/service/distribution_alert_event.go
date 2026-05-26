package service

import (
	"context"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrInvalidDistributionAlertEvent = infraerrors.BadRequest(
		"INVALID_DISTRIBUTION_ALERT_EVENT",
		"invalid distribution alert event",
	)
)

type DistributionAlertEvent struct {
	ID               int64          `json:"id"`
	ChannelOrgID     int64          `json:"channel_org_id"`
	OrganizationName string         `json:"organization_name"`
	OrganizationType string         `json:"organization_type"`
	AlertType        string         `json:"alert_type"`
	Severity         string         `json:"severity"`
	Status           string         `json:"status"`
	Details          map[string]any `json:"details"`
	TriggeredAt      time.Time      `json:"triggered_at"`
	ResolvedAt       *time.Time     `json:"resolved_at,omitempty"`
	LastObservedAt   time.Time      `json:"last_observed_at"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type DistributionAlertEventListFilter struct {
	ChannelOrgID int64
	AlertType    string
	Severity     string
	Status       string
}

func (f DistributionAlertEventListFilter) normalized() DistributionAlertEventListFilter {
	f.AlertType = normalizeDistributionAlertType(f.AlertType)
	f.Severity = normalizeDistributionAlertSeverity(f.Severity)
	f.Status = normalizeDistributionAlertStatus(f.Status)
	if f.ChannelOrgID < 0 {
		f.ChannelOrgID = 0
	}
	return f
}

type DistributionAlertEventRepository interface {
	ListAlertEvents(ctx context.Context, filter DistributionAlertEventListFilter, params pagination.PaginationParams) ([]DistributionAlertEvent, *pagination.PaginationResult, error)
}

func normalizeDistributionAlertType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "low_balance", "balance_exhausted", "consumption_warning", "consumption_exhausted":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func normalizeDistributionAlertSeverity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "info", "warning", "critical":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func normalizeDistributionAlertStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "active", "resolved":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}
