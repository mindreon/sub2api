package handler

import (
	"io"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/voucher"
	"github.com/gin-gonic/gin"
)

// VoucherHandler serves user-facing voucher PIN purchase APIs.
type VoucherHandler struct {
	svc *voucher.Service
}

func NewVoucherHandler(svc *voucher.Service) *VoucherHandler {
	return &VoucherHandler{svc: svc}
}

func (h *VoucherHandler) GetCheckoutInfo(c *gin.Context) {
	info, err := h.svc.CheckoutInfo(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, info)
}

type createVoucherOrderRequest struct {
	ProductID      int64  `json:"product_id" binding:"required"`
	Quantity       int    `json:"quantity" binding:"required"`
	IdempotencyKey string `json:"idempotency_key"`
}

func (h *VoucherHandler) CreateOrder(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	var req createVoucherOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}
	order, err := h.svc.CreateOrder(c.Request.Context(), voucher.CreateOrderInput{
		UserID:         subject.UserID,
		ProductID:      req.ProductID,
		Quantity:       req.Quantity,
		IdempotencyKey: req.IdempotencyKey,
		ClientIP:       c.ClientIP(),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"order": order})
}

func (h *VoucherHandler) ListMyOrders(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "25"))
	orders, total, err := h.svc.ListMyOrders(c.Request.Context(), subject.UserID, page, perPage)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"orders": orders,
		"pagination": gin.H{
			"page":       page,
			"per_page":   perPage,
			"total":      total,
			"total_pages": (total + perPage - 1) / perPage,
		},
	})
}

func (h *VoucherHandler) GetOrder(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}
	includePins := c.Query("include_pins") == "1" || c.Query("include_pins") == "true"
	order, err := h.svc.GetOrder(c.Request.Context(), subject.UserID, id, includePins)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"order": order})
}

func (h *VoucherHandler) SubmitPaymentProof(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}
	paymentRef := c.PostForm("payment_ref")
	var bankID *int
	if raw := c.PostForm("bank_id"); raw != "" {
		if n, perr := strconv.Atoi(raw); perr == nil {
			bankID = &n
		}
	}
	file, header, ferr := c.Request.FormFile("payment_proof")
	var proofName string
	var proofReader io.Reader
	if ferr == nil && file != nil {
		defer file.Close()
		proofName = header.Filename
		proofReader = file
	}
	order, err := h.svc.SubmitPaymentProof(c.Request.Context(), voucher.SubmitProofInput{
		UserID:      subject.UserID,
		OrderID:     id,
		PaymentRef:  paymentRef,
		BankID:      bankID,
		ProofReader: proofReader,
		ProofName:   proofName,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"order": order})
}

func (h *VoucherHandler) CancelOrder(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}
	order, err := h.svc.CancelOrder(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"order": order})
}
