package media

// MediaCostInput 是计费引擎的统一输入。
type MediaCostInput struct {
	Rules          []MediaPricingRule // 该模型的候选定价规则
	Usage          BillingUsage       // 本次调用的实际用量
	RateMultiplier float64            // 费率倍数（<0 视为 0，避免误按 1x 扣费）
}

// MediaCost 是计费结果明细。
type MediaCost struct {
	Metric          BillingMetric // 实际命中的计费度量
	Units           int64         // 实际计费的度量数量（已应用保底）
	UnitPrice       float64       // 命中规则的单位价格
	Currency        Currency      // 计价货币
	TotalCost       float64       // UnitPrice × Units（未乘倍率）
	ActualCost      float64       // TotalCost × RateMultiplier
	MinUnitsApplied bool          // 是否因低于最低用量而按保底计费
	EstimatedUnits  bool          // Units 是否来自估算公式（而非上游真实 usage）
}

// CalculateMediaCost 计算一次多模态调用的费用。
//
// 流程：选规则 → 定数量（上游 usage 优先，公式兜底）→ 应用最低用量保底 → 乘单价与倍率。
// 找不到匹配规则时返回 ErrNoMatchingRule，由调用方决定回退策略。
func CalculateMediaCost(input MediaCostInput) (*MediaCost, error) {
	rule := selectRule(input.Rules, input.Usage)
	if rule == nil {
		return nil, ErrNoMatchingRule
	}

	units, estimated := resolveUnits(rule.Metric, input.Usage)

	minUnitsApplied := false
	if minU := rule.minUnitsFor(input.Usage); units < minU {
		units = minU
		minUnitsApplied = true
	}
	if units < 0 {
		units = 0
	}

	rate := input.RateMultiplier
	if rate < 0 {
		rate = 0
	}

	total := rule.UnitPrice * float64(units)
	return &MediaCost{
		Metric:          rule.Metric,
		Units:           units,
		UnitPrice:       rule.UnitPrice,
		Currency:        rule.Currency,
		TotalCost:       total,
		ActualCost:      total * rate,
		MinUnitsApplied: minUnitsApplied,
		EstimatedUnits:  estimated,
	}, nil
}

// selectRule 在候选规则中挑出与用量匹配且最具体的一条。
// 多条同分时取先出现者，保证结果稳定。
func selectRule(rules []MediaPricingRule, u BillingUsage) *MediaPricingRule {
	var best *MediaPricingRule
	bestScore := -1
	for i := range rules {
		r := &rules[i]
		if !r.matches(u) {
			continue
		}
		if s := r.specificity(); s > bestScore {
			bestScore = s
			best = r
		}
	}
	return best
}

// resolveUnits 根据度量确定计费数量。
// 返回的 estimated 表示数量是否来自兜底估算公式（而非上游返回的真实用量）。
func resolveUnits(metric BillingMetric, u BillingUsage) (units int64, estimated bool) {
	switch metric {
	case MetricVideoToken:
		if u.VideoTokens > 0 {
			return u.VideoTokens, false
		}
		return estimateVideoTokens(u), true
	case MetricVideoSecond:
		return roundUpSeconds(u.VideoOutputSeconds), false
	case MetricAudioSecond:
		return roundUpSeconds(u.AudioSeconds), false
	case MetricAudioToken:
		return u.AudioTokens, false
	case MetricCharacter:
		return int64(u.AudioCharacters), false
	case MetricImageCount:
		return int64(u.ImageCount), false
	case MetricToken:
		return int64(u.InputTokens + u.OutputTokens), false
	case MetricRequest:
		if u.RequestCount <= 0 {
			return 1, false
		}
		return int64(u.RequestCount), false
	}
	return 0, false
}

// estimateVideoTokens 是视频 token 的兜底估算公式：
//
//	(输入视频时长 + 输出视频时长) × 宽 × 高 × 帧率 / 1024
//
// 与火山 Seedance / OpenRouter 公式一致；帧率缺省时用 DefaultVideoFPS。
// 仅在上游未返回真实 token 时使用；准确值应以 API usage 为准。
func estimateVideoTokens(u BillingUsage) int64 {
	fps := u.VideoFPS
	if fps <= 0 {
		fps = DefaultVideoFPS
	}
	seconds := u.VideoInputSeconds + u.VideoOutputSeconds

	// 缺精确宽高时（常见于文生/图生视频只带 resolution），按标准档位推导，
	// 保证预扣估算非零，避免零余额用户绕过额度检查。
	width, height := u.VideoWidth, u.VideoHeight
	if width <= 0 || height <= 0 {
		if dw, dh, ok := resolveVideoDimensions(u.Resolution, u.AspectRatio); ok {
			width, height = dw, dh
		}
	}

	if seconds <= 0 || width <= 0 || height <= 0 {
		return 0
	}
	pixels := float64(width) * float64(height)
	return int64(seconds * pixels * float64(fps) / 1024)
}

// roundUpSeconds 将秒数向上取整为整秒（不足 1 秒按 1 秒），用于按秒计费。
func roundUpSeconds(sec float64) int64 {
	if sec <= 0 {
		return 0
	}
	whole := int64(sec)
	if float64(whole) < sec {
		whole++
	}
	return whole
}
