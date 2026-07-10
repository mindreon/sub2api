package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/media"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// gatewayMediaAccountSelector 复用 GatewayService 的账号调度为 media 包提供端口实现。
type gatewayMediaAccountSelector struct {
	gateway *service.GatewayService
}

// NewGatewayMediaAccountSelector 构造 media.AccountSelector。
func NewGatewayMediaAccountSelector(gateway *service.GatewayService) media.AccountSelector {
	return &gatewayMediaAccountSelector{gateway: gateway}
}

func (s *gatewayMediaAccountSelector) Select(ctx context.Context, in media.AccountSelectInput) (media.AccountSelection, error) {
	if s == nil || s.gateway == nil {
		return media.AccountSelection{}, media.ErrUpstreamNotWired
	}
	if in.GroupID <= 0 {
		return media.AccountSelection{}, fmt.Errorf("media: group_id required for account selection")
	}
	groupID := in.GroupID
	account, err := s.gateway.SelectAccountForModelWithExclusions(ctx, &groupID, "", in.Model, in.ExcludedIDs)
	if err != nil {
		return media.AccountSelection{}, err
	}
	return accountToMediaSelection(account), nil
}

// repoMediaAccountLoader 从账号仓储加载凭证。
type repoMediaAccountLoader struct {
	repo service.AccountRepository
}

// NewRepoMediaAccountLoader 构造 media.AccountCredentialsLoader。
func NewRepoMediaAccountLoader(repo service.AccountRepository) media.AccountCredentialsLoader {
	return &repoMediaAccountLoader{repo: repo}
}

func (l *repoMediaAccountLoader) Load(ctx context.Context, accountID int64) (media.AccountSelection, error) {
	if l == nil || l.repo == nil || accountID <= 0 {
		return media.AccountSelection{}, media.ErrUpstreamNotWired
	}
	account, err := l.repo.GetByID(ctx, accountID)
	if err != nil {
		return media.AccountSelection{}, err
	}
	if account == nil {
		return media.AccountSelection{}, fmt.Errorf("media: account %d not found", accountID)
	}
	return accountToMediaSelection(account), nil
}

func accountToMediaSelection(account *service.Account) media.AccountSelection {
	sel := media.AccountSelection{
		AccountID: account.ID,
		Platform:  account.Platform,
		APIKey:    strings.TrimSpace(account.GetCredential("api_key")),
	}
	if account.Extra != nil {
		if u, ok := account.Extra["base_url"].(string); ok {
			sel.BaseURL = strings.TrimSpace(u)
		}
	}
	return sel
}
