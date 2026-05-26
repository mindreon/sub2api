package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type distributionCommissionAttributionRepoStub struct {
	attribution *DistributionAttribution
}

func (s *distributionCommissionAttributionRepoStub) GetByUserID(ctx context.Context, userID int64) (*DistributionAttribution, error) {
	return s.attribution, nil
}

type distributionCommissionMemberRepoStub struct {
	member       *DistributionMemberView
	memberByID   map[int64]*DistributionMemberView
	memberByRole map[string]*DistributionMemberView
}

func (s *distributionCommissionMemberRepoStub) GetByID(ctx context.Context, memberID int64) (*DistributionMemberView, error) {
	if s.memberByID != nil {
		if member, ok := s.memberByID[memberID]; ok {
			return member, nil
		}
	}
	return s.member, nil
}

func (s *distributionCommissionMemberRepoStub) GetByChannelOrgIDAndRole(ctx context.Context, channelOrgID int64, roleType string) (*DistributionMemberView, error) {
	if s.memberByRole != nil {
		if member, ok := s.memberByRole[roleType]; ok {
			return member, nil
		}
	}
	return s.member, nil
}

type distributionCommissionOrgRepoStub struct {
	org *DistributionOrganization
}

func (s *distributionCommissionOrgRepoStub) GetByID(ctx context.Context, id int64) (*DistributionOrganization, error) {
	return s.org, nil
}

func (s *distributionCommissionOrgRepoStub) GetByOwnerUserID(ctx context.Context, userID int64) (*DistributionOrganization, error) {
	return nil, nil
}

type distributionCommissionRepoStub struct {
	inputs          []DistributionCommissionInput
	totalByUserID   map[int64]float64
	totalByMemberID map[int64]float64
}

func (s *distributionCommissionRepoStub) Create(ctx context.Context, input DistributionCommissionInput) (*DistributionCommissionLedger, error) {
	s.inputs = append(s.inputs, input)
	now := time.Now().UTC()
	return &DistributionCommissionLedger{
		ID:               1001,
		ChannelOrgID:     input.ChannelOrgID,
		MemberID:         input.MemberID,
		UserID:           input.UserID,
		UsageLogID:       input.UsageLogID,
		CommissionType:   input.CommissionType,
		BaseAmount:       input.BaseAmount,
		Rate:             input.Rate,
		Amount:           input.Amount,
		Status:           input.Status,
		SettlementMethod: input.SettlementMethod,
		FrozenUntil:      input.FrozenUntil,
		SettledAt:        input.SettledAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func (s *distributionCommissionRepoStub) GetByID(ctx context.Context, commissionID int64) (*DistributionCommissionLedger, error) {
	return &DistributionCommissionLedger{ID: commissionID, Status: "available"}, nil
}

func (s *distributionCommissionRepoStub) HasCommissionTypeSince(ctx context.Context, channelOrgID int64, commissionType string, since time.Time) (bool, error) {
	return false, nil
}

func (s *distributionCommissionRepoStub) GetTotalCommissionByUserID(ctx context.Context, channelOrgID int64, userID int64) (float64, error) {
	if s.totalByUserID == nil {
		return 0, nil
	}
	return s.totalByUserID[userID], nil
}

func (s *distributionCommissionRepoStub) GetTotalCommissionByMemberID(ctx context.Context, channelOrgID int64, memberID int64) (float64, error) {
	if s.totalByMemberID == nil {
		return 0, nil
	}
	return s.totalByMemberID[memberID], nil
}

type distributionCommissionWalletRepoStub struct {
	reserved []distributionWalletMutationCall
}

type distributionWalletMutationCall struct {
	channelOrgID int64
	amount       float64
}

func (s *distributionCommissionWalletRepoStub) ReserveCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error) {
	s.reserved = append(s.reserved, distributionWalletMutationCall{channelOrgID: channelOrgID, amount: amount})
	return &DistributionWallet{ChannelOrgID: channelOrgID}, nil
}

func (s *distributionCommissionWalletRepoStub) ReleaseCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error) {
	return &DistributionWallet{ChannelOrgID: channelOrgID}, nil
}

func (s *distributionCommissionWalletRepoStub) SettleReservedCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error) {
	return &DistributionWallet{ChannelOrgID: channelOrgID}, nil
}

func (s *distributionCommissionWalletRepoStub) DeductCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error) {
	return &DistributionWallet{ChannelOrgID: channelOrgID}, nil
}

func (s *distributionCommissionWalletRepoStub) RefundCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error) {
	return &DistributionWallet{ChannelOrgID: channelOrgID}, nil
}

func (s *distributionCommissionWalletRepoStub) ConsumeUsage(ctx context.Context, channelOrgID int64, amount float64, consumptionLimit float64) (*DistributionWallet, error) {
	return &DistributionWallet{ChannelOrgID: channelOrgID}, nil
}

func TestDistributionCommissionService_UsesAccountStatsCostAsBaseAmount(t *testing.T) {
	base := 12.5
	accountStats := 8.75
	memberID := int64(42)

	svc := NewDistributionCommissionService(
		&distributionCommissionAttributionRepoStub{
			attribution: &DistributionAttribution{
				UserID:           7,
				ChannelOrgID:     99,
				ReferrerMemberID: &memberID,
			},
		},
		&distributionCommissionMemberRepoStub{
			member: &DistributionMemberView{
				MemberID:       memberID,
				ChannelOrgID:   99,
				CommissionRate: 0.15,
				Status:         "active",
			},
		},
		nil,
		nil,
		&distributionCommissionRepoStub{},
		nil,
		24*time.Hour,
	)

	out, err := svc.AccrueForUsageLog(context.Background(), &UsageLog{
		ID:               555,
		UserID:           7,
		TotalCost:        base,
		AccountStatsCost: &accountStats,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, accountStats, out.BaseAmount)
	require.InDelta(t, accountStats*0.15, out.Amount, 0.0001)
}

func TestDistributionCommissionService_OmitsZeroUsageLogID(t *testing.T) {
	memberID := int64(42)
	repo := &distributionCommissionRepoStub{}
	svc := NewDistributionCommissionService(
		&distributionCommissionAttributionRepoStub{
			attribution: &DistributionAttribution{
				UserID:           7,
				ChannelOrgID:     99,
				ReferrerMemberID: &memberID,
			},
		},
		&distributionCommissionMemberRepoStub{
			member: &DistributionMemberView{
				MemberID:       memberID,
				ChannelOrgID:   99,
				CommissionRate: 0.15,
				Status:         "active",
			},
		},
		nil,
		nil,
		repo,
		nil,
		0,
	)

	out, err := svc.AccrueForUsageLog(context.Background(), &UsageLog{
		UserID:    7,
		TotalCost: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, repo.inputs, 1)
	require.Nil(t, repo.inputs[0].UsageLogID)
}

func TestDistributionCommissionService_ReservesWalletForPayableSettlementModes(t *testing.T) {
	memberID := int64(42)
	repo := &distributionCommissionRepoStub{}
	walletRepo := &distributionCommissionWalletRepoStub{}
	svc := NewDistributionCommissionService(
		&distributionCommissionAttributionRepoStub{
			attribution: &DistributionAttribution{
				UserID:           7,
				ChannelOrgID:     99,
				ReferrerMemberID: &memberID,
			},
		},
		&distributionCommissionMemberRepoStub{
			member: &DistributionMemberView{
				MemberID:       memberID,
				ChannelOrgID:   99,
				CommissionRate: 0.15,
				Status:         "active",
			},
		},
		&distributionCommissionOrgRepoStub{
			org: &DistributionOrganization{
				ID:     99,
				Type:   "reseller",
				Config: map[string]any{"commission_settlement_method": "manual"},
			},
		},
		nil,
		repo,
		nil,
		0,
	)
	svc.SetWalletRepository(walletRepo)

	out, err := svc.AccrueForUsageLog(context.Background(), &UsageLog{
		UserID:    7,
		TotalCost: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, walletRepo.reserved, 1)
	require.Equal(t, int64(99), walletRepo.reserved[0].channelOrgID)
	require.InDelta(t, 1.5, walletRepo.reserved[0].amount, 0.0001)
}

type distributionCommissionAccrualStub struct {
	calls    int
	usageLog *UsageLog
}

func (s *distributionCommissionAccrualStub) AccrueForUsageLog(ctx context.Context, usageLog *UsageLog) (*DistributionCommissionLedger, error) {
	s.calls++
	s.usageLog = usageLog
	return nil, nil
}

func TestRecordDistributionCommissionBestEffort_CallsAccrualService(t *testing.T) {
	usageLog := &UsageLog{UserID: 7, TotalCost: 10}
	accrual := &distributionCommissionAccrualStub{}

	recordDistributionCommissionBestEffort(context.Background(), accrual, usageLog, "test")

	require.Equal(t, 1, accrual.calls)
	require.Same(t, usageLog, accrual.usageLog)
}

func TestDistributionCommissionService_ClampsByOrganizationCap(t *testing.T) {
	memberID := int64(42)
	repo := &distributionCommissionRepoStub{}
	svc := NewDistributionCommissionService(
		&distributionCommissionAttributionRepoStub{
			attribution: &DistributionAttribution{
				UserID:           7,
				ChannelOrgID:     99,
				ReferrerMemberID: &memberID,
			},
		},
		&distributionCommissionMemberRepoStub{
			member: &DistributionMemberView{
				MemberID:       memberID,
				ChannelOrgID:   99,
				CommissionRate: 0.5,
				Status:         "active",
			},
		},
		&distributionCommissionOrgRepoStub{
			org: &DistributionOrganization{
				ID:     99,
				Config: map[string]any{"commission_upper_ratio": 0.2},
			},
		},
		nil,
		repo,
		nil,
		0,
	)

	out, err := svc.AccrueForUsageLog(context.Background(), &UsageLog{
		UserID:    7,
		TotalCost: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.InDelta(t, 2.0, out.Amount, 0.0001)
	require.InDelta(t, 2.0, repo.inputs[0].Amount, 0.0001)
}

func TestDistributionCommissionService_CreatesManagementAndChannelCommissions(t *testing.T) {
	directMemberID := int64(42)
	parentMemberID := int64(41)
	managerMemberID := int64(40)
	repo := &distributionCommissionRepoStub{}
	svc := NewDistributionCommissionService(
		&distributionCommissionAttributionRepoStub{
			attribution: &DistributionAttribution{
				UserID:           7,
				ChannelOrgID:     99,
				ReferrerMemberID: &directMemberID,
			},
		},
		&distributionCommissionMemberRepoStub{
			memberByID: map[int64]*DistributionMemberView{
				directMemberID: {
					MemberID:       directMemberID,
					UserID:         12,
					ChannelOrgID:   99,
					RoleType:       "kol1",
					ParentMemberID: &parentMemberID,
					CommissionRate: 0.15,
					Status:         "active",
				},
				parentMemberID: {
					MemberID:       parentMemberID,
					UserID:         11,
					ChannelOrgID:   99,
					RoleType:       "agent",
					CommissionRate: 0.22,
					Status:         "active",
				},
			},
			memberByRole: map[string]*DistributionMemberView{
				"manager": {
					MemberID:       managerMemberID,
					UserID:         10,
					ChannelOrgID:   99,
					RoleType:       "manager",
					CommissionRate: 0.30,
					Status:         "active",
				},
			},
		},
		&distributionCommissionOrgRepoStub{
			org: &DistributionOrganization{
				ID:     99,
				Type:   "reseller",
				Status: "active",
				Config: map[string]any{
					"channel_commission_rate": 0.05,
					"management_reward_cap":   0.08,
					"commission_upper_ratio":  0.35,
				},
			},
		},
		nil,
		repo,
		nil,
		0,
	)

	out, err := svc.AccrueForUsageLog(context.Background(), &UsageLog{
		ID:        556,
		UserID:    7,
		TotalCost: 100,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, repo.inputs, 3)

	require.Equal(t, "direct", repo.inputs[0].CommissionType)
	require.Equal(t, directMemberID, repo.inputs[0].MemberID)
	require.InDelta(t, 15.0, repo.inputs[0].Amount, 0.0001)

	require.Equal(t, "management_reward", repo.inputs[1].CommissionType)
	require.Equal(t, parentMemberID, repo.inputs[1].MemberID)
	require.InDelta(t, 7.0, repo.inputs[1].Amount, 0.0001)

	require.Equal(t, "channel_commission", repo.inputs[2].CommissionType)
	require.Equal(t, managerMemberID, repo.inputs[2].MemberID)
	require.InDelta(t, 5.0, repo.inputs[2].Amount, 0.0001)
}

func TestDistributionCommissionService_CapShrinksChannelThenManagementThenDirect(t *testing.T) {
	directMemberID := int64(52)
	parentMemberID := int64(51)
	managerMemberID := int64(50)
	repo := &distributionCommissionRepoStub{}
	svc := NewDistributionCommissionService(
		&distributionCommissionAttributionRepoStub{
			attribution: &DistributionAttribution{
				UserID:           7,
				ChannelOrgID:     99,
				ReferrerMemberID: &directMemberID,
			},
		},
		&distributionCommissionMemberRepoStub{
			memberByID: map[int64]*DistributionMemberView{
				directMemberID: {
					MemberID:       directMemberID,
					UserID:         22,
					ChannelOrgID:   99,
					RoleType:       "kol1",
					ParentMemberID: &parentMemberID,
					CommissionRate: 0.20,
					Status:         "active",
				},
				parentMemberID: {
					MemberID:       parentMemberID,
					UserID:         21,
					ChannelOrgID:   99,
					RoleType:       "agent",
					CommissionRate: 0.30,
					Status:         "active",
				},
			},
			memberByRole: map[string]*DistributionMemberView{
				"manager": {
					MemberID:       managerMemberID,
					UserID:         20,
					ChannelOrgID:   99,
					RoleType:       "manager",
					CommissionRate: 0.40,
					Status:         "active",
				},
			},
		},
		&distributionCommissionOrgRepoStub{
			org: &DistributionOrganization{
				ID:     99,
				Type:   "reseller",
				Status: "active",
				Config: map[string]any{
					"channel_commission_rate": 0.10,
					"management_reward_cap":   0.10,
					"commission_upper_ratio":  0.25,
				},
			},
		},
		nil,
		repo,
		nil,
		0,
	)

	_, err := svc.AccrueForUsageLog(context.Background(), &UsageLog{
		UserID:    7,
		TotalCost: 100,
	})
	require.NoError(t, err)
	require.Len(t, repo.inputs, 2)

	require.Equal(t, "direct", repo.inputs[0].CommissionType)
	require.InDelta(t, 20.0, repo.inputs[0].Amount, 0.0001)
	require.Equal(t, "management_reward", repo.inputs[1].CommissionType)
	require.InDelta(t, 5.0, repo.inputs[1].Amount, 0.0001)
}

func TestDistributionCommissionService_ClampsDirectCommissionByConfiguredRateCap(t *testing.T) {
	memberID := int64(42)
	repo := &distributionCommissionRepoStub{}
	svc := NewDistributionCommissionService(
		&distributionCommissionAttributionRepoStub{
			attribution: &DistributionAttribution{
				UserID:           7,
				ChannelOrgID:     99,
				ReferrerMemberID: &memberID,
			},
		},
		&distributionCommissionMemberRepoStub{
			member: &DistributionMemberView{
				MemberID:       memberID,
				ChannelOrgID:   99,
				CommissionRate: 0.2,
				Status:         "active",
			},
		},
		&distributionCommissionOrgRepoStub{
			org: &DistributionOrganization{
				ID:     99,
				Config: map[string]any{"direct_commission_rate_cap": 0.1},
			},
		},
		nil,
		repo,
		nil,
		0,
	)

	out, err := svc.AccrueForUsageLog(context.Background(), &UsageLog{
		UserID:    7,
		TotalCost: 100,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, repo.inputs, 1)
	require.InDelta(t, 0.1, repo.inputs[0].Rate, 0.0001)
	require.InDelta(t, 10.0, repo.inputs[0].Amount, 0.0001)
}

func TestDistributionCommissionService_ClampsDraftsByUserCumulativeCommissionCap(t *testing.T) {
	directMemberID := int64(42)
	parentMemberID := int64(41)
	managerMemberID := int64(40)
	repo := &distributionCommissionRepoStub{
		totalByUserID: map[int64]float64{7: 30},
	}
	svc := NewDistributionCommissionService(
		&distributionCommissionAttributionRepoStub{
			attribution: &DistributionAttribution{
				UserID:           7,
				ChannelOrgID:     99,
				ReferrerMemberID: &directMemberID,
			},
		},
		&distributionCommissionMemberRepoStub{
			memberByID: map[int64]*DistributionMemberView{
				directMemberID: {
					MemberID:       directMemberID,
					UserID:         12,
					ChannelOrgID:   99,
					RoleType:       "kol1",
					ParentMemberID: &parentMemberID,
					CommissionRate: 0.15,
					Status:         "active",
				},
				parentMemberID: {
					MemberID:       parentMemberID,
					UserID:         11,
					ChannelOrgID:   99,
					RoleType:       "agent",
					CommissionRate: 0.22,
					Status:         "active",
				},
			},
			memberByRole: map[string]*DistributionMemberView{
				"manager": {
					MemberID:       managerMemberID,
					UserID:         10,
					ChannelOrgID:   99,
					RoleType:       "manager",
					CommissionRate: 0.30,
					Status:         "active",
				},
			},
		},
		&distributionCommissionOrgRepoStub{
			org: &DistributionOrganization{
				ID:   99,
				Type: "reseller",
				Config: map[string]any{
					"channel_commission_rate":   0.05,
					"management_reward_cap":     0.08,
					"user_commission_total_cap": 35.0,
				},
			},
		},
		nil,
		repo,
		nil,
		0,
	)

	out, err := svc.AccrueForUsageLog(context.Background(), &UsageLog{
		UserID:    7,
		TotalCost: 100,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, repo.inputs, 1)
	require.Equal(t, "direct", repo.inputs[0].CommissionType)
	require.InDelta(t, 5.0, repo.inputs[0].Amount, 0.0001)
}

func TestDistributionCommissionService_ClampsDraftByMemberCumulativeCommissionCap(t *testing.T) {
	memberID := int64(42)
	repo := &distributionCommissionRepoStub{
		totalByMemberID: map[int64]float64{memberID: 14},
	}
	svc := NewDistributionCommissionService(
		&distributionCommissionAttributionRepoStub{
			attribution: &DistributionAttribution{
				UserID:           7,
				ChannelOrgID:     99,
				ReferrerMemberID: &memberID,
			},
		},
		&distributionCommissionMemberRepoStub{
			member: &DistributionMemberView{
				MemberID:       memberID,
				ChannelOrgID:   99,
				CommissionRate: 0.15,
				Status:         "active",
			},
		},
		&distributionCommissionOrgRepoStub{
			org: &DistributionOrganization{
				ID:     99,
				Config: map[string]any{"member_commission_total_cap": 15.0},
			},
		},
		nil,
		repo,
		nil,
		0,
	)

	out, err := svc.AccrueForUsageLog(context.Background(), &UsageLog{
		UserID:    7,
		TotalCost: 100,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, repo.inputs, 1)
	require.InDelta(t, 1.0, repo.inputs[0].Amount, 0.0001)
	require.InDelta(t, 0.01, repo.inputs[0].Rate, 0.0001)
}
