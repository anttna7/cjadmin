package service

import (
	"errors"
	"time"

	"cjadmin/internal/models"

	"gorm.io/gorm"
)

// BannerService 横幅广告服务
type BannerService struct {
	db *gorm.DB
}

// NewBannerService 创建横幅服务实例
func NewBannerService(db *gorm.DB) *BannerService {
	return &BannerService{db: db}
}

// Create 创建横幅（租户隔离）
func (s *BannerService) Create(banner *models.Banner) error {
	if banner.TenantID == 0 {
		return errors.New("租户ID不能为空")
	}
	if banner.Title == "" {
		return errors.New("横幅标题不能为空")
	}
	if banner.ImageURL == "" {
		return errors.New("横幅图片不能为空")
	}

	banner.CreatedAt = time.Now()
	banner.UpdatedAt = time.Now()

	return s.db.Create(banner).Error
}

// Update 更新横幅
func (s *BannerService) Update(tenantID, bannerID int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	result := s.db.Model(&models.Banner{}).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", bannerID, tenantID).
		Updates(updates)

	if result.RowsAffected == 0 {
		return errors.New("横幅不存在")
	}
	return result.Error
}

// Delete 删除横幅（软删除）
func (s *BannerService) Delete(tenantID, bannerID int64) error {
	now := time.Now()
	result := s.db.Model(&models.Banner{}).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", bannerID, tenantID).
		Update("deleted_at", now)

	if result.RowsAffected == 0 {
		return errors.New("横幅不存在")
	}
	return result.Error
}

// GetByID 获取横幅详情
func (s *BannerService) GetByID(tenantID, bannerID int64) (*models.Banner, error) {
	var banner models.Banner
	err := s.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", bannerID, tenantID).
		First(&banner).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("横幅不存在")
		}
		return nil, err
	}
	return &banner, nil
}

// List 获取横幅列表
func (s *BannerService) List(tenantID int64, page, pageSize int, filters map[string]interface{}) ([]models.Banner, int64, error) {
	var banners []models.Banner
	var total int64

	query := s.db.Model(&models.Banner{}).Where("tenant_id = ? AND deleted_at IS NULL", tenantID)

	// 应用筛选条件
	if position, ok := filters["position"].(string); ok && position != "" {
		query = query.Where("position = ?", position)
	}
	if isActive, ok := filters["is_active"].(bool); ok {
		query = query.Where("is_active = ?", isActive)
	}
	if title, ok := filters["title"].(string); ok && title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("sort_order ASC, created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&banners).Error

	return banners, total, err
}

// GetActiveBanners 获取当前有效的横幅（用于前端展示）
func (s *BannerService) GetActiveBanners(tenantID int64, position string) ([]models.Banner, error) {
	var banners []models.Banner
	now := time.Now()

	query := s.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Where("is_active = ?", true).
		Where("(start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", now, now)

	if position != "" {
		query = query.Where("position = ?", position)
	}

	err := query.Order("sort_order ASC").Find(&banners).Error
	return banners, err
}

// ToggleStatus 切换横幅状态
func (s *BannerService) ToggleStatus(tenantID, bannerID int64) error {
	var banner models.Banner
	err := s.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", bannerID, tenantID).
		First(&banner).Error
	if err != nil {
		return errors.New("横幅不存在")
	}

	return s.db.Model(&banner).Updates(map[string]interface{}{
		"is_active":  !banner.IsActive,
		"updated_at": time.Now(),
	}).Error
}

// IncrementViewCount 增加浏览次数
func (s *BannerService) IncrementViewCount(bannerID int64) error {
	return s.db.Model(&models.Banner{}).
		Where("id = ?", bannerID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// IncrementClickCount 增加点击次数
func (s *BannerService) IncrementClickCount(bannerID int64) error {
	return s.db.Model(&models.Banner{}).
		Where("id = ?", bannerID).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
}

// GetStatistics 获取横幅统计
func (s *BannerService) GetStatistics(tenantID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总横幅数
	var totalCount int64
	s.db.Model(&models.Banner{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&totalCount)
	stats["total_count"] = totalCount

	// 启用的横幅数
	var activeCount int64
	s.db.Model(&models.Banner{}).
		Where("tenant_id = ? AND deleted_at IS NULL AND is_active = ?", tenantID, true).
		Count(&activeCount)
	stats["active_count"] = activeCount

	// 总浏览次数
	var totalViews int64
	s.db.Model(&models.Banner{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Select("COALESCE(SUM(view_count), 0)").
		Scan(&totalViews)
	stats["total_views"] = totalViews

	// 总点击次数
	var totalClicks int64
	s.db.Model(&models.Banner{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Select("COALESCE(SUM(click_count), 0)").
		Scan(&totalClicks)
	stats["total_clicks"] = totalClicks

	// 计算点击率
	if totalViews > 0 {
		stats["click_rate"] = float64(totalClicks) / float64(totalViews) * 100
	} else {
		stats["click_rate"] = 0.0
	}

	return stats, nil
}

// UpdateSortOrder 更新排序顺序
func (s *BannerService) UpdateSortOrder(tenantID int64, bannerOrders []map[string]int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range bannerOrders {
			bannerID := item["id"]
			sortOrder := item["sort_order"]
			if err := tx.Model(&models.Banner{}).
				Where("id = ? AND tenant_id = ?", bannerID, tenantID).
				Update("sort_order", sortOrder).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
