package handler

import (
	"cjadmin/internal/models"
	"cjadmin/internal/service"
	"cjadmin/internal/utils"
	"fmt"
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

// VerifyCustomerRequest 验证客户请求
type VerifyCustomerRequest struct {
	TenantCode  string `json:"tenant_code" binding:"required"`
	CompanyName string `json:"company_name" binding:"required"`
}

// VerifyCustomer 验证客户账户
func (h *PublicHandler) VerifyCustomer(c *gin.Context) {
	var req VerifyCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	// 获取租户
	var tenant models.Tenant
	if err := h.db.Where("tenant_code = ? AND status = ?", req.TenantCode, "active").First(&tenant).Error; err != nil {
		utils.NotFound(c, "租户不存在或未激活")
		return
	}

	// 查找客户
	var customer models.Customer
	err := h.db.Where("tenant_id = ? AND company_name = ? AND deleted_at IS NULL", tenant.ID, req.CompanyName).
		First(&customer).Error

	if err != nil {
		utils.NotFound(c, "未找到该公司账户，请确认公司名称是否正确")
		return
	}

	// 检查客户状态
	if customer.Status != "active" {
		utils.BadRequest(c, "该账户状态异常，请联系客服")
		return
	}

	// 返回客户基本信息
	utils.Success(c, gin.H{
		"id":           customer.ID,
		"company_name": customer.CompanyName,
		"status":       customer.Status,
	})
}

// GetPaymentAccountRequest 获取收款账户请求
type GetPaymentAccountRequest struct {
	TenantCode  string `json:"tenant_code" binding:"required"`
	CustomerID  int64  `json:"customer_id" binding:"required"`
	AccountType string `json:"account_type" binding:"required"` // public 或 private
}

// GetPaymentAccount 获取轮询选中的收款账户
func (h *PublicHandler) GetPaymentAccount(c *gin.Context) {
	var req GetPaymentAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	// 获取租户
	var tenant models.Tenant
	if err := h.db.Where("tenant_code = ? AND status = ?", req.TenantCode, "active").First(&tenant).Error; err != nil {
		utils.NotFound(c, "租户不存在或未激活")
		return
	}

	// 获取客户信息（包含返点配置）
	var customer models.Customer
	if err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", req.CustomerID, tenant.ID).
		First(&customer).Error; err != nil {
		utils.NotFound(c, "客户不存在")
		return
	}

	// 使用轮询服务选择收款账户
	rotationService := service.NewPaymentRotationService(h.db)

	// 获取默认轮询策略（可以从系统配置中读取）
	strategy := models.RotateStrategyWeight

	account, err := rotationService.SelectPaymentAccount(tenant.ID, req.AccountType, 0, strategy)
	if err != nil {
		utils.BadRequest(c, "暂无可用的收款账户，请联系客服")
		return
	}

	// 返回账户信息和返点配置
	result := gin.H{
		"account": gin.H{
			"id":             account.ID,
			"account_name":   account.AccountName,
			"account_holder": account.AccountHolder,
			"account_number": account.AccountNumber,
			"bank_name":      account.BankName,
			"bank_branch":    account.BankBranch,
		},
		"rebate_config": gin.H{
			"public_rebate_rate":  customer.PublicRebateRate,
			"private_rebate_rate": customer.PrivateRebateRate,
			"rebate_rate": func() float64 {
				if req.AccountType == "public" {
					return customer.PublicRebateRate
				}
				return customer.PrivateRebateRate
			}(),
		},
	}

	utils.Success(c, result)
}

// SubmitRechargeRequest 提交充值请求
type SubmitRechargeRequest struct {
	TenantCode  string  `json:"tenant_code" binding:"required"`
	CustomerID  int64   `json:"customer_id" binding:"required"`
	AccountID   int64   `json:"account_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	PaymentType string  `json:"payment_type" binding:"required"` // public 或 private
}

// SubmitRecharge 提交充值订单
func (h *PublicHandler) SubmitRecharge(c *gin.Context) {
	var req SubmitRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	// 获取租户
	var tenant models.Tenant
	if err := h.db.Where("tenant_code = ? AND status = ?", req.TenantCode, "active").First(&tenant).Error; err != nil {
		utils.NotFound(c, "租户不存在或未激活")
		return
	}

	// 获取客户信息
	var customer models.Customer
	if err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", req.CustomerID, tenant.ID).
		First(&customer).Error; err != nil {
		utils.NotFound(c, "客户不存在")
		return
	}

	// 获取收款账户
	var account models.PaymentAccount
	if err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", req.AccountID, tenant.ID).
		First(&account).Error; err != nil {
		utils.NotFound(c, "收款账户不存在")
		return
	}

	// 计算返点金额
	var rebateRate float64
	if req.PaymentType == "public" {
		rebateRate = customer.PublicRebateRate
	} else {
		rebateRate = customer.PrivateRebateRate
	}
	giftAmount := req.Amount * rebateRate / 100

	// 生成订单号
	orderNo := fmt.Sprintf("RC%s%06d", time.Now().Format("20060102150405"), customer.ID%1000000)

	// 创建充值订单
	order := models.Order{
		TenantID:     tenant.ID,
		CustomerID:   customer.ID,
		OrderNo:      orderNo,
		OrderType:    "recharge",
		TotalAmount:  req.Amount,
		PaidAmount:   0,
		Status:       "pending",
		PaymentType:  req.PaymentType,
		RebateRate:   rebateRate,
		RebateAmount: giftAmount,
		Remark:       fmt.Sprintf("充值订单 - 收款账户: %s", account.AccountName),
	}

	if err := h.db.Create(&order).Error; err != nil {
		utils.ServerError(c, "创建充值订单失败")
		return
	}

	// 更新收款账户使用统计
	rotationService := service.NewPaymentRotationService(h.db)
	rotationService.UpdateAccountStatistics(account.ID, req.Amount)

	utils.Success(c, gin.H{
		"order_id":     order.ID,
		"order_no":     order.OrderNo,
		"amount":       req.Amount,
		"gift_amount":  giftAmount,
		"total_amount": req.Amount + giftAmount,
	})
}
