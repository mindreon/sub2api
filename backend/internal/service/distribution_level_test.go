package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type distributionLevelSettingRepoStub struct {
	values map[string]string
}

func (s *distributionLevelSettingRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *distributionLevelSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (s *distributionLevelSettingRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *distributionLevelSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *distributionLevelSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *distributionLevelSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *distributionLevelSettingRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestParseDistributionLevelConfigsNormalizesRecords(t *testing.T) {
	levels, err := parseDistributionLevelConfigs([]map[string]any{
		{
			"code":            " vip ",
			"name":            "",
			"commission_rate": 120,
			"active":          true,
			"sort_order":      3,
			"note":            " core ",
		},
		{
			"code": "  ",
		},
	})

	require.NoError(t, err)
	require.Equal(t, []DistributionLevelConfig{
		{
			Code:           "VIP",
			Name:           "VIP",
			CommissionRate: 100,
			Active:         true,
			SortOrder:      3,
			Note:           "core",
		},
	}, levels)
}

func TestSettingServiceGetDistributionGlobalLevels(t *testing.T) {
	repo := &distributionLevelSettingRepoStub{
		values: map[string]string{
			SettingKeyDistributionGlobalLevels: `[{"code":" gold ","name":"","commission_rate":12.5,"active":true,"sort_order":1,"note":" default "}]`,
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	levels := svc.GetDistributionGlobalLevels(context.Background())

	require.Equal(t, []DistributionLevelConfig{
		{
			Code:           "GOLD",
			Name:           "GOLD",
			CommissionRate: 12.5,
			Active:         true,
			SortOrder:      1,
			Note:           "default",
		},
	}, levels)
}
