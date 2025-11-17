package handler

import (
	"strconv"

	"github.com/anttna7/cjadmin/internal/middleware"
	"github.com/anttna7/cjadmin/internal/service"
	"github.com/anttna7/cjadmin/internal/utils"
	"github.com/gin-gonic/gin"
)

type TenantHandler struct {
	tenantService *service.TenantService
}

func NewTenantHandler() *TenantHandler {
	return &TenantHandler{
		tenantService: service.NewTenantService(),
	}
}

// Create 创建租户
// 任务 4.2.1: 创建租户API
func (h *TenantHandler) Create(c *gin.Context) {
	var req service.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	tenant, err := h.tenantService.Create(&req, userID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "租户创建成功", tenant)
}

// Update 更新租户
// 任务 4.2.2: 更新租户API
func (h *TenantHandler) Update(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "租户ID无效")
		return
	}

	var req service.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	if err := h.tenantService.Update(tenantID, &req); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "租户更新成功", nil)
}

// Delete 删除租户
// 任务 4.2.3: 删除租户API
func (h *TenantHandler) Delete(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "租户ID无效")
		return
	}

	if err := h.tenantService.Delete(tenantID); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "租户删除成功", nil)
}

// Get 获取租户详情
// 任务 4.2.4: 获取租户详情API
func (h *TenantHandler) Get(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "租户ID无效")
		return
	}

	tenant, err := h.tenantService.Get(tenantID)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, err.Error())
		return
	}

	utils.Success(c, tenant)
}

// List 获取租户列表
// 任务 4.2.5: 获取租户列表API
func (h *TenantHandler) List(c *gin.Context) {
	var req service.ListTenantRequest
	req.Page = 1
	req.Size = 20

	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	tenants, total, err := h.tenantService.List(&req)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.PageSuccess(c, tenants, total, req.Page, req.Size)
}

// UpdateStatus 修改租户状态
// 任务 4.2.6: 修改租户状态API
func (h *TenantHandler) UpdateStatus(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "租户ID无效")
		return
	}

	type StatusRequest struct {
		Status string `json:"status" binding:"required"`
	}

	var req StatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	if err := h.tenantService.UpdateStatus(tenantID, req.Status); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "状态修改成功", nil)
}

// GetStatistics 获取租户统计信息
func (h *TenantHandler) GetStatistics(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "租户ID无效")
		return
	}

	stats, err := h.tenantService.GetStatistics(tenantID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, stats)
}
