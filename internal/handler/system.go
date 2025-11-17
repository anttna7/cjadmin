package handler

import (
	"strconv"

	"github.com/anttna7/cjadmin/internal/middleware"
	"github.com/anttna7/cjadmin/internal/service"
	"github.com/anttna7/cjadmin/internal/utils"
	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	systemService *service.SystemService
}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{
		systemService: service.NewSystemService(),
	}
}

// SetSetting 设置系统配置
// 任务 11.4.1: 设置系统配置API
func (h *SystemHandler) SetSetting(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.SetSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	setting, err := h.systemService.SetSetting(*tenantID, &req, userID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "配置设置成功", setting)
}

// GetSetting 获取系统配置
// 任务 11.4.2: 获取系统配置API
func (h *SystemHandler) GetSetting(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	settingKey := c.Param("key")
	if settingKey == "" {
		utils.Error(c, utils.CodeInvalidParams, "配置键不能为空")
		return
	}

	setting, err := h.systemService.GetSetting(*tenantID, settingKey)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, err.Error())
		return
	}

	utils.Success(c, setting)
}

// ListSettings 获取系统配置列表
// 任务 11.4.3: 获取配置列表API
func (h *SystemHandler) ListSettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.ListSettingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	settings, err := h.systemService.ListSettings(*tenantID, &req)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, settings)
}

// DeleteSetting 删除系统配置
func (h *SystemHandler) DeleteSetting(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	settingKey := c.Param("key")
	if settingKey == "" {
		utils.Error(c, utils.CodeInvalidParams, "配置键不能为空")
		return
	}

	if err := h.systemService.DeleteSetting(*tenantID, settingKey); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "配置删除成功", nil)
}

// SetPlatformSetting 设置平台级配置（仅平台管理员）
func (h *SystemHandler) SetPlatformSetting(c *gin.Context) {
	var req service.SetSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	setting, err := h.systemService.SetPlatformSetting(&req, userID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "平台配置设置成功", setting)
}

// GetPlatformSetting 获取平台级配置
func (h *SystemHandler) GetPlatformSetting(c *gin.Context) {
	settingKey := c.Param("key")
	if settingKey == "" {
		utils.Error(c, utils.CodeInvalidParams, "配置键不能为空")
		return
	}

	setting, err := h.systemService.GetPlatformSetting(settingKey)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, err.Error())
		return
	}

	utils.Success(c, setting)
}

// CreateAuditLog 创建审计日志
// 任务 11.4.4: 创建审计日志API
func (h *SystemHandler) CreateAuditLog(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.CreateAuditLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.systemService.CreateAuditLog(*tenantID, userID, &req, ipAddress, userAgent); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "审计日志创建成功", nil)
}

// QueryAuditLogs 查询审计日志
// 任务 11.4.5: 查询审计日志API
func (h *SystemHandler) QueryAuditLogs(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.QueryAuditLogRequest
	req.Page = 1
	req.Size = 20

	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	logs, total, err := h.systemService.QueryAuditLogs(*tenantID, &req)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.PageSuccess(c, logs, total, req.Page, req.Size)
}

// GetAuditLog 获取审计日志详情
func (h *SystemHandler) GetAuditLog(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	logID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "日志ID无效")
		return
	}

	log, err := h.systemService.GetAuditLog(*tenantID, logID)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, err.Error())
		return
	}

	utils.Success(c, log)
}

// GetAuditStatistics 获取审计统计信息
func (h *SystemHandler) GetAuditStatistics(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	// 默认统计最近30天
	days := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	stats, err := h.systemService.GetAuditStatistics(*tenantID, days)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, stats)
}

// CleanOldAuditLogs 清理旧的审计日志（仅管理员）
func (h *SystemHandler) CleanOldAuditLogs(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	type CleanRequest struct {
		DaysToKeep int `json:"days_to_keep" binding:"required,min=30"`
	}

	var req CleanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	deletedCount, err := h.systemService.CleanOldAuditLogs(*tenantID, req.DaysToKeep)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "审计日志清理成功", gin.H{
		"deleted_count": deletedCount,
	})
}
