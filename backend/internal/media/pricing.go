package media

import (
	"errors"
	"strings"
)

// DefaultVideoFPS 是视频 token 估算公式使用的默认帧率。
// 火山 Seedance 公式以 24fps 为基准。
const DefaultVideoFPS = 24

// ErrNoMatchingRule 表示在给定用量下找不到任何匹配的定价规则。
var ErrNoMatchingRule = errors.New("media: no matching pricing rule")

// Currency 计价货币。内部规则以原始货币表达，货币换算是 P1 适配层的职责。
type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyCNY Currency = "CNY"
)

// MinUnitsKey 是"最低计费用量"查表的键。
//
// 火山 Seedance 含视频输入时，最低 token 与「分辨率 + 宽高比 + 输出时长」相关。
type MinUnitsKey struct {
	Resolution    string // 归一化小写，如 "720p"
	AspectRatio   string // 如 "16:9"
	OutputSeconds int    // 输出视频时长（秒）
}

// MediaPricingRule 是一条多维定价规则。
//
// 一个模型通常对应一组规则（不同分辨率 / 是否含视频输入各一条）。
// 引擎从候选规则中挑出与实际用量最匹配（最具体）的一条来计价。
type MediaPricingRule struct {
	Metric    BillingMetric // 该规则按什么度量计价
	UnitPrice float64       // 单个度量单位的价格（已是"每 1 单位"，非每百万）
	Currency  Currency      // UnitPrice 的货币

	// --- 匹配条件（零值表示不限制该维度）---
	Resolutions       []string // 命中其一即可；空表示不限分辨率
	RequireVideoInput *bool    // nil 不限；否则要求 HasVideoInput 等于该值
	RequireAudio      *bool    // nil 不限；否则要求 HasAudio 等于该值

	// MinUnits 最低计费用量保底表（可空）。仅当规则命中且查到对应键时生效。
	MinUnits map[MinUnitsKey]int64
}

// PerMillion 用"每百万单位价格"构造单个单位价格，便于按厂商定价表书写规则。
//
// 例如火山 46 元/百万 token：PerMillion(46, CurrencyCNY)。
func PerMillion(pricePerMillion float64, currency Currency) (unitPrice float64, cur Currency) {
	return pricePerMillion / 1_000_000, currency
}

// matches 判断规则是否适用于给定用量。
func (r *MediaPricingRule) matches(u BillingUsage) bool {
	if len(r.Resolutions) > 0 && !containsFold(r.Resolutions, u.Resolution) {
		return false
	}
	if r.RequireVideoInput != nil && *r.RequireVideoInput != u.HasVideoInput {
		return false
	}
	if r.RequireAudio != nil && *r.RequireAudio != u.HasAudio {
		return false
	}
	return true
}

// specificity 返回规则被满足的约束数量，用于在多条匹配规则中挑最具体的一条。
func (r *MediaPricingRule) specificity() int {
	score := 0
	if len(r.Resolutions) > 0 {
		score++
	}
	if r.RequireVideoInput != nil {
		score++
	}
	if r.RequireAudio != nil {
		score++
	}
	return score
}

// minUnitsFor 返回该用量对应的最低计费用量；无保底时返回 0。
func (r *MediaPricingRule) minUnitsFor(u BillingUsage) int64 {
	if len(r.MinUnits) == 0 {
		return 0
	}
	key := MinUnitsKey{
		Resolution:    strings.ToLower(strings.TrimSpace(u.Resolution)),
		AspectRatio:   strings.TrimSpace(u.AspectRatio),
		OutputSeconds: int(u.VideoOutputSeconds + 0.5), // 四舍五入到整秒
	}
	return r.MinUnits[key]
}

func containsFold(list []string, v string) bool {
	for _, item := range list {
		if strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(v)) {
			return true
		}
	}
	return false
}
