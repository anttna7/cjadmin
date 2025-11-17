package handler

import (
	"strconv"

	"github.com/anttna7/cjadmin/internal/middleware"
	"github.com/anttna7/cjadmin/internal/service"
	"github.com/anttna7/cjadmin/internal/utils"
	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	customerService *service.CustomerService
}

func NewCustomerHandler() *CustomerHandler {
	return &CustomerHandler{
		customerService: service.NewCustomerService(),
	}
}

// Create 创建客户
func (h *CustomerHandler) Create(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	customer, err := h.customerService.Create(*tenantID, &req)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "客户创建成功", customer)
}

// Update 更新客户
func (h *CustomerHandler) Update(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "客户ID无效")
		return
	}

	var req service.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	if err := h.customerService.Update(*tenantID, customerID, &req); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "客户更新成功", nil)
}

// Delete 删除客户
func (h *CustomerHandler) Delete(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "客户ID无效")
		return
	}

	if err := h.customerService.Delete(*tenantID, customerID); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "客户删除成功", nil)
}

// Get 获取客户详情
func (h *CustomerHandler) Get(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "客户ID无效")
		return
	}

	customer, err := h.customerService.Get(*tenantID, customerID)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, err.Error())
		return
	}

	utils.Success(c, customer)
}

// List 获取客户列表
func (h *CustomerHandler) List(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.ListCustomerRequest
	req.Page = 1
	req.Size = 20

	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	customers, total, err := h.customerService.List(*tenantID, &req)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.PageSuccess(c, customers, total, req.Page, req.Size)
}

// GetAccounts 获取客户账户信息
func (h *CustomerHandler) GetAccounts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "客户ID无效")
		return
	}

	accounts, err := h.customerService.GetCustomerAccounts(*tenantID, customerID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, accounts)
}
