package handler

import (
	"strconv"

	"github.com/anttna7/cjadmin/internal/middleware"
	"github.com/anttna7/cjadmin/internal/service"
	"github.com/anttna7/cjadmin/internal/utils"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{
		orderService: service.NewOrderService(),
	}
}

// CreateRechargeOrder 创建充值订单
func (h *OrderHandler) CreateRechargeOrder(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.CreateRechargeOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	order, err := h.orderService.CreateRechargeOrder(*tenantID, &req, userID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "充值订单创建成功", order)
}

// PaymentCallback 支付回调
func (h *OrderHandler) PaymentCallback(c *gin.Context) {
	type CallbackRequest struct {
		OrderNo       string `json:"order_no" binding:"required"`
		TransactionID string `json:"transaction_id" binding:"required"`
	}

	var req CallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	if err := h.orderService.PaymentCallback(req.OrderNo, req.TransactionID); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "支付回调处理成功", nil)
}

// VerifyRechargeOrder 财务审核充值订单
func (h *OrderHandler) VerifyRechargeOrder(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "订单ID无效")
		return
	}

	type VerifyRequest struct {
		Approved bool   `json:"approved"`
		Notes    string `json:"notes"`
	}

	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.orderService.VerifyRechargeOrder(*tenantID, orderID, userID, req.Approved, req.Notes); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "审核完成", nil)
}

// CreateTransferOrder 创建转账订单
func (h *OrderHandler) CreateTransferOrder(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	type TransferRequest struct {
		CustomerID int64   `json:"customer_id" binding:"required"`
		Amount     float64 `json:"amount" binding:"required,gt=0"`
	}

	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	order, err := h.orderService.CreateTransferOrder(*tenantID, req.CustomerID, req.Amount, userID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "转账成功", order)
}

// GetOrder 获取订单详情
func (h *OrderHandler) GetOrder(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "订单ID无效")
		return
	}

	order, err := h.orderService.GetOrder(*tenantID, orderID)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, err.Error())
		return
	}

	utils.Success(c, order)
}

// ListOrders 获取订单列表
func (h *OrderHandler) ListOrders(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	orderType := c.Query("order_type")
	status := c.Query("status")

	var customerID *int64
	if cid := c.Query("customer_id"); cid != "" {
		id, _ := strconv.ParseInt(cid, 10, 64)
		customerID = &id
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	orders, total, err := h.orderService.ListOrders(*tenantID, orderType, status, customerID, page, size)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.PageSuccess(c, orders, total, page, size)
}
