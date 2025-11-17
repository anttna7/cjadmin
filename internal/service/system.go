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
	Action       string `json:"action" binding:"required"`
	ResourceType string `json:"resource_type" binding:"required"`
	ResourceID   *int64 `json:"resource_id"`
	Details      string `json:"details"`
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
// 任务 11.3.1: 创建审计日志
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

	return database.DB.Create(auditLog).Error
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
