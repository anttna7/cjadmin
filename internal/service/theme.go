package service

import (
	"errors"
	"time"

	"cjadmin/internal/models"

	"gorm.io/gorm"
)

// ThemeService 主题管理服务
type ThemeService struct {
	db *gorm.DB
}

// NewThemeService 创建主题服务实例
func NewThemeService(db *gorm.DB) *ThemeService {
	return &ThemeService{db: db}
}

// Create 创建主题（租户隔离）
func (s *ThemeService) Create(theme *models.Theme) error {
	if theme.TenantID == 0 {
		return errors.New("租户ID不能为空")
	}
	if theme.Name == "" {
		return errors.New("主题名称不能为空")
	}

	theme.CreatedAt = time.Now()
	theme.UpdatedAt = time.Now()

	// 如果设置为默认主题，先取消其他默认主题
	if theme.IsDefault {
		if err := s.clearDefaultTheme(theme.TenantID); err != nil {
			return err
		}
	}

	return s.db.Create(theme).Error
}

// Update 更新主题
func (s *ThemeService) Update(tenantID, themeID int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	// 如果设置为默认主题，先取消其他默认主题
	if isDefault, ok := updates["is_default"].(bool); ok && isDefault {
		if err := s.clearDefaultTheme(tenantID); err != nil {
			return err
		}
	}

	result := s.db.Model(&models.Theme{}).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", themeID, tenantID).
		Updates(updates)

	if result.RowsAffected == 0 {
		return errors.New("主题不存在")
	}
	return result.Error
}

// Delete 删除主题（软删除）
func (s *ThemeService) Delete(tenantID, themeID int64) error {
	// 检查是否为默认主题
	var theme models.Theme
	err := s.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", themeID, tenantID).
		First(&theme).Error
	if err != nil {
		return errors.New("主题不存在")
	}

	if theme.IsDefault {
		return errors.New("不能删除默认主题")
	}

	now := time.Now()
	return s.db.Model(&models.Theme{}).
		Where("id = ? AND tenant_id = ?", themeID, tenantID).
		Update("deleted_at", now).Error
}

// GetByID 获取主题详情
func (s *ThemeService) GetByID(tenantID, themeID int64) (*models.Theme, error) {
	var theme models.Theme
	err := s.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", themeID, tenantID).
		First(&theme).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("主题不存在")
		}
		return nil, err
	}
	return &theme, nil
}

// List 获取主题列表
func (s *ThemeService) List(tenantID int64, page, pageSize int, filters map[string]interface{}) ([]models.Theme, int64, error) {
	var themes []models.Theme
	var total int64

	query := s.db.Model(&models.Theme{}).Where("tenant_id = ? AND deleted_at IS NULL", tenantID)

	// 应用筛选条件
	if themeType, ok := filters["theme_type"].(string); ok && themeType != "" {
		query = query.Where("theme_type = ?", themeType)
	}
	if isActive, ok := filters["is_active"].(bool); ok {
		query = query.Where("is_active = ?", isActive)
	}
	if name, ok := filters["name"].(string); ok && name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("is_default DESC, created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&themes).Error

	return themes, total, err
}

// GetDefaultTheme 获取默认主题
func (s *ThemeService) GetDefaultTheme(tenantID int64) (*models.Theme, error) {
	var theme models.Theme
	err := s.db.Where("tenant_id = ? AND is_default = ? AND deleted_at IS NULL", tenantID, true).
		First(&theme).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果没有默认主题，返回第一个激活的主题
			err = s.db.Where("tenant_id = ? AND is_active = ? AND deleted_at IS NULL", tenantID, true).
				Order("created_at ASC").
				First(&theme).Error
			if err != nil {
				return nil, nil // 没有任何主题
			}
		} else {
			return nil, err
		}
	}
	return &theme, nil
}

// SetDefault 设置默认主题
func (s *ThemeService) SetDefault(tenantID, themeID int64) error {
	// 检查主题是否存在
	var theme models.Theme
	err := s.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", themeID, tenantID).
		First(&theme).Error
	if err != nil {
		return errors.New("主题不存在")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 取消所有默认主题
		if err := tx.Model(&models.Theme{}).
			Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// 设置新的默认主题
		return tx.Model(&models.Theme{}).
			Where("id = ?", themeID).
			Updates(map[string]interface{}{
				"is_default": true,
				"updated_at": time.Now(),
			}).Error
	})
}

// ToggleStatus 切换主题状态
func (s *ThemeService) ToggleStatus(tenantID, themeID int64) error {
	var theme models.Theme
	err := s.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", themeID, tenantID).
		First(&theme).Error
	if err != nil {
		return errors.New("主题不存在")
	}

	// 不能禁用默认主题
	if theme.IsDefault && theme.IsActive {
		return errors.New("不能禁用默认主题")
	}

	return s.db.Model(&theme).Updates(map[string]interface{}{
		"is_active":  !theme.IsActive,
		"updated_at": time.Now(),
	}).Error
}

// GetFestivalThemes 获取节日主题列表
func (s *ThemeService) GetFestivalThemes(tenantID int64) ([]models.Theme, error) {
	var themes []models.Theme
	err := s.db.Where("tenant_id = ? AND theme_type = ? AND deleted_at IS NULL",
		tenantID, models.ThemeTypeFestival).
		Order("created_at DESC").
		Find(&themes).Error
	return themes, err
}

// CreateFromTemplate 从模板创建主题
func (s *ThemeService) CreateFromTemplate(tenantID, createdBy int64, festivalType string) (*models.Theme, error) {
	// 预设节日主题配置
	templates := map[string]models.Theme{
		models.FestivalTypeNewYear: {
			Name:            "元旦主题",
			Description:     "元旦新年主题，喜庆红色风格",
			ThemeType:       models.ThemeTypeFestival,
			FestivalType:    models.FestivalTypeNewYear,
			PrimaryColor:    "#ff4d4f",
			SecondaryColor:  "#ffd700",
			BackgroundColor: "#fff1f0",
			TextColor:       "#333333",
		},
		models.FestivalTypeSpringFestival: {
			Name:            "春节主题",
			Description:     "春节主题，传统中国红风格",
			ThemeType:       models.ThemeTypeFestival,
			FestivalType:    models.FestivalTypeSpringFestival,
			PrimaryColor:    "#cf1322",
			SecondaryColor:  "#ffd700",
			BackgroundColor: "#fff1f0",
			TextColor:       "#333333",
		},
		models.FestivalTypeMidAutumn: {
			Name:            "中秋主题",
			Description:     "中秋节主题，金黄色风格",
			ThemeType:       models.ThemeTypeFestival,
			FestivalType:    models.FestivalTypeMidAutumn,
			PrimaryColor:    "#fa8c16",
			SecondaryColor:  "#ffd700",
			BackgroundColor: "#fff7e6",
			TextColor:       "#333333",
		},
		models.FestivalTypeChristmas: {
			Name:            "圣诞主题",
			Description:     "圣诞节主题，红绿配色",
			ThemeType:       models.ThemeTypeFestival,
			FestivalType:    models.FestivalTypeChristmas,
			PrimaryColor:    "#389e0d",
			SecondaryColor:  "#cf1322",
			BackgroundColor: "#f6ffed",
			TextColor:       "#333333",
		},
	}

	template, ok := templates[festivalType]
	if !ok {
		return nil, errors.New("未知的节日类型")
	}

	theme := template
	theme.TenantID = tenantID
	theme.CreatedBy = createdBy
	theme.CreatedAt = time.Now()
	theme.UpdatedAt = time.Now()
	theme.IsActive = true
	theme.IsDefault = false

	if err := s.db.Create(&theme).Error; err != nil {
		return nil, err
	}

	return &theme, nil
}

// GetStatistics 获取主题统计
func (s *ThemeService) GetStatistics(tenantID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总主题数
	var totalCount int64
	s.db.Model(&models.Theme{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&totalCount)
	stats["total_count"] = totalCount

	// 启用的主题数
	var activeCount int64
	s.db.Model(&models.Theme{}).
		Where("tenant_id = ? AND deleted_at IS NULL AND is_active = ?", tenantID, true).
		Count(&activeCount)
	stats["active_count"] = activeCount

	// 节日主题数
	var festivalCount int64
	s.db.Model(&models.Theme{}).
		Where("tenant_id = ? AND deleted_at IS NULL AND theme_type = ?", tenantID, models.ThemeTypeFestival).
		Count(&festivalCount)
	stats["festival_count"] = festivalCount

	return stats, nil
}

// clearDefaultTheme 清除默认主题标记
func (s *ThemeService) clearDefaultTheme(tenantID int64) error {
	return s.db.Model(&models.Theme{}).
		Where("tenant_id = ? AND is_default = ? AND deleted_at IS NULL", tenantID, true).
		Update("is_default", false).Error
}
