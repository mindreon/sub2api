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

type distributionWalletRequestRepository struct {
	db *sql.DB
}

type distributionWalletRequestMutationBuilder func(current *service.DistributionWallet, request *service.DistributionWalletRequest) (*distributionWalletMutationState, error)

func NewDistributionWalletRequestRepository(_ *dbent.Client, db *sql.DB) *distributionWalletRequestRepository {
	return &distributionWalletRequestRepository{db: db}
}

func (r *distributionWalletRequestRepository) ListRequests(ctx context.Context, filter service.DistributionWalletRequestListFilter, params pagination.PaginationParams) ([]service.DistributionWalletRequest, *pagination.PaginationResult, error) {
	if r.db == nil {
		return nil, nil, service.ErrInvalidDistributionWalletRequest
	}
	filter.RequestType = strings.ToLower(strings.TrimSpace(filter.RequestType))
	if filter.RequestType != "recharge" && filter.RequestType != "refund" {
		filter.RequestType = ""
	}
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	if filter.Status != "pending" && filter.Status != "approved" && filter.Status != "rejected" {
		filter.Status = ""
	}
	if filter.ChannelOrgID < 0 {
		filter.ChannelOrgID = 0
	}
	whereSQL, args := buildDistributionWalletRequestWhere(filter)

	var total int64
	countQuery := `
SELECT COUNT(*)
FROM channel_wallet_requests r
` + whereSQL
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count distribution wallet requests: %w", err)
	}

	args = append(args, params.Limit(), params.Offset())
	rows, err := r.db.QueryContext(ctx, `
SELECT r.id,
       r.channel_org_id,
       o.name,
       o.type,
       r.request_type,
       r.amount,
       r.reference_no,
       r.note,
       r.status,
       r.created_by_user_id,
       COALESCE(creator.email, ''),
       COALESCE(creator.username, ''),
       r.reviewed_by_user_id,
       COALESCE(reviewer.email, ''),
       COALESCE(reviewer.username, ''),
       r.review_note,
       r.reviewed_at,
       r.created_at
FROM channel_wallet_requests r
JOIN channel_organizations o ON o.id = r.channel_org_id
JOIN users creator ON creator.id = r.created_by_user_id
LEFT JOIN users reviewer ON reviewer.id = r.reviewed_by_user_id
`+whereSQL+`
ORDER BY r.created_at DESC, r.id DESC
LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args)),
		args...,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list distribution wallet requests: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionWalletRequests(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *distributionWalletRequestRepository) CreateRequest(ctx context.Context, input service.DistributionWalletRequestCreateInput) (*service.DistributionWalletRequest, error) {
	if r.db == nil || input.ChannelOrgID <= 0 || input.CreatedByUserID <= 0 || input.Amount <= 0 {
		return nil, service.ErrInvalidDistributionWalletRequest
	}
	requestType := strings.ToLower(strings.TrimSpace(input.RequestType))
	switch requestType {
	case "recharge", "refund":
	default:
		requestType = ""
	}
	if requestType == "" {
		return nil, service.ErrInvalidDistributionWalletRequest
	}
	rows, err := r.db.QueryContext(ctx, `
INSERT INTO channel_wallet_requests (
    channel_org_id,
    request_type,
    amount,
    reference_no,
    note,
    status,
    created_by_user_id,
    created_at
)
VALUES ($1, $2, $3, $4, $5, 'pending', $6, NOW())
RETURNING id`,
		input.ChannelOrgID,
		requestType,
		input.Amount,
		strings.TrimSpace(input.ReferenceNo),
		strings.TrimSpace(input.Note),
		input.CreatedByUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("create distribution wallet request: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var requestID int64
	if !rows.Next() {
		return nil, service.ErrInvalidDistributionWalletRequest
	}
	if err := rows.Scan(&requestID); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return r.GetRequestByID(ctx, requestID)
}

func (r *distributionWalletRequestRepository) GetRequestByID(ctx context.Context, requestID int64) (*service.DistributionWalletRequest, error) {
	if r.db == nil || requestID <= 0 {
		return nil, service.ErrInvalidDistributionWalletRequest
	}
	return getDistributionWalletRequestByID(ctx, r.db, requestID, false)
}

func (r *distributionWalletRequestRepository) ApproveRechargeRequest(ctx context.Context, requestID int64, input service.DistributionWalletRequestReviewInput, walletInput service.DistributionWalletRechargeInput) (*service.DistributionWalletRequest, error) {
	if r.db == nil {
		return nil, service.ErrInvalidDistributionWalletRequest
	}
	return r.approveRequestWithWalletMutation(ctx, requestID, input, walletInput.Amount, func(current *service.DistributionWallet, request *service.DistributionWalletRequest) (*distributionWalletMutationState, error) {
		if current == nil || request == nil {
			return nil, service.ErrInvalidDistributionWalletRequest
		}
		return &distributionWalletMutationState{
			transactionType: "recharge",
			prepaidBalance:  current.PrepaidBalance + request.Amount,
			reservedBalance: current.CommissionReserved,
			totalRecharged:  current.TotalRecharged + request.Amount,
			totalConsumed:   current.TotalConsumed,
			referenceNo:     walletInput.ReferenceNo,
			note:            walletInput.Note,
			operatorUserID:  walletInput.OperatorUserID,
		}, nil
	}, "recharge")
}

func (r *distributionWalletRequestRepository) ApproveRefundRequest(ctx context.Context, requestID int64, input service.DistributionWalletRequestReviewInput, walletInput service.DistributionWalletRefundInput) (*service.DistributionWalletRequest, error) {
	if r.db == nil {
		return nil, service.ErrInvalidDistributionWalletRequest
	}
	return r.approveRequestWithWalletMutation(ctx, requestID, input, walletInput.Amount, func(current *service.DistributionWallet, request *service.DistributionWalletRequest) (*distributionWalletMutationState, error) {
		if current == nil || request == nil {
			return nil, service.ErrInvalidDistributionWalletRequest
		}
		if current.PrepaidBalance-current.CommissionReserved < request.Amount {
			return nil, service.ErrDistributionWalletInsufficientBalance
		}
		return &distributionWalletMutationState{
			transactionType: "refund",
			prepaidBalance:  current.PrepaidBalance - request.Amount,
			reservedBalance: current.CommissionReserved,
			totalRecharged:  current.TotalRecharged,
			totalConsumed:   current.TotalConsumed,
			referenceNo:     walletInput.ReferenceNo,
			note:            walletInput.Note,
			operatorUserID:  walletInput.OperatorUserID,
		}, nil
	}, "refund")
}

func (r *distributionWalletRequestRepository) RejectRequest(ctx context.Context, requestID int64, input service.DistributionWalletRequestReviewInput) (*service.DistributionWalletRequest, error) {
	if r.db == nil || requestID <= 0 || input.ReviewedByUserID <= 0 {
		return nil, service.ErrInvalidDistributionWalletRequest
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin reject wallet request transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	request, err := getDistributionWalletRequestByID(ctx, tx, requestID, true)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(strings.TrimSpace(request.Status), "pending") {
		return nil, service.ErrInvalidDistributionWalletRequest
	}
	out, err := updateDistributionWalletRequestStatus(ctx, tx, requestID, "rejected", input.ReviewNote, input.ReviewedByUserID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit reject wallet request transaction: %w", err)
	}
	return out, nil
}

func (r *distributionWalletRequestRepository) approveRequestWithWalletMutation(
	ctx context.Context,
	requestID int64,
	input service.DistributionWalletRequestReviewInput,
	amount float64,
	build distributionWalletRequestMutationBuilder,
	expectedType string,
) (*service.DistributionWalletRequest, error) {
	if requestID <= 0 || input.ReviewedByUserID <= 0 || amount <= 0 {
		return nil, service.ErrInvalidDistributionWalletRequest
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin approve wallet request transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	request, err := getDistributionWalletRequestByID(ctx, tx, requestID, true)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(strings.TrimSpace(request.Status), "pending") || !strings.EqualFold(strings.TrimSpace(request.RequestType), expectedType) {
		return nil, service.ErrInvalidDistributionWalletRequest
	}
	current, err := getDistributionWalletByChannelOrgID(ctx, tx, request.ChannelOrgID, true)
	if err != nil {
		return nil, err
	}
	org, err := getDistributionOrganizationByID(ctx, tx, request.ChannelOrgID)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, service.ErrInvalidDistributionOrganization
	}
	next, err := build(current, request)
	if err != nil {
		return nil, err
	}
	next.status = resolveDistributionWalletMutationStatus(current, next, org)
	updated, err := updateDistributionWalletState(ctx, tx, request.ChannelOrgID, next)
	if err != nil {
		return nil, err
	}
	if err := syncDistributionAlertEvents(ctx, tx, org, updated); err != nil {
		return nil, err
	}
	if err := insertDistributionWalletTransaction(ctx, tx, request.ChannelOrgID, request.Amount, current, updated, next); err != nil {
		return nil, err
	}
	out, err := updateDistributionWalletRequestStatus(ctx, tx, requestID, "approved", input.ReviewNote, input.ReviewedByUserID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit approve wallet request transaction: %w", err)
	}
	return out, nil
}

func buildDistributionWalletRequestWhere(filter service.DistributionWalletRequestListFilter) (string, []any) {
	whereSQL := "WHERE 1=1"
	args := make([]any, 0, 4)
	if filter.ChannelOrgID > 0 {
		args = append(args, filter.ChannelOrgID)
		whereSQL += fmt.Sprintf(" AND r.channel_org_id = $%d", len(args))
	}
	if filter.RequestType != "" {
		args = append(args, filter.RequestType)
		whereSQL += fmt.Sprintf(" AND r.request_type = $%d", len(args))
	}
	if filter.Status != "" {
		args = append(args, filter.Status)
		whereSQL += fmt.Sprintf(" AND r.status = $%d", len(args))
	}
	return whereSQL, args
}

type distributionWalletRequestQueryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func getDistributionWalletRequestByID(ctx context.Context, q distributionWalletRequestQueryer, requestID int64, forUpdate bool) (*service.DistributionWalletRequest, error) {
	rows, err := q.QueryContext(ctx, `
SELECT r.id,
       r.channel_org_id,
       o.name,
       o.type,
       r.request_type,
       r.amount,
       r.reference_no,
       r.note,
       r.status,
       r.created_by_user_id,
       COALESCE(creator.email, ''),
       COALESCE(creator.username, ''),
       r.reviewed_by_user_id,
       COALESCE(reviewer.email, ''),
       COALESCE(reviewer.username, ''),
       r.review_note,
       r.reviewed_at,
       r.created_at
FROM channel_wallet_requests r
JOIN channel_organizations o ON o.id = r.channel_org_id
JOIN users creator ON creator.id = r.created_by_user_id
LEFT JOIN users reviewer ON reviewer.id = r.reviewed_by_user_id
WHERE r.id = $1`+distributionWalletForUpdateSQL(forUpdate), requestID)
	if err != nil {
		return nil, fmt.Errorf("get distribution wallet request: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionWalletRequests(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, service.ErrInvalidDistributionWalletRequest
	}
	return &items[0], nil
}

func updateDistributionWalletRequestStatus(ctx context.Context, tx *sql.Tx, requestID int64, status string, reviewNote string, reviewedByUserID int64) (*service.DistributionWalletRequest, error) {
	rows, err := tx.QueryContext(ctx, `
UPDATE channel_wallet_requests r
SET status = $1,
    review_note = $2,
    reviewed_by_user_id = $3,
    reviewed_at = NOW()
WHERE r.id = $4
RETURNING r.id,
          r.channel_org_id,
          (SELECT name FROM channel_organizations WHERE id = r.channel_org_id),
          (SELECT type FROM channel_organizations WHERE id = r.channel_org_id),
          r.request_type,
          r.amount,
          r.reference_no,
          r.note,
          r.status,
          r.created_by_user_id,
          COALESCE((SELECT email FROM users WHERE id = r.created_by_user_id), ''),
          COALESCE((SELECT username FROM users WHERE id = r.created_by_user_id), ''),
          r.reviewed_by_user_id,
          COALESCE((SELECT email FROM users WHERE id = r.reviewed_by_user_id), ''),
          COALESCE((SELECT username FROM users WHERE id = r.reviewed_by_user_id), ''),
          r.review_note,
          r.reviewed_at,
          r.created_at`,
		status,
		strings.TrimSpace(reviewNote),
		reviewedByUserID,
		requestID,
	)
	if err != nil {
		return nil, fmt.Errorf("update distribution wallet request status: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanDistributionWalletRequests(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, service.ErrInvalidDistributionWalletRequest
	}
	return &items[0], nil
}

func scanDistributionWalletRequests(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}) ([]service.DistributionWalletRequest, error) {
	items := make([]service.DistributionWalletRequest, 0)
	for rows.Next() {
		var item service.DistributionWalletRequest
		if err := rows.Scan(
			&item.ID,
			&item.ChannelOrgID,
			&item.OrganizationName,
			&item.OrganizationType,
			&item.RequestType,
			&item.Amount,
			&item.ReferenceNo,
			&item.Note,
			&item.Status,
			&item.CreatedByUserID,
			&item.CreatedByUserEmail,
			&item.CreatedByUsername,
			&item.ReviewedByUserID,
			&item.ReviewedByUserEmail,
			&item.ReviewedByUsername,
			&item.ReviewNote,
			&item.ReviewedAt,
			&item.CreatedAt,
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
