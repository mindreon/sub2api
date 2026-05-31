package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/voucher"
	"github.com/gin-gonic/gin"
)

// VoucherHandler serves admin voucher management APIs.
type VoucherHandler struct {
	svc *voucher.Service
}

func NewVoucherHandler(svc *voucher.Service) *VoucherHandler {
	return &VoucherHandler{svc: svc}
}

func (h *VoucherHandler) GetSettings(c *gin.Context) {
	view, err := h.svc.AdminSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, view)
}

type updateVoucherSettingsRequest struct {
	Enabled             *bool              `json:"enabled"`
	UIEnabled           *bool              `json:"ui_enabled"`
	Sandbox             *bool              `json:"sandbox"`
	APIKey              *string            `json:"api_key"`
	APISecret           *string            `json:"api_secret"`
	APIBase             *string            `json:"api_base"`
	BankAccounts        []voucher.BankAccount `json:"bank_accounts"`
	OrderTimeoutHours   *int               `json:"order_timeout_hours"`
	MaxQuantityPerOrder *int               `json:"max_quantity_per_order"`
	ReviewSLAHours      *int               `json:"review_sla_hours"`
	FeeRate             *float64           `json:"fee_rate"`
	HelpText            *string            `json:"help_text"`
	RetailMarkupPercent *float64           `json:"retail_markup_percent"`
}

func (h *VoucherHandler) UpdateSettings(c *gin.Context) {
	var req updateVoucherSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}
	view, err := h.svc.UpdateAdminSettings(c.Request.Context(), voucher.UpdateSettingsInput{
		Enabled:             req.Enabled,
		UIEnabled:           req.UIEnabled,
		APIKey:              req.APIKey,
		APISecret:           req.APISecret,
		APIBase:             req.APIBase,
		Sandbox:             req.Sandbox,
		BankAccounts:        req.BankAccounts,
		OrderTimeoutHours:   req.OrderTimeoutHours,
		MaxQuantityPerOrder: req.MaxQuantityPerOrder,
		ReviewSLAHours:      req.ReviewSLAHours,
		FeeRate:             req.FeeRate,
		HelpText:            req.HelpText,
		RetailMarkupPercent: req.RetailMarkupPercent,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, view)
}

func (h *VoucherHandler) TestConnection(c *gin.Context) {
	cfg, err := h.svc.LoadAPIConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	result := voucher.TestConnection(c.Request.Context(), cfg)
	if !result.OK {
		response.BadRequest(c, result.Message)
		return
	}
	response.Success(c, result)
}

func (h *VoucherHandler) SyncCatalog(c *gin.Context) {
	count, err := h.svc.SyncCatalog(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"synced": count})
}

func (h *VoucherHandler) SyncStock(c *gin.Context) {
	count, err := h.svc.SyncStock(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"updated": count})
}

func (h *VoucherHandler) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "25"))
	status := c.Query("status")
	orders, total, err := h.svc.AdminListOrders(c.Request.Context(), status, page, perPage)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"orders": orders,
		"pagination": gin.H{
			"page":        page,
			"per_page":    perPage,
			"total":       total,
			"total_pages": (total + perPage - 1) / perPage,
		},
	})
}

func (h *VoucherHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}
	order, err := h.svc.AdminGetOrder(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"order": order})
}

type rejectVoucherOrderRequest struct {
	Reason string `json:"reason"`
}

func (h *VoucherHandler) VerifyOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}
	order, err := h.svc.AdminVerifyOrder(c.Request.Context(), id, adminOperator(c))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"order": order})
}

func (h *VoucherHandler) RejectOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}
	var req rejectVoucherOrderRequest
	_ = c.ShouldBindJSON(&req)
	order, err := h.svc.AdminRejectOrder(c.Request.Context(), id, req.Reason, adminOperator(c))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"order": order})
}

func (h *VoucherHandler) RetryFulfill(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}
	order, err := h.svc.AdminRetryFulfill(c.Request.Context(), id, adminOperator(c))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"order": order})
}

func adminOperator(c *gin.Context) string {
	if id, ok := c.Get("admin_user_id"); ok {
		return "admin:" + strconv.FormatInt(id.(int64), 10)
	}
	return "admin"
}
