package media

import "strings"

// 火山方舟 Seedance 系列定价（元/百万 tokens，在线推理）。
// 数据来源：火山方舟「模型价格」文档（docs/design/multimodal-async-billing-2026-07-01.md 附录 A）。
// 计费度量统一为 MetricVideoToken；token 用量优先取上游 usage.completion_tokens，
// 缺失时由引擎按 estimateVideoTokens 估算。
const (
	seedance20Res480720NoVideo = 46.0
	seedance20Res480720Video   = 28.0
	seedance20Res1080NoVideo   = 51.0
	seedance20Res1080Video     = 31.0
	seedance20Res4kNoVideo     = 26.0
	seedance20Res4kVideo       = 16.0

	seedance20FastNoVideo = 37.0
	seedance20FastVideo   = 22.0

	seedance20MiniNoVideo = 23.0
	seedance20MiniVideo   = 14.0

	seedance15ProAudio   = 16.0
	seedance15ProNoAudio = 8.0
)

// boolPtr 返回指向给定 bool 的指针，用于构造规则的可选匹配条件。
func boolPtr(b bool) *bool { return &b }

// SeedancePricingRules 返回给定 Seedance 模型的定价规则集合。
// 模型名大小写不敏感，兼容带/不带 "doubao-" / "dreamina-" / "bytedance/" 前缀。
// 返回 nil 表示不是已知的 Seedance 模型。
func SeedancePricingRules(model string) []MediaPricingRule {
	m := normalizeModelKey(model)
	m = strings.TrimPrefix(m, "doubao-")
	m = strings.TrimPrefix(m, "dreamina-")
	m = strings.TrimPrefix(m, "bytedance/")

	switch {
	case strings.HasPrefix(m, "seedance-2.0-fast"), strings.HasPrefix(m, "seedance-2-0-fast"):
		return flatVideoInputRules(seedance20FastNoVideo, seedance20FastVideo, minTokensSeedance20())
	case strings.HasPrefix(m, "seedance-2.0-mini"), strings.HasPrefix(m, "seedance-2-0-mini"):
		return flatVideoInputRules(seedance20MiniNoVideo, seedance20MiniVideo, nil)
	case strings.HasPrefix(m, "seedance-2.0"), strings.HasPrefix(m, "seedance-2-0"):
		return seedance20Rules()
	case strings.HasPrefix(m, "seedance-1.5-pro"), strings.HasPrefix(m, "seedance-1-5-pro"):
		return seedance15ProRules()
	}
	return nil
}

// openRouterVideoTokenUSDPerM OpenRouter Seedance 2.0 video token 单价（美元/百万 tokens）。
const openRouterVideoTokenUSDPerM = 7.0

// OpenRouterSeedancePricingRules 返回 OpenRouter 侧 Seedance 模型的 USD 定价规则。
func OpenRouterSeedancePricingRules(model string) []MediaPricingRule {
	if !isSeedanceFamily(normalizeModelKey(model)) {
		return nil
	}
	unit, _ := PerMillion(openRouterVideoTokenUSDPerM, CurrencyUSD)
	return []MediaPricingRule{{
		Metric:    MetricVideoToken,
		UnitPrice: unit,
		Currency:  CurrencyUSD,
	}}
}

// seedance20Rules 构造 Seedance 2.0 的分辨率 × 是否含视频输入定价矩阵。
func seedance20Rules() []MediaPricingRule {
	res480720 := []string{"480p", "720p"}
	res1080 := []string{"1080p"}
	res4k := []string{"4k"}

	return []MediaPricingRule{
		videoTokenRule(res480720, false, seedance20Res480720NoVideo, nil),
		videoTokenRule(res480720, true, seedance20Res480720Video, minTokensSeedance20()),
		videoTokenRule(res1080, false, seedance20Res1080NoVideo, nil),
		videoTokenRule(res1080, true, seedance20Res1080Video, nil),
		videoTokenRule(res4k, false, seedance20Res4kNoVideo, nil),
		videoTokenRule(res4k, true, seedance20Res4kVideo, nil),
	}
}

// flatVideoInputRules 构造仅按"是否含视频输入"区分、不分分辨率的两条规则（fast / mini）。
func flatVideoInputRules(noVideoCNYPerM, videoCNYPerM float64, minTokens map[MinUnitsKey]int64) []MediaPricingRule {
	return []MediaPricingRule{
		videoTokenRule(nil, false, noVideoCNYPerM, nil),
		videoTokenRule(nil, true, videoCNYPerM, minTokens),
	}
}

// seedance15ProRules 按"是否有声"区分（1.5-pro 无含视频输入价差）。
func seedance15ProRules() []MediaPricingRule {
	unitAudio, _ := PerMillion(seedance15ProAudio, CurrencyCNY)
	unitNoAudio, _ := PerMillion(seedance15ProNoAudio, CurrencyCNY)
	return []MediaPricingRule{
		{Metric: MetricVideoToken, UnitPrice: unitAudio, Currency: CurrencyCNY, RequireAudio: boolPtr(true)},
		{Metric: MetricVideoToken, UnitPrice: unitNoAudio, Currency: CurrencyCNY, RequireAudio: boolPtr(false)},
	}
}

// videoTokenRule 是构造一条 video_token 规则的便捷函数（价格入参为元/百万 tokens）。
func videoTokenRule(resolutions []string, hasVideoInput bool, cnyPerMillion float64, minTokens map[MinUnitsKey]int64) MediaPricingRule {
	unit, _ := PerMillion(cnyPerMillion, CurrencyCNY)
	return MediaPricingRule{
		Metric:            MetricVideoToken,
		UnitPrice:         unit,
		Currency:          CurrencyCNY,
		Resolutions:       resolutions,
		RequireVideoInput: boolPtr(hasVideoInput),
		MinUnits:          minTokens,
	}
}

// minTokensSeedance20 返回 Seedance 2.0/2.0-fast 含视频输入时的最低 token 表（16:9）。
// 键覆盖 480p 与 720p、输出 4~15 秒。数据来源见设计文档附录。
func minTokensSeedance20() map[MinUnitsKey]int64 {
	min480 := map[int]int64{
		4: 70308, 5: 90396, 6: 100440, 7: 120528, 8: 140616, 9: 150660,
		10: 170748, 11: 190836, 12: 200880, 13: 220968, 14: 241056, 15: 251100,
	}
	min720 := map[int]int64{
		4: 151200, 5: 194400, 6: 216000, 7: 259200, 8: 302400, 9: 324000,
		10: 367200, 11: 410400, 12: 432000, 13: 475200, 14: 518400, 15: 540000,
	}
	out := make(map[MinUnitsKey]int64, len(min480)+len(min720))
	for sec, v := range min480 {
		out[MinUnitsKey{Resolution: "480p", AspectRatio: "16:9", OutputSeconds: sec}] = v
	}
	for sec, v := range min720 {
		out[MinUnitsKey{Resolution: "720p", AspectRatio: "16:9", OutputSeconds: sec}] = v
	}
	return out
}
