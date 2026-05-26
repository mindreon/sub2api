package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type distributionWalletRequestRepoStub struct {
	listFilter           DistributionWalletRequestListFilter
	listParams           pagination.PaginationParams
	createInput          *DistributionWalletRequestCreateInput
	rejectInput          *DistributionWalletRequestReviewInput
	approveRechargeInput *DistributionWalletRechargeInput
	approveRefundInput   *DistributionWalletRefundInput
	getByIDRequest       *DistributionWalletRequest
	listItems            []DistributionWalletRequest
}

func (s *distributionWalletRequestRepoStub) ListRequests(ctx context.Context, filter DistributionWalletRequestListFilter, params pagination.PaginationParams) ([]DistributionWalletRequest, *pagination.PaginationResult, error) {
	s.listFilter = filter
	s.listParams = params
	if s.listItems != nil {
		return s.listItems, &pagination.PaginationResult{Total: int64(len(s.listItems)), Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
	}
	return []DistributionWalletRequest{{
		ID:          1,
		ChannelOrgID: filter.ChannelOrgID,
		RequestType: filter.RequestType,
		Status:      filter.Status,
		Amount:      88,
		CreatedAt:   time.Now().UTC(),
	}}, &pagination.PaginationResult{Total: 1, Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func (s *distributionWalletRequestRepoStub) CreateRequest(ctx context.Context, input DistributionWalletRequestCreateInput) (*DistributionWalletRequest, error) {
	s.createInput = &input
	return &DistributionWalletRequest{
		ID:              9,
		ChannelOrgID:    input.ChannelOrgID,
		RequestType:     input.RequestType,
		Status:          "pending",
		Amount:          input.Amount,
		ReferenceNo:     input.ReferenceNo,
		Note:            input.Note,
		CreatedByUserID: input.CreatedByUserID,
		CreatedAt:       time.Now().UTC(),
	}, nil
}

func (s *distributionWalletRequestRepoStub) GetRequestByID(ctx context.Context, requestID int64) (*DistributionWalletRequest, error) {
	if s.getByIDRequest != nil {
		return s.getByIDRequest, nil
	}
	return &DistributionWalletRequest{
		ID:              requestID,
		ChannelOrgID:    88,
		RequestType:     "recharge",
		Status:          "pending",
		Amount:          120,
		ReferenceNo:     "BANK-1",
		Note:            "bank transfer",
		CreatedByUserID: 7,
		CreatedAt:       time.Now().UTC(),
	}, nil
}

func (s *distributionWalletRequestRepoStub) ApproveRechargeRequest(ctx context.Context, requestID int64, input DistributionWalletRequestReviewInput, walletInput DistributionWalletRechargeInput) (*DistributionWalletRequest, error) {
	s.approveRechargeInput = &walletInput
	return &DistributionWalletRequest{
		ID:               requestID,
		ChannelOrgID:     88,
		RequestType:      "recharge",
		Status:           "approved",
		Amount:           walletInput.Amount,
		ReferenceNo:      walletInput.ReferenceNo,
		Note:             walletInput.Note,
		ReviewedByUserID: &input.ReviewedByUserID,
		ReviewNote:       input.ReviewNote,
		ReviewedAt:       distributionWalletRequestPtrTime(time.Now().UTC()),
		CreatedAt:        time.Now().UTC(),
	}, nil
}

func (s *distributionWalletRequestRepoStub) ApproveRefundRequest(ctx context.Context, requestID int64, input DistributionWalletRequestReviewInput, walletInput DistributionWalletRefundInput) (*DistributionWalletRequest, error) {
	s.approveRefundInput = &walletInput
	return &DistributionWalletRequest{
		ID:               requestID,
		ChannelOrgID:     88,
		RequestType:      "refund",
		Status:           "approved",
		Amount:           walletInput.Amount,
		ReferenceNo:      walletInput.ReferenceNo,
		Note:             walletInput.Note,
		ReviewedByUserID: &input.ReviewedByUserID,
		ReviewNote:       input.ReviewNote,
		ReviewedAt:       distributionWalletRequestPtrTime(time.Now().UTC()),
		CreatedAt:        time.Now().UTC(),
	}, nil
}

func (s *distributionWalletRequestRepoStub) RejectRequest(ctx context.Context, requestID int64, input DistributionWalletRequestReviewInput) (*DistributionWalletRequest, error) {
	s.rejectInput = &input
	return &DistributionWalletRequest{
		ID:               requestID,
		ChannelOrgID:     88,
		RequestType:      "refund",
		Status:           "rejected",
		Amount:           20,
		ReviewedByUserID: &input.ReviewedByUserID,
		ReviewNote:       input.ReviewNote,
		ReviewedAt:       distributionWalletRequestPtrTime(time.Now().UTC()),
		CreatedAt:        time.Now().UTC(),
	}, nil
}

func distributionWalletRequestPtrTime(value time.Time) *time.Time {
	return &value
}

func TestDistributionAdminService_ListAndCreateWalletRequests(t *testing.T) {
	requestRepo := &distributionWalletRequestRepoStub{}
	svc := NewDistributionAdminService(nil, nil, nil, nil, nil, nil, nil)
	svc.SetWalletRequestRepository(requestRepo)

	items, page, err := svc.ListWalletRequests(context.Background(), DistributionWalletRequestListFilter{
		ChannelOrgID: 88,
		RequestType:  "recharge",
		Status:       "pending",
	}, pagination.PaginationParams{Page: 2, PageSize: 10})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, int64(88), requestRepo.listFilter.ChannelOrgID)
	require.Equal(t, "recharge", requestRepo.listFilter.RequestType)
	require.Equal(t, "pending", requestRepo.listFilter.Status)
	require.Equal(t, 2, requestRepo.listParams.Page)

	out, err := svc.CreateWalletRequest(context.Background(), DistributionWalletRequestCreateInput{
		ChannelOrgID:    88,
		RequestType:     "refund",
		Amount:          50,
		ReferenceNo:     "RF-100",
		Note:            "return balance",
		CreatedByUserID: 7,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.NotNil(t, requestRepo.createInput)
	require.Equal(t, int64(7), requestRepo.createInput.CreatedByUserID)
	require.Equal(t, "pending", out.Status)
}

func TestDistributionAdminService_ApproveRechargeWalletRequest(t *testing.T) {
	requestRepo := &distributionWalletRequestRepoStub{
		getByIDRequest: &DistributionWalletRequest{
			ID:              31,
			ChannelOrgID:    88,
			RequestType:     "recharge",
			Status:          "pending",
			Amount:          120,
			ReferenceNo:     "BANK-1",
			Note:            "confirmed by bank",
			CreatedByUserID: 7,
			CreatedAt:       time.Now().UTC(),
		},
	}
	svc := NewDistributionAdminService(nil, nil, nil, nil, nil, nil, nil)
	svc.SetWalletRequestRepository(requestRepo)

	operatorUserID := int64(9)
	out, err := svc.ReviewWalletRequest(context.Background(), 31, DistributionWalletRequestReviewInput{
		Action:           "approve",
		ReviewNote:       "到账确认",
		ReviewedByUserID: operatorUserID,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "approved", out.Status)
	require.NotNil(t, requestRepo.approveRechargeInput)
	require.InDelta(t, 120, requestRepo.approveRechargeInput.Amount, 0.0001)
	require.Equal(t, "BANK-1", requestRepo.approveRechargeInput.ReferenceNo)
	require.Contains(t, requestRepo.approveRechargeInput.Note, "request_id=31")
}

func TestDistributionAdminService_ApproveRefundWalletRequestUsesConfiguredFee(t *testing.T) {
	orgRepo := &distributionAdminOrgRepoStub{
		item: &DistributionOrganization{
			ID:   88,
			Type: "reseller",
			Config: map[string]any{
				"refund_fee_rate": 0.1,
			},
		},
	}
	requestRepo := &distributionWalletRequestRepoStub{
		getByIDRequest: &DistributionWalletRequest{
			ID:              45,
			ChannelOrgID:    88,
			RequestType:     "refund",
			Status:          "pending",
			Amount:          100,
			ReferenceNo:     "RF-1",
			Note:            "return remaining balance",
			CreatedByUserID: 7,
			CreatedAt:       time.Now().UTC(),
		},
	}
	svc := NewDistributionAdminService(orgRepo, nil, nil, nil, nil, nil, nil)
	svc.SetWalletRequestRepository(requestRepo)

	operatorUserID := int64(9)
	out, err := svc.ReviewWalletRequest(context.Background(), 45, DistributionWalletRequestReviewInput{
		Action:           "approve",
		ReviewNote:       "同意退款",
		ReviewedByUserID: operatorUserID,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "approved", out.Status)
	require.NotNil(t, requestRepo.approveRefundInput)
	require.InDelta(t, 100, requestRepo.approveRefundInput.Amount, 0.0001)
	require.Equal(t, "RF-1", requestRepo.approveRefundInput.ReferenceNo)
	require.Contains(t, requestRepo.approveRefundInput.Note, "mock_refund")
	require.Contains(t, requestRepo.approveRefundInput.Note, "fee_amount=10")
	require.Contains(t, requestRepo.approveRefundInput.Note, "net_amount=90")
}

func TestDistributionAdminService_RejectWalletRequest(t *testing.T) {
	requestRepo := &distributionWalletRequestRepoStub{
		getByIDRequest: &DistributionWalletRequest{
			ID:              46,
			ChannelOrgID:    88,
			RequestType:     "refund",
			Status:          "pending",
			Amount:          20,
			CreatedByUserID: 7,
			CreatedAt:       time.Now().UTC(),
		},
	}
	svc := NewDistributionAdminService(nil, nil, nil, nil, nil, nil, nil)
	svc.SetWalletRequestRepository(requestRepo)

	operatorUserID := int64(9)
	out, err := svc.ReviewWalletRequest(context.Background(), 46, DistributionWalletRequestReviewInput{
		Action:           "reject",
		ReviewNote:       "资料不完整",
		ReviewedByUserID: operatorUserID,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "rejected", out.Status)
	require.NotNil(t, requestRepo.rejectInput)
	require.Equal(t, "资料不完整", requestRepo.rejectInput.ReviewNote)
	require.Nil(t, requestRepo.approveRechargeInput)
	require.Nil(t, requestRepo.approveRefundInput)
}

func TestDistributionUserManageService_CreateAndListWalletRequestsForManager(t *testing.T) {
	memberRepo := &distributionUserManageMemberRepoStub{
		byUser: map[int64][]DistributionMemberView{
			7: {{MemberID: 11, UserID: 7, ChannelOrgID: 88, RoleType: "manager", Status: "active"}},
		},
	}
	orgRepo := &distributionUserManageOrgRepoStub{
		orgByID: map[int64]*DistributionOrganization{
			88: {ID: 88, Type: "reseller", Name: "Channel A", Status: "active"},
		},
	}
	requestRepo := &distributionWalletRequestRepoStub{}
	adminSvc := NewDistributionAdminService(nil, nil, nil, nil, nil, nil, nil)
	adminSvc.SetWalletRequestRepository(requestRepo)
	svc := NewDistributionUserManageService(memberRepo, orgRepo, adminSvc, nil)

	created, err := svc.CreateWalletRequestForUser(context.Background(), 7, DistributionWalletRequestCreateInput{
		RequestType: "recharge",
		Amount:      300,
		ReferenceNo: "CRYPTO-1",
		Note:        "USDT transfer",
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotNil(t, requestRepo.createInput)
	require.Equal(t, int64(88), requestRepo.createInput.ChannelOrgID)
	require.Equal(t, int64(7), requestRepo.createInput.CreatedByUserID)

	items, page, err := svc.ListWalletRequestsForUser(context.Background(), 7, DistributionWalletRequestListFilter{
		RequestType: "recharge",
		Status:      "pending",
	}, pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, int64(88), requestRepo.listFilter.ChannelOrgID)
	require.Equal(t, "recharge", requestRepo.listFilter.RequestType)
	require.Equal(t, "pending", requestRepo.listFilter.Status)
}
