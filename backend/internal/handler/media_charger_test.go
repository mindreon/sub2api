package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/media"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type mediaBillingRepoStub struct {
	last *service.UsageBillingCommand
}

func (s *mediaBillingRepoStub) Apply(_ context.Context, cmd *service.UsageBillingCommand) (*service.UsageBillingApplyResult, error) {
	cp := *cmd
	s.last = &cp
	bal := 8.5
	return &service.UsageBillingApplyResult{Applied: true, NewBalance: &bal}, nil
}

type mediaUsageLogWriterStub struct {
	last *service.UsageLog
}

func (s *mediaUsageLogWriterStub) Create(_ context.Context, log *service.UsageLog) (bool, error) {
	cp := *log
	s.last = &cp
	return true, nil
}

type mediaLastUsedSchedulerStub struct {
	ids []int64
}

func (s *mediaLastUsedSchedulerStub) ScheduleLastUsedUpdate(accountID int64) {
	s.ids = append(s.ids, accountID)
}

type mediaAPIKeyGetterStub struct {
	key *service.APIKey
}

func (s mediaAPIKeyGetterStub) GetByID(context.Context, int64) (*service.APIKey, error) {
	return s.key, nil
}

type mediaAccountGetterStub struct {
	account *service.Account
}

func (s mediaAccountGetterStub) GetByID(context.Context, int64) (*service.Account, error) {
	return s.account, nil
}

type mediaGroupGetterStub struct {
	group *service.Group
}

func (s mediaGroupGetterStub) GetByID(context.Context, int64) (*service.Group, error) {
	return s.group, nil
}

type mediaSubscriptionGetterStub struct {
	sub *service.UserSubscription
}

func (s mediaSubscriptionGetterStub) GetActiveByUserIDAndGroupID(context.Context, int64, int64) (*service.UserSubscription, error) {
	return s.sub, nil
}

func TestMediaChargerBalanceWritesUsageLogAndQuotaCosts(t *testing.T) {
	billingRepo := &mediaBillingRepoStub{}
	logs := &mediaUsageLogWriterStub{}
	lastUsed := &mediaLastUsedSchedulerStub{}
	multiplier := 1.25
	charger := newMediaCharger(mediaChargerDeps{
		billingRepo:       billingRepo,
		usageLogs:         logs,
		lastUsedScheduler: lastUsed,
		apiKeys: mediaAPIKeyGetterStub{key: &service.APIKey{
			ID:          2,
			UserID:      1,
			Quota:       100,
			RateLimit5h: 10,
		}},
		accounts: mediaAccountGetterStub{account: &service.Account{
			ID:             3,
			Type:           service.AccountTypeAPIKey,
			RateMultiplier: &multiplier,
			Extra:          map[string]any{"quota_limit": 100.0},
		}},
	})

	groupID := int64(9)
	_, err := charger.Charge(context.Background(), media.ChargeRequest{
		RequestID:           "task-balance",
		UserID:              1,
		APIKeyID:            2,
		AccountID:           3,
		GroupID:             &groupID,
		Model:               "doubao-seedance-2.0",
		MediaType:           "video",
		Metric:              media.MetricVideoToken,
		Units:               12345,
		CostBillingCurrency: 2.0,
		ActualCost:          3.0,
		RateMultiplier:      1.5,
	})
	if err != nil {
		t.Fatalf("charge: %v", err)
	}
	if billingRepo.last == nil {
		t.Fatal("billing command was not applied")
	}
	if billingRepo.last.BalanceCost != 3.0 || billingRepo.last.SubscriptionCost != 0 {
		t.Fatalf("unexpected billing costs: balance=%f subscription=%f", billingRepo.last.BalanceCost, billingRepo.last.SubscriptionCost)
	}
	if billingRepo.last.APIKeyQuotaCost != 3.0 || billingRepo.last.APIKeyRateLimitCost != 3.0 {
		t.Fatalf("api key costs not populated: quota=%f rate_limit=%f", billingRepo.last.APIKeyQuotaCost, billingRepo.last.APIKeyRateLimitCost)
	}
	if billingRepo.last.AccountQuotaCost != 2.5 {
		t.Fatalf("account quota cost = %f, want 2.5", billingRepo.last.AccountQuotaCost)
	}
	if logs.last == nil {
		t.Fatal("usage log was not written")
	}
	if logs.last.BillingType != service.BillingTypeBalance || logs.last.TotalCost != 2.0 || logs.last.ActualCost != 3.0 {
		t.Fatalf("unexpected usage log: billing_type=%d total=%f actual=%f", logs.last.BillingType, logs.last.TotalCost, logs.last.ActualCost)
	}
	if logs.last.MediaType == nil || *logs.last.MediaType != "video" {
		t.Fatalf("media type not recorded in usage log: %#v", logs.last.MediaType)
	}
	if len(lastUsed.ids) != 1 || lastUsed.ids[0] != 3 {
		t.Fatalf("account last-used update was not scheduled: %#v", lastUsed.ids)
	}
}

func TestMediaChargerResolvesSubscriptionGroup(t *testing.T) {
	billingRepo := &mediaBillingRepoStub{}
	logs := &mediaUsageLogWriterStub{}
	groupID := int64(9)
	subID := int64(88)
	charger := newMediaCharger(mediaChargerDeps{
		billingRepo: billingRepo,
		usageLogs:   logs,
		apiKeys:     mediaAPIKeyGetterStub{key: &service.APIKey{ID: 2, UserID: 1}},
		accounts:    mediaAccountGetterStub{account: &service.Account{ID: 3, Type: service.AccountTypeAPIKey}},
		groups:      mediaGroupGetterStub{group: &service.Group{ID: groupID, SubscriptionType: service.SubscriptionTypeSubscription}},
		subscriptions: mediaSubscriptionGetterStub{sub: &service.UserSubscription{
			ID:        subID,
			UserID:    1,
			GroupID:   groupID,
			Status:    service.SubscriptionStatusActive,
			ExpiresAt: time.Now().Add(time.Hour),
		}},
	})

	_, err := charger.Charge(context.Background(), media.ChargeRequest{
		RequestID:           "task-sub",
		UserID:              1,
		APIKeyID:            2,
		AccountID:           3,
		GroupID:             &groupID,
		Model:               "doubao-seedance-2.0",
		MediaType:           "video",
		CostBillingCurrency: 2.0,
		ActualCost:          3.0,
		RateMultiplier:      1.5,
	})
	if err != nil {
		t.Fatalf("charge: %v", err)
	}
	if billingRepo.last == nil {
		t.Fatal("billing command was not applied")
	}
	if billingRepo.last.BillingType != service.BillingTypeSubscription || billingRepo.last.SubscriptionID == nil || *billingRepo.last.SubscriptionID != subID {
		t.Fatalf("subscription was not resolved: cmd=%#v", billingRepo.last)
	}
	if billingRepo.last.SubscriptionCost != 3.0 || billingRepo.last.BalanceCost != 0 {
		t.Fatalf("unexpected billing costs: subscription=%f balance=%f", billingRepo.last.SubscriptionCost, billingRepo.last.BalanceCost)
	}
	if logs.last == nil || logs.last.SubscriptionID == nil || *logs.last.SubscriptionID != subID {
		t.Fatalf("usage log did not record subscription: %#v", logs.last)
	}
}

func TestMediaChargerSubscriptionGroupWithoutActiveSubscriptionDoesNotChargeBalance(t *testing.T) {
	billingRepo := &mediaBillingRepoStub{}
	groupID := int64(9)
	charger := newMediaCharger(mediaChargerDeps{
		billingRepo:   billingRepo,
		apiKeys:       mediaAPIKeyGetterStub{key: &service.APIKey{ID: 2, UserID: 1}},
		accounts:      mediaAccountGetterStub{account: &service.Account{ID: 3, Type: service.AccountTypeAPIKey}},
		groups:        mediaGroupGetterStub{group: &service.Group{ID: groupID, SubscriptionType: service.SubscriptionTypeSubscription}},
		subscriptions: mediaSubscriptionGetterStub{sub: nil},
	})

	_, err := charger.Charge(context.Background(), media.ChargeRequest{
		RequestID:           "task-sub-missing",
		UserID:              1,
		APIKeyID:            2,
		AccountID:           3,
		GroupID:             &groupID,
		Model:               "doubao-seedance-2.0",
		MediaType:           "video",
		CostBillingCurrency: 2.0,
		ActualCost:          3.0,
		RateMultiplier:      1.5,
	})
	if !errors.Is(err, service.ErrSubscriptionInvalid) {
		t.Fatalf("expected subscription invalid error, got %v", err)
	}
	if billingRepo.last != nil {
		t.Fatalf("subscription group without active subscription must not balance charge: %#v", billingRepo.last)
	}
}
