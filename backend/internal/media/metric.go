// Package media 实现多模态（视频/音频/图片）模型的计费能力。
//
// 设计原则（见 docs/design/multimodal-async-billing-2026-07-01.md）：
//   - 本包与上游代码单向解耦：media 可以调用上游 service，但上游任何包都不得 import 本包。
//   - P0 阶段（当前文件所属）只包含纯计费引擎，不依赖任何上游类型，可独立单测。
package media

// BillingMetric 表示"计价的基本单位是什么"。
//
// 与上游的 BillingMode（token/per_request/image）是不同维度的概念：
// BillingMode 描述"计费方式"，BillingMetric 描述"按什么度量计价"。
// 这样视频按 token、按秒，音频按秒、按字符等都能用同一套引擎表达。
type BillingMetric string

const (
	// MetricToken 按文本 token 计价（对齐上游 token 模式，便于统一）。
	MetricToken BillingMetric = "token"
	// MetricImageCount 按生成的图片张数计价。
	MetricImageCount BillingMetric = "image_count"
	// MetricVideoToken 按视频 token 计价（火山 Seedance / OpenRouter 视频）。
	MetricVideoToken BillingMetric = "video_token"
	// MetricVideoSecond 按输出视频秒数计价（部分平台按秒定价）。
	MetricVideoSecond BillingMetric = "video_second"
	// MetricAudioSecond 按音频秒数计价（TTS / 语音生成）。
	MetricAudioSecond BillingMetric = "audio_second"
	// MetricAudioToken 按音频 token 计价。
	MetricAudioToken BillingMetric = "audio_token"
	// MetricCharacter 按字符数计价（部分 TTS 按字符）。
	MetricCharacter BillingMetric = "character"
	// MetricRequest 按次计价（对齐上游 per_request）。
	MetricRequest BillingMetric = "request"
)

// IsValid 判断度量是否为已知合法值。
func (m BillingMetric) IsValid() bool {
	switch m {
	case MetricToken, MetricImageCount, MetricVideoToken, MetricVideoSecond,
		MetricAudioSecond, MetricAudioToken, MetricCharacter, MetricRequest:
		return true
	}
	return false
}

// BillingUsage 描述一次调用产生的全部可计费维度。
//
// 不同模态的字段可以同时存在（例如一条视频既有 VideoTokens 又有音频）。
// 引擎根据被选中规则的 BillingMetric 决定实际取用哪些字段。
type BillingUsage struct {
	// --- Token 维度（文本，对齐上游语义）---
	InputTokens  int
	OutputTokens int

	// --- 图片维度 ---
	ImageCount int

	// --- 视频维度 ---
	// VideoTokens 为上游返回的真实计费 token（如火山 usage.completion_tokens）。
	// > 0 时引擎优先使用它，否则回退到估算公式。
	VideoTokens        int64
	VideoInputSeconds  float64 // 输入/参考视频时长（秒），纯文生/图生时为 0
	VideoOutputSeconds float64 // 输出视频时长（秒）
	VideoWidth         int     // 输出视频宽（像素）
	VideoHeight        int     // 输出视频高（像素）
	VideoFPS           int     // 输出视频帧率，0 时引擎按默认帧率处理

	// --- 音频维度 ---
	AudioSeconds    float64
	AudioTokens     int64
	AudioCharacters int

	// --- 通用 / 分类维度 ---
	RequestCount int    // 按次计费时的次数，<=0 视为 1
	Resolution   string // 分辨率档，如 "480p"/"720p"/"1080p"/"4k"
	AspectRatio  string // 宽高比，如 "16:9"（用于最低 token 保底查表）
	HasVideoInput bool  // 是否包含视频输入（影响单价档：含/不含）
	HasAudio      bool  // 是否生成音频（影响单价档：有声/无声）
}
