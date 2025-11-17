package service

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/anttna7/cjadmin/internal/database"
	"github.com/anttna7/cjadmin/internal/models"
	"gorm.io/gorm"
)

type SystemService struct{}

func NewSystemService() *SystemService {
	return &SystemService{}
}

type SetSettingRequest struct {
	SettingKey   string                 `json:"setting_key" binding:"required"`
	SettingValue map[string]interface{} `json:"setting_value" binding:"required"`
	Description  string                 `json:"description"`
}

type ListSettingRequest struct {
	Category string `form:"category"`
}

type CreateAuditLogRequest struct {
	Action        string                        `json:"action" binding:"required"`
	ResourceType  string                        `json:"resource_type" binding:"required"`
	ResourceID    *int64                        `json:"resource_id"`
	Details       string                        `json:"details"`
	ChangedFields map[string]models.FieldChange `json:"changed_fields,omitempty"` // 字段级变更追踪
}

type CreateImportLogRequest struct {
	ImportType   string                   `json:"import_type" binding:"required"`
	FileName     string                   `json:"file_name" binding:"required"`
	TotalRows    int                      `json:"total_rows"`
	SuccessRows  int                      `json:"success_rows"`
	FailedRows   int                      `json:"failed_rows"`
	ErrorDetails []map[string]interface{} `json:"error_details,omitempty"`
	Status       string                   `json:"status"` // processing, completed, failed
}

type QueryImportLogRequest struct {
	Page       int    `form:"page" binding:"min=1"`
	Size       int    `form:"size" binding:"min=1,max=100"`
	ImportType string `form:"import_type"`
	Status     string `form:"status"`
}

type QueryAuditLogRequest struct {
	Page         int        `form:"page" binding:"min=1"`
	Size         int        `form:"size" binding:"min=1,max=100"`
	UserID       *int64     `form:"user_id"`
	Action       string     `form:"action"`
	ResourceType string     `form:"resource_type"`
	StartDate    *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate      *time.Time `form:"end_date" time_format:"2006-01-02"`
}

// SetSetting 设置系统配置
// 任务 11.2.1: 设置系统配置
func (s *SystemService) SetSetting(tenantID int64, req *SetSettingRequest, updatedBy int64) (*models.SystemSetting, error) {
	// 将值转换为JSONB
	valueBytes, err := json.Marshal(req.SettingValue)
	if err != nil {
		return nil, errors.New("配置值格式错误")
	}

	// 查找现有配置
	var setting models.SystemSetting
	err = database.DB.Where("tenant_id = ? AND setting_key = ?", tenantID, req.SettingKey).
		First(&setting).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新配置
			setting = models.SystemSetting{
				TenantModel: models.TenantModel{
					TenantID: tenantID,
				},
				SettingKey:   req.SettingKey,
				SettingValue: valueBytes,
				Description:  req.Description,
				UpdatedBy:    &updatedBy,
			}
			if err := database.DB.Create(&setting).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		// 更新现有配置
		updates := map[string]interface{}{
			"setting_value": valueBytes,
			"updated_by":    updatedBy,
		}
		if req.Description != "" {
			updates["description"] = req.Description
		}
		if err := database.DB.Model(&setting).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return &setting, nil
}

// GetSetting 获取系统配置
// 任务 11.2.2: 获取系统配置
func (s *SystemService) GetSetting(tenantID int64, settingKey string) (*models.SystemSetting, error) {
	var setting models.SystemSetting
	if err := database.DB.Where("tenant_id = ? AND setting_key = ?", tenantID, settingKey).
		First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("配置不存在")
		}
		return nil, err
	}
	return &setting, nil
}

// ListSettings 获取系统配置列表
// 任务 11.2.3: 获取配置列表
func (s *SystemService) ListSettings(tenantID int64, req *ListSettingRequest) ([]models.SystemSetting, error) {
	var settings []models.SystemSetting

	query := database.DB.Model(&models.SystemSetting{}).Where("tenant_id = ?", tenantID)

	// 按分类过滤
	if req.Category != "" {
		query = query.Where("setting_key LIKE ?", req.Category+".%")
	}

	if err := query.Order("setting_key").Find(&settings).Error; err != nil {
		return nil, err
	}

	return settings, nil
}

// DeleteSetting 删除系统配置
func (s *SystemService) DeleteSetting(tenantID int64, settingKey string) error {
	result := database.DB.Where("tenant_id = ? AND setting_key = ?", tenantID, settingKey).
		Delete(&models.SystemSetting{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("配置不存在")
	}

	return nil
}

// GetPlatformSetting 获取平台级别配置（不区分租户）
func (s *SystemService) GetPlatformSetting(settingKey string) (*models.SystemSetting, error) {
	var setting models.SystemSetting
	if err := database.DB.Where("tenant_id IS NULL AND setting_key = ?", settingKey).
		First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("配置不存在")
		}
		return nil, err
	}
	return &setting, nil
}

// SetPlatformSetting 设置平台级别配置（不区分租户）
func (s *SystemService) SetPlatformSetting(req *SetSettingRequest, updatedBy int64) (*models.SystemSetting, error) {
	// 将值转换为JSONB
	valueBytes, err := json.Marshal(req.SettingValue)
	if err != nil {
		return nil, errors.New("配置值格式错误")
	}

	// 查找现有配置
	var setting models.SystemSetting
	err = database.DB.Where("tenant_id IS NULL AND setting_key = ?", req.SettingKey).
		First(&setting).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新配置
			setting = models.SystemSetting{
				SettingKey:   req.SettingKey,
				SettingValue: valueBytes,
				Description:  req.Description,
				UpdatedBy:    &updatedBy,
			}
			if err := database.DB.Create(&setting).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		// 更新现有配置
		updates := map[string]interface{}{
			"setting_value": valueBytes,
			"updated_by":    updatedBy,
		}
		if req.Description != "" {
			updates["description"] = req.Description
		}
		if err := database.DB.Model(&setting).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return &setting, nil
}

// CreateAuditLog 创建审计日志
// 任务 11.3.1: 创建审计日志（含字段级变更追踪）
func (s *SystemService) CreateAuditLog(tenantID, userID int64, req *CreateAuditLogRequest, ipAddress, userAgent string) error {
	auditLog := &models.AuditLog{
		TenantModel: models.TenantModel{
			TenantID: tenantID,
		},
		UserID:       userID,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Details:      req.Details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	// 如果有字段变更信息，序列化并存储
	if req.ChangedFields != nil && len(req.ChangedFields) > 0 {
		changedFieldsBytes, err := json.Marshal(req.ChangedFields)
		if err == nil {
			auditLog.ChangedFields = changedFieldsBytes
		}
	}

	return database.DB.Create(auditLog).Error
}

// CompareAndTrackChanges 比较两个对象的字段变更
// 用于自动生成字段级别的变更记录
func (s *SystemService) CompareAndTrackChanges(before, after interface{}) map[string]models.FieldChange {
	changes := make(map[string]models.FieldChange)

	beforeBytes, _ := json.Marshal(before)
	afterBytes, _ := json.Marshal(after)

	var beforeMap, afterMap map[string]interface{}
	json.Unmarshal(beforeBytes, &beforeMap)
	json.Unmarshal(afterBytes, &afterMap)

	// 比较所有字段
	for key, afterValue := range afterMap {
		beforeValue, exists := beforeMap[key]
		// 忽略系统字段
		if key == "created_at" || key == "updated_at" || key == "deleted_at" {
			continue
		}
		// 如果字段不存在或值不同，记录变更
		if !exists || !compareValues(beforeValue, afterValue) {
			changes[key] = models.FieldChange{
				Before: beforeValue,
				After:  afterValue,
			}
		}
	}

	return changes
}

// compareValues 比较两个值是否相等
func compareValues(a, b interface{}) bool {
	aBytes, _ := json.Marshal(a)
	bBytes, _ := json.Marshal(b)
	return string(aBytes) == string(bBytes)
}

// GetFieldChangeHistory 获取指定资源的字段变更历史
func (s *SystemService) GetFieldChangeHistory(tenantID int64, resourceType string, resourceID int64, fieldName *string) ([]models.AuditLog, error) {
	query := database.DB.Model(&models.AuditLog{}).
		Where("tenant_id = ? AND resource_type = ? AND resource_id = ? AND changed_fields IS NOT NULL",
			tenantID, resourceType, resourceID)

	var logs []models.AuditLog
	if err := query.Preload("User").Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}

	// 如果指定了字段名，过滤出包含该字段变更的日志
	if fieldName != nil && *fieldName != "" {
		var filteredLogs []models.AuditLog
		for _, log := range logs {
			var changes map[string]models.FieldChange
			if err := json.Unmarshal(log.ChangedFields, &changes); err == nil {
				if _, exists := changes[*fieldName]; exists {
					filteredLogs = append(filteredLogs, log)
				}
			}
		}
		return filteredLogs, nil
	}

	return logs, nil
}

// QueryAuditLogs 查询审计日志
// 任务 11.3.2: 查询审计日志
func (s *SystemService) QueryAuditLogs(tenantID int64, req *QueryAuditLogRequest) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := database.DB.Model(&models.AuditLog{}).Where("tenant_id = ?", tenantID)

	// 过滤条件
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}
	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}
	if req.ResourceType != "" {
		query = query.Where("resource_type = ?", req.ResourceType)
	}
	if req.StartDate != nil {
		query = query.Where("created_at >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		// 结束日期包含当天，所以加1天
		endDate := req.EndDate.Add(24 * time.Hour)
		query = query.Where("created_at < ?", endDate)
	}

	// 统计总数
	query.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.Size
	if err := query.Offset(offset).Limit(req.Size).
		Preload("User").
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetAuditLog 获取审计日志详情
func (s *SystemService) GetAuditLog(tenantID, logID int64) (*models.AuditLog, error) {
	var log models.AuditLog
	if err := database.DB.Where("id = ? AND tenant_id = ?", logID, tenantID).
		Preload("User").
		First(&log).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("审计日志不存在")
		}
		return nil, err
	}
	return &log, nil
}

// GetAuditStatistics 获取审计统计信息
func (s *SystemService) GetAuditStatistics(tenantID int64, days int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 计算时间范围
	startDate := time.Now().AddDate(0, 0, -days)

	// 总操作数
	var totalActions int64
	database.DB.Model(&models.AuditLog{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Count(&totalActions)
	stats["total_actions"] = totalActions

	// 按操作类型统计
	type ActionStat struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}
	var actionStats []ActionStat
	database.DB.Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Group("action").
		Scan(&actionStats)
	stats["by_action"] = actionStats

	// 按资源类型统计
	type ResourceStat struct {
		ResourceType string `json:"resource_type"`
		Count        int64  `json:"count"`
	}
	var resourceStats []ResourceStat
	database.DB.Model(&models.AuditLog{}).
		Select("resource_type, COUNT(*) as count").
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Group("resource_type").
		Scan(&resourceStats)
	stats["by_resource"] = resourceStats

	// 活跃用户数
	var activeUsers int64
	database.DB.Model(&models.AuditLog{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Distinct("user_id").
		Count(&activeUsers)
	stats["active_users"] = activeUsers

	return stats, nil
}

// CleanOldAuditLogs 清理旧的审计日志
func (s *SystemService) CleanOldAuditLogs(tenantID int64, daysToKeep int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -daysToKeep)

	result := database.DB.Where("tenant_id = ? AND created_at < ?", tenantID, cutoffDate).
		Delete(&models.AuditLog{})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// ===== 导入日志相关方法 =====

// CreateImportLog 创建导入日志
func (s *SystemService) CreateImportLog(tenantID, userID int64, req *CreateImportLogRequest) (*models.ImportLog, error) {
	importLog := &models.ImportLog{
		TenantModel: models.TenantModel{
			TenantID: tenantID,
		},
		ImportType:  req.ImportType,
		FileName:    req.FileName,
		TotalRows:   req.TotalRows,
		SuccessRows: req.SuccessRows,
		FailedRows:  req.FailedRows,
		Status:      req.Status,
		ImportedBy:  &userID,
	}

	// 序列化错误详情
	if req.ErrorDetails != nil && len(req.ErrorDetails) > 0 {
		errorDetailsBytes, err := json.Marshal(req.ErrorDetails)
		if err == nil {
			importLog.ErrorDetails = errorDetailsBytes
		}
	}

	if err := database.DB.Create(importLog).Error; err != nil {
		return nil, err
	}

	return importLog, nil
}

// UpdateImportLog 更新导入日志
func (s *SystemService) UpdateImportLog(tenantID, logID int64, req *CreateImportLogRequest) error {
	updates := map[string]interface{}{
		"total_rows":   req.TotalRows,
		"success_rows": req.SuccessRows,
		"failed_rows":  req.FailedRows,
		"status":       req.Status,
	}

	// 更新错误详情
	if req.ErrorDetails != nil {
		errorDetailsBytes, err := json.Marshal(req.ErrorDetails)
		if err == nil {
			updates["error_details"] = errorDetailsBytes
		}
	}

	return database.DB.Model(&models.ImportLog{}).
		Where("id = ? AND tenant_id = ?", logID, tenantID).
		Updates(updates).Error
}

// QueryImportLogs 查询导入日志
func (s *SystemService) QueryImportLogs(tenantID int64, req *QueryImportLogRequest) ([]models.ImportLog, int64, error) {
	var logs []models.ImportLog
	var total int64

	query := database.DB.Model(&models.ImportLog{}).Where("tenant_id = ?", tenantID)

	// 过滤条件
	if req.ImportType != "" {
		query = query.Where("import_type = ?", req.ImportType)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 统计总数
	query.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.Size
	if err := query.Offset(offset).Limit(req.Size).
		Preload("Importer").
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetImportLog 获取导入日志详情
func (s *SystemService) GetImportLog(tenantID, logID int64) (*models.ImportLog, error) {
	var log models.ImportLog
	if err := database.DB.Where("id = ? AND tenant_id = ?", logID, tenantID).
		Preload("Importer").
		First(&log).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("导入日志不存在")
		}
		return nil, err
	}
	return &log, nil
}

// GetImportStatistics 获取导入统计信息
func (s *SystemService) GetImportStatistics(tenantID int64, days int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 计算时间范围
	startDate := time.Now().AddDate(0, 0, -days)

	// 总导入次数
	var totalImports int64
	database.DB.Model(&models.ImportLog{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Count(&totalImports)
	stats["total_imports"] = totalImports

	// 按类型统计
	type TypeStat struct {
		ImportType string `json:"import_type"`
		Count      int64  `json:"count"`
	}
	var typeStats []TypeStat
	database.DB.Model(&models.ImportLog{}).
		Select("import_type, COUNT(*) as count").
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Group("import_type").
		Scan(&typeStats)
	stats["by_type"] = typeStats

	// 按状态统计
	type StatusStat struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var statusStats []StatusStat
	database.DB.Model(&models.ImportLog{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Group("status").
		Scan(&statusStats)
	stats["by_status"] = statusStats

	// 总导入行数
	type RowStat struct {
		TotalRows   int64 `json:"total_rows"`
		SuccessRows int64 `json:"success_rows"`
		FailedRows  int64 `json:"failed_rows"`
	}
	var rowStat RowStat
	database.DB.Model(&models.ImportLog{}).
		Select("SUM(total_rows) as total_rows, SUM(success_rows) as success_rows, SUM(failed_rows) as failed_rows").
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Scan(&rowStat)
	stats["rows"] = rowStat

	return stats, nil
}
