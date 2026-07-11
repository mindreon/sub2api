package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/media"
	"github.com/Wei-Shaw/sub2api/internal/media/providers"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestAccountMediaProviderFactory_SelectsVideoAPIStyle(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		apiStyle   string
		wantCommon bool
	}{
		{name: "empty base URL uses native", wantCommon: false},
		{name: "official Volcengine URL uses native", baseURL: "https://ark.cn-beijing.volces.com", wantCommon: false},
		{name: "official BytePlus URL uses native", baseURL: "https://ark.ap-southeast-1.bytepluses.com", wantCommon: false},
		{name: "custom URL uses common style", baseURL: "https://token.genvia.ai", wantCommon: true},
		{name: "explicit native overrides custom URL", baseURL: "https://proxy.example.com", apiStyle: "native", wantCommon: false},
		{name: "explicit common overrides official URL", baseURL: "https://ark.cn-beijing.volces.com", apiStyle: "common", wantCommon: true},
	}

	factory := NewAccountMediaProviderFactory()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := factory.NewProvider(media.AccountSelection{
				AccountID: 1,
				Platform:  service.PlatformVolcengine,
				APIKey:    "test-key",
				BaseURL:   tt.baseURL,
				APIStyle:  tt.apiStyle,
			}, "seedance2.0-fast-p5")
			if err != nil {
				t.Fatalf("new provider: %v", err)
			}
			_, isCommon := provider.(*providers.CommonStyleVideoProvider)
			if isCommon != tt.wantCommon {
				t.Fatalf("provider type %T, want common=%v", provider, tt.wantCommon)
			}
		})
	}
}

func TestAccountToMediaSelectionReadsAPIStyle(t *testing.T) {
	selection := accountToMediaSelection(&service.Account{
		ID:       7,
		Platform: service.PlatformVolcengine,
		Extra: map[string]any{
			"base_url":        "https://proxy.example.com",
			"media_api_style": "native",
		},
	})

	if selection.BaseURL != "https://proxy.example.com" || selection.APIStyle != "native" {
		t.Fatalf("unexpected selection: %+v", selection)
	}
}
