package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type settingBrandingRepoStub struct {
	values map[string]string
}

func (s *settingBrandingRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *settingBrandingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	panic("unexpected GetValue call")
}

func (s *settingBrandingRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *settingBrandingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (s *settingBrandingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *settingBrandingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *settingBrandingRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

type settingBrandingResolverStub struct {
	orgByHost map[string]*DistributionOrganization
}

func (s *settingBrandingResolverStub) GetByBrandHost(ctx context.Context, host string) (*DistributionOrganization, error) {
	if s == nil {
		return nil, nil
	}
	return s.orgByHost[NormalizeDistributionBrandHost(host)], nil
}

func TestSettingServiceGetPublicSettingsAppliesBrandingOverride(t *testing.T) {
	svc := NewSettingService(&settingBrandingRepoStub{
		values: map[string]string{
			SettingKeySiteName:   "Sub2API",
			SettingKeySiteLogo:   "/logo.png",
			SettingKeyAPIBaseURL: "https://api.example.com",
		},
	}, &config.Config{})
	svc.SetDistributionBrandingResolver(&settingBrandingResolverStub{
		orgByHost: map[string]*DistributionOrganization{
			"brand.example.com": {
				ID:   88,
				Type: "oem",
				Name: "BrandHub",
				BrandConfig: map[string]any{
					"logo_url":   "https://cdn.example.com/brand-logo.png",
					"domain":     "brand.example.com",
					"api_domain": "api.brand.example.com",
				},
			},
		},
	})

	settings, err := svc.GetPublicSettings(WithPublicSettingsRequestMeta(context.Background(), "brand.example.com", "https"))
	require.NoError(t, err)
	require.Equal(t, "BrandHub", settings.SiteName)
	require.Equal(t, "https://cdn.example.com/brand-logo.png", settings.SiteLogo)
	require.Equal(t, "https://api.brand.example.com", settings.APIBaseURL)
}
