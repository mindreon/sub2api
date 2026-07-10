package media

import "context"

// AccountSelectInput 是媒体任务提交时的账号调度入参。
type AccountSelectInput struct {
	GroupID     int64
	Model       string
	ExcludedIDs map[int64]struct{}
}

// AccountSelection 是选中账号的凭证视图（由组合根从 service.Account 翻译而来）。
type AccountSelection struct {
	AccountID int64
	Platform  string
	APIKey    string
	BaseURL   string
}

// AccountSelector 按分组与模型选择上游账号（复用 Gateway 调度逻辑）。
type AccountSelector interface {
	Select(ctx context.Context, in AccountSelectInput) (AccountSelection, error)
}

// AccountCredentialsLoader 按账号 ID 加载提交/轮询所需的凭证。
type AccountCredentialsLoader interface {
	Load(ctx context.Context, accountID int64) (AccountSelection, error)
}

// ProviderFactory 根据账号凭证构造厂商适配器。
type ProviderFactory interface {
	NewProvider(sel AccountSelection, model string) (Provider, error)
}
