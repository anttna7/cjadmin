package handler

import (
	"net/http"
	"strconv"

	"cjadmin/internal/middleware"
	"cjadmin/internal/models"
	"cjadmin/internal/service"
	"cjadmin/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PaymentAccountHandler 收款账户处理器
type PaymentAccountHandler struct {
	db                     *gorm.DB
	paymentRotationService *service.PaymentRotationService
}

// NewPaymentAccountHandler 创建收款账户处理器
func NewPaymentAccountHandler(db *gorm.DB) *PaymentAccountHandler {
	return &PaymentAccountHandler{
		db:                     db,
		paymentRotationService: service.NewPaymentRotationService(db),
	}
}

// CreatePaymentAccountRequest 创建收款账户请求
type CreatePaymentAccountRequest struct {
	AccountName    string  `json:"account_name" binding:"required"`
	AccountType    string  `json:"account_type" binding:"required,oneof=public private"`
	PaymentMethod  string  `json:"payment_method" binding:"required,oneof=bank_card alipay wechat api"`
	AccountNumber  string  `json:"account_number"`
	AccountHolder  string  `json:"account_holder"`
	BankName       string  `json:"bank_name"`
	BankBranch     string  `json:"bank_branch"`
	APIConfigID    *int64  `json:"api_config_id"`
	APIType        string  `json:"api_type"`
	APIMerchantID  string  `json:"api_merchant_id"`
	APIAppID       string  `json:"api_app_id"`
	APISecretKey   string  `json:"api_secret_key"`
	APIPublicKey   string  `json:"api_public_key"`
	APIEndpoint    string  `json:"api_endpoint"`
	IsAutoRotate   *bool   `json:"is_auto_rotate"`
	RotateStrategy string  `json:"rotate_strategy,omitempty"`
	RotateWeight   int     `json:"rotate_weight,omitempty"`
	RotatePriority int     `json:"rotate_priority,omitempty"`
	DailyLimit     float64 `json:"daily_limit,omitempty"`
	SingleLimit    float64 `json:"single_limit,omitempty"`
	MonthlyLimit   float64 `json:"monthly_limit,omitempty"`
	Description    string  `json:"description"`
	Remark         string  `json:"remark"`
}

// CreatePaymentAccount 创建收款账户
func (h *PaymentAccountHandler) CreatePaymentAccount(c *gin.Context) {
	var req CreatePaymentAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	tenantID := middleware.GetTenantID(c)

	account := &models.PaymentAccount{
		TenantID:       tenantID,
		AccountName:    req.AccountName,
		AccountType:    req.AccountType,
		PaymentMethod:  req.PaymentMethod,
		AccountNumber:  req.AccountNumber,
		AccountHolder:  req.AccountHolder,
		BankName:       req.BankName,
		BankBranch:     req.BankBranch,
		APIConfigID:    req.APIConfigID,
		APIType:        req.APIType,
		APIMerchantID:  req.APIMerchantID,
		APIAppID:       req.APIAppID,
		APISecretKey:   req.APISecretKey,
		APIPublicKey:   req.APIPublicKey,
		APIEndpoint:    req.APIEndpoint,
		IsAutoRotate:   req.IsAutoRotate != nil && *req.IsAutoRotate,
		RotateStrategy: req.RotateStrategy,
		RotateWeight:   req.RotateWeight,
		RotatePriority: req.RotatePriority,
		DailyLimit:     req.DailyLimit,
		SingleLimit:    req.SingleLimit,
		MonthlyLimit:   req.MonthlyLimit,
		Description:    req.Description,
		Remark:         req.Remark,
		Status:         models.PaymentAccountStatusActive,
	}

	// 设置默认值
	if account.RotateStrategy == "" {
		account.RotateStrategy = models.RotateStrategyWeight
	}
	if account.RotateWeight == 0 {
		account.RotateWeight = 1
	}

	if err := h.db.Create(account).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建收款账户失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, account)
}

// GetPaymentAccounts 获取收款账户列表
func (h *PaymentAccountHandler) GetPaymentAccounts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	accountType := c.Query("account_type")
	status := c.Query("status")
	paymentMethod := c.Query("payment_method")

	var accounts []models.PaymentAccount
	var total int64

	query := h.db.Model(&models.PaymentAccount{}).
		Where("tenant_id = ?", tenantID).
		Where("deleted_at IS NULL")

	if accountType != "" {
		query = query.Where("account_type = ?", accountType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if paymentMethod != "" {
		query = query.Where("payment_method = ?", paymentMethod)
	}

	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}

	offset := (page - 1) * pageSize
	if err := query.
		Preload("APIConfig").
		Order("rotate_priority DESC, created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&accounts).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}

	utils.PaginatedResponse(c, accounts, total, page, pageSize)
}

// GetPaymentAccount 获取收款账户详情
func (h *PaymentAccountHandler) GetPaymentAccount(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	accountID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var account models.PaymentAccount
	if err := h.db.
		Preload("APIConfig").
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", accountID, tenantID).
		First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "收款账户不存在")
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+err.Error())
		}
		return
	}

	utils.SuccessResponse(c, account)
}

// UpdatePaymentAccount 更新收款账户
func (h *PaymentAccountHandler) UpdatePaymentAccount(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	accountID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req CreatePaymentAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 检查账户是否存在
	var account models.PaymentAccount
	if err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", accountID, tenantID).
		First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "收款账户不存在")
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+err.Error())
		}
		return
	}

	// 更新字段
	updates := map[string]interface{}{
		"account_name":    req.AccountName,
		"account_type":    req.AccountType,
		"payment_method":  req.PaymentMethod,
		"account_number":  req.AccountNumber,
		"account_holder":  req.AccountHolder,
		"bank_name":       req.BankName,
		"bank_branch":     req.BankBranch,
		"api_config_id":   req.APIConfigID,
		"api_type":        req.APIType,
		"api_merchant_id": req.APIMerchantID,
		"api_app_id":      req.APIAppID,
		"api_secret_key":  req.APISecretKey,
		"api_public_key":  req.APIPublicKey,
		"api_endpoint":    req.APIEndpoint,
		"rotate_strategy": req.RotateStrategy,
		"rotate_weight":   req.RotateWeight,
		"rotate_priority": req.RotatePriority,
		"daily_limit":     req.DailyLimit,
		"single_limit":    req.SingleLimit,
		"monthly_limit":   req.MonthlyLimit,
		"description":     req.Description,
		"remark":          req.Remark,
	}

	if req.IsAutoRotate != nil {
		updates["is_auto_rotate"] = *req.IsAutoRotate
	}

	if err := h.db.Model(&account).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新失败: "+err.Error())
		return
	}

	// 重新查询更新后的数据
	h.db.Preload("APIConfig").First(&account, accountID)

	utils.SuccessResponse(c, account)
}

// DeletePaymentAccount 删除收款账户
func (h *PaymentAccountHandler) DeletePaymentAccount(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	accountID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	result := h.db.Where("id = ? AND tenant_id = ?", accountID, tenantID).
		Delete(&models.PaymentAccount{})

	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除失败: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "收款账户不存在")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "删除成功"})
}

// UpdatePaymentAccountStatus 更新收款账户状态
func (h *PaymentAccountHandler) UpdatePaymentAccountStatus(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	accountID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive maintenance"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	result := h.db.Model(&models.PaymentAccount{}).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", accountID, tenantID).
		Update("status", req.Status)

	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新失败: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "收款账户不存在")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "状态更新成功"})
}

// GetPaymentAccountStatistics 获取收款账户统计
func (h *PaymentAccountHandler) GetPaymentAccountStatistics(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	accountType := c.Query("account_type")

	statistics, err := h.paymentRotationService.GetAccountStatistics(tenantID, accountType)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "统计查询失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, statistics)
}

// SelectPaymentAccount 选择收款账户（用于充值订单）
func (h *PaymentAccountHandler) SelectPaymentAccount(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	var req struct {
		AccountType string  `json:"account_type" binding:"required,oneof=public private"`
		Amount      float64 `json:"amount" binding:"required,gt=0"`
		Strategy    string  `json:"strategy"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 默认策略
	if req.Strategy == "" {
		req.Strategy = models.RotateStrategyWeight
	}

	account, err := h.paymentRotationService.SelectPaymentAccount(
		tenantID,
		req.AccountType,
		req.Amount,
		req.Strategy,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "选择收款账户失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, account)
}
