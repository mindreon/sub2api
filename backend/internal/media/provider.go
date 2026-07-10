package media

import (
	"context"
	"errors"
)

// 本文件定义"厂商适配器"端口。P3 只声明契约；具体的火山方舟 / OpenRouter
// 实现放在 P5 的 internal/media/providers/ 子包，通过 ProviderRegistry 注入，
// task_service.go / poller.go 面向本接口编排，不知道任何厂商细节。

// ProviderState 是厂商侧异步任务的粗粒度状态。
type ProviderState string

const (
	ProviderInProgress ProviderState = "in_progress"
	ProviderSucceeded  ProviderState = "succeeded"
	ProviderFailed     ProviderState = "failed"
)

// ProviderStatus 是一次状态查询的结果。
//
// Usage 是结算所需的用量（优先使用厂商返回的真实 usage，如
// completion_tokens），交给 Ledger.Settle 重新计费，保证价格计算逻辑
// 只有一处（引擎），不会在各厂商适配器里重复实现。
type ProviderStatus struct {
	State        ProviderState
	Usage        BillingUsage
	RawUsage     map[string]any // 原始 usage，原样存入 Task.UpstreamUsage 便于审计
	ResultURL    string         // 厂商侧生成结果地址（视频 URL），成功时提取；可能有过期时限
	ErrorMessage string
}

// Provider 是厂商适配器需要实现的最小接口。
type Provider interface {
	// Submit 把任务提交到厂商侧，返回厂商任务 ID。
	Submit(ctx context.Context, task *Task) (upstreamTaskID string, err error)
	// QueryStatus 查询厂商侧任务当前状态。
	QueryStatus(ctx context.Context, task *Task) (*ProviderStatus, error)
}

// ErrProviderNotFound 表示没有为该模型注册厂商适配器。
var ErrProviderNotFound = errors.New("media: no provider registered for model")

// ProviderRegistry 按模型名解析出对应的厂商适配器。
type ProviderRegistry interface {
	ProviderFor(model string) (Provider, error)
}

// ProviderRegistryFunc 让普通函数满足 ProviderRegistry。
type ProviderRegistryFunc func(model string) (Provider, error)

func (f ProviderRegistryFunc) ProviderFor(model string) (Provider, error) { return f(model) }

// StaticProviderRegistry 是按模型名精确匹配的最简单实现，供 P3 单测与
// P5 少量模型场景直接使用。
type StaticProviderRegistry map[string]Provider

func (r StaticProviderRegistry) ProviderFor(model string) (Provider, error) {
	if p, ok := r[model]; ok {
		return p, nil
	}
	return nil, ErrProviderNotFound
}
