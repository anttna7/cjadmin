package handler

import (
	"net/http"
	"strconv"
	"time"

	"cjadmin/internal/middleware"
	"cjadmin/internal/models"
	"cjadmin/internal/service"
	"cjadmin/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MarketingHandler 营销管理处理器
type MarketingHandler struct {
	activityService *service.ActivityService
	bannerService   *service.BannerService
	themeService    *service.ThemeService
	reportService   *service.ReportService
}

// NewMarketingHandler 创建营销处理器实例
func NewMarketingHandler(db *gorm.DB) *MarketingHandler {
	return &MarketingHandler{
		activityService: service.NewActivityService(db),
		bannerService:   service.NewBannerService(db),
		themeService:    service.NewThemeService(db),
		reportService:   service.NewReportService(db),
	}
}

// ==================== 活动管理 ====================

// CreateActivity 创建活动
func (h *MarketingHandler) CreateActivity(c *gin.Context) {
	var activity models.Activity
	if err := c.ShouldBindJSON(&activity); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	activity.TenantID = middleware.GetTenantID(c)
	activity.CreatedBy = middleware.GetUserID(c)

	if err := h.activityService.Create(&activity); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, activity)
}

// UpdateActivity 更新活动
func (h *MarketingHandler) UpdateActivity(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.activityService.Update(tenantID, id, updates); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteActivity 删除活动
func (h *MarketingHandler) DeleteActivity(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	if err := h.activityService.Delete(tenantID, id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetActivity 获取活动详情
func (h *MarketingHandler) GetActivity(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	activity, err := h.activityService.GetByID(tenantID, id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, activity)
}

// ListActivities 获取活动列表
func (h *MarketingHandler) ListActivities(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	filters := make(map[string]interface{})
	if activityType := c.Query("activity_type"); activityType != "" {
		filters["activity_type"] = activityType
	}
	if position := c.Query("display_position"); position != "" {
		filters["display_position"] = position
	}
	if isActive := c.Query("is_active"); isActive != "" {
		filters["is_active"] = isActive == "true"
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}

	activities, total, err := h.activityService.List(tenantID, page, pageSize, filters)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取活动列表失败")
		return
	}

	response.SuccessWithPagination(c, activities, total, page, pageSize)
}

// GetActiveActivities 获取当前有效活动（前端展示用）
func (h *MarketingHandler) GetActiveActivities(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	position := c.Query("position")

	activities, err := h.activityService.GetActiveActivities(tenantID, position)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取活动失败")
		return
	}

	response.Success(c, activities)
}

// ToggleActivityStatus 切换活动状态
func (h *MarketingHandler) ToggleActivityStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	if err := h.activityService.ToggleStatus(tenantID, id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// ParticipateActivity 参与活动
func (h *MarketingHandler) ParticipateActivity(c *gin.Context) {
	var req struct {
		ActivityID        int64                  `json:"activity_id" binding:"required"`
		CustomerID        int64                  `json:"customer_id" binding:"required"`
		ParticipationData map[string]interface{} `json:"participation_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	tenantID := middleware.GetTenantID(c)

	participant := &models.ActivityParticipant{
		TenantID:   tenantID,
		ActivityID: req.ActivityID,
		CustomerID: req.CustomerID,
	}

	if err := h.activityService.RecordParticipation(participant); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, participant)
}

// GetActivityParticipants 获取活动参与者
func (h *MarketingHandler) GetActivityParticipants(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	participants, total, err := h.activityService.GetParticipants(tenantID, id, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取参与者失败")
		return
	}

	response.SuccessWithPagination(c, participants, total, page, pageSize)
}

// GetActivityStatistics 获取活动统计
func (h *MarketingHandler) GetActivityStatistics(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	stats, err := h.activityService.GetStatistics(tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取统计失败")
		return
	}

	response.Success(c, stats)
}

// ==================== 横幅管理 ====================

// CreateBanner 创建横幅
func (h *MarketingHandler) CreateBanner(c *gin.Context) {
	var banner models.Banner
	if err := c.ShouldBindJSON(&banner); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	banner.TenantID = middleware.GetTenantID(c)
	banner.CreatedBy = middleware.GetUserID(c)

	if err := h.bannerService.Create(&banner); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, banner)
}

// UpdateBanner 更新横幅
func (h *MarketingHandler) UpdateBanner(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.bannerService.Update(tenantID, id, updates); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteBanner 删除横幅
func (h *MarketingHandler) DeleteBanner(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	if err := h.bannerService.Delete(tenantID, id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetBanner 获取横幅详情
func (h *MarketingHandler) GetBanner(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	banner, err := h.bannerService.GetByID(tenantID, id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, banner)
}

// ListBanners 获取横幅列表
func (h *MarketingHandler) ListBanners(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	filters := make(map[string]interface{})
	if position := c.Query("position"); position != "" {
		filters["position"] = position
	}
	if isActive := c.Query("is_active"); isActive != "" {
		filters["is_active"] = isActive == "true"
	}
	if title := c.Query("title"); title != "" {
		filters["title"] = title
	}

	banners, total, err := h.bannerService.List(tenantID, page, pageSize, filters)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取横幅列表失败")
		return
	}

	response.SuccessWithPagination(c, banners, total, page, pageSize)
}

// GetActiveBanners 获取当前有效横幅（前端展示用）
func (h *MarketingHandler) GetActiveBanners(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	position := c.Query("position")

	banners, err := h.bannerService.GetActiveBanners(tenantID, position)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取横幅失败")
		return
	}

	response.Success(c, banners)
}

// ToggleBannerStatus 切换横幅状态
func (h *MarketingHandler) ToggleBannerStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	if err := h.bannerService.ToggleStatus(tenantID, id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// RecordBannerClick 记录横幅点击
func (h *MarketingHandler) RecordBannerClick(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.bannerService.IncrementClickCount(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "记录失败")
		return
	}

	response.Success(c, nil)
}

// GetBannerStatistics 获取横幅统计
func (h *MarketingHandler) GetBannerStatistics(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	stats, err := h.bannerService.GetStatistics(tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取统计失败")
		return
	}

	response.Success(c, stats)
}

// ==================== 主题管理 ====================

// CreateTheme 创建主题
func (h *MarketingHandler) CreateTheme(c *gin.Context) {
	var theme models.Theme
	if err := c.ShouldBindJSON(&theme); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	theme.TenantID = middleware.GetTenantID(c)
	theme.CreatedBy = middleware.GetUserID(c)

	if err := h.themeService.Create(&theme); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, theme)
}

// UpdateTheme 更新主题
func (h *MarketingHandler) UpdateTheme(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.themeService.Update(tenantID, id, updates); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteTheme 删除主题
func (h *MarketingHandler) DeleteTheme(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	if err := h.themeService.Delete(tenantID, id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetTheme 获取主题详情
func (h *MarketingHandler) GetTheme(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	theme, err := h.themeService.GetByID(tenantID, id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, theme)
}

// ListThemes 获取主题列表
func (h *MarketingHandler) ListThemes(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	filters := make(map[string]interface{})
	if themeType := c.Query("theme_type"); themeType != "" {
		filters["theme_type"] = themeType
	}
	if isActive := c.Query("is_active"); isActive != "" {
		filters["is_active"] = isActive == "true"
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}

	themes, total, err := h.themeService.List(tenantID, page, pageSize, filters)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取主题列表失败")
		return
	}

	response.SuccessWithPagination(c, themes, total, page, pageSize)
}

// GetDefaultTheme 获取默认主题（前端展示用）
func (h *MarketingHandler) GetDefaultTheme(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	theme, err := h.themeService.GetDefaultTheme(tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取主题失败")
		return
	}

	response.Success(c, theme)
}

// SetDefaultTheme 设置默认主题
func (h *MarketingHandler) SetDefaultTheme(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	if err := h.themeService.SetDefault(tenantID, id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// ToggleThemeStatus 切换主题状态
func (h *MarketingHandler) ToggleThemeStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	if err := h.themeService.ToggleStatus(tenantID, id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// CreateThemeFromTemplate 从模板创建主题
func (h *MarketingHandler) CreateThemeFromTemplate(c *gin.Context) {
	var req struct {
		FestivalType string `json:"festival_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)

	theme, err := h.themeService.CreateFromTemplate(tenantID, userID, req.FestivalType)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, theme)
}

// GetThemeStatistics 获取主题统计
func (h *MarketingHandler) GetThemeStatistics(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	stats, err := h.themeService.GetStatistics(tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取统计失败")
		return
	}

	response.Success(c, stats)
}

// ==================== 报表管理 ====================

// CreateReport 创建报表
func (h *MarketingHandler) CreateReport(c *gin.Context) {
	var report models.CustomReport
	if err := c.ShouldBindJSON(&report); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	report.TenantID = middleware.GetTenantID(c)
	report.CreatedBy = middleware.GetUserID(c)

	if err := h.reportService.Create(&report); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, report)
}

// UpdateReport 更新报表
func (h *MarketingHandler) UpdateReport(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.reportService.Update(tenantID, id, updates); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteReport 删除报表
func (h *MarketingHandler) DeleteReport(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	if err := h.reportService.Delete(tenantID, id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetReport 获取报表详情
func (h *MarketingHandler) GetReport(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	report, err := h.reportService.GetByID(tenantID, id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, report)
}

// ListReports 获取报表列表
func (h *MarketingHandler) ListReports(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	filters := make(map[string]interface{})
	if category := c.Query("category"); category != "" {
		filters["category"] = category
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}

	reports, total, err := h.reportService.List(tenantID, page, pageSize, filters)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取报表列表失败")
		return
	}

	response.SuccessWithPagination(c, reports, total, page, pageSize)
}

// ExecuteReport 执行报表
func (h *MarketingHandler) ExecuteReport(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)

	var parameters map[string]interface{}
	c.ShouldBindJSON(&parameters)

	results, rowCount, err := h.reportService.Execute(tenantID, id, userID, parameters)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"results":   results,
		"row_count": rowCount,
	})
}

// GetDataSources 获取可用数据源
func (h *MarketingHandler) GetDataSources(c *gin.Context) {
	sources := h.reportService.GetDataSources()
	response.Success(c, sources)
}

// GetTableFields 获取表字段
func (h *MarketingHandler) GetTableFields(c *gin.Context) {
	tableName := c.Param("table")

	fields, err := h.reportService.GetTableFields(tableName)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, fields)
}

// GetReportExecutionHistory 获取执行历史
func (h *MarketingHandler) GetReportExecutionHistory(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	executions, total, err := h.reportService.GetExecutionHistory(tenantID, id, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取执行历史失败")
		return
	}

	response.SuccessWithPagination(c, executions, total, page, pageSize)
}

// GetReportStatistics 获取报表统计
func (h *MarketingHandler) GetReportStatistics(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	stats, err := h.reportService.GetStatistics(tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取统计失败")
		return
	}

	response.Success(c, stats)
}

// 辅助方法
func parseTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return t
}
