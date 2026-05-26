package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type DistributionHandler struct {
	adminService        *service.DistributionAdminService
	organizationService *service.DistributionOrganizationService
	memberService       *service.DistributionMemberService
	promotionService    *service.DistributionPromotionService
}

func NewDistributionHandler(
	adminService *service.DistributionAdminService,
	organizationService *service.DistributionOrganizationService,
	memberService *service.DistributionMemberService,
	promotionService *service.DistributionPromotionService,
) *DistributionHandler {
	return &DistributionHandler{
		adminService:        adminService,
		organizationService: organizationService,
		memberService:       memberService,
		promotionService:    promotionService,
	}
}

type createDistributionOrganizationRequest struct {
	Type        string         `json:"type" binding:"required,oneof=platform reseller oem"`
	Name        string         `json:"name" binding:"required,max=255"`
	OwnerUserID *int64         `json:"owner_user_id"`
	Status      string         `json:"status" binding:"omitempty,oneof=active inactive disabled"`
	Config      map[string]any `json:"config"`
	BrandConfig map[string]any `json:"brand_config"`
}

type updateDistributionOrganizationRequest = createDistributionOrganizationRequest

func (h *DistributionHandler) ListOrganizations(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.adminService.ListOrganizations(c.Request.Context(), pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) CreateOrganization(c *gin.Context) {
	if h == nil || h.organizationService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}

	var req createDistributionOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrInvalidDistributionOrganization)
		return
	}

	org, err := h.organizationService.CreateOrganization(c.Request.Context(), service.DistributionOrganizationInput{
		Type:        req.Type,
		Name:        req.Name,
		OwnerUserID: req.OwnerUserID,
		Status:      req.Status,
		Config:      req.Config,
		BrandConfig: req.BrandConfig,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Created(c, org)
}

func (h *DistributionHandler) UpdateOrganization(c *gin.Context) {
	if h == nil || h.organizationService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.ErrorFrom(c, service.ErrInvalidDistributionOrganization)
		return
	}

	var req updateDistributionOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrInvalidDistributionOrganization)
		return
	}

	org, err := h.organizationService.UpdateOrganization(c.Request.Context(), id, service.DistributionOrganizationInput{
		Type:        req.Type,
		Name:        req.Name,
		OwnerUserID: req.OwnerUserID,
		Status:      req.Status,
		Config:      req.Config,
		BrandConfig: req.BrandConfig,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, org)
}

type createDistributionMemberRequest struct {
	ChannelOrgID   int64   `json:"channel_org_id" binding:"required,min=1"`
	UserID         int64   `json:"user_id" binding:"required,min=1"`
	RoleType       string  `json:"role_type" binding:"required,oneof=manager agent kol1 kol2"`
	ParentMemberID *int64  `json:"parent_member_id"`
	LevelCode      string  `json:"level_code" binding:"omitempty,max=20"`
	CommissionRate float64 `json:"commission_rate" binding:"min=0"`
	Status         string  `json:"status" binding:"omitempty,oneof=active inactive disabled"`
}

type createDistributionPromotionLinkRequest struct {
	MemberID   int64  `json:"member_id" binding:"required,min=1"`
	Code       string `json:"code"`
	TargetType string `json:"target_type" binding:"omitempty,oneof=registration oauth manual"`
	Status     string `json:"status" binding:"omitempty,oneof=active inactive disabled"`
}

type updateDistributionWalletRequest struct {
	WarningThreshold float64 `json:"warning_threshold" binding:"min=0"`
}

type createDistributionWalletRechargeRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	ReferenceNo string  `json:"reference_no" binding:"max=120"`
	Note        string  `json:"note" binding:"max=2000"`
}

type createDistributionWalletRefundRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	ReferenceNo string  `json:"reference_no" binding:"max=120"`
	Note        string  `json:"note" binding:"max=2000"`
}

type createDistributionWalletRequestRequest struct {
	RequestType string  `json:"request_type" binding:"required,oneof=recharge refund"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	ReferenceNo string  `json:"reference_no" binding:"max=120"`
	Note        string  `json:"note" binding:"max=2000"`
}

type reviewDistributionWalletRequestRequest struct {
	Action     string `json:"action" binding:"required,oneof=approve reject"`
	ReviewNote string `json:"review_note" binding:"max=2000"`
}

type updateDistributionAttributionRequest struct {
	ChannelOrgID     int64  `json:"channel_org_id" binding:"required,min=1"`
	ReferrerMemberID *int64 `json:"referrer_member_id"`
	PromotionLinkID  *int64 `json:"promotion_link_id"`
	Note             string `json:"note" binding:"max=2000"`
}

type settleDistributionCommissionRequest struct {
	SettlementMethod      string `json:"settlement_method" binding:"omitempty,oneof=balance auto manual offline"`
	SettlementReferenceNo string `json:"settlement_reference_no" binding:"max=120"`
	SettlementNote        string `json:"settlement_note" binding:"max=2000"`
}

func (h *DistributionHandler) ListMembers(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	page, pageSize := response.ParsePagination(c)
	filter := parseAdminDistributionFilter(c)
	items, total, err := h.adminService.ListMembers(c.Request.Context(), filter, pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) CreateMember(c *gin.Context) {
	if h == nil || h.memberService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}

	var req createDistributionMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrInvalidDistributionMember)
		return
	}

	member, err := h.memberService.CreateMember(c.Request.Context(), service.DistributionMemberInput{
		ChannelOrgID:   req.ChannelOrgID,
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

func (h *DistributionHandler) ListAttributions(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	page, pageSize := response.ParsePagination(c)
	filter := parseAdminDistributionFilter(c)
	items, total, err := h.adminService.ListAttributions(c.Request.Context(), filter, pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) ListAttributionAudits(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	page, pageSize := response.ParsePagination(c)
	filter := parseAdminDistributionFilter(c)
	items, total, err := h.adminService.ListAttributionAudits(c.Request.Context(), filter, pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) ListCommissions(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	page, pageSize := response.ParsePagination(c)
	filter := parseAdminDistributionFilter(c)
	items, total, err := h.adminService.ListCommissions(c.Request.Context(), filter, pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) UpdateAttribution(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		response.ErrorFrom(c, service.ErrInvalidDistributionAttribution)
		return
	}

	var req updateDistributionAttributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrInvalidDistributionAttribution)
		return
	}

	var operatorUserID *int64
	if subject, ok := middleware.GetAuthSubjectFromContext(c); ok && subject.UserID > 0 {
		operatorUserID = &subject.UserID
	}

	attribution, err := h.adminService.UpdateAttribution(c.Request.Context(), userID, service.DistributionAttributionAdminUpdateInput{
		UserID:           userID,
		ChannelOrgID:     req.ChannelOrgID,
		ReferrerMemberID: req.ReferrerMemberID,
		PromotionLinkID:  req.PromotionLinkID,
		OperatorUserID:   operatorUserID,
		Note:             req.Note,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, attribution)
}

func (h *DistributionHandler) ListWallets(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	page, pageSize := response.ParsePagination(c)
	filter := parseAdminDistributionFilter(c)
	items, total, err := h.adminService.ListWallets(c.Request.Context(), filter, pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) ListWalletTransactions(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	page, pageSize := response.ParsePagination(c)
	filter := parseAdminDistributionFilter(c)
	items, total, err := h.adminService.ListWalletTransactions(c.Request.Context(), filter, pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) ListWalletRequests(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	page, pageSize := response.ParsePagination(c)
	filter := parseDistributionWalletRequestFilter(c)
	items, total, err := h.adminService.ListWalletRequests(c.Request.Context(), filter, pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) ListAlertEvents(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	page, pageSize := response.ParsePagination(c)
	filter := parseDistributionAlertEventFilter(c)
	items, total, err := h.adminService.ListAlertEvents(c.Request.Context(), filter, pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) GetStats(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	stats, err := h.adminService.GetStats(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, stats)
}

func (h *DistributionHandler) RechargeWallet(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	channelOrgID, err := strconv.ParseInt(c.Param("channel_org_id"), 10, 64)
	if err != nil || channelOrgID <= 0 {
		response.ErrorFrom(c, service.ErrInvalidDistributionWallet)
		return
	}

	var req createDistributionWalletRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrInvalidDistributionWallet)
		return
	}

	var operatorUserID *int64
	if subject, ok := middleware.GetAuthSubjectFromContext(c); ok && subject.UserID > 0 {
		operatorUserID = &subject.UserID
	}

	wallet, err := h.adminService.RechargeWallet(c.Request.Context(), channelOrgID, service.DistributionWalletRechargeInput{
		Amount:         req.Amount,
		ReferenceNo:    req.ReferenceNo,
		Note:           req.Note,
		OperatorUserID: operatorUserID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, wallet)
}

func (h *DistributionHandler) UpdateWalletWarningThreshold(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	channelOrgID, err := strconv.ParseInt(c.Param("channel_org_id"), 10, 64)
	if err != nil || channelOrgID <= 0 {
		response.ErrorFrom(c, service.ErrInvalidDistributionWallet)
		return
	}
	var req updateDistributionWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrInvalidDistributionWallet)
		return
	}
	wallet, err := h.adminService.UpdateWalletWarningThreshold(c.Request.Context(), channelOrgID, req.WarningThreshold)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, wallet)
}

func (h *DistributionHandler) RefundWallet(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	channelOrgID, err := strconv.ParseInt(c.Param("channel_org_id"), 10, 64)
	if err != nil || channelOrgID <= 0 {
		response.ErrorFrom(c, service.ErrInvalidDistributionWallet)
		return
	}

	var req createDistributionWalletRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrInvalidDistributionWallet)
		return
	}

	var operatorUserID *int64
	if subject, ok := middleware.GetAuthSubjectFromContext(c); ok && subject.UserID > 0 {
		operatorUserID = &subject.UserID
	}

	result, err := h.adminService.RefundWallet(c.Request.Context(), channelOrgID, service.DistributionWalletRefundInput{
		Amount:         req.Amount,
		ReferenceNo:    req.ReferenceNo,
		Note:           req.Note,
		OperatorUserID: operatorUserID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, result)
}

func (h *DistributionHandler) ReviewWalletRequest(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	requestID, err := strconv.ParseInt(c.Param("request_id"), 10, 64)
	if err != nil || requestID <= 0 {
		response.ErrorFrom(c, service.ErrInvalidDistributionWalletRequest)
		return
	}
	var req reviewDistributionWalletRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrInvalidDistributionWalletRequest)
		return
	}

	var reviewedByUserID int64
	if subject, ok := middleware.GetAuthSubjectFromContext(c); ok && subject.UserID > 0 {
		reviewedByUserID = subject.UserID
	}

	out, err := h.adminService.ReviewWalletRequest(c.Request.Context(), requestID, service.DistributionWalletRequestReviewInput{
		Action:           req.Action,
		ReviewNote:       req.ReviewNote,
		ReviewedByUserID: reviewedByUserID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, out)
}

func (h *DistributionHandler) SettleCommission(c *gin.Context) {
	if h == nil || h.adminService == nil {
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
	var settledByUserID *int64
	if subject, ok := middleware.GetAuthSubjectFromContext(c); ok && subject.UserID > 0 {
		settledByUserID = &subject.UserID
	}
	ledger, err := h.adminService.SettleCommission(c.Request.Context(), commissionID, service.DistributionCommissionSettlementInput{
		SettlementMethod:      req.SettlementMethod,
		SettlementReferenceNo: req.SettlementReferenceNo,
		SettlementNote:        req.SettlementNote,
		SettledByUserID:       settledByUserID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, ledger)
}

func (h *DistributionHandler) ReverseCommission(c *gin.Context) {
	if h == nil || h.adminService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	commissionID, err := strconv.ParseInt(c.Param("commission_id"), 10, 64)
	if err != nil || commissionID <= 0 {
		response.ErrorFrom(c, service.ErrInvalidDistributionCommission)
		return
	}
	ledger, err := h.adminService.ReverseCommission(c.Request.Context(), commissionID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, ledger)
}

func (h *DistributionHandler) ListPromotionLinks(c *gin.Context) {
	if h == nil || h.promotionService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}
	page, pageSize := response.ParsePagination(c)
	filter := parseAdminDistributionFilter(c)
	items, total, err := h.promotionService.ListLinks(c.Request.Context(), filter, pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) CreatePromotionLink(c *gin.Context) {
	if h == nil || h.promotionService == nil {
		response.ErrorFrom(c, service.ErrServiceUnavailable)
		return
	}

	var req createDistributionPromotionLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrInvalidDistributionPromotionLink)
		return
	}

	link, err := h.promotionService.CreateLink(c.Request.Context(), service.DistributionPromotionLinkInput{
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

func parseAdminDistributionFilter(c *gin.Context) service.DistributionAdminListFilter {
	var channelOrgID int64
	var userID int64
	if raw := c.Query("channel_org_id"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed > 0 {
			channelOrgID = parsed
		}
	}
	if raw := c.Query("user_id"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed > 0 {
			userID = parsed
		}
	}
	return service.DistributionAdminListFilter{
		ChannelOrgID:    channelOrgID,
		UserID:          userID,
		RoleType:        c.Query("role_type"),
		TransactionType: c.Query("transaction_type"),
		Q:               c.Query("q"),
	}
}

func parseDistributionWalletRequestFilter(c *gin.Context) service.DistributionWalletRequestListFilter {
	var channelOrgID int64
	if raw := c.Query("channel_org_id"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed > 0 {
			channelOrgID = parsed
		}
	}
	return service.DistributionWalletRequestListFilter{
		ChannelOrgID: channelOrgID,
		RequestType:  c.Query("request_type"),
		Status:       c.Query("status"),
	}
}

func parseDistributionAlertEventFilter(c *gin.Context) service.DistributionAlertEventListFilter {
	var channelOrgID int64
	if raw := c.Query("channel_org_id"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed > 0 {
			channelOrgID = parsed
		}
	}
	return service.DistributionAlertEventListFilter{
		ChannelOrgID: channelOrgID,
		AlertType:    c.Query("alert_type"),
		Severity:     c.Query("severity"),
		Status:       c.Query("status"),
	}
}

func paginated(c *gin.Context, items any, result *pagination.PaginationResult, fallbackPage int, fallbackPageSize int) {
	if result == nil {
		response.Paginated(c, items, 0, fallbackPage, fallbackPageSize)
		return
	}
	response.Paginated(c, items, result.Total, result.Page, result.PageSize)
}
