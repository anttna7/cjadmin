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

// ExportAuditLogs 导出审计日志
func (h *SystemHandler) ExportAuditLogs(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10000")) // 导出时使用较大的页面大小
	operationType := c.Query("operation_type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var userID *int64
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if uid, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			userID = &uid
		}
	}

	// 获取审计日志
	logs, total, err := h.systemService.GetAuditLogs(*tenantID, page, pageSize, operationType, userID, startDate, endDate)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	// 设置响应头
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=audit_logs_export.csv")

	// 写入BOM以支持Excel正确识别UTF-8
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	// 写入CSV头
	c.Writer.Write([]byte("ID,操作类型,操作描述,资源类型,资源ID,用户ID,请求方法,请求路径,请求参数,响应代码,耗时(ms),IP地址,User-Agent,创建时间\n"))

	// 写入数据行
	for _, log := range logs {
		resourceID := ""
		if log.ResourceID != nil {
			resourceID = *log.ResourceID
		}

		row := []string{
			strconv.FormatInt(log.ID, 10),
			log.OperationType,
			log.OperationDesc,
			log.ResourceType,
			resourceID,
			strconv.FormatInt(log.UserID, 10),
			log.RequestMethod,
			log.RequestPath,
			log.RequestParams,
			strconv.Itoa(log.ResponseCode),
			strconv.Itoa(log.Duration),
			log.IPAddress,
			log.UserAgent,
			log.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// CSV转义（处理逗号和引号）
		for i, field := range row {
			if i > 0 {
				c.Writer.Write([]byte(","))
			}
			// 如果字段包含逗号或引号，需要用引号包围并转义引号
			if containsSpecialChar(field) {
				field = `"` + escapeQuotes(field) + `"`
			}
			c.Writer.Write([]byte(field))
		}
		c.Writer.Write([]byte("\n"))
	}
}

// containsSpecialChar 检查字符串是否包含特殊字符
func containsSpecialChar(s string) bool {
	for _, char := range s {
		if char == ',' || char == '"' || char == '\n' || char == '\r' {
			return true
		}
	}
	return false
}

// escapeQuotes 转义引号
func escapeQuotes(s string) string {
	result := ""
	for _, char := range s {
		if char == '"' {
			result += `""`
		} else {
			result += string(char)
		}
	}
	return result
}

// ===== 字段变更历史接口 =====

// GetFieldChangeHistory 获取字段变更历史
func (h *SystemHandler) GetFieldChangeHistory(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	resourceType := c.Query("resource_type")
	resourceIDStr := c.Query("resource_id")
	fieldName := c.Query("field_name")

	if resourceType == "" || resourceIDStr == "" {
		utils.Error(c, utils.CodeInvalidParams, "资源类型和资源ID不能为空")
		return
	}

	resourceID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "资源ID格式错误")
		return
	}

	var fieldNamePtr *string
	if fieldName != "" {
		fieldNamePtr = &fieldName
	}

	logs, err := h.systemService.GetFieldChangeHistory(*tenantID, resourceType, resourceID, fieldNamePtr)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"logs": logs,
	})
}

// ===== 导入日志接口 =====

// QueryImportLogs 查询导入日志
func (h *SystemHandler) QueryImportLogs(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.QueryImportLogRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, err.Error())
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}

	logs, total, err := h.systemService.QueryImportLogs(*tenantID, &req)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"list":  logs,
		"total": total,
		"page":  req.Page,
		"size":  req.Size,
	})
}

// GetImportLog 获取导入日志详情
func (h *SystemHandler) GetImportLog(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "ID格式错误")
		return
	}

	log, err := h.systemService.GetImportLog(*tenantID, id)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, err.Error())
		return
	}

	utils.Success(c, log)
}

// GetImportStatistics 获取导入统计
func (h *SystemHandler) GetImportStatistics(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}

	stats, err := h.systemService.GetImportStatistics(*tenantID, days)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, stats)
}
