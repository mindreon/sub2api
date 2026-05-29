package handler

import (
	"errors"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type DistributionHandler struct {
	scopeService     *service.DistributionScopeService
	promotionService *service.DistributionPromotionService
	memberService    *service.DistributionMemberService
	manageService    *service.DistributionUserManageService
	analyticsService *service.DistributionAnalyticsService
}

func NewDistributionHandler(
	scopeService *service.DistributionScopeService,
	promotionService *service.DistributionPromotionService,
	memberService *service.DistributionMemberService,
	manageService *service.DistributionUserManageService,
	analyticsService *service.DistributionAnalyticsService,
) *DistributionHandler {
	return &DistributionHandler{
		scopeService:     scopeService,
		promotionService: promotionService,
		memberService:    memberService,
		manageService:    manageService,
		analyticsService: analyticsService,
	}
}

func (h *DistributionHandler) currentUserID(c *gin.Context) (int64, bool) {
	if h == nil || h.scopeService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return 0, false
	}
	subject, ok := servermiddleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.ErrorFrom(c, service.ErrInvalidDistributionAttribution)
		return 0, false
	}
	return subject.UserID, true
}

func (h *DistributionHandler) GetOverview(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	channelOrgID, err := h.scopeService.ResolveUserChannelOrgID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrDistributionAttributionNotFound) {
			response.Success(c, distributionOverviewPayload(userID, 0, false, nil))
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	canManageChannel, err := h.scopeService.CanManageChannelForUser(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var summary any
	if canManageChannel {
		resolved, err := h.scopeService.GetOverviewForUser(c.Request.Context(), userID)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		summary = resolved
	}
	response.Success(c, distributionOverviewPayload(userID, channelOrgID, canManageChannel, summary))
}

func distributionOverviewPayload(userID, channelOrgID int64, canManageChannel bool, summary any) gin.H {
	return gin.H{
		"user_id":            userID,
		"channel_org_id":     channelOrgID,
		"can_manage_channel": canManageChannel,
		"summary":            summary,
	}
}

func (h *DistributionHandler) ListMembers(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	roleType := c.Query("role_type")
	items, total, err := h.scopeService.ListMembersForUser(c.Request.Context(), userID, pagination.PaginationParams{Page: page, PageSize: pageSize}, roleType)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if total == nil {
		response.Paginated(c, items, 0, page, pageSize)
		return
	}
	response.Paginated(c, items, total.Total, total.Page, total.PageSize)
}

func (h *DistributionHandler) ListAttributions(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.scopeService.ListAttributionsForUser(c.Request.Context(), userID, pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if total == nil {
		response.Paginated(c, items, 0, page, pageSize)
		return
	}
	response.Paginated(c, items, total.Total, total.Page, total.PageSize)
}

func (h *DistributionHandler) ListCommissions(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.scopeService.ListCommissionsForUser(c.Request.Context(), userID, pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if total == nil {
		response.Paginated(c, items, 0, page, pageSize)
		return
	}
	response.Paginated(c, items, total.Total, total.Page, total.PageSize)
}

func (h *DistributionHandler) ListWholesalePricing(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	if h == nil || h.manageService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, discountRate, err := h.manageService.ListWholesalePricingForUser(
		c.Request.Context(),
		userID,
		pagination.PaginationParams{Page: page, PageSize: pageSize},
		c.Query("q"),
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if total == nil {
		response.Success(c, gin.H{
			"discount_rate": discountRate,
			"items":         items,
			"total":         0,
			"page":          page,
			"page_size":     pageSize,
			"pages":         0,
		})
		return
	}
	response.Success(c, gin.H{
		"discount_rate": discountRate,
		"items":         items,
		"total":         total.Total,
		"page":          total.Page,
		"page_size":     total.PageSize,
		"pages":         total.Pages,
	})
}

func (h *DistributionHandler) GetAnalytics(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	if h == nil || h.analyticsService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}

	filter, err := parseDistributionAnalyticsFilter(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	analytics, err := h.analyticsService.GetAnalyticsForUser(c.Request.Context(), userID, filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, analytics)
}

func (h *DistributionHandler) ListWalletTransactions(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.scopeService.ListWalletTransactionsForUser(c.Request.Context(), userID, pagination.PaginationParams{Page: page, PageSize: pageSize}, c.Query("transaction_type"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if total == nil {
		response.Paginated(c, items, 0, page, pageSize)
		return
	}
	response.Paginated(c, items, total.Total, total.Page, total.PageSize)
}

func (h *DistributionHandler) ListWalletRequests(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	if h == nil || h.manageService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.manageService.ListWalletRequestsForUser(c.Request.Context(), userID, service.DistributionWalletRequestListFilter{
		RequestType: c.Query("request_type"),
		Status:      c.Query("status"),
	}, pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if total == nil {
		response.Paginated(c, items, 0, page, pageSize)
		return
	}
	response.Paginated(c, items, total.Total, total.Page, total.PageSize)
}

func (h *DistributionHandler) ListAlertEvents(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	if h == nil || h.manageService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.manageService.ListAlertEventsForUser(c.Request.Context(), userID, service.DistributionAlertEventListFilter{
		AlertType: c.Query("alert_type"),
		Severity:  c.Query("severity"),
		Status:    c.Query("status"),
	}, pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if total == nil {
		response.Paginated(c, items, 0, page, pageSize)
		return
	}
	response.Paginated(c, items, total.Total, total.Page, total.PageSize)
}

type createDistributionPromotionLinkRequest struct {
	MemberID   int64  `json:"member_id" binding:"required,min=1"`
	Code       string `json:"code"`
	TargetType string `json:"target_type" binding:"omitempty,oneof=registration oauth manual"`
	Status     string `json:"status" binding:"omitempty,oneof=active inactive disabled"`
}

type createDistributionMemberRequest struct {
	UserID         int64   `json:"user_id" binding:"required,min=1"`
	RoleType       string  `json:"role_type" binding:"required,oneof=agent kol1 kol2"`
	ParentMemberID *int64  `json:"parent_member_id"`
	LevelCode      string  `json:"level_code" binding:"omitempty,max=20"`
	CommissionRate float64 `json:"commission_rate" binding:"min=0"`
	Status         string  `json:"status" binding:"omitempty,oneof=active inactive disabled"`
}

type updateDistributionOrganizationRequest struct {
	Name        string         `json:"name" binding:"omitempty,max=255"`
	Config      map[string]any `json:"config"`
	BrandConfig map[string]any `json:"brand_config"`
}

type settleDistributionCommissionRequest struct {
	SettlementMethod      string `json:"settlement_method" binding:"omitempty,oneof=balance auto manual offline"`
	SettlementReferenceNo string `json:"settlement_reference_no" binding:"max=120"`
	SettlementNote        string `json:"settlement_note" binding:"max=2000"`
}

type createDistributionWalletRequestRequest struct {
	RequestType string  `json:"request_type" binding:"required,oneof=recharge refund"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	ReferenceNo string  `json:"reference_no" binding:"max=120"`
	Note        string  `json:"note" binding:"max=2000"`
}

func (h *DistributionHandler) GetOrganization(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	if h == nil || h.manageService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	org, err := h.manageService.ResolveChannelOrganization(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, org)
}

func (h *DistributionHandler) UpdateOrganization(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	if h == nil || h.manageService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	var req updateDistributionOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrInvalidDistributionOrganization)
		return
	}
	org, err := h.manageService.UpdateOrganizationForUser(c.Request.Context(), userID, service.DistributionOrganizationInput{
		Name:        req.Name,
		Config:      req.Config,
		BrandConfig: req.BrandConfig,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, org)
}

func (h *DistributionHandler) CreateMember(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	if h == nil || h.memberService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}

	var req createDistributionMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrInvalidDistributionMember)
		return
	}

	member, err := h.memberService.CreateMemberForUser(c.Request.Context(), userID, service.DistributionMemberInput{
		UserID:         req.UserID,
		RoleType:       req.RoleType,
		ParentMemberID: req.ParentMemberID,
		LevelCode:      req.LevelCode,
		CommissionRate: req.CommissionRate,
		Status:         req.Status,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Created(c, member)
}

func (h *DistributionHandler) CreateWalletRequest(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	if h == nil || h.manageService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	var req createDistributionWalletRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrInvalidDistributionWalletRequest)
		return
	}
	out, err := h.manageService.CreateWalletRequestForUser(c.Request.Context(), userID, service.DistributionWalletRequestCreateInput{
		RequestType: req.RequestType,
		Amount:      req.Amount,
		ReferenceNo: req.ReferenceNo,
		Note:        req.Note,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, out)
}

func (h *DistributionHandler) SettleCommission(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	if h == nil || h.manageService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	commissionID, err := strconv.ParseInt(c.Param("commission_id"), 10, 64)
	if err != nil || commissionID <= 0 {
		response.ErrorFrom(c, service.ErrInvalidDistributionCommission)
		return
	}
	var req settleDistributionCommissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrInvalidDistributionCommission)
		return
	}
	ledger, err := h.manageService.SettleCommissionForUser(c.Request.Context(), userID, commissionID, service.DistributionCommissionSettlementInput{
		SettlementMethod:      req.SettlementMethod,
		SettlementReferenceNo: req.SettlementReferenceNo,
		SettlementNote:        req.SettlementNote,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, ledger)
}

func (h *DistributionHandler) ListPromotionLinks(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	if h == nil || h.promotionService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.promotionService.ListLinksForUser(c.Request.Context(), userID, pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if total == nil {
		response.Paginated(c, items, 0, page, pageSize)
		return
	}
	response.Paginated(c, items, total.Total, total.Page, total.PageSize)
}

func (h *DistributionHandler) CreatePromotionLink(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	if h == nil || h.promotionService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}

	var req createDistributionPromotionLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrInvalidDistributionPromotionLink)
		return
	}

	link, err := h.promotionService.CreateLinkForUser(c.Request.Context(), userID, service.DistributionPromotionLinkInput{
		MemberID:   req.MemberID,
		Code:       req.Code,
		TargetType: req.TargetType,
		Status:     req.Status,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Created(c, link)
}

func parseDistributionAnalyticsFilter(c *gin.Context) (service.DistributionAnalyticsFilter, error) {
	userTZ := c.Query("timezone")
	now := timezone.NowInUserLocation(userTZ)

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	startTime := timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -6), userTZ)
	endTime := timezone.StartOfDayInUserLocation(now.AddDate(0, 0, 1), userTZ)

	if startDate != "" {
		parsed, err := timezone.ParseInUserLocation("2006-01-02", startDate, userTZ)
		if err != nil {
			return service.DistributionAnalyticsFilter{}, service.ErrInvalidDistributionStats
		}
		startTime = parsed
	}
	if endDate != "" {
		parsed, err := timezone.ParseInUserLocation("2006-01-02", endDate, userTZ)
		if err != nil {
			return service.DistributionAnalyticsFilter{}, service.ErrInvalidDistributionStats
		}
		endTime = parsed.AddDate(0, 0, 1)
	}
	if !endTime.After(startTime) {
		return service.DistributionAnalyticsFilter{}, service.ErrInvalidDistributionStats
	}

	granularity := c.DefaultQuery("granularity", "day")
	limit := parseDistributionAnalyticsLimit(c.DefaultQuery("limit", "10"))

	return service.DistributionAnalyticsFilter{
		StartTime:   startTime,
		EndTime:     endTime,
		Granularity: granularity,
		Limit:       limit,
	}, nil
}

func parseDistributionAnalyticsLimit(raw string) int {
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return 10
	}
	if limit > 100 {
		return 100
	}
	return limit
}
