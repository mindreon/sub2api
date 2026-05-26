package service

import (
	"context"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrInvalidDistributionWalletRequest = infraerrors.BadRequest(
		"INVALID_DISTRIBUTION_WALLET_REQUEST",
		"invalid distribution wallet request",
	)
)

type DistributionWalletRequest struct {
	ID                  int64      `json:"id"`
	ChannelOrgID        int64      `json:"channel_org_id"`
	OrganizationName    string     `json:"organization_name"`
	OrganizationType    string     `json:"organization_type"`
	RequestType         string     `json:"request_type"`
	Amount              float64    `json:"amount"`
	ReferenceNo         string     `json:"reference_no"`
	Note                string     `json:"note"`
	Status              string     `json:"status"`
	CreatedByUserID     int64      `json:"created_by_user_id"`
	CreatedByUserEmail  string     `json:"created_by_user_email"`
	CreatedByUsername   string     `json:"created_by_username"`
	ReviewedByUserID    *int64     `json:"reviewed_by_user_id,omitempty"`
	ReviewedByUserEmail string     `json:"reviewed_by_user_email"`
	ReviewedByUsername  string     `json:"reviewed_by_username"`
	ReviewNote          string     `json:"review_note"`
	ReviewedAt          *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

type DistributionWalletRequestListFilter struct {
	ChannelOrgID int64
	RequestType  string
	Status       string
}

func (f DistributionWalletRequestListFilter) normalized() DistributionWalletRequestListFilter {
	f.RequestType = normalizeDistributionWalletRequestType(f.RequestType)
	f.Status = normalizeDistributionWalletRequestStatus(f.Status)
	if f.ChannelOrgID < 0 {
		f.ChannelOrgID = 0
	}
	return f
}

type DistributionWalletRequestCreateInput struct {
	ChannelOrgID    int64
	RequestType     string
	Amount          float64
	ReferenceNo     string
	Note            string
	CreatedByUserID int64
}

type DistributionWalletRequestReviewInput struct {
	Action           string
	ReviewNote       string
	ReviewedByUserID int64
}

type DistributionWalletRequestRepository interface {
	ListRequests(ctx context.Context, filter DistributionWalletRequestListFilter, params pagination.PaginationParams) ([]DistributionWalletRequest, *pagination.PaginationResult, error)
	CreateRequest(ctx context.Context, input DistributionWalletRequestCreateInput) (*DistributionWalletRequest, error)
	GetRequestByID(ctx context.Context, requestID int64) (*DistributionWalletRequest, error)
	ApproveRechargeRequest(ctx context.Context, requestID int64, input DistributionWalletRequestReviewInput, walletInput DistributionWalletRechargeInput) (*DistributionWalletRequest, error)
	ApproveRefundRequest(ctx context.Context, requestID int64, input DistributionWalletRequestReviewInput, walletInput DistributionWalletRefundInput) (*DistributionWalletRequest, error)
	RejectRequest(ctx context.Context, requestID int64, input DistributionWalletRequestReviewInput) (*DistributionWalletRequest, error)
}

func normalizeDistributionWalletRequestType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "recharge", "refund":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func normalizeDistributionWalletRequestStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "pending", "approved", "rejected":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func normalizeDistributionWalletRequestAction(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "approve", "reject":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}
