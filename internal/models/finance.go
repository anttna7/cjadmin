package models

import "time"

// Invoice 发票管理表
type Invoice struct {
	TenantModel
	CustomerID   int64      `gorm:"not null;index" json:"customer_id"`
	OrderID      *int64     `gorm:"index" json:"order_id"`
	InvoiceNo    string     `gorm:"size:100;unique;not null" json:"invoice_no"`
	InvoiceType  string     `gorm:"size:20" json:"invoice_type"` // 增值税专用发票、普通发票
	InvoiceTitle string     `gorm:"size:255" json:"invoice_title"`
	TaxNo        string     `gorm:"size:100" json:"tax_no"`
	Amount       float64    `gorm:"type:decimal(15,2);not null" json:"amount"`
	TaxAmount    float64    `gorm:"type:decimal(15,2)" json:"tax_amount"`
	Status       string     `gorm:"size:20;default:pending" json:"status"` // pending, issued, cancelled
	IssueDate    *time.Time `gorm:"type:date" json:"issue_date"`
	FileURL      string     `gorm:"size:500" json:"file_url"`
	Notes        string     `gorm:"type:text" json:"notes"`

	// 关联
	Customer *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Order    *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// 发票状态常量
const (
	InvoiceStatusPending   = "pending"   // 待开具
	InvoiceStatusIssued    = "issued"    // 已开具
	InvoiceStatusCancelled = "cancelled" // 已作废
)

func (Invoice) TableName() string {
	return "invoices"
}
