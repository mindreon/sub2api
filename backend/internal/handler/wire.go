package handler

import (
	"context"
	"log/slog"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/media"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/internal/voucher"

	"github.com/google/wire"
)

// ProvideAdminHandlers creates the AdminHandlers struct
func ProvideAdminHandlers(
	dashboardHandler *admin.DashboardHandler,
	userHandler *admin.UserHandler,
	groupHandler *admin.GroupHandler,
	accountHandler *admin.AccountHandler,
	announcementHandler *admin.AnnouncementHandler,
	dataManagementHandler *admin.DataManagementHandler,
	backupHandler *admin.BackupHandler,
	oauthHandler *admin.OAuthHandler,
	openaiOAuthHandler *admin.OpenAIOAuthHandler,
	geminiOAuthHandler *admin.GeminiOAuthHandler,
	antigravityOAuthHandler *admin.AntigravityOAuthHandler,
	grokOAuthHandler *admin.GrokOAuthHandler,
	proxyHandler *admin.ProxyHandler,
	redeemHandler *admin.RedeemHandler,
	promoHandler *admin.PromoHandler,
	settingHandler *admin.SettingHandler,
	opsHandler *admin.OpsHandler,
	systemHandler *admin.SystemHandler,
	subscriptionHandler *admin.SubscriptionHandler,
	usageHandler *admin.UsageHandler,
	userAttributeHandler *admin.UserAttributeHandler,
	errorPassthroughHandler *admin.ErrorPassthroughHandler,
	tlsFingerprintProfileHandler *admin.TLSFingerprintProfileHandler,
	apiKeyHandler *admin.AdminAPIKeyHandler,
	scheduledTestHandler *admin.ScheduledTestHandler,
	channelHandler *admin.ChannelHandler,
	channelMonitorHandler *admin.ChannelMonitorHandler,
	channelMonitorTemplateHandler *admin.ChannelMonitorRequestTemplateHandler,
	contentModerationHandler *admin.ContentModerationHandler,
	paymentHandler *admin.PaymentHandler,
	affiliateHandler *admin.AffiliateHandler,
	distributionHandler *admin.DistributionHandler,
	catalogModelHandler *admin.CatalogModelHandler,
	voucherAdminHandler *admin.VoucherHandler,
	complianceHandler *admin.ComplianceHandler,
	mediaAdminHandler *admin.MediaHandler,
) *AdminHandlers {
	return &AdminHandlers{
		Dashboard:              dashboardHandler,
		User:                   userHandler,
		Group:                  groupHandler,
		Account:                accountHandler,
		Announcement:           announcementHandler,
		DataManagement:         dataManagementHandler,
		Backup:                 backupHandler,
		OAuth:                  oauthHandler,
		OpenAIOAuth:            openaiOAuthHandler,
		GeminiOAuth:            geminiOAuthHandler,
		AntigravityOAuth:       antigravityOAuthHandler,
		GrokOAuth:              grokOAuthHandler,
		Proxy:                  proxyHandler,
		Redeem:                 redeemHandler,
		Promo:                  promoHandler,
		Setting:                settingHandler,
		Ops:                    opsHandler,
		System:                 systemHandler,
		Subscription:           subscriptionHandler,
		Usage:                  usageHandler,
		UserAttribute:          userAttributeHandler,
		ErrorPassthrough:       errorPassthroughHandler,
		TLSFingerprintProfile:  tlsFingerprintProfileHandler,
		APIKey:                 apiKeyHandler,
		ScheduledTest:          scheduledTestHandler,
		Channel:                channelHandler,
		ChannelMonitor:         channelMonitorHandler,
		ChannelMonitorTemplate: channelMonitorTemplateHandler,
		ContentModeration:      contentModerationHandler,
		Payment:                paymentHandler,
		Affiliate:              affiliateHandler,
		Distribution:           distributionHandler,
		CatalogModel:           catalogModelHandler,
		Voucher:                voucherAdminHandler,
		Compliance:             complianceHandler,
		Media:                  mediaAdminHandler,
	}
}

// ProvideSystemHandler creates admin.SystemHandler with UpdateService
func ProvideSystemHandler(updateService *service.UpdateService, lockService *service.SystemOperationLockService) *admin.SystemHandler {
	return admin.NewSystemHandler(updateService, lockService)
}

// ProvideSettingHandler creates SettingHandler with version from BuildInfo
func ProvideSettingHandler(settingService *service.SettingService, buildInfo BuildInfo, notificationEmailService *service.NotificationEmailService) *SettingHandler {
	h := NewSettingHandler(settingService, buildInfo.Version)
	h.SetNotificationEmailService(notificationEmailService)
	return h
}

// ProvideAdminSettingHandler creates admin.SettingHandler with notification template APIs.
func ProvideAdminSettingHandler(settingService *service.SettingService, emailService *service.EmailService, turnstileService *service.TurnstileService, opsService *service.OpsService, paymentConfigService *service.PaymentConfigService, paymentService *service.PaymentService, userAttributeService *service.UserAttributeService, notificationEmailService *service.NotificationEmailService) *admin.SettingHandler {
	h := admin.NewSettingHandler(settingService, emailService, turnstileService, opsService, paymentConfigService, paymentService, userAttributeService)
	h.SetNotificationEmailService(notificationEmailService)
	return h
}

// ProvideAdminAPIKeyHandler creates the admin API key handler with user-key creation support.
func ProvideAdminAPIKeyHandler(adminService service.AdminService, apiKeyService *service.APIKeyService) *admin.AdminAPIKeyHandler {
	return admin.NewAdminAPIKeyHandler(adminService, apiKeyService)
}

// ProvideVoucherService creates the KVoucher PIN purchase service.
func ProvideVoucherService(entClient *dbent.Client, settingRepo service.SettingRepository) *voucher.Service {
	return voucher.NewService(entClient, voucher.NewConfigStore(settingRepo))
}

// --- 多模态异步计费组合根（P4）---
//
// media 包对上游零 import（见 internal/media 包注释）。这里是唯一把
// media 的端口接口（TaskStore/HoldStore/BalanceReader/Charger）接到
// 上游具体实现（ent/UsageBillingRepository/BillingCacheService）的地方。
// 这些 Provide 函数全部是新增代码，不会与上游产生 merge 冲突。

// defaultMediaCNYToUSDRate 是 CNY→USD 的兜底汇率（Seedance 单价以人民币计价）。
// TODO(P5): 改为从 settingService 读取可配置汇率，而不是编译期常量。
const defaultMediaCNYToUSDRate = 0.14

// defaultMediaPollInterval 是轮询 Worker 检查未终结任务的间隔。
const defaultMediaPollInterval = 20 * time.Second

// ProvideMediaTaskStore 用 ent 客户端构造任务存储。
func ProvideMediaTaskStore(client *dbent.Client) media.TaskStore {
	return media.NewEntTaskStore(client)
}

// ProvideMediaHoldStore 用 ent 客户端构造预扣台账存储。
func ProvideMediaHoldStore(client *dbent.Client) media.HoldStore {
	return media.NewEntHoldStore(client)
}

// ProvideMediaBalanceReader 把上游余额读取包装成 HoldAware 版本：
// 可用额度 = 上游真实余额 − 本包未结算的预扣之和。
func ProvideMediaBalanceReader(billingCache *service.BillingCacheService, holds media.HoldStore) media.BalanceReader {
	raw := media.NewBalanceReader(billingCache)
	return media.NewHoldAwareBalance(raw, holds)
}

type mediaBillingApplier interface {
	Apply(ctx context.Context, cmd *service.UsageBillingCommand) (*service.UsageBillingApplyResult, error)
}

type mediaUsageLogWriter interface {
	Create(ctx context.Context, log *service.UsageLog) (bool, error)
}

type mediaUsageLogBestEffortWriter interface {
	CreateBestEffort(ctx context.Context, log *service.UsageLog) error
}

type mediaBillingCacheUpdater interface {
	QueueDeductBalance(userID int64, amount float64)
	InvalidateUserBalance(ctx context.Context, userID int64) error
	QueueUpdateSubscriptionUsage(userID, groupID int64, costUSD float64)
	QueueUpdateAPIKeyRateLimitUsage(apiKeyID int64, cost float64)
}

type mediaAPIKeyGetter interface {
	GetByID(ctx context.Context, id int64) (*service.APIKey, error)
}

type mediaAccountGetter interface {
	GetByID(ctx context.Context, id int64) (*service.Account, error)
}

type mediaGroupGetter interface {
	GetByID(ctx context.Context, id int64) (*service.Group, error)
}

type mediaSubscriptionGetter interface {
	GetActiveByUserIDAndGroupID(ctx context.Context, userID, groupID int64) (*service.UserSubscription, error)
}

type mediaAuthCacheInvalidator interface {
	InvalidateAuthCacheByKey(ctx context.Context, key string)
}

type mediaLastUsedScheduler interface {
	ScheduleLastUsedUpdate(accountID int64)
}

type mediaChargerDeps struct {
	billingRepo       mediaBillingApplier
	usageLogs         mediaUsageLogWriter
	billingCache      mediaBillingCacheUpdater
	apiKeys           mediaAPIKeyGetter
	accounts          mediaAccountGetter
	groups            mediaGroupGetter
	subscriptions     mediaSubscriptionGetter
	authInvalidator   mediaAuthCacheInvalidator
	lastUsedScheduler mediaLastUsedScheduler
}

// ProvideMediaCharger 把 media.ChargeRequest 翻译为上游 UsageBillingCommand，
// 通过已有的 UsageBillingRepository.Apply 完成真实扣费（幂等、进 usage_logs 报表）。
func ProvideMediaCharger(
	repo service.UsageBillingRepository,
	usageLogs service.UsageLogRepository,
	billingCache *service.BillingCacheService,
	apiKeys service.APIKeyRepository,
	accounts service.AccountRepository,
	groups service.GroupRepository,
	subscriptions service.UserSubscriptionRepository,
	authInvalidator service.APIKeyAuthCacheInvalidator,
	deferredService *service.DeferredService,
) media.Charger {
	return newMediaCharger(mediaChargerDeps{
		billingRepo:       repo,
		usageLogs:         usageLogs,
		billingCache:      billingCache,
		apiKeys:           apiKeys,
		accounts:          accounts,
		groups:            groups,
		subscriptions:     subscriptions,
		authInvalidator:   authInvalidator,
		lastUsedScheduler: deferredService,
	})
}

func newMediaCharger(deps mediaChargerDeps) media.Charger {
	return media.ChargerFunc(func(ctx context.Context, req media.ChargeRequest) (*media.ChargeResult, error) {
		if deps.billingRepo == nil {
			return nil, media.ErrUpstreamNotWired
		}
		apiKey, err := loadMediaChargeAPIKey(ctx, deps.apiKeys, req.APIKeyID)
		if err != nil {
			return nil, err
		}
		account, err := loadMediaChargeAccount(ctx, deps.accounts, req.AccountID)
		if err != nil {
			return nil, err
		}
		billingType, subscriptionID, err := resolveMediaChargeBillingType(ctx, deps, req)
		if err != nil {
			return nil, err
		}
		cmd := &service.UsageBillingCommand{
			RequestID:      req.RequestID,
			APIKeyID:       req.APIKeyID,
			UserID:         req.UserID,
			AccountID:      req.AccountID,
			SubscriptionID: subscriptionID,
			Model:          req.Model,
			MediaType:      req.MediaType,
			BillingType:    billingType,
		}
		if account != nil {
			cmd.AccountType = account.Type
		}
		if req.Metric == media.MetricVideoToken {
			cmd.OutputTokens = intFromInt64(req.Units)
		}
		if billingType == service.BillingTypeSubscription && subscriptionID != nil {
			cmd.SubscriptionCost = req.ActualCost
		} else {
			cmd.BalanceCost = req.ActualCost
		}
		if req.ActualCost > 0 && apiKey != nil {
			if apiKey.Quota > 0 {
				cmd.APIKeyQuotaCost = req.ActualCost
			}
			if apiKey.HasRateLimits() {
				cmd.APIKeyRateLimitCost = req.ActualCost
			}
		}
		accountRateMultiplier := 1.0
		if account != nil {
			accountRateMultiplier = account.BillingRateMultiplier()
			if req.CostBillingCurrency > 0 && account.IsAPIKeyOrBedrock() && account.HasAnyQuotaLimit() {
				cmd.AccountQuotaCost = req.CostBillingCurrency * accountRateMultiplier
			}
		}
		cmd.Normalize()

		result, err := deps.billingRepo.Apply(ctx, cmd)
		if err != nil {
			return nil, err
		}
		applied := result != nil && result.Applied
		if applied {
			syncMediaChargeCaches(ctx, deps, req, apiKey, billingType, result)
			scheduleMediaAccountLastUsed(deps.lastUsedScheduler, req.AccountID)
		}
		writeMediaUsageLog(ctx, deps.usageLogs, buildMediaUsageLog(req, billingType, subscriptionID, accountRateMultiplier))
		var newBalance *float64
		if result != nil {
			newBalance = result.NewBalance
		}
		return &media.ChargeResult{Applied: applied, NewBalance: newBalance}, nil
	})
}

func scheduleMediaAccountLastUsed(scheduler mediaLastUsedScheduler, accountID int64) {
	if scheduler == nil || accountID <= 0 {
		return
	}
	scheduler.ScheduleLastUsedUpdate(accountID)
}

func loadMediaChargeAPIKey(ctx context.Context, repo mediaAPIKeyGetter, apiKeyID int64) (*service.APIKey, error) {
	if repo == nil || apiKeyID <= 0 {
		return nil, nil
	}
	return repo.GetByID(ctx, apiKeyID)
}

func loadMediaChargeAccount(ctx context.Context, repo mediaAccountGetter, accountID int64) (*service.Account, error) {
	if repo == nil || accountID <= 0 {
		return nil, nil
	}
	return repo.GetByID(ctx, accountID)
}

func resolveMediaChargeBillingType(ctx context.Context, deps mediaChargerDeps, req media.ChargeRequest) (int8, *int64, error) {
	if req.IsSubscription && req.SubscriptionID != nil {
		return service.BillingTypeSubscription, req.SubscriptionID, nil
	}
	if req.GroupID == nil || *req.GroupID <= 0 || deps.groups == nil {
		return service.BillingTypeBalance, nil, nil
	}
	group, err := deps.groups.GetByID(ctx, *req.GroupID)
	if err != nil {
		return service.BillingTypeBalance, nil, err
	}
	if group == nil || !group.IsSubscriptionType() {
		return service.BillingTypeBalance, nil, nil
	}
	if deps.subscriptions == nil {
		return service.BillingTypeBalance, nil, media.ErrUpstreamNotWired
	}
	sub, err := deps.subscriptions.GetActiveByUserIDAndGroupID(ctx, req.UserID, *req.GroupID)
	if err != nil {
		return service.BillingTypeBalance, nil, err
	}
	if sub == nil {
		return service.BillingTypeBalance, nil, service.ErrSubscriptionInvalid
	}
	return service.BillingTypeSubscription, &sub.ID, nil
}

func syncMediaChargeCaches(ctx context.Context, deps mediaChargerDeps, req media.ChargeRequest, apiKey *service.APIKey, billingType int8, result *service.UsageBillingApplyResult) {
	if deps.billingCache == nil || req.ActualCost <= 0 {
		return
	}
	if billingType == service.BillingTypeSubscription {
		if req.GroupID != nil && *req.GroupID > 0 {
			deps.billingCache.QueueUpdateSubscriptionUsage(req.UserID, *req.GroupID, req.ActualCost)
		}
	} else if result != nil && result.NewBalance != nil && *result.NewBalance <= 0 {
		if err := deps.billingCache.InvalidateUserBalance(ctx, req.UserID); err != nil {
			slog.Warn("invalidate media balance cache after deduction failed", "user_id", req.UserID, "error", err)
		}
	} else {
		deps.billingCache.QueueDeductBalance(req.UserID, req.ActualCost)
	}

	if apiKey != nil && apiKey.HasRateLimits() {
		deps.billingCache.QueueUpdateAPIKeyRateLimitUsage(apiKey.ID, req.ActualCost)
	}
	if result != nil && result.APIKeyQuotaExhausted && deps.authInvalidator != nil && apiKey != nil && apiKey.Key != "" {
		deps.authInvalidator.InvalidateAuthCacheByKey(ctx, apiKey.Key)
	}
}

func buildMediaUsageLog(req media.ChargeRequest, billingType int8, subscriptionID *int64, accountRateMultiplier float64) *service.UsageLog {
	mediaType := req.MediaType
	billingMode := string(service.BillingModeToken)
	log := &service.UsageLog{
		UserID:                req.UserID,
		APIKeyID:              req.APIKeyID,
		AccountID:             req.AccountID,
		RequestID:             req.RequestID,
		Model:                 req.Model,
		RequestedModel:        req.Model,
		GroupID:               req.GroupID,
		SubscriptionID:        subscriptionID,
		OutputTokens:          mediaUsageTokens(req),
		OutputCost:            req.CostBillingCurrency,
		TotalCost:             req.CostBillingCurrency,
		ActualCost:            req.ActualCost,
		RateMultiplier:        req.RateMultiplier,
		AccountRateMultiplier: &accountRateMultiplier,
		BillingType:           billingType,
		BillingMode:           &billingMode,
		RequestType:           service.RequestTypeSync,
		Stream:                false,
		MediaType:             &mediaType,
		CreatedAt:             time.Now(),
	}
	log.SyncRequestTypeAndLegacyFields()
	return log
}

func mediaUsageTokens(req media.ChargeRequest) int {
	if req.Metric != media.MetricVideoToken {
		return 0
	}
	return intFromInt64(req.Units)
}

func intFromInt64(v int64) int {
	if v <= 0 {
		return 0
	}
	maxInt := int64(^uint(0) >> 1)
	if v > maxInt {
		return int(maxInt)
	}
	return int(v)
}

func writeMediaUsageLog(ctx context.Context, repo mediaUsageLogWriter, usageLog *service.UsageLog) {
	if repo == nil || usageLog == nil {
		return
	}
	if bestEffort, ok := repo.(mediaUsageLogBestEffortWriter); ok {
		if err := bestEffort.CreateBestEffort(ctx, usageLog); err == nil {
			return
		} else {
			slog.Warn("media usage log best-effort write failed", "request_id", usageLog.RequestID, "error", err)
		}
	}
	if _, err := repo.Create(ctx, usageLog); err != nil {
		slog.Warn("media usage log write failed", "request_id", usageLog.RequestID, "error", err)
	}
}

// mediaSettingsKV 把 service.SettingRepository 适配为 media.SettingsKV。
type mediaSettingsKV struct {
	repo service.SettingRepository
}

func (a mediaSettingsKV) GetValue(ctx context.Context, key string) (string, error) {
	return a.repo.GetValue(ctx, key)
}

func (a mediaSettingsKV) Set(ctx context.Context, key, value string) error {
	return a.repo.Set(ctx, key, value)
}

// ProvideMediaConfigStore 从 setting 表加载 media_cny_to_usd_rate。
func ProvideMediaConfigStore(settingRepo service.SettingRepository) *media.ConfigStore {
	fallback := media.BillingConfig{CNYToUSDRate: defaultMediaCNYToUSDRate}
	return media.NewConfigStore(mediaSettingsKV{repo: settingRepo}, fallback)
}

// ProvideMediaQuoter 组装计费门面：内置 Seedance 单价表 + 可配置汇率 + 余额检查。
func ProvideMediaQuoter(balance media.BalanceReader, configs *media.ConfigStore) *media.Quoter {
	converter := media.NewConfigBackedConverter(configs, media.CurrencyUSD, defaultMediaCNYToUSDRate)
	rules := media.NewConfigBackedRuleProvider(configs, media.SeedanceRuleProvider())
	return media.NewQuoter(rules, converter, balance, media.CurrencyUSD)
}

// media 结果转存 S3 配置的 setting 键。
const (
	mediaAssetBucketKey         = "media_asset_bucket"
	mediaAssetEndpointKey       = "media_asset_endpoint"
	mediaAssetRegionKey         = "media_asset_region"
	mediaAssetAccessKeyIDKey    = "media_asset_access_key_id"
	mediaAssetSecretKey         = "media_asset_secret_access_key"
	mediaAssetPrefixKey         = "media_asset_prefix"
	mediaAssetPublicBaseURLKey  = "media_asset_public_base_url"
	mediaAssetForcePathStyleKey = "media_asset_force_path_style"
)

// ProvideMediaAssetStore 从 setting 表读取媒体转存 S3 配置并构造 AssetStore。
// 未配置（缺 bucket/凭证）时返回 nil，结算逻辑自动降级为保留上游直链。
func ProvideMediaAssetStore(settingRepo service.SettingRepository) media.AssetStore {
	ctx := context.Background()
	get := func(key string) string {
		v, err := settingRepo.GetValue(ctx, key)
		if err != nil {
			return ""
		}
		return v
	}
	cfg := repository.MediaAssetS3Config{
		Endpoint:        get(mediaAssetEndpointKey),
		Region:          get(mediaAssetRegionKey),
		Bucket:          get(mediaAssetBucketKey),
		AccessKeyID:     get(mediaAssetAccessKeyIDKey),
		SecretAccessKey: get(mediaAssetSecretKey),
		Prefix:          get(mediaAssetPrefixKey),
		PublicBaseURL:   get(mediaAssetPublicBaseURLKey),
		ForcePathStyle:  get(mediaAssetForcePathStyleKey) == "true",
	}
	store, err := repository.NewMediaAssetS3Store(ctx, cfg)
	if err != nil {
		return nil // 配置有误时降级，不阻断服务启动
	}
	return store
}

// ProvideMediaReservation 用 ent 事务实现严格预扣（原子建库 + 锁定重算余额）。
func ProvideMediaReservation(client *dbent.Client) media.Reservation {
	return media.NewEntReservation(client)
}

// ProvideMediaLedger 组装预扣/结算/释放三段式台账编排器。
// 注入 AssetStore（结果转存）与严格预扣事务；rawBalance 用未扣预扣的上游余额，
// held 由事务内重算，避免与 HoldAwareBalance 双减。
func ProvideMediaLedger(
	quoter *media.Quoter,
	charger media.Charger,
	tasks media.TaskStore,
	holds media.HoldStore,
	balance media.BalanceReader,
	assets media.AssetStore,
	reservation media.Reservation,
	billingCache *service.BillingCacheService,
) *media.Ledger {
	rawBalance := media.NewBalanceReader(billingCache)
	return media.NewLedger(quoter, charger, tasks, holds, balance).
		WithAssetStore(assets).
		WithReservation(reservation, rawBalance)
}

// ProvideMediaAccountSelector 复用 Gateway 账号调度。
func ProvideMediaAccountSelector(gateway *service.GatewayService) media.AccountSelector {
	return NewGatewayMediaAccountSelector(gateway)
}

// ProvideMediaAccountLoader 从账号仓储加载凭证。
func ProvideMediaAccountLoader(repo service.AccountRepository) media.AccountCredentialsLoader {
	return NewRepoMediaAccountLoader(repo)
}

// ProvideMediaProviderFactory 根据账号凭证构造厂商适配器（含 env 回退）。
func ProvideMediaProviderFactory() media.ProviderFactory {
	return NewEnvFallbackMediaProviderFactory(NewAccountMediaProviderFactory())
}

// ProvideMediaTaskService 组装任务提交编排器（预扣 → 提交上游 → 标记进行中）。
func ProvideMediaTaskService(
	ledger *media.Ledger,
	accounts media.AccountSelector,
	creds media.AccountCredentialsLoader,
	factory media.ProviderFactory,
	subscriptions media.SubscriptionResolver,
) *media.TaskService {
	return media.NewTaskService(ledger, accounts, creds, factory, subscriptions)
}

type mediaSubscriptionResolverAdapter struct {
	groups        service.GroupRepository
	subscriptions service.UserSubscriptionRepository
}

func ProvideMediaSubscriptionResolver(groups service.GroupRepository, subscriptions service.UserSubscriptionRepository) media.SubscriptionResolver {
	return mediaSubscriptionResolverAdapter{groups: groups, subscriptions: subscriptions}
}

func (a mediaSubscriptionResolverAdapter) ResolveSubscription(ctx context.Context, userID int64, groupID *int64) (*media.SubscriptionBilling, error) {
	if groupID == nil || *groupID <= 0 || a.groups == nil {
		return nil, nil
	}
	group, err := a.groups.GetByID(ctx, *groupID)
	if err != nil {
		return nil, err
	}
	if group == nil || !group.IsSubscriptionType() {
		return nil, nil
	}
	if a.subscriptions == nil {
		return nil, media.ErrUpstreamNotWired
	}
	sub, err := a.subscriptions.GetActiveByUserIDAndGroupID(ctx, userID, *groupID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, service.ErrSubscriptionInvalid
	}
	return &media.SubscriptionBilling{SubscriptionID: sub.ID, IsSubscription: true}, nil
}

// ProvideMediaPoller 组装并启动轮询 Worker（回收路径兜底，见设计文档 §5）。
func ProvideMediaPoller(
	tasks media.TaskStore,
	ledger *media.Ledger,
	creds media.AccountCredentialsLoader,
	factory media.ProviderFactory,
	subscriptions media.SubscriptionResolver,
) *media.Poller {
	poller := media.NewPoller(tasks, ledger, creds, factory, media.PollerConfig{SubscriptionResolver: subscriptions})
	go poller.Run(context.Background(), defaultMediaPollInterval)
	return poller
}

// ProvideHandlers creates the Handlers struct
func ProvideHandlers(
	authHandler *AuthHandler,
	distributionHandler *DistributionHandler,
	userHandler *UserHandler,
	apiKeyHandler *APIKeyHandler,
	usageHandler *UsageHandler,
	redeemHandler *RedeemHandler,
	subscriptionHandler *SubscriptionHandler,
	announcementHandler *AnnouncementHandler,
	channelMonitorUserHandler *ChannelMonitorUserHandler,
	adminHandlers *AdminHandlers,
	gatewayHandler *GatewayHandler,
	openaiGatewayHandler *OpenAIGatewayHandler,
	settingHandler *SettingHandler,
	totpHandler *TotpHandler,
	paymentHandler *PaymentHandler,
	paymentWebhookHandler *PaymentWebhookHandler,
	voucherHandler *VoucherHandler,
	availableChannelHandler *AvailableChannelHandler,
	batchImageHandler *BatchImageHandler,
	publicCatalogHandler *PublicCatalogHandler,
	mediaHandler *MediaHandler,
	_ *service.IdempotencyCoordinator,
	_ *service.IdempotencyCleanupService,
	_ *media.Poller,
) *Handlers {
	return &Handlers{
		Auth:             authHandler,
		Distribution:     distributionHandler,
		User:             userHandler,
		APIKey:           apiKeyHandler,
		Usage:            usageHandler,
		Redeem:           redeemHandler,
		Subscription:     subscriptionHandler,
		Announcement:     announcementHandler,
		ChannelMonitor:   channelMonitorUserHandler,
		Admin:            adminHandlers,
		Gateway:          gatewayHandler,
		OpenAIGateway:    openaiGatewayHandler,
		Setting:          settingHandler,
		Totp:             totpHandler,
		Payment:          paymentHandler,
		PaymentWebhook:   paymentWebhookHandler,
		Voucher:          voucherHandler,
		AvailableChannel: availableChannelHandler,
		BatchImage:       batchImageHandler,
		PublicCatalog:    publicCatalogHandler,
		Media:            mediaHandler,
	}
}

// ProviderSet is the Wire provider set for all handlers
var ProviderSet = wire.NewSet(
	// Top-level handlers
	NewAuthHandler,
	NewDistributionHandler,
	NewUserHandler,
	NewAPIKeyHandler,
	NewUsageHandler,
	NewRedeemHandler,
	NewSubscriptionHandler,
	NewAnnouncementHandler,
	NewChannelMonitorUserHandler,
	NewGatewayHandler,
	NewOpenAIGatewayHandler,
	NewTotpHandler,
	ProvideSettingHandler,
	NewPaymentHandler,
	NewPaymentWebhookHandler,
	NewVoucherHandler,
	NewAvailableChannelHandler,
	NewBatchImageHandler,
	NewPublicCatalogHandler,
	ProvideVoucherService,

	// 多模态异步计费（视频/音频/图片），见文件顶部"组合根"注释
	NewMediaHandler,
	ProvideMediaTaskStore,
	ProvideMediaHoldStore,
	ProvideMediaBalanceReader,
	ProvideMediaCharger,
	ProvideMediaConfigStore,
	ProvideMediaAssetStore,
	ProvideMediaReservation,
	ProvideMediaQuoter,
	ProvideMediaLedger,
	ProvideMediaAccountSelector,
	ProvideMediaAccountLoader,
	ProvideMediaProviderFactory,
	ProvideMediaTaskService,
	ProvideMediaSubscriptionResolver,
	ProvideMediaPoller,
	admin.NewMediaHandler,

	// Admin handlers
	admin.NewDashboardHandler,
	admin.NewUserHandler,
	admin.NewGroupHandler,
	admin.NewAccountHandler,
	admin.NewAnnouncementHandler,
	admin.NewDataManagementHandler,
	admin.NewBackupHandler,
	admin.NewOAuthHandler,
	admin.NewOpenAIOAuthHandler,
	admin.NewGeminiOAuthHandler,
	admin.NewAntigravityOAuthHandler,
	admin.NewGrokOAuthHandler,
	admin.NewProxyHandler,
	admin.NewRedeemHandler,
	admin.NewPromoHandler,
	ProvideAdminSettingHandler,
	admin.NewOpsHandler,
	ProvideSystemHandler,
	admin.NewSubscriptionHandler,
	admin.NewUsageHandler,
	admin.NewUserAttributeHandler,
	admin.NewErrorPassthroughHandler,
	admin.NewTLSFingerprintProfileHandler,
	ProvideAdminAPIKeyHandler,
	admin.NewScheduledTestHandler,
	admin.NewChannelHandler,
	admin.NewChannelMonitorHandler,
	admin.NewChannelMonitorRequestTemplateHandler,
	admin.NewContentModerationHandler,
	admin.NewPaymentHandler,
	admin.NewAffiliateHandler,
	admin.NewDistributionHandler,
	admin.NewCatalogModelHandler,
	admin.NewVoucherHandler,
	admin.NewComplianceHandler,

	// AdminHandlers and Handlers constructors
	ProvideAdminHandlers,
	ProvideHandlers,
)
