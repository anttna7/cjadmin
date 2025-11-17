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

// CreditHandler 授信处理器
type CreditHandler struct {
	db            *gorm.DB
	creditService *service.CreditService
}

// NewCreditHandler 创建授信处理器
func NewCreditHandler(db *gorm.DB) *CreditHandler {
	return &CreditHandler{
		db:            db,
		creditService: service.NewCreditService(db),
	}
}

// ApplyCreditRequest 申请授信请求
type ApplyCreditRequest struct {
	CustomerID         int64   `json:"customer_id" binding:"required"`
	Amount             float64 `json:"amount" binding:"required,gt=0"`
	Reason             string  `json:"reason" binding:"required"`
	AssignedApproverID *int64  `json:"assigned_approver_id"`
}

// ApplyCredit 申请授信
func (h *CreditHandler) ApplyCredit(c *gin.Context) {
	var req ApplyCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)

	record, err := h.creditService.ApplyCredit(
		tenantID,
		req.CustomerID,
		userID,
		req.Amount,
		req.Reason,
		req.AssignedApproverID,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "申请授信失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, record)
}

// ApproveCreditRequest 审批授信请求
type ApproveCreditRequest struct {
	Approved bool   `json:"approved"`
	Remark   string `json:"remark"`
}

// ApproveCredit 审批授信
func (h *CreditHandler) ApproveCredit(c *gin.Context) {
	recordID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)

	var req ApproveCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	err := h.creditService.ApproveCredit(
		tenantID,
		recordID,
		userID,
		req.Approved,
		req.Remark,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "审批失败: "+err.Error())
		return
	}

	message := "审批通过"
	if !req.Approved {
		message = "审批拒绝"
	}

	utils.SuccessResponse(c, gin.H{"message": message})
}

// CancelCredit 取消授信申请
func (h *CreditHandler) CancelCredit(c *gin.Context) {
	recordID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)

	err := h.creditService.CancelCredit(tenantID, recordID, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "取消失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "已取消授信申请"})
}

// AdjustCreditRequest 调整授信请求
type AdjustCreditRequest struct {
	CustomerID   int64   `json:"customer_id" binding:"required"`
	AdjustAmount float64 `json:"adjust_amount" binding:"required"`
	Reason       string  `json:"reason" binding:"required"`
}

// AdjustCredit 调整授信额度
func (h *CreditHandler) AdjustCredit(c *gin.Context) {
	var req AdjustCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)

	record, err := h.creditService.AdjustCredit(
		tenantID,
		req.CustomerID,
		userID,
		req.AdjustAmount,
		req.Reason,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "调整授信失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, record)
}

// RepayCreditRequest 还款请求
type RepayCreditRequest struct {
	CustomerID  int64   `json:"customer_id" binding:"required"`
	RepayAmount float64 `json:"repay_amount" binding:"required,gt=0"`
	RepayMethod string  `json:"repay_method" binding:"required,oneof=cash transfer deduction"`
	OrderID     *int64  `json:"order_id"`
}

// RepayCredit 还款
func (h *CreditHandler) RepayCredit(c *gin.Context) {
	var req RepayCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	tenantID := middleware.GetTenantID(c)

	record, err := h.creditService.RepayCredit(
		tenantID,
		req.CustomerID,
		req.RepayAmount,
		req.RepayMethod,
		req.OrderID,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "还款失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, record)
}

// GetCreditRecords 获取授信记录列表
func (h *CreditHandler) GetCreditRecords(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var customerID *int64
	if cid := c.Query("customer_id"); cid != "" {
		id, _ := strconv.ParseInt(cid, 10, 64)
		customerID = &id
	}

	recordType := c.Query("record_type")
	status := c.Query("status")

	records, total, err := h.creditService.GetCreditRecords(
		tenantID,
		customerID,
		recordType,
		status,
		page,
		pageSize,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}

	utils.PaginatedResponse(c, records, total, page, pageSize)
}

// GetCreditRecord 获取授信记录详情
func (h *CreditHandler) GetCreditRecord(c *gin.Context) {
	recordID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	record, err := h.creditService.GetCreditRecord(tenantID, recordID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, record)
}

// GetCustomerCreditInfo 获取客户授信信息
func (h *CreditHandler) GetCustomerCreditInfo(c *gin.Context) {
	customerID, _ := strconv.ParseInt(c.Param("customer_id"), 10, 64)

	info, err := h.creditService.GetCustomerCreditInfo(customerID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, info)
}

// GetApprovers 获取审批人列表
func (h *CreditHandler) GetApprovers(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	approvers, err := h.creditService.GetApprovers(tenantID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, approvers)
}

// GetCreditStatistics 获取授信统计
func (h *CreditHandler) GetCreditStatistics(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	statistics, err := h.creditService.GetCreditStatistics(tenantID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "统计查询失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, statistics)
}
