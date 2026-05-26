package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

var (
	ErrInvalidDistributionCommission = infraerrors.BadRequest("INVALID_DISTRIBUTION_COMMISSION", "invalid distribution commission")
)

type DistributionCommissionLedger struct {
	ID                    int64      `json:"id"`
	ChannelOrgID          int64      `json:"channel_org_id"`
	MemberID              int64      `json:"member_id"`
	UserID                int64      `json:"user_id"`
	UsageLogID            *int64     `json:"usage_log_id,omitempty"`
	CommissionType        string     `json:"commission_type"`
	BaseAmount            float64    `json:"base_amount"`
	Rate                  float64    `json:"rate"`
	Amount                float64    `json:"amount"`
	Status                string     `json:"status"`
	SettlementMethod      string     `json:"settlement_method"`
	SettlementReferenceNo string     `json:"settlement_reference_no"`
	SettlementNote        string     `json:"settlement_note"`
	FrozenUntil           *time.Time `json:"frozen_until,omitempty"`
	SettledAt             *time.Time `json:"settled_at,omitempty"`
	SettledByUserID       *int64     `json:"settled_by_user_id,omitempty"`
	ReversedFromID        *int64     `json:"reversed_from_id,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type DistributionCommissionInput struct {
	ChannelOrgID          int64
	MemberID              int64
	UserID                int64
	UsageLogID            *int64
	CommissionType        string
	BaseAmount            float64
	Rate                  float64
	Amount                float64
	Status                string
	SettlementMethod      string
	SettlementReferenceNo string
	SettlementNote        string
	FrozenUntil           *time.Time
	SettledAt             *time.Time
	SettledByUserID       *int64
	ReversedFromID        *int64
}

type DistributionCommissionRepository interface {
	Create(ctx context.Context, input DistributionCommissionInput) (*DistributionCommissionLedger, error)
	GetByID(ctx context.Context, commissionID int64) (*DistributionCommissionLedger, error)
	HasCommissionTypeSince(ctx context.Context, channelOrgID int64, commissionType string, since time.Time) (bool, error)
	GetTotalCommissionByUserID(ctx context.Context, channelOrgID int64, userID int64) (float64, error)
	GetTotalCommissionByMemberID(ctx context.Context, channelOrgID int64, memberID int64) (float64, error)
}

type DistributionAttributionLookupRepository interface {
	GetByUserID(ctx context.Context, userID int64) (*DistributionAttribution, error)
}

type DistributionMemberLookupRepository interface {
	GetByID(ctx context.Context, memberID int64) (*DistributionMemberView, error)
	GetByChannelOrgIDAndRole(ctx context.Context, channelOrgID int64, roleType string) (*DistributionMemberView, error)
}

type DistributionOrganizationLookupRepository interface {
	GetByID(ctx context.Context, id int64) (*DistributionOrganization, error)
	GetByOwnerUserID(ctx context.Context, userID int64) (*DistributionOrganization, error)
}

type DistributionConsumptionRepository interface {
	GetMonthlyChannelConsumption(ctx context.Context, channelOrgID int64, since time.Time) (float64, error)
}

type DistributionCommissionHistoryRepository interface {
	HasCommissionTypeSince(ctx context.Context, channelOrgID int64, commissionType string, since time.Time) (bool, error)
}

type DistributionCommissionService struct {
	attributionRepo  DistributionAttributionLookupRepository
	memberRepo       DistributionMemberLookupRepository
	organizationRepo DistributionOrganizationLookupRepository
	consumptionRepo  DistributionConsumptionRepository
	commissionRepo   DistributionCommissionRepository
	historyRepo      DistributionCommissionHistoryRepository
	walletRepo       DistributionWalletMutationRepository
	settingService   *SettingService
	freezeDuration   time.Duration
}

type distributionCommissionDraft struct {
	MemberID         int64
	CommissionType   string
	BaseAmount       float64
	Rate             float64
	Amount           float64
	SettlementMethod string
}

type distributionCommissionAccrual interface {
	AccrueForUsageLog(ctx context.Context, usageLog *UsageLog) (*DistributionCommissionLedger, error)
}

func NewDistributionCommissionService(
	attributionRepo DistributionAttributionLookupRepository,
	memberRepo DistributionMemberLookupRepository,
	organizationRepo DistributionOrganizationLookupRepository,
	consumptionRepo DistributionConsumptionRepository,
	commissionRepo DistributionCommissionRepository,
	historyRepo DistributionCommissionHistoryRepository,
	freezeDuration time.Duration,
) *DistributionCommissionService {
	return &DistributionCommissionService{
		attributionRepo:  attributionRepo,
		memberRepo:       memberRepo,
		organizationRepo: organizationRepo,
		consumptionRepo:  consumptionRepo,
		commissionRepo:   commissionRepo,
		historyRepo:      historyRepo,
		freezeDuration:   freezeDuration,
	}
}

func (s *DistributionCommissionService) SetSettingService(settingService *SettingService) {
	if s == nil {
		return
	}
	s.settingService = settingService
}

func (s *DistributionCommissionService) SetWalletRepository(walletRepo DistributionWalletMutationRepository) {
	if s == nil {
		return
	}
	s.walletRepo = walletRepo
}

func (s *DistributionCommissionService) AccrueForUsageLog(ctx context.Context, usageLog *UsageLog) (*DistributionCommissionLedger, error) {
	if s == nil || s.attributionRepo == nil || s.memberRepo == nil || s.commissionRepo == nil {
		return nil, ErrInvalidDistributionCommission
	}
	if usageLog == nil || usageLog.UserID <= 0 {
		return nil, ErrInvalidDistributionCommission
	}

	attribution, err := s.attributionRepo.GetByUserID(ctx, usageLog.UserID)
	if err != nil {
		if errors.Is(err, ErrDistributionAttributionNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if attribution == nil || attribution.ReferrerMemberID == nil || *attribution.ReferrerMemberID <= 0 {
		return nil, nil
	}

	member, err := s.memberRepo.GetByID(ctx, *attribution.ReferrerMemberID)
	if err != nil {
		return nil, err
	}
	if member == nil || member.ChannelOrgID != attribution.ChannelOrgID || !strings.EqualFold(strings.TrimSpace(member.Status), "active") {
		return nil, nil
	}

	baseAmount := distributionCommissionBaseAmount(usageLog)
	rate := member.CommissionRate
	if strings.EqualFold(strings.TrimSpace(member.RoleType), "kol2") && s.settingService != nil {
		if globalRate := s.settingService.GetDistributionKol2Rate(ctx); globalRate > 0 {
			rate = globalRate / 100
		}
	}
	if baseAmount <= 0 || rate <= 0 {
		return nil, nil
	}

	var org *DistributionOrganization
	if s.organizationRepo != nil {
		if loadedOrg, err := s.organizationRepo.GetByID(ctx, attribution.ChannelOrgID); err == nil && loadedOrg != nil {
			org = loadedOrg
		}
	}
	if cap := resolveDistributionDirectCommissionRateCap(org); cap > 0 && rate > cap {
		rate = cap
	}

	drafts := []distributionCommissionDraft{
		{
			MemberID:         member.MemberID,
			CommissionType:   "direct",
			BaseAmount:       baseAmount,
			Rate:             rate,
			Amount:           baseAmount * rate,
			SettlementMethod: resolveDistributionCommissionSettlementMethod(org),
		},
	}
	if managementDraft, ok := s.buildManagementRewardDraft(ctx, org, member, baseAmount, rate); ok {
		drafts = append(drafts, managementDraft)
	}
	if channelDraft, ok := s.buildChannelCommissionDraft(ctx, org, member, baseAmount); ok {
		drafts = append(drafts, channelDraft)
	}
	if err := s.applyMemberCommissionTotalCap(ctx, attribution.ChannelOrgID, drafts, org); err != nil {
		return nil, err
	}

	capRatio := s.resolveCommissionUpperRatio(ctx, org)
	if capRatio > 0 {
		applyDistributionCommissionCap(drafts, baseAmount*capRatio)
	}
	if err := s.applyUserCommissionTotalCap(ctx, attribution.ChannelOrgID, usageLog.UserID, drafts, org); err != nil {
		return nil, err
	}

	filteredDrafts := make([]distributionCommissionDraft, 0, len(drafts))
	for _, draft := range drafts {
		if draft.Amount > 0 && draft.Rate > 0 {
			filteredDrafts = append(filteredDrafts, draft)
		}
	}
	if len(filteredDrafts) == 0 {
		return nil, nil
	}

	reservedAmount := 0.0
	if shouldAffectDistributionWallet(org) && s.walletRepo != nil {
		for _, draft := range filteredDrafts {
			if requiresDistributionWalletReserve(draft.SettlementMethod) {
				reservedAmount += draft.Amount
			}
		}
		if reservedAmount > 0 {
			if _, err := s.walletRepo.ReserveCommission(ctx, attribution.ChannelOrgID, reservedAmount); err != nil {
				return nil, err
			}
		}
	}

	status := "available"
	var frozenUntil *time.Time
	if freezeDuration := s.resolveFreezeDuration(ctx, attribution.ChannelOrgID); freezeDuration > 0 {
		status = "frozen"
		t := time.Now().UTC().Add(freezeDuration)
		frozenUntil = &t
	}

	var usageLogID *int64
	if usageLog.ID > 0 {
		id := usageLog.ID
		usageLogID = &id
	}

	var ledger *DistributionCommissionLedger
	for _, draft := range filteredDrafts {
		created, err := s.commissionRepo.Create(ctx, DistributionCommissionInput{
			ChannelOrgID:     attribution.ChannelOrgID,
			MemberID:         draft.MemberID,
			UserID:           usageLog.UserID,
			UsageLogID:       usageLogID,
			CommissionType:   draft.CommissionType,
			BaseAmount:       draft.BaseAmount,
			Rate:             draft.Rate,
			Amount:           draft.Amount,
			Status:           status,
			SettlementMethod: draft.SettlementMethod,
			FrozenUntil:      frozenUntil,
		})
		if err != nil {
			if reservedAmount > 0 && shouldAffectDistributionWallet(org) && s.walletRepo != nil {
				if _, releaseErr := s.walletRepo.ReleaseCommission(ctx, attribution.ChannelOrgID, reservedAmount); releaseErr != nil {
					logger.LegacyPrintf("service.distribution_commission", "wallet release after create failure: %v", releaseErr)
				}
			}
			return nil, err
		}
		if ledger == nil && draft.CommissionType == "direct" {
			ledger = created
		}
	}
	if ledger == nil {
		return nil, nil
	}
	s.recordTeamRewardBestEffort(ctx, attribution.ChannelOrgID, frozenUntil, usageLog)
	return ledger, nil
}

func shouldAffectDistributionWallet(org *DistributionOrganization) bool {
	if org == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(org.Type)) {
	case "reseller", "oem":
		return true
	default:
		return false
	}
}

func requiresDistributionWalletReserve(settlementMethod string) bool {
	switch strings.ToLower(strings.TrimSpace(settlementMethod)) {
	case "balance", "auto", "manual":
		return true
	default:
		return false
	}
}

func (s *DistributionCommissionService) buildManagementRewardDraft(
	ctx context.Context,
	org *DistributionOrganization,
	member *DistributionMemberView,
	baseAmount float64,
	directRate float64,
) (distributionCommissionDraft, bool) {
	if s == nil || s.memberRepo == nil || member == nil || member.ParentMemberID == nil || *member.ParentMemberID <= 0 {
		return distributionCommissionDraft{}, false
	}
	switch strings.ToLower(strings.TrimSpace(member.RoleType)) {
	case "kol1", "kol2":
	default:
		return distributionCommissionDraft{}, false
	}

	parent, err := s.memberRepo.GetByID(ctx, *member.ParentMemberID)
	if err != nil || parent == nil {
		return distributionCommissionDraft{}, false
	}
	if parent.ChannelOrgID != member.ChannelOrgID || !strings.EqualFold(strings.TrimSpace(parent.Status), "active") {
		return distributionCommissionDraft{}, false
	}

	rate := parent.CommissionRate - directRate
	if cap := distributionOrganizationConfigFloat(nilSafeDistributionOrgConfig(org), "management_reward_rate"); cap > 0 && rate > cap {
		rate = cap
	}
	if cap := distributionOrganizationConfigFloat(nilSafeDistributionOrgConfig(org), "management_reward_cap"); cap > 0 && rate > cap {
		rate = cap
	}
	if rate <= 0 {
		return distributionCommissionDraft{}, false
	}

	return distributionCommissionDraft{
		MemberID:         parent.MemberID,
		CommissionType:   "management_reward",
		BaseAmount:       baseAmount,
		Rate:             rate,
		Amount:           baseAmount * rate,
		SettlementMethod: resolveDistributionCommissionSettlementMethod(org),
	}, true
}

func (s *DistributionCommissionService) buildChannelCommissionDraft(
	ctx context.Context,
	org *DistributionOrganization,
	member *DistributionMemberView,
	baseAmount float64,
) (distributionCommissionDraft, bool) {
	if s == nil || s.memberRepo == nil || member == nil {
		return distributionCommissionDraft{}, false
	}
	rate := distributionOrganizationConfigFloat(nilSafeDistributionOrgConfig(org), "channel_commission_rate")
	if rate <= 0 {
		return distributionCommissionDraft{}, false
	}

	manager, err := s.memberRepo.GetByChannelOrgIDAndRole(ctx, member.ChannelOrgID, "manager")
	if err != nil || manager == nil {
		return distributionCommissionDraft{}, false
	}
	if manager.ChannelOrgID != member.ChannelOrgID || !strings.EqualFold(strings.TrimSpace(manager.Status), "active") {
		return distributionCommissionDraft{}, false
	}
	if manager.MemberID == member.MemberID {
		return distributionCommissionDraft{}, false
	}

	return distributionCommissionDraft{
		MemberID:         manager.MemberID,
		CommissionType:   "channel_commission",
		BaseAmount:       baseAmount,
		Rate:             rate,
		Amount:           baseAmount * rate,
		SettlementMethod: resolveDistributionCommissionSettlementMethod(org),
	}, true
}

func (s *DistributionCommissionService) resolveCommissionUpperRatio(ctx context.Context, org *DistributionOrganization) float64 {
	capRatio := distributionOrganizationConfigFloat(nilSafeDistributionOrgConfig(org), "commission_upper_ratio", "total_commission_ratio", "commission_limit_ratio")
	if capRatio <= 0 && s != nil && s.settingService != nil {
		capRatio = s.settingService.GetDistributionCommissionUpperRatio(ctx) / 100
	}
	return capRatio
}

func resolveDistributionDirectCommissionRateCap(org *DistributionOrganization) float64 {
	return distributionOrganizationConfigFloat(
		nilSafeDistributionOrgConfig(org),
		"direct_commission_rate_cap",
		"direct_commission_cap",
		"direct_rate_cap",
	)
}

func resolveDistributionUserCommissionTotalCap(org *DistributionOrganization) float64 {
	return distributionOrganizationConfigFloat(
		nilSafeDistributionOrgConfig(org),
		"user_commission_total_cap",
		"single_user_commission_cap",
		"per_user_commission_cap",
	)
}

func resolveDistributionMemberCommissionTotalCap(org *DistributionOrganization) float64 {
	return distributionOrganizationConfigFloat(
		nilSafeDistributionOrgConfig(org),
		"member_commission_total_cap",
		"single_member_commission_cap",
		"promoter_commission_cap",
		"per_member_commission_cap",
	)
}

func (s *DistributionCommissionService) applyUserCommissionTotalCap(
	ctx context.Context,
	channelOrgID int64,
	userID int64,
	drafts []distributionCommissionDraft,
	org *DistributionOrganization,
) error {
	if s == nil || s.commissionRepo == nil || channelOrgID <= 0 || userID <= 0 || len(drafts) == 0 {
		return nil
	}
	capAmount := resolveDistributionUserCommissionTotalCap(org)
	if capAmount <= 0 {
		return nil
	}
	currentTotal, err := s.commissionRepo.GetTotalCommissionByUserID(ctx, channelOrgID, userID)
	if err != nil {
		return err
	}
	remaining := capAmount - currentTotal
	if remaining <= 0 {
		clearDistributionCommissionDrafts(drafts)
		return nil
	}
	applyDistributionCommissionCap(drafts, remaining)
	return nil
}

func (s *DistributionCommissionService) applyMemberCommissionTotalCap(
	ctx context.Context,
	channelOrgID int64,
	drafts []distributionCommissionDraft,
	org *DistributionOrganization,
) error {
	if s == nil || s.commissionRepo == nil || channelOrgID <= 0 || len(drafts) == 0 {
		return nil
	}
	capAmount := resolveDistributionMemberCommissionTotalCap(org)
	if capAmount <= 0 {
		return nil
	}
	runningTotals := make(map[int64]float64)
	for i := range drafts {
		memberID := drafts[i].MemberID
		if memberID <= 0 || drafts[i].Amount <= 0 {
			continue
		}
		total, ok := runningTotals[memberID]
		if !ok {
			loadedTotal, err := s.commissionRepo.GetTotalCommissionByMemberID(ctx, channelOrgID, memberID)
			if err != nil {
				return err
			}
			total = loadedTotal
		}
		remaining := capAmount - total
		if remaining <= 0 {
			drafts[i].Amount = 0
			drafts[i].Rate = 0
			runningTotals[memberID] = total
			continue
		}
		if drafts[i].Amount > remaining {
			drafts[i].Amount = remaining
			if drafts[i].BaseAmount > 0 {
				drafts[i].Rate = drafts[i].Amount / drafts[i].BaseAmount
			}
		}
		runningTotals[memberID] = total + drafts[i].Amount
	}
	return nil
}

func clearDistributionCommissionDrafts(drafts []distributionCommissionDraft) {
	for i := range drafts {
		drafts[i].Amount = 0
		drafts[i].Rate = 0
	}
}

func applyDistributionCommissionCap(drafts []distributionCommissionDraft, maxAmount float64) {
	if maxAmount <= 0 || len(drafts) == 0 {
		return
	}
	total := 0.0
	for _, draft := range drafts {
		total += draft.Amount
	}
	excess := total - maxAmount
	if excess <= 0 {
		return
	}

	for _, commissionType := range []string{"team_reward", "channel_commission", "management_reward", "direct"} {
		if excess <= 0 {
			break
		}
		for i := range drafts {
			if drafts[i].CommissionType != commissionType || drafts[i].Amount <= 0 {
				continue
			}
			shrink := excess
			if drafts[i].Amount < shrink {
				shrink = drafts[i].Amount
			}
			drafts[i].Amount -= shrink
			if drafts[i].BaseAmount > 0 {
				drafts[i].Rate = drafts[i].Amount / drafts[i].BaseAmount
			}
			excess -= shrink
			if excess <= 0 {
				break
			}
		}
	}
}

func resolveDistributionCommissionSettlementMethod(org *DistributionOrganization) string {
	method := strings.TrimSpace(strings.ToLower(distributionOrganizationConfigString(nilSafeDistributionOrgConfig(org),
		"commission_settlement_method",
		"settlement_method",
		"commission_settlement_mode",
	)))
	switch method {
	case "balance", "auto", "manual", "offline":
		return method
	default:
		return "balance"
	}
}

func (s *DistributionCommissionService) resolveFreezeDuration(ctx context.Context, channelOrgID int64) time.Duration {
	if s == nil {
		return 0
	}
	if s.organizationRepo != nil && channelOrgID > 0 {
		if org, err := s.organizationRepo.GetByID(ctx, channelOrgID); err == nil && org != nil {
			if hours := distributionOrganizationConfigFloat(org.Config, "freeze_hours"); hours > 0 {
				return time.Duration(hours * float64(time.Hour))
			}
			if days := distributionOrganizationConfigFloat(org.Config, "freeze_days"); days > 0 {
				return time.Duration(days * 24 * float64(time.Hour))
			}
		}
	}
	if s.settingService != nil {
		if hours := s.settingService.GetDistributionFreezeHours(ctx); hours > 0 {
			return time.Duration(hours) * time.Hour
		}
	}
	return s.freezeDuration
}

func distributionOrganizationConfigFloat(config map[string]any, keys ...string) float64 {
	for _, key := range keys {
		if config == nil {
			continue
		}
		if v, ok := config[key]; ok {
			switch n := v.(type) {
			case float64:
				return n
			case float32:
				return float64(n)
			case int:
				return float64(n)
			case int64:
				return float64(n)
			case json.Number:
				if f, err := n.Float64(); err == nil {
					return f
				}
			case string:
				if f, err := strconv.ParseFloat(strings.TrimSpace(n), 64); err == nil {
					return f
				}
			}
		}
	}
	return 0
}

func distributionOrganizationConfigString(config map[string]any, keys ...string) string {
	for _, key := range keys {
		if config == nil {
			continue
		}
		if v, ok := config[key]; ok {
			if s, ok := v.(string); ok {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

func nilSafeDistributionOrgConfig(org *DistributionOrganization) map[string]any {
	if org == nil {
		return nil
	}
	return org.Config
}

func (s *DistributionCommissionService) recordTeamRewardBestEffort(ctx context.Context, channelOrgID int64, frozenUntil *time.Time, usageLog *UsageLog) {
	if s == nil || s.organizationRepo == nil || s.commissionRepo == nil || s.consumptionRepo == nil || s.memberRepo == nil {
		return
	}
	if usageLog == nil || channelOrgID <= 0 {
		return
	}

	org, err := s.organizationRepo.GetByID(ctx, channelOrgID)
	if err != nil || org == nil {
		return
	}
	teamRewardRate := distributionOrganizationConfigFloat(org.Config, "team_reward_rate")
	teamRewardThreshold := distributionOrganizationConfigFloat(org.Config, "team_reward_threshold")
	if teamRewardRate <= 0 || teamRewardThreshold <= 0 {
		return
	}

	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	if s.historyRepo != nil {
		paid, err := s.historyRepo.HasCommissionTypeSince(ctx, channelOrgID, "team_reward", monthStart)
		if err == nil && paid {
			return
		}
	}

	monthlyTotal, err := s.consumptionRepo.GetMonthlyChannelConsumption(ctx, channelOrgID, monthStart)
	if err != nil {
		return
	}
	if monthlyTotal < teamRewardThreshold {
		return
	}

	manager, err := s.memberRepo.GetByChannelOrgIDAndRole(ctx, channelOrgID, "manager")
	if err != nil || manager == nil {
		return
	}

	amount := monthlyTotal * teamRewardRate
	if amount <= 0 {
		return
	}
	var usageLogID *int64
	if usageLog.ID > 0 {
		id := usageLog.ID
		usageLogID = &id
	}
	if _, err := s.commissionRepo.Create(ctx, DistributionCommissionInput{
		ChannelOrgID:     channelOrgID,
		MemberID:         manager.MemberID,
		UserID:           manager.UserID,
		UsageLogID:       usageLogID,
		CommissionType:   "team_reward",
		BaseAmount:       monthlyTotal,
		Rate:             teamRewardRate,
		Amount:           amount,
		Status:           "available",
		SettlementMethod: "manual",
		FrozenUntil:      frozenUntil,
	}); err != nil {
		logger.LegacyPrintf("service.distribution_commission", "Create team reward commission failed: %v", err)
	}
}

func distributionCommissionBaseAmount(usageLog *UsageLog) float64 {
	if usageLog == nil {
		return 0
	}
	if usageLog.AccountStatsCost != nil && *usageLog.AccountStatsCost > 0 {
		return *usageLog.AccountStatsCost
	}
	if usageLog.ActualCost > 0 {
		return usageLog.ActualCost
	}
	return usageLog.TotalCost
}

func recordDistributionCommissionBestEffort(ctx context.Context, svc distributionCommissionAccrual, usageLog *UsageLog, logKey string) {
	if svc == nil || usageLog == nil {
		return
	}
	commissionCtx, cancel := detachedBillingContext(ctx)
	defer cancel()

	if _, err := svc.AccrueForUsageLog(commissionCtx, usageLog); err != nil {
		if strings.TrimSpace(logKey) == "" {
			logKey = "service.distribution_commission"
		}
		logger.LegacyPrintf(logKey, "Accrue distribution commission failed: %v", err)
	}
}
