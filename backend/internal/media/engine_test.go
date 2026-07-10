package media

import (
	"errors"
	"math"
	"testing"
)

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

// 720p 16:9、5 秒、无视频输入：估算 token 与官方示例价（4.97 元）对齐。
func TestEstimateVideoTokens_720p5s(t *testing.T) {
	u := BillingUsage{
		VideoOutputSeconds: 5,
		VideoWidth:         1280,
		VideoHeight:        720,
		VideoFPS:           24,
	}
	got := estimateVideoTokens(u)
	if got != 108000 {
		t.Fatalf("expected 108000 tokens, got %d", got)
	}
}

func TestEstimateVideoTokens_DefaultFPS(t *testing.T) {
	withFPS := estimateVideoTokens(BillingUsage{VideoOutputSeconds: 5, VideoWidth: 1280, VideoHeight: 720, VideoFPS: 24})
	noFPS := estimateVideoTokens(BillingUsage{VideoOutputSeconds: 5, VideoWidth: 1280, VideoHeight: 720})
	if withFPS != noFPS {
		t.Fatalf("default fps should equal 24: withFPS=%d noFPS=%d", withFPS, noFPS)
	}
}

func TestEstimateVideoTokens_IncompleteInputs(t *testing.T) {
	// 既无宽高也无可识别分辨率：仍返回 0。
	if got := estimateVideoTokens(BillingUsage{VideoOutputSeconds: 5}); got != 0 {
		t.Fatalf("missing dimensions and resolution should yield 0, got %d", got)
	}
}

// 缺宽高但给了 resolution：应按标准档位推导维度，得到非零预扣估算。
// 720p 16:9 → 1280×720，与显式传宽高结果一致（108000）。
func TestEstimateVideoTokens_DerivesFromResolution(t *testing.T) {
	got := estimateVideoTokens(BillingUsage{
		VideoOutputSeconds: 5,
		Resolution:         "720p",
		AspectRatio:        "16:9",
		VideoFPS:           24,
	})
	if got != 108000 {
		t.Fatalf("expected 108000 from derived 1280x720, got %d", got)
	}
	if got == 0 {
		t.Fatalf("text-to-video reserve must not be zero when resolution known")
	}
}

// 无宽高比时默认 16:9 横向。
func TestEstimateVideoTokens_DefaultsAspectRatio(t *testing.T) {
	got := estimateVideoTokens(BillingUsage{VideoOutputSeconds: 5, Resolution: "720p", VideoFPS: 24})
	if got != 108000 {
		t.Fatalf("expected default 16:9 to yield 108000, got %d", got)
	}
}

func TestResolveVideoDimensions(t *testing.T) {
	cases := []struct {
		res, ar string
		wantW   int
		wantH   int
		wantOK  bool
	}{
		{"480p", "16:9", 854, 480, true},
		{"720p", "16:9", 1280, 720, true},
		{"1080p", "16:9", 1920, 1080, true},
		{"4k", "16:9", 3840, 2160, true},
		{"720p", "9:16", 720, 1280, true}, // 竖向：短边仍为 720，宽高交换
		{"720p", "1:1", 720, 720, true},   // 方形
		{"720p", "", 1280, 720, true},     // 缺比例默认 16:9
		{"unknown", "16:9", 0, 0, false},  // 未知分辨率
	}
	for _, tc := range cases {
		t.Run(tc.res+"_"+tc.ar, func(t *testing.T) {
			w, h, ok := resolveVideoDimensions(tc.res, tc.ar)
			if ok != tc.wantOK || w != tc.wantW || h != tc.wantH {
				t.Fatalf("resolveVideoDimensions(%q,%q)=(%d,%d,%v), want (%d,%d,%v)",
					tc.res, tc.ar, w, h, ok, tc.wantW, tc.wantH, tc.wantOK)
			}
		})
	}
}

// 无视频输入命中 46 元/M 档，成本对齐官方示例 4.968 元。
func TestCalculateMediaCost_Seedance20_NoVideoInput(t *testing.T) {
	cost, err := CalculateMediaCost(MediaCostInput{
		Rules: SeedancePricingRules("doubao-seedance-2.0"),
		Usage: BillingUsage{
			Resolution:         "720p",
			AspectRatio:        "16:9",
			VideoOutputSeconds: 5,
			VideoWidth:         1280,
			VideoHeight:        720,
			VideoFPS:           24,
			HasVideoInput:      false,
		},
		RateMultiplier: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cost.Metric != MetricVideoToken {
		t.Fatalf("expected video_token metric, got %s", cost.Metric)
	}
	if cost.Units != 108000 {
		t.Fatalf("expected 108000 units, got %d", cost.Units)
	}
	if !cost.EstimatedUnits {
		t.Fatalf("expected estimated units (no upstream tokens)")
	}
	if !approxEqual(cost.TotalCost, 4.968) {
		t.Fatalf("expected total 4.968, got %f", cost.TotalCost)
	}
}

// 上游返回真实 token 时优先使用，不再估算。
func TestCalculateMediaCost_PrefersUpstreamTokens(t *testing.T) {
	cost, err := CalculateMediaCost(MediaCostInput{
		Rules: SeedancePricingRules("doubao-seedance-2.0"),
		Usage: BillingUsage{
			Resolution:    "720p",
			VideoTokens:   200000,
			HasVideoInput: false,
		},
		RateMultiplier: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cost.EstimatedUnits {
		t.Fatalf("should not be estimated when upstream tokens present")
	}
	if cost.Units != 200000 {
		t.Fatalf("expected 200000 units, got %d", cost.Units)
	}
}

// 含视频输入且低于最低 token：按 720p/5s 保底 194400 计费，成本 5.4432 元。
func TestCalculateMediaCost_MinUnitsFloor(t *testing.T) {
	cost, err := CalculateMediaCost(MediaCostInput{
		Rules: SeedancePricingRules("doubao-seedance-2.0"),
		Usage: BillingUsage{
			Resolution:         "720p",
			AspectRatio:        "16:9",
			VideoOutputSeconds: 5,
			VideoTokens:        1000, // 远低于保底
			HasVideoInput:      true,
		},
		RateMultiplier: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cost.MinUnitsApplied {
		t.Fatalf("expected min units applied")
	}
	if cost.Units != 194400 {
		t.Fatalf("expected floor 194400 units, got %d", cost.Units)
	}
	if !approxEqual(cost.TotalCost, 5.4432) {
		t.Fatalf("expected total 5.4432, got %f", cost.TotalCost)
	}
}

// 是否含视频输入命中不同单价档。
func TestSelectRule_VideoInputSwitchesPrice(t *testing.T) {
	rules := SeedancePricingRules("doubao-seedance-2.0")

	noVideo := selectRule(rules, BillingUsage{Resolution: "720p", HasVideoInput: false})
	withVideo := selectRule(rules, BillingUsage{Resolution: "720p", HasVideoInput: true})

	unitNoVideo, _ := PerMillion(46, CurrencyCNY)
	unitVideo, _ := PerMillion(28, CurrencyCNY)
	if noVideo == nil || !approxEqual(noVideo.UnitPrice, unitNoVideo) {
		t.Fatalf("no-video rule mismatch: %+v", noVideo)
	}
	if withVideo == nil || !approxEqual(withVideo.UnitPrice, unitVideo) {
		t.Fatalf("video rule mismatch: %+v", withVideo)
	}
}

// fast 不区分分辨率，720p/5s 无视频输入成本对齐官方 4.00 元。
func TestCalculateMediaCost_Seedance20Fast(t *testing.T) {
	cost, err := CalculateMediaCost(MediaCostInput{
		Rules: SeedancePricingRules("doubao-seedance-2.0-fast"),
		Usage: BillingUsage{
			Resolution:         "720p",
			VideoOutputSeconds: 5,
			VideoWidth:         1280,
			VideoHeight:        720,
			HasVideoInput:      false,
		},
		RateMultiplier: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approxEqual(cost.TotalCost, 3.996) {
		t.Fatalf("expected total 3.996, got %f", cost.TotalCost)
	}
}

func TestSeedancePricingRules_UnknownModel(t *testing.T) {
	if rules := SeedancePricingRules("gpt-5.1"); rules != nil {
		t.Fatalf("expected nil for non-seedance model, got %d rules", len(rules))
	}
}

func TestSeedancePricingRules_MiniAndProResolve(t *testing.T) {
	if len(SeedancePricingRules("seedance-2.0-mini")) == 0 {
		t.Fatalf("mini should resolve rules")
	}
	if len(SeedancePricingRules("doubao-seedance-1.5-pro")) == 0 {
		t.Fatalf("1.5-pro should resolve rules")
	}
}

func TestSeedancePricingRules_DreaminaAliasesResolve(t *testing.T) {
	if len(SeedancePricingRules("dreamina-seedance-2-0-260128")) == 0 {
		t.Fatalf("dreamina seedance should resolve rules")
	}
	if len(SeedancePricingRules("dreamina-seedance-2-0-fast-260128")) == 0 {
		t.Fatalf("dreamina seedance fast should resolve rules")
	}
}

// 1.5-pro 按有声/无声区分单价。
func TestCalculateMediaCost_Seedance15Pro_Audio(t *testing.T) {
	rules := SeedancePricingRules("doubao-seedance-1.5-pro")
	audio := selectRule(rules, BillingUsage{HasAudio: true})
	silent := selectRule(rules, BillingUsage{HasAudio: false})

	unitAudio, _ := PerMillion(16, CurrencyCNY)
	unitSilent, _ := PerMillion(8, CurrencyCNY)
	if audio == nil || !approxEqual(audio.UnitPrice, unitAudio) {
		t.Fatalf("audio rule mismatch: %+v", audio)
	}
	if silent == nil || !approxEqual(silent.UnitPrice, unitSilent) {
		t.Fatalf("silent rule mismatch: %+v", silent)
	}
}

func TestCalculateMediaCost_NoMatchingRule(t *testing.T) {
	rules := []MediaPricingRule{
		{Metric: MetricVideoToken, UnitPrice: 1, Resolutions: []string{"1080p"}},
	}
	_, err := CalculateMediaCost(MediaCostInput{
		Rules: rules,
		Usage: BillingUsage{Resolution: "480p"},
	})
	if !errors.Is(err, ErrNoMatchingRule) {
		t.Fatalf("expected ErrNoMatchingRule, got %v", err)
	}
}

func TestCalculateMediaCost_NegativeRateClampedToZero(t *testing.T) {
	cost, err := CalculateMediaCost(MediaCostInput{
		Rules:          SeedancePricingRules("doubao-seedance-2.0"),
		Usage:          BillingUsage{Resolution: "720p", VideoTokens: 100000, HasVideoInput: false},
		RateMultiplier: -1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cost.ActualCost != 0 {
		t.Fatalf("expected actual cost 0 for negative rate, got %f", cost.ActualCost)
	}
	if cost.TotalCost <= 0 {
		t.Fatalf("total cost should still be positive, got %f", cost.TotalCost)
	}
}

func TestResolveUnits_ByMetric(t *testing.T) {
	cases := []struct {
		name   string
		metric BillingMetric
		usage  BillingUsage
		want   int64
	}{
		{"image", MetricImageCount, BillingUsage{ImageCount: 3}, 3},
		{"video_second_roundup", MetricVideoSecond, BillingUsage{VideoOutputSeconds: 4.2}, 5},
		{"audio_second", MetricAudioSecond, BillingUsage{AudioSeconds: 10}, 10},
		{"character", MetricCharacter, BillingUsage{AudioCharacters: 120}, 120},
		{"token", MetricToken, BillingUsage{InputTokens: 10, OutputTokens: 5}, 15},
		{"request_default", MetricRequest, BillingUsage{}, 1},
		{"request_count", MetricRequest, BillingUsage{RequestCount: 4}, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := resolveUnits(tc.metric, tc.usage)
			if got != tc.want {
				t.Fatalf("%s: expected %d, got %d", tc.name, tc.want, got)
			}
		})
	}
}
