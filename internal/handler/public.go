package handler

import (
	"cjadmin/internal/models"
	"cjadmin/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PublicHandler 公共API处理器（无需认证）
type PublicHandler struct {
	db *gorm.DB
}

// NewPublicHandler 创建公共处理器实例
func NewPublicHandler(db *gorm.DB) *PublicHandler {
	return &PublicHandler{db: db}
}

// GetPublicTheme 获取租户公共主题
func (h *PublicHandler) GetPublicTheme(c *gin.Context) {
	tenantCode := c.Query("tenant")
	if tenantCode == "" {
		utils.BadRequest(c, "租户标识不能为空")
		return
	}

	// 根据租户编码获取租户ID
	var tenant models.Tenant
	if err := h.db.Where("tenant_code = ? AND status = ?", tenantCode, "active").First(&tenant).Error; err != nil {
		utils.NotFound(c, "租户不存在或未激活")
		return
	}

	// 获取租户默认主题
	var theme models.Theme
	err := h.db.Where("tenant_id = ? AND is_default = ? AND deleted_at IS NULL", tenant.ID, true).
		First(&theme).Error

	if err != nil {
		// 如果没有默认主题，返回空数据
		utils.Success(c, nil)
		return
	}

	// 返回主题信息（只返回前端需要的字段）
	result := gin.H{
		"primary_color":    theme.PrimaryColor,
		"secondary_color":  theme.SecondaryColor,
		"accent_color":     theme.AccentColor,
		"background_color": theme.BackgroundColor,
		"background_image": theme.BackgroundImage,
		"welcome_message":  theme.WelcomeMessage,
	}

	utils.Success(c, result)
}

// GetPublicActivities 获取租户公共活动
func (h *PublicHandler) GetPublicActivities(c *gin.Context) {
	tenantCode := c.Query("tenant")
	position := c.Query("position")

	if tenantCode == "" {
		utils.BadRequest(c, "租户标识不能为空")
		return
	}

	// 根据租户编码获取租户ID
	var tenant models.Tenant
	if err := h.db.Where("tenant_code = ? AND status = ?", tenantCode, "active").First(&tenant).Error; err != nil {
		utils.NotFound(c, "租户不存在或未激活")
		return
	}

	// 获取当前有效的活动
	now := time.Now()
	query := h.db.Where("tenant_id = ? AND deleted_at IS NULL", tenant.ID).
		Where("is_active = ?", true).
		Where("start_time <= ? AND end_time >= ?", now, now)

	if position != "" {
		query = query.Where("display_position = ?", position)
	}

	var activities []models.Activity
	if err := query.Order("sort_order ASC").Find(&activities).Error; err != nil {
		utils.ServerError(c, "获取活动列表失败")
		return
	}

	// 简化返回数据
	var result []gin.H
	for _, a := range activities {
		result = append(result, gin.H{
			"id":            a.ID,
			"name":          a.Name,
			"activity_type": a.ActivityType,
			"description":   a.Description,
			"start_time":    a.StartTime,
			"end_time":      a.EndTime,
			"content":       a.Content,
		})
	}

	utils.Success(c, result)
}

// GetPublicBanners 获取租户公共横幅
func (h *PublicHandler) GetPublicBanners(c *gin.Context) {
	tenantCode := c.Query("tenant")
	position := c.Query("position")

	if tenantCode == "" {
		utils.BadRequest(c, "租户标识不能为空")
		return
	}

	// 根据租户编码获取租户ID
	var tenant models.Tenant
	if err := h.db.Where("tenant_code = ? AND status = ?", tenantCode, "active").First(&tenant).Error; err != nil {
		utils.NotFound(c, "租户不存在或未激活")
		return
	}

	// 获取当前有效的横幅
	now := time.Now()
	query := h.db.Where("tenant_id = ? AND deleted_at IS NULL", tenant.ID).
		Where("is_active = ?", true).
		Where("start_time <= ? AND end_time >= ?", now, now)

	if position != "" {
		query = query.Where("position = ?", position)
	}

	var banners []models.Banner
	if err := query.Order("sort_order ASC").Find(&banners).Error; err != nil {
		utils.ServerError(c, "获取横幅列表失败")
		return
	}

	// 简化返回数据
	var result []gin.H
	for _, b := range banners {
		// 增加浏览计数
		h.db.Model(&models.Banner{}).Where("id = ?", b.ID).
			Update("view_count", gorm.Expr("view_count + 1"))

		result = append(result, gin.H{
			"id":        b.ID,
			"title":     b.Title,
			"image_url": b.ImageURL,
			"link_url":  b.LinkURL,
		})
	}

	utils.Success(c, result)
}
