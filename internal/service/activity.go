package service

import (
	"errors"
	"time"

	"cjadmin/internal/models"

	"gorm.io/gorm"
)

// ActivityService 活动管理服务
type ActivityService struct {
	db *gorm.DB
}

// NewActivityService 创建活动服务实例
func NewActivityService(db *gorm.DB) *ActivityService {
	return &ActivityService{db: db}
}

// Create 创建活动（租户隔离）
func (s *ActivityService) Create(activity *models.Activity) error {
	if activity.TenantID == 0 {
		return errors.New("租户ID不能为空")
	}
	if activity.Name == "" {
		return errors.New("活动名称不能为空")
	}
	if activity.StartTime.IsZero() || activity.EndTime.IsZero() {
		return errors.New("活动时间不能为空")
	}
	if activity.EndTime.Before(activity.StartTime) {
		return errors.New("结束时间不能早于开始时间")
	}

	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()

	return s.db.Create(activity).Error
}

// Update 更新活动
func (s *ActivityService) Update(tenantID, activityID int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	result := s.db.Model(&models.Activity{}).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", activityID, tenantID).
		Updates(updates)

	if result.RowsAffected == 0 {
		return errors.New("活动不存在")
	}
	return result.Error
}

// Delete 删除活动（软删除）
func (s *ActivityService) Delete(tenantID, activityID int64) error {
	now := time.Now()
	result := s.db.Model(&models.Activity{}).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", activityID, tenantID).
		Update("deleted_at", now)

	if result.RowsAffected == 0 {
		return errors.New("活动不存在")
	}
	return result.Error
}

// GetByID 获取活动详情
func (s *ActivityService) GetByID(tenantID, activityID int64) (*models.Activity, error) {
	var activity models.Activity
	err := s.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", activityID, tenantID).
		First(&activity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("活动不存在")
		}
		return nil, err
	}
	return &activity, nil
}

// List 获取活动列表
func (s *ActivityService) List(tenantID int64, page, pageSize int, filters map[string]interface{}) ([]models.Activity, int64, error) {
	var activities []models.Activity
	var total int64

	query := s.db.Model(&models.Activity{}).Where("tenant_id = ? AND deleted_at IS NULL", tenantID)

	// 应用筛选条件
	if activityType, ok := filters["activity_type"].(string); ok && activityType != "" {
		query = query.Where("activity_type = ?", activityType)
	}
	if position, ok := filters["display_position"].(string); ok && position != "" {
		query = query.Where("display_position = ?", position)
	}
	if isActive, ok := filters["is_active"].(bool); ok {
		query = query.Where("is_active = ?", isActive)
	}
	if name, ok := filters["name"].(string); ok && name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	// 时间范围筛选
	if startTime, ok := filters["start_time"].(time.Time); ok && !startTime.IsZero() {
		query = query.Where("start_time >= ?", startTime)
	}
	if endTime, ok := filters["end_time"].(time.Time); ok && !endTime.IsZero() {
		query = query.Where("end_time <= ?", endTime)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("sort_order ASC, created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&activities).Error

	return activities, total, err
}

// GetActiveActivities 获取当前有效的活动（用于前端展示）
func (s *ActivityService) GetActiveActivities(tenantID int64, position string) ([]models.Activity, error) {
	var activities []models.Activity
	now := time.Now()

	query := s.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Where("is_active = ?", true).
		Where("start_time <= ? AND end_time >= ?", now, now)

	if position != "" {
		query = query.Where("display_position = ?", position)
	}

	err := query.Order("sort_order ASC").Find(&activities).Error
	return activities, err
}

// ToggleStatus 切换活动状态
func (s *ActivityService) ToggleStatus(tenantID, activityID int64) error {
	var activity models.Activity
	err := s.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", activityID, tenantID).
		First(&activity).Error
	if err != nil {
		return errors.New("活动不存在")
	}

	return s.db.Model(&activity).Updates(map[string]interface{}{
		"is_active":  !activity.IsActive,
		"updated_at": time.Now(),
	}).Error
}

// IncrementViewCount 增加浏览次数
func (s *ActivityService) IncrementViewCount(activityID int64) error {
	return s.db.Model(&models.Activity{}).
		Where("id = ?", activityID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// RecordParticipation 记录活动参与
func (s *ActivityService) RecordParticipation(participant *models.ActivityParticipant) error {
	// 检查活动是否存在且有效
	var activity models.Activity
	err := s.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL",
		participant.ActivityID, participant.TenantID).First(&activity).Error
	if err != nil {
		return errors.New("活动不存在")
	}

	// 检查活动是否在有效期内
	now := time.Now()
	if now.Before(activity.StartTime) || now.After(activity.EndTime) {
		return errors.New("活动不在有效期内")
	}

	// 检查是否已参与
	var count int64
	s.db.Model(&models.ActivityParticipant{}).
		Where("activity_id = ? AND customer_id = ?", participant.ActivityID, participant.CustomerID).
		Count(&count)
	if count > 0 {
		return errors.New("已参与过此活动")
	}

	// 检查参与人数限制
	if activity.MaxParticipants > 0 && activity.ParticipantCount >= activity.MaxParticipants {
		return errors.New("活动参与人数已达上限")
	}

	// 事务处理
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 创建参与记录
		participant.ParticipatedAt = now
		if err := tx.Create(participant).Error; err != nil {
			return err
		}

		// 更新参与人数
		return tx.Model(&models.Activity{}).
			Where("id = ?", participant.ActivityID).
			UpdateColumn("participant_count", gorm.Expr("participant_count + 1")).Error
	})
}

// GetParticipants 获取活动参与者列表
func (s *ActivityService) GetParticipants(tenantID, activityID int64, page, pageSize int) ([]models.ActivityParticipant, int64, error) {
	var participants []models.ActivityParticipant
	var total int64

	query := s.db.Model(&models.ActivityParticipant{}).
		Where("tenant_id = ? AND activity_id = ?", tenantID, activityID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("Customer").
		Order("participated_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&participants).Error

	return participants, total, err
}

// GetStatistics 获取活动统计
func (s *ActivityService) GetStatistics(tenantID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总活动数
	var totalCount int64
	s.db.Model(&models.Activity{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&totalCount)
	stats["total_count"] = totalCount

	// 进行中的活动数
	now := time.Now()
	var activeCount int64
	s.db.Model(&models.Activity{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Where("is_active = ? AND start_time <= ? AND end_time >= ?", true, now, now).
		Count(&activeCount)
	stats["active_count"] = activeCount

	// 总参与人次
	var totalParticipants int64
	s.db.Model(&models.ActivityParticipant{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalParticipants)
	stats["total_participants"] = totalParticipants

	// 总浏览次数
	var totalViews int64
	s.db.Model(&models.Activity{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Select("COALESCE(SUM(view_count), 0)").
		Scan(&totalViews)
	stats["total_views"] = totalViews

	return stats, nil
}

// GrantReward 发放奖励
func (s *ActivityService) GrantReward(tenantID, participantID int64) error {
	now := time.Now()
	result := s.db.Model(&models.ActivityParticipant{}).
		Where("id = ? AND tenant_id = ? AND reward_status = ?", participantID, tenantID, models.RewardStatusPending).
		Updates(map[string]interface{}{
			"reward_status":     models.RewardStatusGranted,
			"reward_granted_at": now,
		})

	if result.RowsAffected == 0 {
		return errors.New("参与记录不存在或奖励已发放")
	}
	return result.Error
}
