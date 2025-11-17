package handler

import (
	"net/http"
	"strconv"

	"cjadmin/internal/middleware"
	"cjadmin/internal/service"
	"cjadmin/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RebateHandler 返点处理器
type RebateHandler struct {
	db            *gorm.DB
	rebateService *service.RebateService
}

// NewRebateHandler 创建返点处理器
func NewRebateHandler(db *gorm.DB) *RebateHandler {
	return &RebateHandler{
		db:            db,
		rebateService: service.NewRebateService(db),
	}
}

// CalculateRebateRequest 计算返点请求
type CalculateRebateRequest struct {
	CustomerID  int64   `json:"customer_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	PaymentType string  `json:"payment_type" binding:"required,oneof=public private"`
}

// CalculateRebate 计算返点
func (h *RebateHandler) CalculateRebate(c *gin.Context) {
	var req CalculateRebateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	result, err := h.rebateService.CalculateRebate(
		req.CustomerID,
		req.Amount,
		req.PaymentType,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, result)
}

// GetCustomerRebateConfig 获取客户返点配置
func (h *RebateHandler) GetCustomerRebateConfig(c *gin.Context) {
	customerID, _ := strconv.ParseInt(c.Param("customer_id"), 10, 64)

	config, err := h.rebateService.GetCustomerRebateConfig(customerID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, config)
}

// UpdateRebateConfigRequest 更新返点配置请求
type UpdateRebateConfigRequest struct {
	PaymentType string  `json:"payment_type" binding:"required,oneof=public private"`
	RebateRate  float64 `json:"rebate_rate" binding:"gte=0,lte=100"`
	CashRate    float64 `json:"cash_rate" binding:"gte=0,lte=100"`
	GiftRate    float64 `json:"gift_rate" binding:"gte=0,lte=100"`
}

// UpdateCustomerRebateConfig 更新客户返点配置
func (h *RebateHandler) UpdateCustomerRebateConfig(c *gin.Context) {
	customerID, _ := strconv.ParseInt(c.Param("customer_id"), 10, 64)

	var req UpdateRebateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	err := h.rebateService.UpdateCustomerRebateConfig(
		customerID,
		req.PaymentType,
		req.RebateRate,
		req.CashRate,
		req.GiftRate,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "更新成功"})
}

// BatchUpdateRebateConfigRequest 批量更新返点配置请求
type BatchUpdateRebateConfigRequest struct {
	CustomerIDs []int64 `json:"customer_ids" binding:"required,min=1"`
	PaymentType string  `json:"payment_type" binding:"required,oneof=public private"`
	RebateRate  float64 `json:"rebate_rate" binding:"gte=0,lte=100"`
	CashRate    float64 `json:"cash_rate" binding:"gte=0,lte=100"`
	GiftRate    float64 `json:"gift_rate" binding:"gte=0,lte=100"`
}

// BatchUpdateRebateConfig 批量更新客户返点配置
func (h *RebateHandler) BatchUpdateRebateConfig(c *gin.Context) {
	var req BatchUpdateRebateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	err := h.rebateService.BatchUpdateRebateConfig(
		req.CustomerIDs,
		req.PaymentType,
		req.RebateRate,
		req.CashRate,
		req.GiftRate,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "批量更新成功",
		"count":   len(req.CustomerIDs),
	})
}

// PreviewRebateRequest 预览返点请求
type PreviewRebateRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	RebateRate  float64 `json:"rebate_rate" binding:"gte=0,lte=100"`
	CashRate    float64 `json:"cash_rate" binding:"gte=0,lte=100"`
	GiftRate    float64 `json:"gift_rate" binding:"gte=0,lte=100"`
	PaymentType string  `json:"payment_type" binding:"required,oneof=public private"`
}

// PreviewRebate 预览返点（不保存）
func (h *RebateHandler) PreviewRebate(c *gin.Context) {
	var req PreviewRebateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	result, err := h.rebateService.PreviewRebate(
		req.Amount,
		req.RebateRate,
		req.CashRate,
		req.GiftRate,
		req.PaymentType,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, result)
}

// GetRebateStatistics 获取返点统计
func (h *RebateHandler) GetRebateStatistics(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	statistics, err := h.rebateService.GetRebateStatistics(tenantID, startDate, endDate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "统计查询失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, statistics)
}
