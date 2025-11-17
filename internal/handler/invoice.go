package handler

import (
	"strconv"

	"github.com/anttna7/cjadmin/internal/middleware"
	"github.com/anttna7/cjadmin/internal/service"
	"github.com/anttna7/cjadmin/internal/utils"
	"github.com/gin-gonic/gin"
)

type InvoiceHandler struct {
	invoiceService *service.InvoiceService
}

func NewInvoiceHandler() *InvoiceHandler {
	return &InvoiceHandler{
		invoiceService: service.NewInvoiceService(),
	}
}

// Create 创建发票
// 任务 9.3.1: 创建发票API
func (h *InvoiceHandler) Create(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	invoice, err := h.invoiceService.Create(*tenantID, &req)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "发票创建成功", invoice)
}

// Update 更新发票
// 任务 9.3.2: 更新发票API
func (h *InvoiceHandler) Update(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	invoiceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "发票ID无效")
		return
	}

	var req service.UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	if err := h.invoiceService.Update(*tenantID, invoiceID, &req); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "发票更新成功", nil)
}

// Delete 删除发票
// 任务 9.3.3: 删除发票API
func (h *InvoiceHandler) Delete(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	invoiceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "发票ID无效")
		return
	}

	if err := h.invoiceService.Delete(*tenantID, invoiceID); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "发票删除成功", nil)
}

// Get 获取发票详情
// 任务 9.3.4: 获取发票详情API
func (h *InvoiceHandler) Get(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	invoiceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "发票ID无效")
		return
	}

	invoice, err := h.invoiceService.Get(*tenantID, invoiceID)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, err.Error())
		return
	}

	utils.Success(c, invoice)
}

// List 获取发票列表
// 任务 9.3.5: 获取发票列表API
func (h *InvoiceHandler) List(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.ListInvoiceRequest
	req.Page = 1
	req.Size = 20

	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	invoices, total, err := h.invoiceService.List(*tenantID, &req)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.PageSuccess(c, invoices, total, req.Page, req.Size)
}

// Issue 开具发票
// 任务 9.3.6: 开具发票API
func (h *InvoiceHandler) Issue(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	invoiceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "发票ID无效")
		return
	}

	type IssueRequest struct {
		FileURL string `json:"file_url"`
	}

	var req IssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	if err := h.invoiceService.Issue(*tenantID, invoiceID, req.FileURL); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "发票开具成功", nil)
}

// Cancel 作废发票
// 任务 9.3.7: 作废发票API
func (h *InvoiceHandler) Cancel(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	invoiceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "发票ID无效")
		return
	}

	type CancelRequest struct {
		Reason string `json:"reason"`
	}

	var req CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	if err := h.invoiceService.Cancel(*tenantID, invoiceID, req.Reason); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "发票作废成功", nil)
}

// GetByCustomer 获取客户的发票列表
func (h *InvoiceHandler) GetByCustomer(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	customerID, err := strconv.ParseInt(c.Param("customer_id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "客户ID无效")
		return
	}

	invoices, err := h.invoiceService.GetByCustomer(*tenantID, customerID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, invoices)
}

// GetByOrder 获取订单的发票列表
func (h *InvoiceHandler) GetByOrder(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	orderID, err := strconv.ParseInt(c.Param("order_id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "订单ID无效")
		return
	}

	invoices, err := h.invoiceService.GetByOrder(*tenantID, orderID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, invoices)
}

// GetStatistics 获取发票统计信息
func (h *InvoiceHandler) GetStatistics(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var customerID *int64
	if cidStr := c.Query("customer_id"); cidStr != "" {
		cid, err := strconv.ParseInt(cidStr, 10, 64)
		if err != nil {
			utils.Error(c, utils.CodeInvalidParams, "客户ID无效")
			return
		}
		customerID = &cid
	}

	stats, err := h.invoiceService.GetStatistics(*tenantID, customerID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, stats)
}
