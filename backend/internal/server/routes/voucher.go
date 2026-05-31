package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// RegisterVoucherRoutes registers KVoucher PIN purchase routes.
func RegisterVoucherRoutes(
	v1 *gin.RouterGroup,
	voucherHandler *handler.VoucherHandler,
	adminVoucherHandler *admin.VoucherHandler,
	jwtAuth middleware.JWTAuthMiddleware,
	adminAuth middleware.AdminAuthMiddleware,
	settingService *service.SettingService,
) {
	user := v1.Group("/voucher")
	user.Use(gin.HandlerFunc(jwtAuth))
	user.Use(middleware.BackendModeUserGuard(settingService))
	{
		user.GET("/checkout-info", voucherHandler.GetCheckoutInfo)
		user.POST("/orders", voucherHandler.CreateOrder)
		user.GET("/orders", voucherHandler.ListMyOrders)
		user.GET("/orders/:id", voucherHandler.GetOrder)
		user.POST("/orders/:id/payment-proof", voucherHandler.SubmitPaymentProof)
		user.POST("/orders/:id/cancel", voucherHandler.CancelOrder)
	}

	adminGroup := v1.Group("/admin/voucher")
	adminGroup.Use(gin.HandlerFunc(adminAuth))
	{
		adminGroup.GET("/settings", adminVoucherHandler.GetSettings)
		adminGroup.PUT("/settings", adminVoucherHandler.UpdateSettings)
		adminGroup.POST("/test-connection", adminVoucherHandler.TestConnection)
		adminGroup.POST("/sync-catalog", adminVoucherHandler.SyncCatalog)
		adminGroup.POST("/sync-stock", adminVoucherHandler.SyncStock)
		adminGroup.GET("/orders", adminVoucherHandler.ListOrders)
		adminGroup.GET("/orders/:id", adminVoucherHandler.GetOrder)
		adminGroup.POST("/orders/:id/verify", adminVoucherHandler.VerifyOrder)
		adminGroup.POST("/orders/:id/reject", adminVoucherHandler.RejectOrder)
		adminGroup.POST("/orders/:id/retry-fulfill", adminVoucherHandler.RetryFulfill)
	}
}
