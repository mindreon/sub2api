package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type distributionAutoSettlementRepoStub struct {
	ids   []int64
	limit int
	err   error
}

func (s *distributionAutoSettlementRepoStub) ListAutoSettleCommissionIDs(ctx context.Context, limit int) ([]int64, error) {
	s.limit = limit
	if s.err != nil {
		return nil, s.err
	}
	return append([]int64(nil), s.ids...), nil
}

type distributionAutoSettlementExecutorStub struct {
	ids    []int64
	inputs []DistributionCommissionSettlementInput
	errs   map[int64]error
}

func (s *distributionAutoSettlementExecutorStub) SettleCommission(ctx context.Context, commissionID int64, input DistributionCommissionSettlementInput) (*DistributionCommissionLedger, error) {
	s.ids = append(s.ids, commissionID)
	s.inputs = append(s.inputs, input)
	if err := s.errs[commissionID]; err != nil {
		return nil, err
	}
	return &DistributionCommissionLedger{ID: commissionID, Status: "settled", SettlementMethod: input.SettlementMethod}, nil
}

func TestDistributionAutoSettlementServiceRunOnce_SettlesAvailableAutoCommissions(t *testing.T) {
	repo := &distributionAutoSettlementRepoStub{ids: []int64{11, 12, 13}}
	executor := &distributionAutoSettlementExecutorStub{}
	svc := NewDistributionAutoSettlementService(repo, executor, time.Minute)

	svc.runOnce()

	require.Equal(t, 100, repo.limit)
	require.Equal(t, []int64{11, 12, 13}, executor.ids)
	require.Len(t, executor.inputs, 3)
	for _, input := range executor.inputs {
		require.Equal(t, "auto", input.SettlementMethod)
	}
}

func TestDistributionAutoSettlementServiceRunOnce_ContinuesAfterSingleSettlementFailure(t *testing.T) {
	repo := &distributionAutoSettlementRepoStub{ids: []int64{21, 22, 23}}
	executor := &distributionAutoSettlementExecutorStub{
		errs: map[int64]error{
			22: errors.New("boom"),
		},
	}
	svc := NewDistributionAutoSettlementService(repo, executor, time.Minute)

	svc.runOnce()

	require.Equal(t, []int64{21, 22, 23}, executor.ids)
	require.Len(t, executor.inputs, 3)
}
