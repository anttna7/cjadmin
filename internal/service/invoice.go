package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/anttna7/cjadmin/internal/database"
	"github.com/anttna7/cjadmin/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InvoiceService struct{}

func NewInvoiceService() *InvoiceService {
	return &InvoiceService{}
}

type CreateInvoiceRequest struct {
	CustomerID   int64      `json:"customer_id" binding:"required"`
	OrderID      *int64     `json:"order_id"`
	InvoiceNo    string     `json:"invoice_no"`
	InvoiceType  string     `json:"invoice_type"`
	InvoiceTitle string     `json:"invoice_title" binding:"required"`
	TaxNo        string     `json:"tax_no"`
	Amount       float64    `json:"amount" binding:"required,gt=0"`
	TaxAmount    float64    `json:"tax_amount"`
	IssueDate    *time.Time `json:"issue_date"`
	FileURL      string     `json:"file_url"`
	Notes        string     `json:"notes"`
}

type UpdateInvoiceRequest struct {
	InvoiceType  string     `json:"invoice_type"`
	InvoiceTitle string     `json:"invoice_title"`
	TaxNo        string     `json:"tax_no"`
	Amount       float64    `json:"amount"`
	TaxAmount    float64    `json:"tax_amount"`
	IssueDate    *time.Time `json:"issue_date"`
	FileURL      string     `json:"file_url"`
	Notes        string     `json:"notes"`
}

type ListInvoiceRequest struct {
	Page       int    `form:"page" binding:"min=1"`
	Size       int    `form:"size" binding:"min=1,max=100"`
	CustomerID *int64 `form:"customer_id"`
	OrderID    *int64 `form:"order_id"`
	Status     string `form:"status"`
	Keyword    string `form:"keyword"`
}

// Create 创建发票
// 任务 9.2.1: 创建发票
func (s *InvoiceService) Create(tenantID int64, req *CreateInvoiceRequest) (*models.Invoice, error) {
	// 验证客户是否存在
	var customer models.Customer
	if err := database.DB.Where("id = ? AND tenant_id = ?", req.CustomerID, tenantID).
		First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("客户不存在")
		}
		return nil, err
	}

	// 如果关联订单，验证订单是否存在
	if req.OrderID != nil {
		var order models.Order
		if err := database.DB.Where("id = ? AND tenant_id = ?", *req.OrderID, tenantID).
			First(&order).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("订单不存在")
			}
			return nil, err
		}
	}

	// 生成发票号（如果没有提供）
	invoiceNo := req.InvoiceNo
	if invoiceNo == "" {
		invoiceNo = fmt.Sprintf("INV%s%s", time.Now().Format("20060102"), uuid.New().String()[:8])
	} else {
		// 检查发票号是否重复
		var count int64
		database.DB.Model(&models.Invoice{}).
			Where("invoice_no = ?", invoiceNo).
			Count(&count)
		if count > 0 {
			return nil, errors.New("发票号已存在")
		}
	}

	invoice := &models.Invoice{
		TenantModel: models.TenantModel{
			TenantID: tenantID,
		},
		CustomerID:   req.CustomerID,
		OrderID:      req.OrderID,
		InvoiceNo:    invoiceNo,
		InvoiceType:  req.InvoiceType,
		InvoiceTitle: req.InvoiceTitle,
		TaxNo:        req.TaxNo,
		Amount:       req.Amount,
		TaxAmount:    req.TaxAmount,
		Status:       models.InvoiceStatusPending,
		IssueDate:    req.IssueDate,
		FileURL:      req.FileURL,
		Notes:        req.Notes,
	}

	if err := database.DB.Create(invoice).Error; err != nil {
		return nil, err
	}

	// 加载关联数据
	database.DB.Preload("Customer").Preload("Order").First(invoice, invoice.ID)

	return invoice, nil
}

// Update 更新发票
// 任务 9.2.2: 更新发票
func (s *InvoiceService) Update(tenantID, invoiceID int64, req *UpdateInvoiceRequest) error {
	var invoice models.Invoice
	if err := database.DB.Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("发票不存在")
		}
		return err
	}

	// 已开具或已作废的发票不能修改
	if invoice.Status == models.InvoiceStatusIssued || invoice.Status == models.InvoiceStatusCancelled {
		return errors.New("已开具或已作废的发票不能修改")
	}

	updates := map[string]interface{}{}
	if req.InvoiceType != "" {
		updates["invoice_type"] = req.InvoiceType
	}
	if req.InvoiceTitle != "" {
		updates["invoice_title"] = req.InvoiceTitle
	}
	if req.TaxNo != "" {
		updates["tax_no"] = req.TaxNo
	}
	if req.Amount > 0 {
		updates["amount"] = req.Amount
	}
	if req.TaxAmount > 0 {
		updates["tax_amount"] = req.TaxAmount
	}
	if req.IssueDate != nil {
		updates["issue_date"] = req.IssueDate
	}
	if req.FileURL != "" {
		updates["file_url"] = req.FileURL
	}
	if req.Notes != "" {
		updates["notes"] = req.Notes
	}

	return database.DB.Model(&invoice).Updates(updates).Error
}

// Delete 删除发票（软删除）
func (s *InvoiceService) Delete(tenantID, invoiceID int64) error {
	var invoice models.Invoice
	if err := database.DB.Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("发票不存在")
		}
		return err
	}

	// 已开具的发票不能删除，只能作废
	if invoice.Status == models.InvoiceStatusIssued {
		return errors.New("已开具的发票不能删除，请使用作废功能")
	}

	return database.DB.Delete(&invoice).Error
}

// Get 获取发票详情
func (s *InvoiceService) Get(tenantID, invoiceID int64) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := database.DB.Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		Preload("Customer").
		Preload("Order").
		First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("发票不存在")
		}
		return nil, err
	}
	return &invoice, nil
}

// List 获取发票列表
// 任务 9.2.4: 获取发票列表
func (s *InvoiceService) List(tenantID int64, req *ListInvoiceRequest) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var total int64

	query := database.DB.Model(&models.Invoice{}).Where("tenant_id = ?", tenantID)

	// 过滤条件
	if req.CustomerID != nil {
		query = query.Where("customer_id = ?", *req.CustomerID)
	}
	if req.OrderID != nil {
		query = query.Where("order_id = ?", *req.OrderID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		query = query.Where("invoice_no LIKE ? OR invoice_title LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 统计总数
	query.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.Size
	if err := query.Offset(offset).Limit(req.Size).
		Preload("Customer").
		Preload("Order").
		Order("created_at DESC").
		Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

// Issue 开具发票
func (s *InvoiceService) Issue(tenantID, invoiceID int64, fileURL string) error {
	var invoice models.Invoice
	if err := database.DB.Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("发票不存在")
		}
		return err
	}

	// 只有待开具的发票可以开具
	if invoice.Status != models.InvoiceStatusPending {
		return errors.New("只有待开具的发票可以开具")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":     models.InvoiceStatusIssued,
		"issue_date": &now,
	}
	if fileURL != "" {
		updates["file_url"] = fileURL
	}

	return database.DB.Model(&invoice).Updates(updates).Error
}

// Cancel 作废发票
// 任务 9.2.3: 作废发票
func (s *InvoiceService) Cancel(tenantID, invoiceID int64, reason string) error {
	var invoice models.Invoice
	if err := database.DB.Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("发票不存在")
		}
		return err
	}

	// 已作废的发票不能再次作废
	if invoice.Status == models.InvoiceStatusCancelled {
		return errors.New("发票已经作废")
	}

	updates := map[string]interface{}{
		"status": models.InvoiceStatusCancelled,
	}
	if reason != "" {
		currentNotes := invoice.Notes
		if currentNotes != "" {
			updates["notes"] = currentNotes + "\n作废原因：" + reason
		} else {
			updates["notes"] = "作废原因：" + reason
		}
	}

	return database.DB.Model(&invoice).Updates(updates).Error
}

// GetByCustomer 获取客户的所有发票
func (s *InvoiceService) GetByCustomer(tenantID, customerID int64) ([]models.Invoice, error) {
	var invoices []models.Invoice
	if err := database.DB.Where("tenant_id = ? AND customer_id = ?", tenantID, customerID).
		Preload("Order").
		Order("created_at DESC").
		Find(&invoices).Error; err != nil {
		return nil, err
	}
	return invoices, nil
}

// GetByOrder 获取订单的发票
func (s *InvoiceService) GetByOrder(tenantID, orderID int64) ([]models.Invoice, error) {
	var invoices []models.Invoice
	if err := database.DB.Where("tenant_id = ? AND order_id = ?", tenantID, orderID).
		Preload("Customer").
		Order("created_at DESC").
		Find(&invoices).Error; err != nil {
		return nil, err
	}
	return invoices, nil
}

// GetStatistics 获取发票统计信息
func (s *InvoiceService) GetStatistics(tenantID int64, customerID *int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := database.DB.Model(&models.Invoice{}).Where("tenant_id = ?", tenantID)
	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}

	// 发票总数
	var totalCount int64
	query.Count(&totalCount)
	stats["total_count"] = totalCount

	// 待开具发票数
	var pendingCount int64
	database.DB.Model(&models.Invoice{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.InvoiceStatusPending).
		Count(&pendingCount)
	stats["pending_count"] = pendingCount

	// 已开具发票数
	var issuedCount int64
	database.DB.Model(&models.Invoice{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.InvoiceStatusIssued).
		Count(&issuedCount)
	stats["issued_count"] = issuedCount

	// 已作废发票数
	var cancelledCount int64
	database.DB.Model(&models.Invoice{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.InvoiceStatusCancelled).
		Count(&cancelledCount)
	stats["cancelled_count"] = cancelledCount

	// 发票总金额（已开具）
	var totalAmount float64
	database.DB.Model(&models.Invoice{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.InvoiceStatusIssued).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalAmount)
	stats["total_amount"] = totalAmount

	return stats, nil
}
